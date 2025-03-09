package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/util/wait"
)

type GRPCServer struct {
	GRPC *grpc.Server
	addr string
}

func NewGRPCServer(addr string) *GRPCServer {
	return &GRPCServer{GRPC: grpc.NewServer(), addr: addr}
}

func (s *GRPCServer) Start(ctx context.Context) error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("net.Listen: %w", err)
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("GRPC server start listening on: %s", s.addr)
		if err := s.GRPC.Serve(lis); err != nil {
			errCh <- fmt.Errorf("s.GRPC.Serve: %w", err)
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		log.Println("shutting down GRPC server")
		s.GRPC.GracefulStop()
		return nil
	case err := <-errCh:
		return err
	}
}

type HTTPServer struct {
	HTTP *http.Server
}

func NewHTTPServer(addr string) *HTTPServer {
	srv := &HTTPServer{}
	srv.HTTP = &http.Server{Addr: addr, Handler: srv.routes()}
	return srv
}

func (s *HTTPServer) Start(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		log.Printf("HTTP server start listening on: %s", s.HTTP.Addr)
		if err := s.HTTP.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("s.HTTP.ListenAndServe: %w", err)
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		log.Println("shutting down HTTP server")
		if err := s.HTTP.Shutdown(context.Background()); err != nil {
			return fmt.Errorf("s.HTTP.Shutdown: %w", err)
		}
		return nil
	case err := <-errCh:
		return err
	}
}

func (s *HTTPServer) routes() *chi.Mux {
	OkHandler := func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}

	r := chi.NewRouter()
	r.Handle("/metrics", promhttp.Handler())
	r.Mount("/debug", middleware.Profiler())
	r.Get("/healthz", OkHandler)
	r.Get("/readyz", OkHandler)
	return r
}

func WorkerF(ctx context.Context) error {
	log.Println("worker starting")
	wait.UntilWithContext(ctx, func(ctx context.Context) { log.Println("worker does useful things") }, 10*time.Second)
	return nil
}
