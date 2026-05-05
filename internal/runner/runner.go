package runner

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

// Result holds the outcome of a command execution.
type Result struct {
	Command  string
	ExitCode int
	Stdout   string
	Stderr   string
	Duration time.Duration
	Err      error
}

// Options configures how a command is run.
type Options struct {
	Timeout    time.Duration
	MaxRetries int
	RetryDelay time.Duration
}

// DefaultOptions returns sensible defaults for command execution.
func DefaultOptions() Options {
	return Options{
		Timeout:    30 * time.Minute,
		MaxRetries: 0,
		RetryDelay: 5 * time.Second,
	}
}

// Run executes the given command with the provided options, retrying on failure
// up to MaxRetries times. It returns the Result of the last attempt.
func Run(ctx context.Context, command string, args []string, opts Options) Result {
	var last Result
	attempts := opts.MaxRetries + 1

	for i := 0; i < attempts; i++ {
		if i > 0 {
			select {
			case <-time.After(opts.RetryDelay):
			case <-ctx.Done():
				last.Err = fmt.Errorf("context cancelled before retry %d: %w", i+1, ctx.Err())
				return last
			}
		}

		last = runOnce(ctx, command, args, opts.Timeout)
		if last.Err == nil && last.ExitCode == 0 {
			return last
		}
	}

	return last
}

func runOnce(ctx context.Context, command string, args []string, timeout time.Duration) Result {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()
	cmd := exec.CommandContext(ctx, command, args...)

	var stdoutBuf, stderrBuf strings.Builder
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	duration := time.Since(start)

	exitCode := 0
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}

	return Result{
		Command:  command,
		ExitCode: exitCode,
		Stdout:   stdoutBuf.String(),
		Stderr:   stderrBuf.String(),
		Duration: duration,
		Err:      err,
	}
}
