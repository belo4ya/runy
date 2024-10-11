package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"runy"
	"time"

	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/util/wait"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := runy.SetupSignalHandler(context.Background()) // setup handler for SIGTERM and SIGINT

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

	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		return fmt.Errorf("net.Listen: %w", err)
	}

	httpSrv := &http.Server{Addr: ":8080"}
	grpcSrv := grpc.NewServer()

	// register all Runnable app components
	runy.SAddF(func(ctx context.Context) error {
		slog.Info("HTTP server start listening on: " + httpSrv.Addr)
		return runy.IgnoreHTTPServerClosed(httpSrv.ListenAndServe())
	}, func(ctx context.Context) error {
		slog.Info("shutting down HTTP server")
		return httpSrv.Shutdown(ctx)
	})
	runy.SAddF(func(ctx context.Context) error {
		slog.Info("GRPC server start listening on: " + lis.Addr().String())
		return grpcSrv.Serve(lis)
	}, func(ctx context.Context) error {
		slog.Info("shutting down GRPC server")
		grpcSrv.GracefulStop()
		return nil
	})
	runy.AddF(func(ctx context.Context) error {
		wait.UntilWithContext(ctx, func(ctx context.Context) { slog.Info("worker does useful things") }, 10*time.Second)
		return nil
	})

	slog.InfoContext(ctx, "starting app")
	if err := runy.Start(ctx); err != nil { // run app
		return fmt.Errorf("problem with running app: %w", err)
	}
	return nil
}

func initWithCleanup(i int) (any, func(), error) {
	slog.Info(fmt.Sprintf("init %d", i))
	return nil, func() {
		slog.Info(fmt.Sprintf("cleanup %d", i))
	}, nil
}
