package async

import (
	"context"
	"sync"
	"time"
)

// Run `asyncFunc` in go-routine, automatically increments WaitGroup counter before start and decrements it after
// exit from go-routine.
func Run(wg *sync.WaitGroup, asyncFunc func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		asyncFunc()
	}()
}

// RunPeriodically runs `f` periodically every `d`. If `f` returns error, then periodic execution will be terminated.
func RunPeriodically(wg *sync.WaitGroup, ctx context.Context, d time.Duration, f func(context.Context) error) {
	Run(wg, func() { _ = runPeriodically(ctx, d, f) })
}

func runPeriodically(ctx context.Context, d time.Duration, f func(context.Context) error) error {
	ticker := time.NewTicker(d)
	defer ticker.Stop()

	if err := f(ctx); err != nil {
		return err
	}

	for {
		select {
		case <-ticker.C:
			if err := f(ctx); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
