package middleware_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cronwrap/cronwrap/internal/circuit"
	"github.com/cronwrap/cronwrap/internal/logger"
	"github.com/cronwrap/cronwrap/internal/middleware"
)

func newCircuitLogger() *logger.Logger {
	return logger.NewDefault()
}

func newTestBreaker(threshold int, cooldown time.Duration) *circuit.Breaker {
	return circuit.New(circuit.Options{
		FailureThreshold: threshold,
		Cooldown:         cooldown,
	})
}

func TestCircuitMiddlewareAllowsHealthyJob(t *testing.T) {
	breaker := newTestBreaker(3, time.Second)
	log := newCircuitLogger()
	mw := middleware.CircuitMiddleware(breaker, log)

	job := func(ctx context.Context) error { return nil }
	wrapped := mw(job)

	if err := wrapped(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCircuitMiddlewareBlocksWhenOpen(t *testing.T) {
	breaker := newTestBreaker(2, 10*time.Second)
	log := newCircuitLogger()
	mw := middleware.CircuitMiddleware(breaker, log)

	failJob := func(ctx context.Context) error { return errors.New("boom") }
	wrapped := mw(failJob)

	// Trip the breaker.
	_ = wrapped(context.Background())
	_ = wrapped(context.Background())

	// Now it should be open.
	err := wrapped(context.Background())
	if err == nil {
		t.Fatal("expected circuit-open error, got nil")
	}
}

func TestCircuitMiddlewareWrapsError(t *testing.T) {
	breaker := newTestBreaker(1, 10*time.Second)
	log := newCircuitLogger()
	mw := middleware.CircuitMiddleware(breaker, log)

	failJob := func(ctx context.Context) error { return errors.New("fail") }
	wrapped := mw(failJob)
	_ = wrapped(context.Background())

	err := wrapped(context.Background())
	if err == nil {
		t.Fatal("expected wrapped error")
	}
	if err.Error()[:8] != "circuit:" {
		t.Errorf("expected error to start with 'circuit:', got %q", err.Error())
	}
}

func TestCircuitMiddlewareRecoveryAfterCooldown(t *testing.T) {
	breaker := newTestBreaker(1, 50*time.Millisecond)
	log := newCircuitLogger()
	mw := middleware.CircuitMiddleware(breaker, log)

	failJob := func(ctx context.Context) error { return errors.New("fail") }
	successJob := func(ctx context.Context) error { return nil }
	wrappedFail := mw(failJob)
	wrappedOk := mw(successJob)

	_ = wrappedFail(context.Background())
	time.Sleep(80 * time.Millisecond)

	if err := wrappedOk(context.Background()); err != nil {
		t.Fatalf("expected recovery after cooldown, got %v", err)
	}
}
