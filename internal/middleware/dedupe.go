package middleware

import (
	"context"
	"fmt"

	"github.com/yourorg/cronwrap/internal/lock"
)

// DedupeMiddleware prevents concurrent execution of the same job by acquiring
// a file-based lock before running. If the lock is already held, the job is
// skipped rather than queued.
func DedupeMiddleware(lk *lock.Lock) func(next func(ctx context.Context) error) func(ctx context.Context) error {
	return func(next func(ctx context.Context) error) func(ctx context.Context) error {
		return func(ctx context.Context) error {
			acquired, err := lk.TryAcquire()
			if err != nil {
				return fmt.Errorf("dedupe: lock check failed: %w", err)
			}
			if !acquired {
				return ErrJobAlreadyRunning
			}
			defer lk.Release() //nolint:errcheck
			return next(ctx)
		}
	}
}

// ErrJobAlreadyRunning is returned when a deduplicated job is skipped because
// another instance is already running.
var ErrJobAlreadyRunning = fmt.Errorf("dedupe: job already running")
