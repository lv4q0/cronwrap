package timeout_test

import (
	"context"
	"testing"
	"time"

	"github.com/cronwrap/internal/timeout"
)

func TestDefaultPolicy(t *testing.T) {
	p := timeout.Default()
	if p.Limit != 30*time.Minute {
		t.Errorf("expected 30m limit, got %v", p.Limit)
	}
	if p.GracePeriod != 5*time.Second {
		t.Errorf("expected 5s grace, got %v", p.GracePeriod)
	}
}

func TestWithContextCancelsAfterLimit(t *testing.T) {
	p := timeout.Policy{Limit: 50 * time.Millisecond}
	ctx, cancel := p.WithContext(context.Background())
	defer cancel()

	select {
	case <-ctx.Done():
		// expected
	case <-time.After(500 * time.Millisecond):
		t.Fatal("context was not cancelled within expected window")
	}
}

func TestWithContextZeroLimitUsesCancel(t *testing.T) {
	p := timeout.Policy{Limit: 0}
	ctx, cancel := p.WithContext(context.Background())

	// Should not be done yet
	select {
	case <-ctx.Done():
		t.Fatal("context should not be done before cancel is called")
	default:
	}
	cancel()

	select {
	case <-ctx.Done():
		// expected
	case <-time.After(100 * time.Millisecond):
		t.Fatal("context should be done after cancel")
	}
}

func TestExceeded(t *testing.T) {
	if timeout.Exceeded(nil) {
		t.Error("nil error should not be exceeded")
	}
	if !timeout.Exceeded(context.DeadlineExceeded) {
		t.Error("DeadlineExceeded should be detected")
	}
	if !timeout.Exceeded(timeout.ErrTimeout) {
		t.Error("ErrTimeout should be detected")
	}
}

func TestWrap(t *testing.T) {
	if timeout.Wrap(context.DeadlineExceeded) != timeout.ErrTimeout {
		t.Error("DeadlineExceeded should be wrapped to ErrTimeout")
	}
	if timeout.Wrap(nil) != nil {
		t.Error("nil should remain nil")
	}
	other := context.Canceled
	if timeout.Wrap(other) != other {
		t.Error("non-deadline errors should pass through unchanged")
	}
}
