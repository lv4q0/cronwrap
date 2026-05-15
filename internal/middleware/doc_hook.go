// Package middleware provides composable middleware for cronwrap job execution.
//
// # Hook Middleware
//
// HookMiddleware allows callers to register arbitrary before/after callbacks
// around job execution without modifying the job itself. This is useful for
// side-effects such as emitting custom metrics, updating external systems, or
// integrating with third-party observability platforms.
//
// Usage:
//
//	before := func(ctx context.Context, job string, _ time.Duration, _ error) {
//		fmt.Printf("starting job %s\n", job)
//	}
//	after := func(ctx context.Context, job string, elapsed time.Duration, err error) {
//		fmt.Printf("job %s finished in %v err=%v\n", job, elapsed, err)
//	}
//
//	chain := middleware.Chain(
//		middleware.HookMiddleware("backup", before, after),
//	)
//	chain.Apply(myJob)(ctx)
//
// Either hook may be nil, in which case it is silently skipped.
package middleware
