package server

import (
	"context"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// serve grpc request
func (s *Server) grpcServe(ctx context.Context, l net.Listener) error {
	opts := grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
		grpc_prometheus.UnaryServerInterceptor,
		grpc_recovery.UnaryServerInterceptor(),
	))

	server := grpc.NewServer(opts)
	s.register.grpc(server) // register handler
	reflection.Register(server)

	grpc_prometheus.Register(server)
	grpc_prometheus.EnableHandlingTimeHistogram()

	s.server.grpc = server
	return server.Serve(l)
}
