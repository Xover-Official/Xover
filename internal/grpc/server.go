package grpc

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
)

// Server manages internal gRPC communication
type Server struct {
	port int
	srv  *grpc.Server
}

func NewServer(port int) *Server {
	return &Server{
		port: port,
		srv:  grpc.NewServer(),
	}
}

// Start runs the gRPC listener
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return err
	}

	fmt.Printf("🔌 gRPC Internal Layer started on port %d\n", s.port)
	return s.srv.Serve(lis)
}

// Stop gracefully shuts down
func (s *Server) Stop() {
	s.srv.GracefulStop()
}
