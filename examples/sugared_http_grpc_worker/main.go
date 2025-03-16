package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/belo4ya/runy"

	"google.golang.org/grpc"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// Create a context that's canceled when SIGINT or SIGTERM is received.
	ctx := runy.SetupSignalHandler()

	_, cleanup, err := initWithCleanup(1)
	if err != nil {
		return fmt.Errorf("failed to init 1: %w", err)
	}
	defer cleanup()
	_, cleanup, err = initWithCleanup(2)
	if err != nil {
		return fmt.Errorf("failed to init 2: %w", err)
	}
	defer cleanup()

	httpSrv := &http.Server{Addr: ":8080"}
	grpcSrv := grpc.NewServer()
	grpcAddr := ":9090"

	// Register HTTP server.
	runy.SAddF(func(_ context.Context) error {
		log.Printf("http server starts listening on: %s", httpSrv.Addr)
		return runy.IgnoreHTTPServerClosed(httpSrv.ListenAndServe())
	}, func(_ context.Context) error {
		log.Println("shutting down http server")
		return httpSrv.Shutdown(context.Background())
	})

	// Register GRPC server.
	runy.SAddF(func(_ context.Context) error {
		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			return fmt.Errorf("net listen: %w", err)
		}
		log.Printf("grpc server starts listening on: %s", grpcAddr)
		return grpcSrv.Serve(lis)
	}, func(_ context.Context) error {
		log.Println("shutting down grpc server")
		grpcSrv.GracefulStop()
		return nil
	})

	// Register a background worker.
	runy.AddF(func(ctx context.Context) error {
		t := time.NewTicker(2 * time.Second)
		defer t.Stop()
		log.Println("worker starting")
		for {
			select {
			case <-ctx.Done():
				log.Println("shutting down worker")
				return nil
			case <-t.C:
				log.Println("worker is doing something useful")
			}
		}
	})

	// Start all components and block until shutdown.
	log.Println("starting app")
	if err := runy.Start(ctx); err != nil {
		return fmt.Errorf("app error: %w", err)
	}
	return nil
}

func initWithCleanup(i int) (any, func(), error) {
	log.Printf("init %d", i)
	return nil, func() {
		log.Printf("cleanup %d", i)
	}, nil
}
