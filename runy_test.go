package runy

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func setupTest(t *testing.T) *group {
	t.Helper()
	t.Cleanup(func() { _g = NewGroup() })
	return _g.(*group)
}

func TestRuny(t *testing.T) {
	t.Run("registration", func(t *testing.T) {
		g := setupTest(t)

		runF := func(ctx context.Context) error { return nil }
		run := RunnableFunc(runF)

		// Test individual registration methods
		Add(run)
		AddF(runF)
		assert.Equal(t, 2, len(g.runnables))

		// Test method chaining with multiple arguments
		Add(run).Add(run, run)
		assert.Equal(t, 5, len(g.runnables))

		AddF(runF).AddF(runF, runF)
		assert.Equal(t, 8, len(g.runnables))

		// Test mixed chaining
		AddF(runF).Add(run).AddF(runF).Add(run)
		assert.Equal(t, 12, len(g.runnables))
	})

	t.Run("registration sugared", func(t *testing.T) {
		g := setupTest(t)

		runStart := func(ctx context.Context) error { return nil }
		runStop := func(ctx context.Context) error { return nil }
		runS := SugaredFromFuncs(runStart, runStop)

		// Test individual registration methods
		SAdd(runS)
		SAddF(runStart, runStop)
		assert.Equal(t, 2, len(g.runnables))

		// Test method chaining
		SAdd(runS).SAdd(runS)
		assert.Equal(t, 4, len(g.runnables))

		SAddF(runStart, runStop).SAddF(runStart, runStop)
		assert.Equal(t, 6, len(g.runnables))

		// Test mixed chaining
		SAddF(runStart, runStop).SAdd(runS).SAddF(runStart, runStop).SAdd(runS)
		assert.Equal(t, 10, len(g.runnables))
	})

	t.Run("concurrent execution", func(t *testing.T) {
		setupTest(t)

		var readyWg, doneWg sync.WaitGroup
		readyWg.Add(2)
		doneWg.Add(2)

		AddF(
			func(ctx context.Context) error {
				readyWg.Done()
				<-ctx.Done()
				doneWg.Done()
				return nil
			},
			func(ctx context.Context) error {
				readyWg.Done()
				<-ctx.Done()
				doneWg.Done()
				return nil
			},
		)

		ctx, cancel := context.WithCancel(context.Background())

		errCh := make(chan error, 1)
		go func() {
			errCh <- Start(ctx)
			close(errCh)
		}()

		readyCh, doneCh := make(chan struct{}), make(chan struct{})
		go func() {
			readyWg.Wait()
			close(readyCh)
			doneWg.Wait()
			close(doneCh)
		}()

		select {
		case <-readyCh:
			// All runnables started successfully
		case <-time.After(time.Second):
			assert.Fail(t, "runnables didn't start in time")
		}

		// Cancel context and verify termination
		cancel()

		select {
		case <-doneCh:
			// All runnables stopped successfully
		case <-time.After(time.Second):
			assert.Fail(t, "runnables didn't terminate in time")
		}

		select {
		case err := <-errCh:
			assert.NoError(t, err, "expected no errors after cancellation")
		case <-time.After(time.Second):
			assert.Fail(t, "runy.Start() didn't return in time")
		}
	})

	t.Run("error propagation", func(t *testing.T) {
		tests := []struct {
			name      string
			runnables []RunnableFunc
			wantErr   assert.ErrorAssertionFunc
		}{
			{
				name: "no errors",
				runnables: []RunnableFunc{
					func(ctx context.Context) error { return nil },
					func(ctx context.Context) error { return nil },
				},
				wantErr: assert.NoError,
			},
			{
				name: "with error",
				runnables: []RunnableFunc{
					func(ctx context.Context) error { return nil },
					func(ctx context.Context) error { return assert.AnError },
				},
				wantErr: assert.Error,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				setupTest(t)
				AddF(tt.runnables...)
				tt.wantErr(t, Start(context.Background()))
			})
		}
	})
}
