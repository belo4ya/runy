package runy_test

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/belo4ya/runy"
)

// Example_sugared demonstrates how to use runy's sugared API to manage
// multiple components using a more concise functional approach.
func Example_sugared() {
	// Create a context that's canceled when SIGINT or SIGTERM is received.
	ctx := runy.SetupSignalHandler()

	// Initialize application components.
	httpSrv := &http.Server{Addr: ":8080"} // main API server
	mgmtSrv := &http.Server{Addr: ":8081"} // /metrics, /debug/pprof, /healthz, /readyz
	shutdownTimeout := 3 * time.Second

	// Register HTTP server using the sugared functional API (SAddF).
	// First function is the Start function, second is the Stop function.
	runy.SAddF(func(_ context.Context) error {
		log.Printf("http server starts listening on: %s", httpSrv.Addr)
		return runy.IgnoreHTTPServerClosed(httpSrv.ListenAndServe())
	}, func(_ context.Context) error {
		// Graceful shutdown with timeout when context is canceled.
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		return httpSrv.Shutdown(ctx)
	})

	// Register management server using the same pattern.
	runy.SAddF(func(_ context.Context) error {
		log.Printf("mgmt server start listening on: %s", mgmtSrv.Addr)
		return runy.IgnoreHTTPServerClosed(mgmtSrv.ListenAndServe())
	}, func(_ context.Context) error {
		// Graceful shutdown with timeout when context is canceled.
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		return mgmtSrv.Shutdown(ctx)
	})

	// Register a background worker using the standard functional API (AddF).
	// This is appropriate for components that don't need a separate stop function.
	runy.AddF(func(ctx context.Context) error {
		t := time.NewTicker(1 * time.Second)
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
		log.Fatalf("app error: %v", err)
	}
}
