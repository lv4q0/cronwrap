package middleware

import (
	"fmt"
	"time"

	"github.com/cronwrap/internal/logger"
)

// LogMiddleware returns a middleware that logs job start, finish, duration,
// and any error produced by the wrapped handler.
func LogMiddleware(log *logger.Logger, jobName string) func(func() error) func() error {
	return func(next func() error) func() error {
		return func() error {
			log.Info("job started", map[string]any{
				"job": jobName,
			})

			start := time.Now()
			err := next()
			duration := time.Since(start)

			fields := map[string]any{
				"job":      jobName,
				"duration": fmt.Sprintf("%.3fs", duration.Seconds()),
			}

			if err != nil {
				fields["error"] = err.Error()
				log.Error("job failed", fields)
			} else {
				log.Info("job finished", fields)
			}

			return err
		}
	}
}
