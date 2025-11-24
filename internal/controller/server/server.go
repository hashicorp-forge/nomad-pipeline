package server

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hashicorp/nomad/api"
	"go.uber.org/zap"

	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/coordinator"
	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/server/http"
	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/server/rpc"
	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/server/state"
	stateImpl "github.com/hashicorp-forge/nomad-pipeline/internal/controller/state"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/logger"
	sharedstate "github.com/hashicorp-forge/nomad-pipeline/internal/pkg/state"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/version"
)

type Server struct {
	baseLogger   *zap.Logger
	serverLogger *zap.Logger

	nomadClient *api.Client

	state state.State

	//
	httpServer *http.Server

	//
	rpcServer *rpc.Server

	runnerController *coordinator.Coordinator
}

func NewServer(cfg *Config) (*Server, error) {

	zapLogger, err := logger.NewZap(cfg.Log)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	zapLogger.Info("starting server", zap.String("version", version.Get()))

	server := Server{
		baseLogger:   zapLogger,
		serverLogger: zapLogger.Named(logger.ComponentNameServer),
	}

	if info, err := os.Stat(cfg.Data.Path); os.IsNotExist(err) {
		if err := os.MkdirAll(cfg.Data.Path, 0755); err != nil {
			return nil, fmt.Errorf("failed to create data directory: %w", err)
		}

	} else if !info.IsDir() {
		return nil, fmt.Errorf("data path exists and is not a directory")
	}

	nomadClient, err := generateNomadClient(cfg.Nomad)
	if err != nil {
		return nil, fmt.Errorf("failed to create Nomad client: %w", err)
	}
	server.nomadClient = nomadClient

	stateBackend, err := stateImpl.NewBackend(cfg.State, zapLogger, nomadClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create state backend: %w", err)
	}
	server.state = stateBackend

	if err := server.setupStateDefaultObjects(); err != nil {
		return nil, fmt.Errorf("failed to setup default state objects: %w", err)
	}

	server.runnerController = coordinator.New(&coordinator.CoordinatorConfig{
		Logger:      zapLogger,
		NomadClient: server.nomadClient,
		State:       server.state,
		DataDir:     cfg.Data.Path,
		RPCAddr:     cfg.RPC.Addr,
	})

	//
	rpcServerReq := rpc.ServerReq{
		Coordinator: server.runnerController,
		Logger:      zapLogger,
		RPCAddr:     cfg.RPC.Addr,
		State:       server.state,
	}

	rpcServer, err := rpc.NewServer(&rpcServerReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC server: %w", err)
	}
	server.rpcServer = rpcServer

	//
	httpServerReq := http.ServerReq{
		Coordinator:        server.runnerController,
		Logger:             zapLogger,
		HTTPAddr:           cfg.HTTP.Addr,
		HTPPAccessLogLevel: cfg.HTTP.AccessLogLevel,
		State:              server.state,
	}

	httpServer, err := http.NewServer(&httpServerReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP server: %w", err)
	}
	server.httpServer = httpServer

	return &server, nil
}

func (s *Server) setupStateDefaultObjects() error {

	stateResp, err := s.state.Namespaces().Get(&state.NamespacesGetReq{Name: "default"})
	if err != nil && err.StatusCode() != 404 {
		return fmt.Errorf("failed to check for default namespace: %w", err)
	}
	if stateResp != nil {
		return nil
	}

	if _, err := s.state.Namespaces().Create(&state.NamespacesCreateReq{Namespace: &sharedstate.Namespace{
		ID:          "default",
		Description: "built-in namespace",
	}}); err != nil {
		return fmt.Errorf("failed to create default namespace: %w", err)
	}

	return nil
}

func (s *Server) Start() {
	if err := s.runnerController.Start(); err != nil {
		s.serverLogger.Fatal("failed to start coordinator", zap.Error(err))
	}

	s.rpcServer.Start()
	s.httpServer.Start()
}

func (s *Server) Stop() {
	s.runnerController.Stop()
	s.httpServer.Stop()
	s.rpcServer.Stop()
}

func (s *Server) WaitForSignals() {

	signalCh := make(chan os.Signal, 3)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	// Wait to receive a signal. This blocks until we are notified.
	for {
		s.serverLogger.Debug("wait for signal handler started")

		sig := <-signalCh
		s.serverLogger.Info("received signal", zap.String("signal", sig.String()))

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
