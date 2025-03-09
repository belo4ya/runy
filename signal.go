package runy

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// SetupSignalHandler registers for SIGINT and SIGTERM.
// A context is returned which is canceled on one of these signals.
// If a second signal is caught, the program is terminated with exit code 1.
func SetupSignalHandler() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	ch := make(chan os.Signal, 2)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-ch
		cancel()
		<-ch
		os.Exit(1)
	}()

	return ctx
}
