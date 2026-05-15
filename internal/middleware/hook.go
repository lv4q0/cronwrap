package middleware

import (
	"context"
	"time"
)

// HookFn is a function called before or after a job runs.
type HookFn func(ctx context.Context, job string, elapsed time.Duration, err error)

// HookMiddleware calls optional before/after hooks around a job execution.
// BeforeFn receives a zero elapsed duration and a nil error.
// AfterFn receives the actual elapsed duration and any error returned by the job.
func HookMiddleware(job string, before, after HookFn) func(JobFunc) JobFunc {
	return func(next JobFunc) JobFunc {
		return func(ctx context.Context) error {
			if before != nil {
				before(ctx, job, 0, nil)
			}

			start := time.Now()
			err := next(ctx)
			elapsed := time.Since(start)

			if after != nil {
				after(ctx, job, elapsed, err)
			}

			return err
		}
	}
}
