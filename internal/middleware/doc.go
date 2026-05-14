// Package middleware provides composable job execution middleware for cronwrap.
//
// Each middleware wraps a job function (func(ctx context.Context) error) and
// adds cross-cutting behaviour such as logging, metrics recording, panic
// recovery, retry logic, notification dispatch, audit logging, and concurrency
// control.
//
// Middlewares are combined using Chain and Apply:
//
//	chain := middleware.Chain(
//		middleware.RecoverMiddleware,
//		middleware.LogMiddleware(log),
//		middleware.NewMetricsMiddleware(store).Wrap,
//		middleware.NewConcurrencyMiddleware(4, false).Wrap,
//		middleware.RetryMiddleware(opts),
//	)
//	if err := middleware.Apply(ctx, chain, jobFn); err != nil {
//		// handle
//	}
//
// Available middleware:
//
//   - Chain / Apply        — compose and execute a slice of middlewares
//   - RecoverMiddleware    — convert panics into errors
//   - LogMiddleware        — structured start/finish logging
//   - NewMetricsMiddleware — record success/failure counters and duration
//   - NewNotifyMiddleware  — dispatch notifications on job completion
//   - RetryMiddleware      — configurable retry with backoff
//   - NewConcurrencyMiddleware — cap simultaneous executions
package middleware
