package middleware

import (
	"context"
	"fmt"

	"github.com/cronwrap/cronwrap/internal/circuit"
	"github.com/cronwrap/cronwrap/internal/logger"
)

// CircuitMiddleware wraps a job with a circuit breaker guard. If the breaker
// is open (too many recent failures) the job is skipped and an error is
// returned so callers can react (e.g. alert, log, skip retry).
func CircuitMiddleware(breaker *circuit.Breaker, log *logger.Logger) func(next func(context.Context) error) func(context.Context) error {
	guard := circuit.NewGuard(breaker, log)
	return func(next func(context.Context) error) func(context.Context) error {
		return func(ctx context.Context) error {
			err := guard.Run(ctx, next)
			if err != nil {
				return fmt.Errorf("circuit: %w", err)
			}
			return nil
		}
	}
}
