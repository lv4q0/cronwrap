// Package middleware provides composable middleware for cronwrap job execution.
//
// # Env Middleware
//
// EnvMiddleware injects a static map of environment variables into the job
// context so that downstream handlers can read them via EnvFromContext.
//
// When setProcess is true the variables are also written into the OS process
// environment for the duration of the job and restored (or unset) afterwards.
// This is useful when wrapping external commands that read from the environment
// directly.
//
// Example usage:
//
//	vars := map[string]string{
//		"APP_ENV":  "production",
//		"LOG_LEVEL": "info",
//	}
//	chain := middleware.Chain(
//		middleware.EnvMiddleware(vars, true),
//		middleware.LogMiddleware(logger),
//	)
//	runner.Run(ctx, cmd, opts, chain)
package middleware
