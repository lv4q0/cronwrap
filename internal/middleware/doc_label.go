// Package middleware provides composable middleware for cronwrap job execution.
//
// # Label middleware
//
// LabelMiddleware injects a map of key/value string pairs into the job context
// so that downstream middleware and job code can read structured metadata about
// the current execution.
//
// Usage:
//
//	chain := middleware.Chain(
//		middleware.LabelMiddleware(map[string]string{
//			"env":  "production",
//			"team": "platform",
//		}),
//		middleware.LogMiddleware(logger),
//	)
//
// Multiple LabelMiddleware layers are merged, with inner layers taking
// precedence over outer ones for duplicate keys.
package middleware
