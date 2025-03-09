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
	"k8s.io/apimachinery/pkg/util/wait"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := runy.SetupSignalHandler() // setup handler for SIGINT and SIGTERM

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
		log.Println("HTTP server start listening on: " + httpSrv.Addr)
		return runy.IgnoreHTTPServerClosed(httpSrv.ListenAndServe())
	}, func(ctx context.Context) error {
		log.Println("shutting down HTTP server")
		return httpSrv.Shutdown(ctx)
	})
	runy.SAddF(func(ctx context.Context) error {
		log.Println("GRPC server start listening on: " + lis.Addr().String())
		return grpcSrv.Serve(lis)
	}, func(ctx context.Context) error {
		log.Println("shutting down GRPC server")
		grpcSrv.GracefulStop()
		return nil
	})
	runy.AddF(func(ctx context.Context) error {
		wait.UntilWithContext(ctx, func(ctx context.Context) { log.Println("worker does useful things") }, 10*time.Second)
		return nil
	})

	log.Println("starting app")
	if err := runy.Start(ctx); err != nil { // run app
		return fmt.Errorf("problem with running app: %w", err)
	}
	return nil
}

func initWithCleanup(i int) (any, func(), error) {
	log.Println(fmt.Sprintf("init %d", i))
	return nil, func() {
		log.Println(fmt.Sprintf("cleanup %d", i))
	}, nil
}
