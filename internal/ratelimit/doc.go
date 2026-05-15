// Package ratelimit provides rate-limiting primitives for cronwrap jobs.
//
// # Overview
//
// The package exposes two main building blocks:
//
//   - [Store] — a thread-safe, in-memory store that tracks the last run time
//     for each job key and answers "is this key allowed to run now?" queries.
//
//   - [Guard] — a higher-level helper that wraps a Store and exposes a single
//     Allow(key) method suitable for use inside middleware chains.
//
// # Usage
//
// Typical usage via the middleware layer:
//
//	store := ratelimit.New()
//	guard := ratelimit.NewGuard("my-job", 5*time.Minute, store, logger)
//
//	// inside a middleware chain:
//	if err := guard.Allow(ctx); err != nil {
//		// job was skipped
//	}
//
// For finer control (e.g. obtaining the next-allowed time for a skip
// callback), use Store.Check and Store.Record directly.
package ratelimit
