// Package middleware provides a composable pipeline for wrapping job execution
// with cross-cutting concerns such as locking, rate limiting, circuit breaking,
// and timeout enforcement.
package middleware

import "context"

// RunFunc is the signature of a job execution function.
type RunFunc func(ctx context.Context) error

// Middleware wraps a RunFunc with additional behaviour.
type Middleware func(next RunFunc) RunFunc

// Chain composes multiple middlewares into a single Middleware.
// Middlewares are applied in the order they are provided, so the first
// middleware in the slice is the outermost wrapper.
func Chain(mws ...Middleware) Middleware {
	return func(next RunFunc) RunFunc {
		for i := len(mws) - 1; i >= 0; i-- {
			next = mws[i](next)
		}
		return next
	}
}

// Apply wraps fn with the provided middlewares and returns the resulting RunFunc.
func Apply(fn RunFunc, mws ...Middleware) RunFunc {
	return Chain(mws...)(fn)
}
