// Package middleware provides composable middleware for cronwrap job execution.
//
// # Tag Middleware
//
// The tag sub-feature allows arbitrary key/value labels to be attached to a
// job execution context and consumed by downstream middleware.
//
// Typical usage:
//
//	chain := middleware.Chain(
//		middleware.TagMiddleware(middleware.Tags{
//			"env":   "production",
//			"owner": "platform-team",
//		}),
//		middleware.TagLoggingMiddleware(log),
//		middleware.LogMiddleware(log),
//	)
//
// Tags are stored in the context under [TagKey] and can be retrieved anywhere
// in the call stack with [TagsFromContext].
package middleware
