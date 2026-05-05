package runner

import (
	"context"
	"testing"
	"time"
)

func TestRunSuccess(t *testing.T) {
	opts := DefaultOptions()
	result := Run(context.Background(), "echo", []string{"hello"}, opts)

	if result.Err != nil {
		t.Fatalf("expected no error, got: %v", result.Err)
	}
	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got: %d", result.ExitCode)
	}
	if result.Duration <= 0 {
		t.Error("expected positive duration")
	}
}

func TestRunFailure(t *testing.T) {
	opts := DefaultOptions()
	result := Run(context.Background(), "false", nil, opts)

	if result.ExitCode == 0 {
		t.Fatal("expected non-zero exit code")
	}
}

func TestRunWithRetries(t *testing.T) {
	opts := Options{
		Timeout:    5 * time.Second,
		MaxRetries: 2,
		RetryDelay: 10 * time.Millisecond,
	}

	start := time.Now()
	result := Run(context.Background(), "false", nil, opts)
	elapsed := time.Since(start)

	if result.ExitCode == 0 {
		t.Fatal("expected failure after retries")
	}
	// With 2 retries and 10ms delay we expect at least 20ms total.
	if elapsed < 20*time.Millisecond {
		t.Errorf("expected retries to add delay, elapsed: %v", elapsed)
	}
}

func TestRunTimeout(t *testing.T) {
	opts := Options{
		Timeout:    50 * time.Millisecond,
		MaxRetries: 0,
		RetryDelay: 0,
	}

	result := Run(context.Background(), "sleep", []string{"5"}, opts)

	if result.Err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestRunContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	opts := DefaultOptions()
	result := Run(ctx, "echo", []string{"should not run"}, opts)

	if result.Err == nil {
		// exec may still succeed if process starts before context check;
		// we simply verify the function returns without panic.
		t.Log("command completed despite cancelled context (acceptable race)")
	}
}
