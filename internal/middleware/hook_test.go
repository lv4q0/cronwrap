package middleware_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cronwrap/internal/middleware"
)

func TestHookMiddlewareCallsBeforeAndAfter(t *testing.T) {
	var beforeCalled, afterCalled bool

	before := func(_ context.Context, job string, elapsed time.Duration, err error) {
		beforeCalled = true
		if job != "myjob" {
			t.Errorf("before: unexpected job name %q", job)
		}
		if elapsed != 0 {
			t.Errorf("before: expected zero elapsed, got %v", elapsed)
		}
		if err != nil {
			t.Errorf("before: expected nil error, got %v", err)
		}
	}

	after := func(_ context.Context, job string, elapsed time.Duration, err error) {
		afterCalled = true
		if job != "myjob" {
			t.Errorf("after: unexpected job name %q", job)
		}
		if elapsed <= 0 {
			t.Errorf("after: expected positive elapsed, got %v", elapsed)
		}
		if err != nil {
			t.Errorf("after: expected nil error, got %v", err)
		}
	}

	job := func(_ context.Context) error { return nil }
	wrapped := middleware.HookMiddleware("myjob", before, after)(job)

	if err := wrapped(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !beforeCalled {
		t.Error("before hook was not called")
	}
	if !afterCalled {
		t.Error("after hook was not called")
	}
}

func TestHookMiddlewareAfterReceivesJobError(t *testing.T) {
	sentinel := errors.New("job failed")
	var gotErr error

	after := func(_ context.Context, _ string, _ time.Duration, err error) {
		gotErr = err
	}

	job := func(_ context.Context) error { return sentinel }
	wrapped := middleware.HookMiddleware("x", nil, after)(job)

	err := wrapped(context.Background())
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if !errors.Is(gotErr, sentinel) {
		t.Errorf("after hook did not receive job error, got %v", gotErr)
	}
}

func TestHookMiddlewareNilHooksAreNoop(t *testing.T) {
	job := func(_ context.Context) error { return nil }
	wrapped := middleware.HookMiddleware("safe", nil, nil)(job)

	if err := wrapped(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHookMiddlewarePreservesError(t *testing.T) {
	sentinel := errors.New("boom")
	job := func(_ context.Context) error { return sentinel }
	wrapped := middleware.HookMiddleware("j", nil, nil)(job)

	if err := wrapped(context.Background()); !errors.Is(err, sentinel) {
		t.Errorf("expected %v, got %v", sentinel, err)
	}
}
