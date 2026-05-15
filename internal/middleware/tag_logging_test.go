package middleware

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/cronwrap/cronwrap/internal/logger"
)

func newTagLogger(buf *bytes.Buffer) *logger.Logger {
	return logger.New(buf, logger.InfoLevel)
}

func TestTagLoggingMiddlewareLogsTagsOnSuccess(t *testing.T) {
	var buf bytes.Buffer
	log := newTagLogger(&buf)

	chain := Chain(
		TagMiddleware(Tags{"env": "test", "owner": "qa"}),
		TagLoggingMiddleware(log),
	)

	if err := Apply(chain, func(ctx context.Context) error { return nil })(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "env") {
		t.Errorf("expected tag key 'env' in log output, got: %s", out)
	}
	if !strings.Contains(out, "test") {
		t.Errorf("expected tag value 'test' in log output, got: %s", out)
	}
}

func TestTagLoggingMiddlewareLogsErrorOnFailure(t *testing.T) {
	var buf bytes.Buffer
	log := newTagLogger(&buf)
	want := errors.New("something failed")

	chain := Chain(
		TagMiddleware(Tags{"env": "prod"}),
		TagLoggingMiddleware(log),
	)

	err := Apply(chain, func(ctx context.Context) error { return want })(context.Background())
	if !errors.Is(err, want) {
		t.Fatalf("expected original error, got %v", err)
	}

	if !strings.Contains(buf.String(), "something failed") {
		t.Errorf("expected error message in log output")
	}
}

func TestTagLoggingMiddlewareNoTagsIsNoop(t *testing.T) {
	var buf bytes.Buffer
	log := newTagLogger(&buf)

	// No TagMiddleware in chain — tags will be empty.
	mw := TagLoggingMiddleware(log)
	if err := mw(func(ctx context.Context) error { return nil })(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should still log start/finish without crashing.
	if !strings.Contains(buf.String(), "job starting") {
		t.Errorf("expected start log line even with no tags")
	}
}

func TestTagsToFieldsEmpty(t *testing.T) {
	if f := tagsToFields(Tags{}); f != nil {
		t.Errorf("expected nil for empty tags, got %v", f)
	}
}
