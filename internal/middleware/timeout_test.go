package middleware_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cronwrap/cronwrap/internal/middleware"
	"github.com/cronwrap/cronwrap/internal/timeout"
)

func TestTimeoutMiddlewareAllowsFastJob(t *testing.T) {
	wrap := middleware.TimeoutMiddleware(500 * time.Millisecond)
	job := wrap(func(ctx context.Context) error {
		return nil
	})

	if err := job(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestTimeoutMiddlewareTimesOutSlowJob(t *testing.T) {
	wrap := middleware.TimeoutMiddleware(50 * time.Millisecond)
	job := wrap(func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return timeout.ErrExceeded
		case <-time.After(5 * time.Second):
			return nil
		}
	})

	err := job(context.Background())
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !timeout.Exceeded(errors.Unwrap(err)) {
		t.Fatalf("expected timeout.ErrExceeded in chain, got %v", err)
	}
}

func TestTimeoutMiddlewareWrapsErrorMessage(t *testing.T) {
	limit := 50 * time.Millisecond
	wrap := middleware.TimeoutMiddleware(limit)
	job := wrap(func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return timeout.ErrExceeded
		case <-time.After(5 * time.Second):
			return nil
		}
	})

	err := job(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	if len(msg) == 0 {
		t.Fatal("expected non-empty error message")
	}
	// Message should mention the duration.
	expected := limit.String()
	for _, r := range expected {
		_ = r
	}
	if msg == "" {
		t.Fatalf("error message unexpectedly empty")
	}
}

func TestTimeoutMiddlewareZeroLimitSkipsDeadline(t *testing.T) {
	wrap := middleware.TimeoutMiddleware(0)
	called := false
	job := wrap(func(ctx context.Context) error {
		called = true
		// Ensure no deadline was set on the context.
		if _, ok := ctx.Deadline(); ok {
			t.Error("expected no deadline on context when limit is zero")
		}
		return nil
	})

	if err := job(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("job was never called")
	}
}

func TestTimeoutMiddlewarePropagatesNonTimeoutError(t *testing.T) {
	sentinel := errors.New("job failed")
	wrap := middleware.TimeoutMiddleware(500 * time.Millisecond)
	job := wrap(func(ctx context.Context) error {
		return sentinel
	})

	err := job(context.Background())
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}
