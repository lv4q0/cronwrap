package middleware_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cronwrap/internal/backoff"
	"github.com/cronwrap/internal/middleware"
)

var errBoom = errors.New("boom")

func TestRetrySucceedsOnFirstAttempt(t *testing.T) {
	cfg := middleware.RetryConfig{MaxAttempts: 3, Strategy: backoff.Default}
	wrap := middleware.RetryMiddleware(cfg)

	calls := 0
	job := wrap(func(_ context.Context) error {
		calls++
		return nil
	})

	if err := job(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestRetryRetriesOnFailure(t *testing.T) {
	cfg := middleware.RetryConfig{
		MaxAttempts: 3,
		Strategy:    backoff.Constant(0),
	}
	wrap := middleware.RetryMiddleware(cfg)

	var calls int32
	job := wrap(func(_ context.Context) error {
		atomic.AddInt32(&calls, 1)
		return errBoom
	})

	err := job(context.Background())
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if !errors.Is(err, errBoom) {
		t.Fatalf("expected wrapped errBoom, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestRetrySucceedsOnSecondAttempt(t *testing.T) {
	cfg := middleware.RetryConfig{MaxAttempts: 3, Strategy: backoff.Constant(0)}
	wrap := middleware.RetryMiddleware(cfg)

	attempts := 0
	job := wrap(func(_ context.Context) error {
		attempts++
		if attempts < 2 {
			return errBoom
		}
		return nil
	})

	if err := job(context.Background()); err != nil {
		t.Fatalf("expected success on second attempt, got %v", err)
	}
	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
}

func TestRetryAbortsOnContextCancel(t *testing.T) {
	cfg := middleware.RetryConfig{
		MaxAttempts: 5,
		Strategy:    backoff.Constant(200 * time.Millisecond),
	}
	wrap := middleware.RetryMiddleware(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	calls := 0
	job := wrap(func(_ context.Context) error {
		calls++
		return errBoom
	})

	err := job(ctx)
	if err == nil {
		t.Fatal("expected error due to context cancellation")
	}
	if calls > 2 {
		t.Fatalf("expected at most 2 calls before cancel, got %d", calls)
	}
}

func TestRetryDefaultsOnZeroMaxAttempts(t *testing.T) {
	cfg := middleware.RetryConfig{MaxAttempts: 0, Strategy: backoff.Constant(0)}
	wrap := middleware.RetryMiddleware(cfg)

	calls := 0
	job := wrap(func(_ context.Context) error {
		calls++
		return errBoom
	})

	_ = job(context.Background())
	if calls != 1 {
		t.Fatalf("expected exactly 1 call for zero MaxAttempts, got %d", calls)
	}
}
