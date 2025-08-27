package server

import (
	"fmt"
	"net"

	"github.com/am0xff/metrics/api/proto"
	"github.com/am0xff/metrics/internal/storage"
	"google.golang.org/grpc"
)

func RunGRPCServer(addr string, storageProvider storage.StorageProvider) error {
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	grpcServer := grpc.NewServer()
	metricsServer := NewMetricsServer(storageProvider)

	gen.RegisterMetricsServiceServer(grpcServer, metricsServer)

	fmt.Printf("gRPC server listening on %s\n", addr)
	return grpcServer.Serve(listen)
}
