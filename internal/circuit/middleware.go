package circuit

import (
	"context"
	"fmt"
	"time"

	"github.com/yourorg/cronwrap/internal/logger"
)

// Guard wraps a Breaker and enforces it around a job execution function.
type Guard struct {
	breaker *Breaker
	log     *logger.Logger
	jobName string
}

// NewGuard creates a Guard for the named job.
func NewGuard(jobName string, threshold int, cooldown time.Duration, log *logger.Logger) *Guard {
	return &Guard{
		breaker: New(threshold, cooldown),
		log:     log,
		jobName: jobName,
	}
}

// RunFunc is the signature of the function the Guard will protect.
type RunFunc func(ctx context.Context) error

// Execute checks the circuit before calling fn. It records success or
// failure and logs state transitions.
func (g *Guard) Execute(ctx context.Context, fn RunFunc) error {
	ok, state := g.breaker.Allow()
	if !ok {
		snap := g.breaker.Snapshot()
		g.log.Warn("circuit open, skipping job", map[string]any{
			"job":      g.jobName,
			"failures": snap.Failures,
			"openedAt": snap.OpenedAt.Format(time.RFC3339),
		})
		return fmt.Errorf("circuit breaker open for job %q (%d consecutive failures)", g.jobName, snap.Failures)
	}

	if state == StateHalfOpen {
		g.log.Info("circuit half-open, probing job", map[string]any{"job": g.jobName})
	}

	err := fn(ctx)
	if err != nil {
		g.breaker.RecordFailure()
		snap := g.breaker.Snapshot()
		g.log.Warn("job failed, circuit updated", map[string]any{
			"job":      g.jobName,
			"failures": snap.Failures,
			"state":    snap.State.String(),
		})
		return err
	}

	g.breaker.RecordSuccess()
	if state == StateHalfOpen {
		g.log.Info("circuit closed after successful probe", map[string]any{"job": g.jobName})
	}
	return nil
}

// Breaker exposes the underlying Breaker for inspection.
func (g *Guard) Breaker() *Breaker { return g.breaker }
