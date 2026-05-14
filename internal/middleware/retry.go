package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/cronwrap/internal/backoff"
)

// RetryConfig controls how the retry middleware behaves.
type RetryConfig struct {
	// MaxAttempts is the total number of attempts (1 = no retry).
	MaxAttempts int
	// Strategy determines the delay between attempts.
	Strategy backoff.Strategy
}

// DefaultRetryConfig returns a sensible default retry configuration.
var DefaultRetryConfig = RetryConfig{
	MaxAttempts: 3,
	Strategy:    backoff.Default,
}

// RetryMiddleware wraps a job and retries it on failure according to cfg.
func RetryMiddleware(cfg RetryConfig) func(func(context.Context) error) func(context.Context) error {
	if cfg.MaxAttempts < 1 {
		cfg.MaxAttempts = 1
	}
	if cfg.Strategy == nil {
		cfg.Strategy = backoff.Default
	}

	return func(next func(context.Context) error) func(context.Context) error {
		return func(ctx context.Context) error {
			var lastErr error
			for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
				if attempt > 0 {
					delay := cfg.Strategy.Delay(attempt)
					select {
					case <-time.After(delay):
					case <-ctx.Done():
						return fmt.Errorf("retry aborted after %d attempt(s): %w", attempt, ctx.Err())
					}
				}
				lastErr = next(ctx)
				if lastErr == nil {
					return nil
				}
			}
			return fmt.Errorf("job failed after %d attempt(s): %w", cfg.MaxAttempts, lastErr)
		}
	}
}
