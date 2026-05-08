package timeout_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cronwrap/internal/timeout"
)

type testLogger struct {
	warnCalled bool
	warnMsg    string
}

func (l *testLogger) Warn(msg string, _ map[string]any)  { l.warnCalled = true; l.warnMsg = msg }
func (l *testLogger) Error(msg string, _ map[string]any) {}

func TestGuardAllowsFastJob(t *testing.T) {
	log := &testLogger{}
	g := timeout.NewGuard(timeout.Policy{Limit: 500 * time.Millisecond}, log)

	err := g.Run(context.Background(), "fast-job", func(ctx context.Context) error {
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if log.warnCalled {
		t.Error("warn should not have been called for a fast job")
	}
}

func TestGuardTimesOutSlowJob(t *testing.T) {
	log := &testLogger{}
	g := timeout.NewGuard(timeout.Policy{Limit: 50 * time.Millisecond}, log)

	err := g.Run(context.Background(), "slow-job", func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
			return nil
		}
	})

	if !errors.Is(err, timeout.ErrTimeout) {
		t.Errorf("expected ErrTimeout, got %v", err)
	}
	if !log.warnCalled {
		t.Error("warn should have been called on timeout")
	}
}

func TestGuardPropagatesJobError(t *testing.T) {
	log := &testLogger{}
	g := timeout.NewGuard(timeout.Policy{Limit: 500 * time.Millisecond}, log)
	sentinel := errors.New("job failed")

	err := g.Run(context.Background(), "failing-job", func(ctx context.Context) error {
		return sentinel
	})

	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestGuardZeroLimitSkipsTimeout(t *testing.T) {
	log := &testLogger{}
	g := timeout.NewGuard(timeout.Policy{Limit: 0}, log)

	err := g.Run(context.Background(), "no-limit-job", func(ctx context.Context) error {
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
