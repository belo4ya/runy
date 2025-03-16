package runy

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"
)

var _g = NewGroup()

// Add registers the provided Runnables to the default Group.
// Returns the Group for method chaining.
func Add(rns ...Runnable) Group {
	return _g.Add(rns...)
}

// AddF registers the provided RunnableFuncs to the default Group.
// Returns the Group for method chaining.
func AddF(rns ...RunnableFunc) Group {
	return _g.AddF(rns...)
}

// SAdd registers a SugaredRunnable to the default Group.
// The SugaredRunnable is converted to a standard Runnable using FromSugared.
// Returns the Group for method chaining.
func SAdd(rn SugaredRunnable, opts ...FromSugaredOption) Group {
	return _g.SAdd(rn, opts...)
}

// SAddF registers a SugaredRunnable created from the provided start and stop functions to the default Group.
// Returns the Group for method chaining.
func SAddF(start StartFunc, stop StopFunc, opts ...FromSugaredOption) Group {
	return _g.SAddF(start, stop, opts...)
}

// Start runs all registered Runnables in the default Group concurrently.
// This function blocks until all Runnables complete or the context is canceled.
func Start(ctx context.Context) (err error) {
	return _g.Start(ctx)
}

// Group manages a collection of Runnables that can be started together.
type Group interface {
	// Add registers the provided Runnables to the Group.
	// Returns the Group for method chaining.
	Add(...Runnable) Group

	// AddF registers the provided RunnableFuncs to the Group.
	// Returns the Group for method chaining.
	AddF(...RunnableFunc) Group

	// SAdd registers a SugaredRunnable to the Group.
	// The SugaredRunnable is converted to a standard Runnable using FromSugared.
	// Returns the Group for method chaining.
	SAdd(SugaredRunnable, ...FromSugaredOption) Group

	// SAddF registers a SugaredRunnable created from the provided start and stop functions to the Group.
	// Returns the Group for method chaining.
	SAddF(StartFunc, StopFunc, ...FromSugaredOption) Group

	// Start runs all registered Runnables concurrently.
	// This function blocks until all Runnables complete or the context is canceled.
	Start(context.Context) error
}

// NewGroup creates a new empty Group.
func NewGroup() Group {
	return &group{}
}

type group struct {
	mu        sync.Mutex
	once      sync.Once
	runnables []Runnable
}

func (g *group) Add(rns ...Runnable) Group {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.runnables = append(g.runnables, rns...)
	return g
}

func (g *group) AddF(rns ...RunnableFunc) Group {
	for _, rn := range rns {
		g.Add(rn)
	}
	return g
}

func (g *group) SAdd(rn SugaredRunnable, opts ...FromSugaredOption) Group {
	return g.Add(FromSugared(rn, opts...))
}

func (g *group) SAddF(start StartFunc, stop StopFunc, opts ...FromSugaredOption) Group {
	return g.Add(FromSugared(SugaredFromFuncs(start, stop), opts...))
}

func (g *group) Start(ctx context.Context) (err error) {
	g.once.Do(func() {
		eg, ctx := errgroup.WithContext(ctx)
		for _, rn := range g.runnables {
			rn := rn
			eg.Go(func() error {
				return rn.Start(ctx)
			})
		}
		err = eg.Wait()
	})
	return err
}
