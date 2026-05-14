package middleware

import (
	"context"
	"time"

	"github.com/cronwrap/internal/notify"
)

// NotifyMiddleware sends a notification event after a job completes.
// It always calls next and fires the notification regardless of success or failure.
type NotifyMiddleware struct {
	dispatcher *notify.Dispatcher
	jobName    string
}

// NewNotifyMiddleware creates a middleware that dispatches job events via the given dispatcher.
func NewNotifyMiddleware(d *notify.Dispatcher, jobName string) *NotifyMiddleware {
	return &NotifyMiddleware{dispatcher: d, jobName: jobName}
}

// Wrap implements the Middleware interface.
func (m *NotifyMiddleware) Wrap(next func(ctx context.Context) error) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		start := time.Now()
		err := next(ctx)
		duration := time.Since(start)

		level := notify.LevelInfo
		message := "job completed successfully"
		if err != nil {
			level = notify.LevelError
			message = err.Error()
		}

		event := notify.Event{
			JobName:  m.jobName,
			Level:    level,
			Message:  message,
			Duration: duration,
			Err:      err,
		}

		_ = m.dispatcher.Dispatch(ctx, event)
		return err
	}
}
