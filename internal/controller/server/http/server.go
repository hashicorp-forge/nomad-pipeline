package http

import (
	"context"
	"fmt"
	"net"
	stdHTTP "net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/coordinator"
	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/server/state"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/logger"
)

type ServerReq struct {
	Coordinator        *coordinator.Coordinator
	Logger             *zap.Logger
	HTTPAddr           string
	HTPPAccessLogLevel string
	State              state.State
}

type Server struct {
	logger *zap.Logger
	ln     net.Listener
	mux    *chi.Mux
	server *stdHTTP.Server
}

func NewServer(req *ServerReq) (*Server, error) {

	parsedURL, err := url.Parse(req.HTTPAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTTP address: %w", err)
	}

	ln, err := net.Listen("tcp", parsedURL.Host)
	if err != nil {
		return nil, fmt.Errorf("failed to setup HTTP listener: %w", err)
	}

	s := &Server{
		logger: req.Logger.Named(logger.ComponentNameHTTPServer).With(
			zap.String("bind_addr", ln.Addr().String()),
			zap.String("network", "tcp"),
		),
		ln:  ln,
		mux: newRouter(req),
	}

	s.server = &stdHTTP.Server{
		Addr:         req.HTTPAddr,
		Handler:      s.mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	s.logger.Info("successfully initialized HTTP server")

	return s, nil
}

func (s *Server) Start() {
	s.logger.Info("starting HTTP server")
	go s.serve()
}

func (s *Server) serve() {
	_ = s.server.Serve(s.ln)
}

func (s *Server) Stop() {
	s.logger.Info("stopped HTTP server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("failed to gracefully stop HTTP server", zap.Error(err))
	} else {
		s.logger.Info("successfully stopped HTTP server")
	}
}
