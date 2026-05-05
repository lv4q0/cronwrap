package logger_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cronwrap/cronwrap/internal/logger"
)

func TestLoggerInfo(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(&buf, logger.LevelInfo, "test-job")
	l.Info("job started", nil)

	out := buf.String()
	if !strings.Contains(out, "level=INFO") {
		t.Errorf("expected level=INFO in output, got: %s", out)
	}
	if !strings.Contains(out, "job=test-job") {
		t.Errorf("expected job=test-job in output, got: %s", out)
	}
	if !strings.Contains(out, `msg="job started"`) {
		t.Errorf("expected msg in output, got: %s", out)
	}
}

func TestLoggerDebugSuppressedAtInfoLevel(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(&buf, logger.LevelInfo, "test-job")
	l.Debug("verbose detail", nil)

	if buf.Len() != 0 {
		t.Errorf("expected no output for DEBUG when level=INFO, got: %s", buf.String())
	}
}

func TestLoggerErrorWithFields(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(&buf, logger.LevelDebug, "my-job")
	l.Error("command failed", map[string]any{"exit_code": 1, "attempt": 3})

	out := buf.String()
	if !strings.Contains(out, "level=ERROR") {
		t.Errorf("expected level=ERROR in output, got: %s", out)
	}
	if !strings.Contains(out, "exit_code=1") {
		t.Errorf("expected exit_code field in output, got: %s", out)
	}
}

func TestLoggerWarn(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(&buf, logger.LevelWarn, "warn-job")
	l.Info("should be suppressed", nil)
	l.Warn("retrying", map[string]any{"attempt": 2})

	out := buf.String()
	if strings.Contains(out, "should be suppressed") {
		t.Errorf("INFO should be suppressed at WARN level")
	}
	if !strings.Contains(out, "level=WARN") {
		t.Errorf("expected WARN message, got: %s", out)
	}
}
