// Package middleware provides composable middleware for wrapping cron job execution.
//
// Each middleware implements a Wrap method that accepts and returns a job function
// of the form func(ctx context.Context) error. Middlewares can be composed using
// the Chain helper:
//
//	chain := middleware.Chain(
//		middleware.RecoverMiddleware,
//		middleware.LogMiddleware(logger, "my-job"),
//		middleware.NewMetricsMiddleware(store).Wrap,
//		middleware.NewNotifyMiddleware(dispatcher, "my-job").Wrap,
//	)
//	
//	if err := middleware.Apply(chain, jobFunc)(ctx); err != nil {
//		// handle error
//	}
//
// Available middleware:
//
//   - RecoverMiddleware  – catches panics and converts them to errors
//   - LogMiddleware      – logs job start, finish, and errors
//   - MetricsMiddleware  – records success/failure counts and duration
//   - NotifyMiddleware   – dispatches notification events on completion
package middleware
