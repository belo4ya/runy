package runnables

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
)

type GRPCServerConfig struct {
	Addr string
}

type GRPCServer struct {
	GRPC *grpc.Server
	conf GRPCServerConfig
}

func NewGRPCServer(conf GRPCServerConfig) *GRPCServer {
	return &GRPCServer{GRPC: grpc.NewServer(), conf: conf}
}

func (s *GRPCServer) Start(ctx context.Context) error {
	lis, err := net.Listen("tcp", s.conf.Addr)
	if err != nil {
		return fmt.Errorf("net listen: %w", err)
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("grpc server start listening on: %s", s.conf.Addr)
		if err := s.GRPC.Serve(lis); err != nil {
			errCh <- fmt.Errorf("grpc serve: %w", err)
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		log.Println("shutting down grpc server")
		s.GRPC.GracefulStop()
		return nil
	case err := <-errCh:
		return err
	}
}
