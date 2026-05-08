package ratelimit

import (
	"fmt"
	"time"

	"github.com/yourorg/cronwrap/internal/logger"
)

// Guard wraps a Limiter with logging support and exposes a checked execution
// helper used by the runner before dispatching a job.
type Guard struct {
	limiter *Limiter
	log     *logger.Logger
}

// NewGuard creates a Guard using the provided interval and logger.
func NewGuard(interval time.Duration, log *logger.Logger) *Guard {
	return &Guard{
		limiter: New(interval),
		log:     log,
	}
}

// Check returns nil if the job is permitted to run, or a descriptive error
// if it is rate-limited. It logs a warning when a job is suppressed.
func (g *Guard) Check(jobName string) error {
	if g.limiter.Allow(jobName) {
		return nil
	}
	next := g.limiter.NextAllowed(jobName)
	msg := fmt.Sprintf("job %q is rate-limited; next allowed at %s", jobName, next.Format(time.RFC3339))
	g.log.Warn(msg, nil)
	return fmt.Errorf("%s", msg)
}

// Reset delegates to the underlying Limiter.
func (g *Guard) Reset(jobName string) {
	g.limiter.Reset(jobName)
}
