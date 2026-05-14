package middleware_test

import (
	"errors"
	"testing"

	"github.com/cronwrap/internal/metrics"
	"github.com/cronwrap/internal/middleware"
)

const testJob = "my-job"

func newStore() *metrics.Store {
	s := &metrics.Store{}
	s.Reset()
	return s
}

func TestMetricsMiddlewareRecordsSuccess(t *testing.T) {
	store := newStore()
	mw := middleware.NewMetricsMiddleware(store, testJob)

	wrapped := mw.Wrap(func() error { return nil })
	if err := wrapped(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	snap := store.Snapshot(testJob)
	if snap.Successes != 1 {
		t.Errorf("expected 1 success, got %d", snap.Successes)
	}
	if snap.Failures != 0 {
		t.Errorf("expected 0 failures, got %d", snap.Failures)
	}
}

func TestMetricsMiddlewareRecordsFailure(t *testing.T) {
	store := newStore()
	mw := middleware.NewMetricsMiddleware(store, testJob)

	jobErr := errors.New("boom")
	wrapped := mw.Wrap(func() error { return jobErr })
	if err := wrapped(); err == nil {
		t.Fatal("expected error, got nil")
	}

	snap := store.Snapshot(testJob)
	if snap.Failures != 1 {
		t.Errorf("expected 1 failure, got %d", snap.Failures)
	}
	if snap.Successes != 0 {
		t.Errorf("expected 0 successes, got %d", snap.Successes)
	}
}

func TestMetricsMiddlewarePreservesError(t *testing.T) {
	store := newStore()
	mw := middleware.NewMetricsMiddleware(store, testJob)

	sentinel := errors.New("sentinel")
	wrapped := mw.Wrap(func() error { return sentinel })
	err := wrapped()
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestMetricsMiddlewareRecordsDuration(t *testing.T) {
	store := newStore()
	mw := middleware.NewMetricsMiddleware(store, testJob)

	wrapped := mw.Wrap(func() error { return nil })
	_ = wrapped()

	snap := store.Snapshot(testJob)
	if snap.TotalDuration <= 0 {
		t.Errorf("expected positive total duration, got %v", snap.TotalDuration)
	}
}
