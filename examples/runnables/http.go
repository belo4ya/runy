package runnables

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

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
