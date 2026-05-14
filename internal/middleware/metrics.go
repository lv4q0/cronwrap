package middleware

import (
	"time"

	"github.com/cronwrap/internal/metrics"
)

// MetricsMiddleware records success/failure and duration for each job run
// into the provided metrics.Store.
type MetricsMiddleware struct {
	store *metrics.Store
	jobName string
}

// NewMetricsMiddleware creates a MetricsMiddleware that records results
// for the given job name into store.
func NewMetricsMiddleware(store *metrics.Store, jobName string) *MetricsMiddleware {
	return &MetricsMiddleware{store: store, jobName: jobName}
}

// Wrap implements the Middleware interface. It times the inner function,
// then records a success or failure in the metrics store.
func (m *MetricsMiddleware) Wrap(next func() error) func() error {
	return func() error {
		start := time.Now()
		err := next()
		duration := time.Since(start)

		if err != nil {
			m.store.RecordFailure(m.jobName, duration)
		} else {
			m.store.RecordSuccess(m.jobName, duration)
		}

		return err
	}
}
