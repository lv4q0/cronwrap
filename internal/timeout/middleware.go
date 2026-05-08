package timeout

import (
	"context"
	"fmt"
	"time"
)

// Logger is the minimal interface required by the timeout guard.
type Logger interface {
	Warn(msg string, fields map[string]any)
	Error(msg string, fields map[string]any)
}

// Guard wraps a job runner function with timeout enforcement.
type Guard struct {
	policy Policy
	log    Logger
}

// NewGuard creates a Guard with the given policy and logger.
func NewGuard(p Policy, log Logger) *Guard {
	return &Guard{policy: p, log: log}
}

// RunFunc is the signature of a job execution function.
type RunFunc func(ctx context.Context) error

// Run executes fn within the policy's time limit.
// If the limit is exceeded the context is cancelled, ErrTimeout is returned,
// and a warning is emitted on the logger.
func (g *Guard) Run(ctx context.Context, jobName string, fn RunFunc) error {
	if g.policy.Limit <= 0 {
		return fn(ctx)
	}

	derived, cancel := g.policy.WithContext(ctx)
	defer cancel()

	start := time.Now()
	err := fn(derived)
	elapsed := time.Since(start)

	if Exceeded(err) || derived.Err() == context.DeadlineExceeded {
		g.log.Warn("job timed out", map[string]any{
			"job":     jobName,
			"limit":   g.policy.Limit.String(),
			"elapsed": elapsed.String(),
		})
		return fmt.Errorf("%w: job=%s limit=%s elapsed=%s",
			ErrTimeout, jobName, g.policy.Limit, elapsed)
	}
	return err
}
