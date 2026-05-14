package audit_test

import (
	"context"
	"errors"
	"testing"

	"github.com/cronwrap/cronwrap/internal/audit"
)

func newMiddleware(t *testing.T, trigger audit.TriggerKind) (*audit.Log, *audit.Middleware) {
	t.Helper()
	log := audit.NewLog(tempPath(t))
	mw := audit.NewMiddleware(log, "test-job", trigger)
	return log, mw
}

func TestAuditMiddlewareRecordsSuccess(t *testing.T) {
	log, mw := newMiddleware(t, audit.TriggerScheduled)

	job := mw.Wrap(func(ctx context.Context) error { return nil })
	if err := job(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries, err := log.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if !entries[0].Success {
		t.Error("expected success=true")
	}
	if entries[0].Message != "" {
		t.Errorf("expected empty message, got %q", entries[0].Message)
	}
}

func TestAuditMiddlewareRecordsFailure(t *testing.T) {
	log, mw := newMiddleware(t, audit.TriggerManual)
	wantErr := errors.New("boom")

	job := mw.Wrap(func(ctx context.Context) error { return wantErr })
	err := job(context.Background())
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected original error, got %v", err)
	}

	entries, _ := log.ReadAll()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Success {
		t.Error("expected success=false")
	}
	if entries[0].Message != "boom" {
		t.Errorf("expected message 'boom', got %q", entries[0].Message)
	}
}

func TestAuditMiddlewarePreservesOriginalError(t *testing.T) {
	_, mw := newMiddleware(t, audit.TriggerRetry)
	sentinel := errors.New("sentinel")

	job := mw.Wrap(func(ctx context.Context) error { return sentinel })
	if got := job(context.Background()); !errors.Is(got, sentinel) {
		t.Errorf("expected sentinel error, got %v", got)
	}
}

func TestAuditMiddlewareRecordsTriggerKind(t *testing.T) {
	log, mw := newMiddleware(t, audit.TriggerRetry)
	job := mw.Wrap(func(ctx context.Context) error { return nil })
	_ = job(context.Background())

	entries, _ := log.ReadAll()
	if entries[0].Trigger != audit.TriggerRetry {
		t.Errorf("expected trigger=retry, got %s", entries[0].Trigger)
	}
}
