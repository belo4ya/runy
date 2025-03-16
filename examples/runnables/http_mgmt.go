package runnables

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MgmtServer struct {
	HTTP *http.Server
	conf MgmtServerConfig
}

type MgmtServerConfig struct {
	Addr            string
	ShutdownTimeout time.Duration
}

func NewMgmtServer(conf MgmtServerConfig) *MgmtServer {
	srv := &MgmtServer{conf: conf}
	srv.HTTP = &http.Server{Addr: conf.Addr, Handler: srv.routes()}
	return srv
}

func (s *MgmtServer) Start(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		log.Printf("management server starts listening on: %s", s.conf.Addr)
		if err := s.HTTP.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("http listen and serve: %w", err)
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		log.Println("shutting down management server")
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

func (s *MgmtServer) routes() *chi.Mux {
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
