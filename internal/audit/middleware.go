package audit

import (
	"context"
	"time"
)

// JobFunc is the signature expected by middleware chains in cronwrap.
type JobFunc func(ctx context.Context) error

// Middleware wraps a JobFunc to record an audit entry after each execution.
type Middleware struct {
	log     *Log
	jobName string
	trigger TriggerKind
}

// NewMiddleware returns an audit Middleware for the given job.
// trigger should reflect how the job was initiated (scheduled, manual, retry).
func NewMiddleware(log *Log, jobName string, trigger TriggerKind) *Middleware {
	return &Middleware{log: log, jobName: jobName, trigger: trigger}
}

// Wrap returns a new JobFunc that records an audit entry and then calls next.
func (m *Middleware) Wrap(next JobFunc) JobFunc {
	return func(ctx context.Context) error {
		start := time.Now()
		err := next(ctx)
		dur := time.Since(start).Seconds()

		e := Entry{
			JobName:  m.jobName,
			Trigger:  m.trigger,
			Success:  err == nil,
			Duration: dur,
		}
		if err != nil {
			e.Message = err.Error()
		}

		// Best-effort: do not mask the original job error.
		_ = m.log.Record(e)
		return err
	}
}
