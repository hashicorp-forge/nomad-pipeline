package rpc

import (
	"fmt"
	"net"
	"net/rpc"

	"go.uber.org/zap"

	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/coordinator"
	"github.com/hashicorp-forge/nomad-pipeline/internal/controller/server/state"
	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/logger"
)

type ServerReq struct {
	Coordinator *coordinator.Coordinator
	Logger      *zap.Logger
	RPCAddr     string
	State       state.State
}

type Server struct {
	logger   *zap.Logger
	listener net.Listener
	server   *rpc.Server
}

func NewServer(req *ServerReq) (*Server, error) {

	ln, err := net.Listen("tcp", req.RPCAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	s := &Server{
		logger: req.Logger.Named(logger.ComponentNameRPCServer).With(
			zap.String("bind_addr", ln.Addr().String()),
			zap.String("network", "tcp"),
		),
		listener: ln,
		server:   rpc.NewServer(),
	}

	rpcEndpoints := map[string]any{
		"Runner": &RunnerEndpoint{
			coordinator: req.Coordinator,
			state:       req.State,
		},
	}

	for name, endpoint := range rpcEndpoints {
		if err := s.server.RegisterName(name, endpoint); err != nil {
			return nil, fmt.Errorf("failed to register ping service: %w", err)
		}
		s.logger.Info("registered RPC endpoint", zap.String("endpoint_name", name))
	}

	s.logger.Info("successfully initialized RPC server")

	return s, nil
}

func (s *Server) Start() {
	s.logger.Info("starting RPC server")
	go s.serve()
}

func (s *Server) serve() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.logger.Error("failed to accept connection", zap.Error(err))
			continue
		}

		s.logger.Debug("accepted RPC connection", zap.String("remote_addr", conn.RemoteAddr().String()))

		go func() {
			s.server.ServeConn(conn)
			s.logger.Debug("RPC connection closed", zap.String("remote_addr", conn.RemoteAddr().String()))
		}()
	}
}

func (s *Server) Stop() {
	s.logger.Debug("stopping RPC server")

	if err := s.listener.Close(); err != nil {
		s.logger.Error("failed to gracefully stop RPC server", zap.Error(err))
	} else {
		s.logger.Info("successfully stopped RPC server")
	}
}
