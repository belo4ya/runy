package runy

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"
)

var g = NewGroup()

func Add(rns ...Runnable) Group {
	return g.Add(rns...)
}

func AddF(rns ...RunnableFunc) Group {
	return g.AddF(rns...)
}

func SAdd(rn SugaredRunnable, opts ...FromSugaredOption) Group {
	return g.SAdd(rn, opts...)
}

func SAddF(start StartFunc, stop StopFunc, opts ...FromSugaredOption) Group {
	return g.SAddF(start, stop, opts...)
}

func Start(ctx context.Context) (err error) {
	return g.Start(ctx)
}

type Group interface {
	Add(...Runnable) Group
	AddF(...RunnableFunc) Group
	SAdd(SugaredRunnable, ...FromSugaredOption) Group
	SAddF(StartFunc, StopFunc, ...FromSugaredOption) Group
	Start(context.Context) error
}

func NewGroup() Group {
	return &group{}
}

type group struct {
	mu        sync.Mutex
	once      sync.Once
	runnables []Runnable
}

func (m *group) Add(rns ...Runnable) Group {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.runnables = append(m.runnables, rns...)
	return m
}

func (m *group) AddF(rns ...RunnableFunc) Group {
	for _, rn := range rns {
		m.Add(rn)
	}
	return m
}

func (m *group) SAdd(rn SugaredRunnable, opts ...FromSugaredOption) Group {
	return m.Add(FromSugared(rn, opts...))
}

func (m *group) SAddF(start StartFunc, stop StopFunc, opts ...FromSugaredOption) Group {
	return m.Add(FromSugared(SugaredFromFuncs(start, stop), opts...))
}

func (m *group) Start(ctx context.Context) (err error) {
	m.once.Do(func() {
		eg, ctx := errgroup.WithContext(ctx)
		for _, rn := range m.runnables {
			rn := rn
			eg.Go(func() error {
				return rn.Start(ctx)
			})
		}
		err = eg.Wait()
	})
	return err
}
