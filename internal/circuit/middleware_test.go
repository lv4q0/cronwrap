package circuit

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/yourorg/cronwrap/internal/logger"
)

func newTestLogger() *logger.Logger {
	return logger.NewDefault()
}

func TestGuardAllowsSuccessfulRun(t *testing.T) {
	g := NewGuard("myjob", 3, 100*time.Millisecond, newTestLogger())
	err := g.Execute(context.Background(), func(_ context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap := g.Breaker().Snapshot(); snap.Failures != 0 {
		t.Errorf("expected 0 failures after success, got %d", snap.Failures)
	}
}

func TestGuardRecordsFailure(t *testing.T) {
	g := NewGuard("myjob", 3, 100*time.Millisecond, newTestLogger())
	wantErr := errors.New("boom")
	err := g.Execute(context.Background(), func(_ context.Context) error {
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected original error, got %v", err)
	}
	if snap := g.Breaker().Snapshot(); snap.Failures != 1 {
		t.Errorf("expected 1 failure, got %d", snap.Failures)
	}
}

func TestGuardBlocksWhenOpen(t *testing.T) {
	g := NewGuard("myjob", 1, 200*time.Millisecond, newTestLogger())
	// Trip the breaker
	_ = g.Execute(context.Background(), func(_ context.Context) error {
		return errors.New("fail")
	})

	err := g.Execute(context.Background(), func(_ context.Context) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error when circuit is open")
	}
	if !strings.Contains(err.Error(), "circuit breaker open") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestGuardRecoveryAfterCooldown(t *testing.T) {
	g := NewGuard("myjob", 1, 50*time.Millisecond, newTestLogger())
	_ = g.Execute(context.Background(), func(_ context.Context) error {
		return errors.New("fail")
	})

	time.Sleep(60 * time.Millisecond)

	err := g.Execute(context.Background(), func(_ context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatalf("expected recovery after cooldown, got: %v", err)
	}
	if snap := g.Breaker().Snapshot(); snap.State != StateClosed {
		t.Errorf("expected closed after probe success, got %v", snap.State)
	}
}
