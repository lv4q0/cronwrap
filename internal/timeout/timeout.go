// Package timeout provides utilities for enforcing execution time limits
// on cron job commands, with configurable grace periods and kill signals.
package timeout

import (
	"context"
	"fmt"
	"time"
)

// ErrTimeout is returned when a job exceeds its allowed duration.
var ErrTimeout = fmt.Errorf("job exceeded time limit")

// Policy defines how a timeout is enforced.
type Policy struct {
	// Limit is the maximum duration before the context is cancelled.
	Limit time.Duration
	// GracePeriod is additional time allowed after cancellation before
	// the caller should consider the job forcibly killed.
	GracePeriod time.Duration
}

// Default returns a Policy with sensible defaults.
func Default() Policy {
	return Policy{
		Limit:       30 * time.Minute,
		GracePeriod: 5 * time.Second,
	}
}

// WithContext returns a derived context that is cancelled after p.Limit.
// The cancel function must be called to release resources.
func (p Policy) WithContext(parent context.Context) (context.Context, context.CancelFunc) {
	if p.Limit <= 0 {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, p.Limit)
}

// Exceeded reports whether the given error represents a timeout condition.
func Exceeded(err error) bool {
	if err == nil {
		return false
	}
	return err == context.DeadlineExceeded || err == ErrTimeout
}

// Wrap converts a context.DeadlineExceeded error into ErrTimeout so callers
// can use a single sentinel value regardless of how the timeout was triggered.
func Wrap(err error) error {
	if err == context.DeadlineExceeded {
		return ErrTimeout
	}
	return err
}
