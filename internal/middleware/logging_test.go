package middleware_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/cronwrap/internal/logger"
	"github.com/cronwrap/internal/middleware"
)

func newLoggerWithBuffer() (*logger.Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	log := logger.New(buf, logger.INFO)
	return log, buf
}

func TestLogMiddlewareLogsJobStart(t *testing.T) {
	log, buf := newLoggerWithBuffer()
	mw := middleware.LogMiddleware(log, "my-job")

	wrapped := mw(func() error { return nil })
	_ = wrapped()

	if !strings.Contains(buf.String(), "job started") {
		t.Errorf("expected 'job started' in log output, got: %s", buf.String())
	}
}

func TestLogMiddlewareLogsJobFinished(t *testing.T) {
	log, buf := newLoggerWithBuffer()
	mw := middleware.LogMiddleware(log, "my-job")

	wrapped := mw(func() error { return nil })
	_ = wrapped()

	out := buf.String()
	if !strings.Contains(out, "job finished") {
		t.Errorf("expected 'job finished' in log output, got: %s", out)
	}
	if !strings.Contains(out, "my-job") {
		t.Errorf("expected job name in log output, got: %s", out)
	}
}

func TestLogMiddlewareLogsErrorOnFailure(t *testing.T) {
	log, buf := newLoggerWithBuffer()
	mw := middleware.LogMiddleware(log, "failing-job")

	expectedErr := errors.New("something went wrong")
	wrapped := mw(func() error { return expectedErr })
	err := wrapped()

	if err != expectedErr {
		t.Fatalf("expected error to propagate, got: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "job failed") {
		t.Errorf("expected 'job failed' in log output, got: %s", out)
	}
	if !strings.Contains(out, "something went wrong") {
		t.Errorf("expected error message in log output, got: %s", out)
	}
}

func TestLogMiddlewareIncludesDuration(t *testing.T) {
	log, buf := newLoggerWithBuffer()
	mw := middleware.LogMiddleware(log, "timed-job")

	wrapped := mw(func() error { return nil })
	_ = wrapped()

	if !strings.Contains(buf.String(), "duration") {
		t.Errorf("expected 'duration' in log output, got: %s", buf.String())
	}
}
