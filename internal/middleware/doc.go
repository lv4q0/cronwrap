// Package middleware provides composable middleware for cronwrap job execution.
//
// Middleware wraps a job function of the form:
//
//	func(ctx context.Context) error
//
// and adds cross-cutting behaviour such as logging, metrics, retry, timeout,
// panic recovery, notifications, audit logging, concurrency limiting, and
// deduplication.
//
// # Composition
//
// Use [Chain] to combine multiple middlewares into a single wrapper, and
// [Apply] to apply the chain to a concrete job function:
//
//	chain := middleware.Chain(
//		middleware.RecoverMiddleware,
//		middleware.LogMiddleware(logger),
//		middleware.NewMetricsMiddleware(store),
//		middleware.RetryMiddleware(opts, backoff),
//		middleware.DedupeMiddleware(lock),
//	)
//	runnable := middleware.Apply(chain, myJob)
//
// # Deduplication
//
// [DedupeMiddleware] uses a file-based lock (see internal/lock) to ensure that
// only one instance of a job runs at a time. If the lock cannot be acquired the
// job is skipped and [ErrJobAlreadyRunning] is returned. The lock is always
// released when the wrapped job returns, even on error or panic (when combined
// with [RecoverMiddleware]).
package middleware
