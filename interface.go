package runy

import (
	"context"
)

// Runnable allows a component to be started.
// It's very important that Start blocks until it's done running.
type Runnable interface {
	// Start starts running the component.
	// The component will stop running when the context is closed.
	// Start blocks until the context is closed or an error occurs.
	Start(ctx context.Context) error
}

// RunnableFunc implements Runnable using a function.
// It's very important that the given function block until it's done running.
type RunnableFunc func(context.Context) error

// Start implements Runnable.
func (r RunnableFunc) Start(ctx context.Context) error {
	return r(ctx)
}

// SugaredRunnable represents a simplified (sugared) version of Runnable with separate
// Start and Stop methods instead of a single blocking Start method.
// This allows for more explicit control over the lifecycle of a component.
type SugaredRunnable interface {
	// Start begins the component operation.
	Start(ctx context.Context) error
	// Stop gracefully terminates the component operation.
	// This method is called when the context passed to FromSugared is canceled.
	Stop(ctx context.Context) error
}

// StartFunc is a function type that implements the Start method of a SugaredRunnable.
type StartFunc func(ctx context.Context) error

// StopFunc is a function type that implements the Stop method of a SugaredRunnable.
type StopFunc func(ctx context.Context) error

// SugaredFromFuncs creates a SugaredRunnable implementation from separate start and stop functions.
// If either function is nil, it will be replaced with a no-op function that returns nil.
func SugaredFromFuncs(start StartFunc, stop StopFunc) SugaredRunnable {
	if start == nil {
		start = nilFunc
	}
	if stop == nil {
		stop = nilFunc
	}
	return &sugaredFromFuncs{start: start, stop: stop}
}

func nilFunc(_ context.Context) error {
	return nil
}

type sugaredFromFuncs struct {
	start StartFunc
	stop  StopFunc
}

func (f *sugaredFromFuncs) Start(ctx context.Context) error {
	return f.start(ctx)
}

func (f *sugaredFromFuncs) Stop(ctx context.Context) error {
	return f.stop(ctx)
}

// FromSugared converts a SugaredRunnable into a standard Runnable.
// The returned Runnable will run the Start method of the SugaredRunnable
// and will call the Stop method when the context is canceled.
// This allows SugaredRunnable implementations to be used anywhere a Runnable is required.
func FromSugared(rn SugaredRunnable, opts ...FromSugaredOption) Runnable {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}
	return RunnableFunc(func(ctx context.Context) error {
		errCh := make(chan error, 1)
		go func() {
			errCh <- rn.Start(ctx)
		}()
		select {
		case <-ctx.Done():
			return rn.Stop(ctx)
		case err := <-errCh:
			return err
		}
	})
}

type fromSugaredOptions struct {
}

func defaultOptions() fromSugaredOptions {
	return fromSugaredOptions{}
}

// FromSugaredOption is a function that modifies the behavior of FromSugared.
type FromSugaredOption func(o *fromSugaredOptions)
