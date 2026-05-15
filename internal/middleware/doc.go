// Package middleware provides composable middleware functions for cron job
// execution. Each middleware wraps a job function with a specific concern:
//
//   - Chain / Apply  – compose multiple middlewares in order
//   - LogMiddleware  – structured start/finish/error logging
//   - RecoverMiddleware – catch panics and return them as errors
//   - RetryMiddleware – retry failed jobs with configurable back-off
//   - TimeoutMiddleware – enforce a per-run deadline
//   - ConcurrencyMiddleware – cap the number of simultaneous runs
//   - DedupeMiddleware – prevent overlapping runs via a file lock
//   - ThrottleMiddleware – rate-limit runs to a minimum interval
//   - CircuitMiddleware – skip runs when too many recent failures occur
//   - NewMetricsMiddleware – record success/failure counters and durations
//   - NewNotifyMiddleware – dispatch notifications on job completion
//
// Middlewares are intended to be composed with Chain so the execution order
// is explicit and predictable.
package middleware
