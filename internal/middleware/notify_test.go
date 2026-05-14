package middleware_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cronwrap/internal/middleware"
	"github.com/cronwrap/internal/notify"
)

type captureDispatcher struct {
	events []notify.Event
}

func (c *captureDispatcher) Dispatch(_ context.Context, e notify.Event) error {
	c.events = append(c.events, e)
	return nil
}

func newNotifyDispatcher(cap *captureDispatcher) *notify.Dispatcher {
	return notify.NewDispatcher(cap, notify.LevelInfo)
}

func TestNotifyMiddlewareSendsEventOnSuccess(t *testing.T) {
	cap := &captureDispatcher{}
	d := newNotifyDispatcher(cap)
	m := middleware.NewNotifyMiddleware(d, "backup")

	wrapped := m.Wrap(func(ctx context.Context) error { return nil })
	if err := wrapped(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cap.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(cap.events))
	}
	if cap.events[0].Level != notify.LevelInfo {
		t.Errorf("expected info level, got %v", cap.events[0].Level)
	}
	if cap.events[0].JobName != "backup" {
		t.Errorf("expected job name 'backup', got %q", cap.events[0].JobName)
	}
}

func TestNotifyMiddlewareSendsEventOnFailure(t *testing.T) {
	cap := &captureDispatcher{}
	d := newNotifyDispatcher(cap)
	m := middleware.NewNotifyMiddleware(d, "cleanup")

	expectedErr := errors.New("disk full")
	wrapped := m.Wrap(func(ctx context.Context) error { return expectedErr })
	err := wrapped(context.Background())

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected original error to propagate, got %v", err)
	}
	if len(cap.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(cap.events))
	}
	if cap.events[0].Level != notify.LevelError {
		t.Errorf("expected error level, got %v", cap.events[0].Level)
	}
}

func TestNotifyMiddlewareRecordsDuration(t *testing.T) {
	cap := &captureDispatcher{}
	d := newNotifyDispatcher(cap)
	m := middleware.NewNotifyMiddleware(d, "report")

	wrapped := m.Wrap(func(ctx context.Context) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})
	_ = wrapped(context.Background())

	if len(cap.events) == 0 {
		t.Fatal("no events dispatched")
	}
	if cap.events[0].Duration < 10*time.Millisecond {
		t.Errorf("expected duration >= 10ms, got %v", cap.events[0].Duration)
	}
}
