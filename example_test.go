package runy_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/belo4ya/runy"
)

// HTTPServer implements the Runnable interface for an HTTP server.
type HTTPServer struct {
	HTTP *http.Server
	conf HTTPServerConfig
}

type HTTPServerConfig struct {
	Addr            string
	ShutdownTimeout time.Duration
}

func NewHTTPServer(conf HTTPServerConfig) *HTTPServer {
	mux := http.NewServeMux()
	return &HTTPServer{
		HTTP: &http.Server{Addr: conf.Addr, Handler: mux},
		conf: conf,
	}
}

// Start implements Runnable interface.
// It starts the HTTP server and blocks until context cancellation or server error.
func (s *HTTPServer) Start(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		log.Printf("http server starts listening on: %s", s.conf.Addr)
		if err := s.HTTP.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("http listen and serve: %w", err)
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		log.Println("shutting down http server")
		// Handle cleanup after context cancellation with a reasonable shutdown timeout.
		ctx, cancel := context.WithTimeout(context.Background(), s.conf.ShutdownTimeout)
		defer cancel()
		if err := s.HTTP.Shutdown(ctx); err != nil {
			return fmt.Errorf("http shutdown: %w", err)
		}
		return nil
	case err := <-errCh:
		return err
	}
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

func Example() {
	// Create a context that's canceled when SIGINT or SIGTERM is received.
	ctx := runy.SetupSignalHandler()

	httpSrv := NewHTTPServer(HTTPServerConfig{ // main API server
		Addr:            ":8080",
		ShutdownTimeout: 3 * time.Second,
	})
	mgmtSrv := NewHTTPServer(HTTPServerConfig{ // /metrics, /debug/pprof, /healthz, /readyz
		Addr:            ":8081",
		ShutdownTimeout: 3 * time.Second,
	})
	worker := NewWorker(1 * time.Second) // background worker

	// Register application components with runy.
	runy.Add(httpSrv, mgmtSrv, worker)

	// Start all components and block until shutdown.
	log.Println("starting app")
	if err := runy.Start(ctx); err != nil {
		log.Fatalf("app error: %v", err)
	}
}
