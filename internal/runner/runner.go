package runner

import (
	"context"
	"errors"
	"os/exec"
	"time"

	"github.com/cronwrap/cronwrap/internal/logger"
)

// Options configures the behaviour of Run.
type Options struct {
	// MaxAttempts is the total number of times the command will be tried.
	MaxAttempts int
	// RetryDelay is the pause between consecutive attempts.
	RetryDelay time.Duration
	// Timeout is the maximum duration for a single attempt (0 = no timeout).
	Timeout time.Duration
	// Logger receives structured log output. If nil, a default stderr logger is used.
	Logger *logger.Logger
	// JobID is an identifier included in log lines.
	JobID string
}

// DefaultOptions returns sensible defaults: 1 attempt, no retry delay, no timeout.
func DefaultOptions() Options {
	return Options{
		MaxAttempts: 1,
		RetryDelay:  0,
		Timeout:     0,
	}
}

// Run executes the given command according to opts, retrying on failure.
// It returns the last error encountered, or nil on success.
func Run(ctx context.Context, opts Options, name string, args ...string) error {
	if opts.MaxAttempts < 1 {
		opts.MaxAttempts = 1
	}
	log := opts.Logger
	if log == nil {
		jobID := opts.JobID
		if jobID == "" {
			jobID = name
		}
		log = logger.NewDefault(jobID)
	}

	var lastErr error
	for attempt := 1; attempt <= opts.MaxAttempts; attempt++ {
		log.Info("starting attempt", map[string]any{"attempt": attempt, "of": opts.MaxAttempts})
		lastErr = runOnce(ctx, opts, log, attempt, name, args...)
		if lastErr == nil {
			log.Info("command succeeded", map[string]any{"attempt": attempt})
			return nil
		}
		log.Error("attempt failed", map[string]any{"attempt": attempt, "error": lastErr})
		if attempt < opts.MaxAttempts && opts.RetryDelay > 0 {
			log.Info("waiting before retry", map[string]any{"delay": opts.RetryDelay})
			select {
			case <-time.After(opts.RetryDelay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	return lastErr
}

// runOnce executes the command once, honouring the per-attempt timeout.
func runOnce(ctx context.Context, opts Options, log *logger.Logger, attempt int, name string, args ...string) error {
	runCtx := ctx
	var cancel context.CancelFunc
	if opts.Timeout > 0 {
		runCtx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(runCtx, name, args...)
	out, err := cmd.CombinedOutput()
	if len(out) > 0 {
		log.Debug("command output", map[string]any{"output": string(out), "attempt": attempt})
	}
	if err != nil {
		if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
			return context.DeadlineExceeded
		}
		return err
	}
	return nil
}
