package grpcserver

import (
	"context"
	"net"

	"google.golang.org/grpc"
)

type Server struct {
	srv  *grpc.Server
	addr string
}

func New(addr string, opts ...grpc.ServerOption) *Server {
	return &Server{srv: grpc.NewServer(opts...), addr: addr}
}

func (s *Server) Server() *grpc.Server {
	return s.srv
}

func (s *Server) Start(_ context.Context) error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	return s.srv.Serve(ln)
}

func (s *Server) Stop(_ context.Context) error {
	s.srv.GracefulStop()
	return nil
}

func (s *Server) Addr() string {
	return s.addr
}
