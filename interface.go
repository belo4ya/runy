package runy

import (
	"context"
)

type Runnable interface {
	Start(ctx context.Context) error
}

type RunnableFunc func(context.Context) error

func (r RunnableFunc) Start(ctx context.Context) error {
	return r(ctx)
}

type SugaredRunnable interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type StartFunc func(ctx context.Context) error

type StopFunc func(ctx context.Context) error

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

type FromSugaredOption func(o *fromSugaredOptions)
