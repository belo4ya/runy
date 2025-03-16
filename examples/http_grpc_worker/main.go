package main

import (
	"context"
	"examples/runnables"
	"fmt"
	"log"
	"time"

	"github.com/belo4ya/runy"
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

	// Initialize and register application components with runy.
	httpSrv := runnables.NewHTTPServer(runnables.HTTPServerConfig{
		Addr:            ":8080",
		ShutdownTimeout: 3 * time.Second,
	})
	mgmtSrv := runnables.NewMgmtServer(runnables.MgmtServerConfig{
		Addr:            ":8081",
		ShutdownTimeout: 3 * time.Second,
	})
	grpcSrv := runnables.NewGRPCServer(runnables.GRPCServerConfig{
		Addr: ":9090",
	})
	worker := NewWorker(2 * time.Second)
	runy.Add(httpSrv, grpcSrv, mgmtSrv, worker)

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

// Worker implements the Runnable interface for a background worker.
type Worker struct {
	t *time.Ticker
}

func NewWorker(every time.Duration) *Worker {
	return &Worker{t: time.NewTicker(every)}
}

// Start implements Runnable interface.
// It performs periodic work until context cancellation.
func (w *Worker) Start(ctx context.Context) error {
	defer w.t.Stop()
	log.Println("worker starting")
	for {
		select {
		case <-ctx.Done():
			log.Println("shutting down worker")
			return nil
		case <-w.t.C:
			log.Println("worker is doing something useful")
		}
	}
}
