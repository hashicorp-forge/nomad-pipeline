package server

import (
	"context"
	"fmt"
	"net"
	stdHTTP "net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/api"

	"github.com/hashicorp-forge/nomad-pipeline/internal/logger"
	"github.com/hashicorp-forge/nomad-pipeline/internal/runner"
	"github.com/hashicorp-forge/nomad-pipeline/internal/server/http"
	"github.com/hashicorp-forge/nomad-pipeline/internal/server/state"
	stateImpl "github.com/hashicorp-forge/nomad-pipeline/internal/state"
)

type Server struct {
	baseLogger   hclog.Logger
	serverLogger hclog.Logger
	srvs         []*httpServer

	nomadClient *api.Client

	state state.State

	runnerController *runner.Controller
}

type httpServer struct {
	logger hclog.Logger
	ln     net.Listener
	mux    *chi.Mux
	server *stdHTTP.Server
}

func NewServer(cfg *Config) (*Server, error) {

	hcLogger := logger.New(cfg.Log)

	server := Server{
		baseLogger:   hcLogger,
		serverLogger: hcLogger.Named(logger.ComponentNameServer),
		state:        stateImpl.NewBackend(cfg.State),
	}

	if info, err := os.Stat(cfg.Data.Path); os.IsNotExist(err) {
		if err := os.MkdirAll(cfg.Data.Path, 0755); err != nil {
			return nil, fmt.Errorf("failed to create data directory: %w", err)
		}

	} else if !info.IsDir() {
		return nil, fmt.Errorf("data path exists and is not a directory")
	}

	nomadClient, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to create Nomad client: %w", err)
	}
	server.nomadClient = nomadClient

	server.runnerController = runner.NewController(&runner.ControllerConfig{
		Logger:      hcLogger,
		NomadClient: server.nomadClient,
		State:       server.state,
		DataDir:     cfg.Data.Path,
	})

	for _, bind := range cfg.HTTP.Binds {

		srv := httpServer{
			logger: hcLogger,
			mux:    http.NewRouter(hcLogger, cfg.HTTP.AccessLogLevel, cfg.Data.Path, server.runnerController, server.state),
		}

		srv.server = &stdHTTP.Server{
			Addr:         bind.Addr,
			Handler:      srv.mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  15 * time.Second,
		}

		parsedURL, err := url.Parse(srv.server.Addr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse bind address: %w", err)
		}

		network := "tcp"
		if parsedURL.Scheme == "unix" {
			network = parsedURL.Scheme
		}

		ln, err := net.Listen(network, parsedURL.Host)
		if err != nil {
			return nil, fmt.Errorf("failed to setup HTTP listener: %w", err)
		}
		srv.ln = ln

		server.srvs = append(server.srvs, &srv)
		hcLogger.Info("successfully setup HTTP server")
	}

	return &server, nil
}

func (s *Server) Start() {
	for _, srv := range s.srvs {
		srv.logger.Info("server now listening for connections")
		go func() {
			_ = srv.server.Serve(srv.ln)
		}()
	}
}

func (s *Server) Stop() {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, srv := range s.srvs {
		if err := srv.server.Shutdown(ctx); err != nil {
			srv.logger.Error("failed to gracefully shutdown HTTP server", "error", err)
		} else {
			srv.logger.Info("successfully shutdown HTTP server")
		}
	}
}

func (s *Server) WaitForSignals() {

	signalCh := make(chan os.Signal, 3)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	// Wait to receive a signal. This blocks until we are notified.
	for {
		s.serverLogger.Debug("wait for signal handler started")

		sig := <-signalCh
		s.serverLogger.Info("received signal", "signal", sig.String())

		// Check the signal we received. If it was a SIGHUP when the
		// functionality is added, we perform the reload tasks and then
		// continue to wait for another signal. Everything else means exit.
		switch sig {
		case syscall.SIGHUP:
		default:
			s.Stop()
			return
		}
	}
}
