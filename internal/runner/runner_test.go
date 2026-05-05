package runner_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/cronwrap/cronwrap/internal/logger"
	"github.com/cronwrap/cronwrap/internal/runner"
)

func makeOpts(buf *bytes.Buffer, attempts int, delay, timeout time.Duration) runner.Options {
	opts := runner.DefaultOptions()
	opts.MaxAttempts = attempts
	opts.RetryDelay = delay
	opts.Timeout = timeout
	opts.Logger = logger.New(buf, logger.LevelDebug, "test-job")
	return opts
}

func TestRunSuccess(t *testing.T) {
	var buf bytes.Buffer
	opts := makeOpts(&buf, 1, 0, 0)
	err := runner.Run(context.Background(), opts, "echo", "hello")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRunFailure(t *testing.T) {
	var buf bytes.Buffer
	opts := makeOpts(&buf, 1, 0, 0)
	err := runner.Run(context.Background(), opts, "false")
	if err == nil {
		t.Fatal("expected error from failing command")
	}
}

func TestRunWithRetries(t *testing.T) {
	var buf bytes.Buffer
	opts := makeOpts(&buf, 3, 10*time.Millisecond, 0)
	start := time.Now()
	err := runner.Run(context.Background(), opts, "false")
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected error after all retries exhausted")
	}
	// 2 delays of 10ms each between 3 attempts
	if elapsed < 20*time.Millisecond {
		t.Errorf("expected at least 20ms elapsed for retries, got %v", elapsed)
	}
}

func TestRunTimeout(t *testing.T) {
	var buf bytes.Buffer
	opts := makeOpts(&buf, 1, 0, 50*time.Millisecond)
	err := runner.Run(context.Background(), opts, "sleep", "5")
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestRunContextCancelled(t *testing.T) {
	var buf bytes.Buffer
	opts := makeOpts(&buf, 1, 0, 0)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := runner.Run(ctx, opts, "echo", "should not run")
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}
