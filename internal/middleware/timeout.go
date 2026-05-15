package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/cronwrap/cronwrap/internal/timeout"
)

// TimeoutMiddleware wraps a job function with a deadline. If the job does not
// complete within the configured duration, the context is cancelled and an
// error describing the timeout is returned.
//
// A zero or negative limit is treated as "no timeout" and the job is executed
// without any deadline.
func TimeoutMiddleware(limit time.Duration) func(next func(ctx context.Context) error) func(ctx context.Context) error {
	return func(next func(ctx context.Context) error) func(ctx context.Context) error {
		return func(ctx context.Context) error {
			if limit <= 0 {
				return next(ctx)
			}

			policy := timeout.Policy{Limit: limit}
			ctxWithDeadline, cancel := timeout.WithContext(ctx, policy)
			defer cancel()

			err := next(ctxWithDeadline)
			if err != nil && timeout.Exceeded(err) {
				return fmt.Errorf("job timed out after %s: %w", limit, err)
			}
			return err
		}
	}
}
