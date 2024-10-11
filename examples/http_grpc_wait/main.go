package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"runy"
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

	httpSrv := NewHTTPServer(":8080")
	grpcSrv := NewGRPCServer(":9090")

	runy.Add(httpSrv, grpcSrv).AddF(WorkerF) // register all Runnable app components

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
