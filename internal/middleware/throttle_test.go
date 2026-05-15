package middleware_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/cronwrap/internal/middleware"
	"github.com/cronwrap/internal/ratelimit"
)

func newThrottleStore(t *testing.T) *ratelimit.Store {
	t.Helper()
	return ratelimit.New()
}

func TestThrottleAllowsFirstRun(t *testing.T) {
	store := newThrottleStore(t)
	cfg := middleware.ThrottleConfig{JobName: "job1", Interval: time.Minute}
	mw := middleware.ThrottleMiddleware(cfg, store)

	called := false
	job := mw(func(ctx context.Context) error {
		called = true
		return nil
	})

	if err := job(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Fatal("expected job to be called")
	}
}

func TestThrottleBlocksWithinInterval(t *testing.T) {
	store := newThrottleStore(t)
	cfg := middleware.ThrottleConfig{JobName: "job2", Interval: time.Minute}
	mw := middleware.ThrottleMiddleware(cfg, store)

	job := mw(func(ctx context.Context) error { return nil })

	// First run records the timestamp.
	_ = job(context.Background())

	// Second run within interval should be blocked.
	err := job(context.Background())
	if err == nil {
		t.Fatal("expected throttle error, got nil")
	}
	if !strings.Contains(err.Error(), "throttle") {
		t.Errorf("error should mention throttle, got: %v", err)
	}
}

func TestThrottleOnSkipCallback(t *testing.T) {
	store := newThrottleStore(t)
	skipCalled := false
	cfg := middleware.ThrottleConfig{
		JobName:  "job3",
		Interval: time.Minute,
		OnSkip: func(name string, next time.Time) {
			skipCalled = true
		},
	}
	mw := middleware.ThrottleMiddleware(cfg, store)
	job := mw(func(ctx context.Context) error { return nil })

	_ = job(context.Background())
	_ = job(context.Background())

	if !skipCalled {
		t.Fatal("expected OnSkip to be called")
	}
}

func TestThrottleJobErrorPropagated(t *testing.T) {
	store := newThrottleStore(t)
	cfg := middleware.ThrottleConfig{JobName: "job4", Interval: time.Minute}
	mw := middleware.ThrottleMiddleware(cfg, store)

	want := errors.New("job failed")
	job := mw(func(ctx context.Context) error { return want })

	if err := job(context.Background()); !errors.Is(err, want) {
		t.Fatalf("expected %v, got %v", want, err)
	}
}
