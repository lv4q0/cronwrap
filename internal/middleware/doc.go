// Package middleware defines a composable RunFunc pipeline for cronwrap jobs.
//
// # Overview
//
// Each job is ultimately represented as a RunFunc:
//
//	type RunFunc func(ctx context.Context) error
//
// A Middleware wraps a RunFunc to add cross-cutting behaviour before or after
// execution. Multiple middlewares are composed with Chain or Apply:
//
//	fn := middleware.Apply(
//		job.Run,
//		middleware.RecoverMiddleware(),
//		timeout.NewGuard(...),
//		circuit.NewGuard(...),
//		ratelimit.NewGuard(...),
//	)
//
// Middlewares are applied outermost-first, matching the order in which they
// appear in the argument list.
package middleware
