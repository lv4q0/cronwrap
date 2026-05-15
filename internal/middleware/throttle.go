package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/cronwrap/internal/ratelimit"
)

// ThrottleConfig controls how the throttle middleware behaves.
type ThrottleConfig struct {
	// JobName is used as the rate-limit key.
	JobName string
	// Interval is the minimum duration between allowed runs.
	Interval time.Duration
	// OnSkip is called when a run is skipped due to throttling (optional).
	OnSkip func(jobName string, next time.Time)
}

// ThrottleMiddleware prevents a job from running more frequently than the
// configured interval. It wraps the ratelimit.Guard to provide a clean
// middleware-compatible interface with optional skip callbacks.
func ThrottleMiddleware(cfg ThrottleConfig, store *ratelimit.Store) func(JobFunc) JobFunc {
	guard := ratelimit.NewGuard(cfg.JobName, cfg.Interval, store, nil)
	return func(next JobFunc) JobFunc {
		return func(ctx context.Context) error {
			allowed, nextAllowed := store.Check(cfg.JobName, cfg.Interval)
			if !allowed {
				if cfg.OnSkip != nil {
					cfg.OnSkip(cfg.JobName, nextAllowed)
				}
				return fmt.Errorf("throttle: job %q skipped, next allowed at %s",
					cfg.JobName, nextAllowed.Format(time.RFC3339))
			}
			_ = guard // guard handles Allow internally; we used Check for skip info
			return next(ctx)
		}
	}
}
