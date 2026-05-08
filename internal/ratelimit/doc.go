// Package ratelimit provides a lightweight token-bucket-style rate limiter
// for cronwrap job executions.
//
// The Limiter type enforces a configurable minimum interval between successive
// runs of any named job, preventing runaway retries or overlapping executions
// when a cron schedule fires more frequently than intended.
//
// The Guard type wraps a Limiter with structured logging and integrates
// cleanly with the runner pipeline: call Guard.Check before executing a job
// and it will either permit the run or return a descriptive error and emit a
// warning log entry.
//
// Typical usage:
//
//	guard := ratelimit.NewGuard(5*time.Minute, log)
//	if err := guard.Check(jobName); err != nil {
//		// skip this execution
//		return
//	}
//	// proceed with job
package ratelimit
