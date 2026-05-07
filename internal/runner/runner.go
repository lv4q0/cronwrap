// Package runner executes commands with retry logic, timeout, and optional locking.
package runner

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/yourorg/cronwrap/internal/lock"
)

// Options configures how a command is run.
type Options struct {
	Command     string
	Args        []string
	Timeout     time.Duration
	MaxRetries  int
	RetryDelay  time.Duration
	LockDir     string   // empty disables locking
	JobName     string
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Timeout:    30 * time.Minute,
		MaxRetries: 0,
		RetryDelay: 5 * time.Second,
	}
}

// Result holds the outcome of a command execution.
type Result struct {
	Attempts int
	Stdout   string
	Stderr   string
	Err      error
	Duration time.Duration
}

// Run executes the command described by opts, retrying on failure.
// If LockDir is set, it acquires a file lock before running.
func Run(opts Options) Result {
	if opts.LockDir != "" {
		l := lock.New(opts.LockDir, opts.JobName)
		if err := l.Acquire(); err != nil {
			return Result{Err: fmt.Errorf("lock: %w", err)}
		}
		defer l.Release()
	}

	var res Result
	maxAttempts := opts.MaxRetries + 1

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		res = runOnce(opts)
		res.Attempts = attempt
		if res.Err == nil {
			return res
		}
		if attempt < maxAttempts {
			time.Sleep(opts.RetryDelay)
		}
	}
	return res
}

func runOnce(opts Options) Result {
	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, opts.Command, opts.Args...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	if ctx.Err() == context.DeadlineExceeded {
		err = fmt.Errorf("command timed out after %s", opts.Timeout)
	}

	return Result{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Err:      err,
		Duration: duration,
	}
}
