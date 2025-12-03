package server

import (
	"fmt"
	"igm-svc/internal/handlers"
	"log"
	"net"

	pb "github.com/parthav-effimove/ONDC-Protos/protos/ondc/igm/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GRPCServer struct {
	server  *grpc.Server
	port    string
	handler *handlers.IssueHandler
}

func NewGRPCServer(port string, handler *handlers.IssueHandler) *GRPCServer {
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			LoggingInterceptor(),
			RecoveryInterceptor(),
		),
	)

	pb.RegisterIssueServiceServer(server, handler)

	reflection.Register(server)

	return &GRPCServer{
		server:  server,
		port:    port,
		handler: handler,
	}
}

func (s *GRPCServer) Start() error {
	listner, err := net.Listen("tcp", s.port)
	if err != nil {
		return fmt.Errorf("failed to listen to %s:%w", s.port, err)
	}
	log.Printf("gRPC server listneing to %s", s.port)
	log.Print("reflection enabled")
	return s.server.Serve(listner)
}

func (s *GRPCServer) Stop() {
	log.Printf("shutting down gRPC server")
	s.server.GracefulStop()
	log.Printf("grpc server stopped")
}
