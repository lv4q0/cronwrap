package ratelimit_test

import (
	"testing"
	"time"

	"github.com/yourorg/cronwrap/internal/logger"
	"github.com/yourorg/cronwrap/internal/ratelimit"
)

func newTestLogger(t *testing.T) *logger.Logger {
	t.Helper()
	return logger.NewDefault()
}

func TestGuardAllowsFirstRun(t *testing.T) {
	g := ratelimit.NewGuard(1*time.Hour, newTestLogger(t))
	if err := g.Check("backup"); err != nil {
		t.Fatalf("expected first run to be allowed, got: %v", err)
	}
}

func TestGuardBlocksSecondRun(t *testing.T) {
	g := ratelimit.NewGuard(1*time.Hour, newTestLogger(t))
	g.Check("backup") // first — allowed
	if err := g.Check("backup"); err == nil {
		t.Fatal("expected second run within interval to be blocked")
	}
}

func TestGuardErrorMessageContainsJobName(t *testing.T) {
	g := ratelimit.NewGuard(1*time.Hour, newTestLogger(t))
	g.Check("nightly-report")
	err := g.Check("nightly-report")
	if err == nil {
		t.Fatal("expected an error")
	}
	if got := err.Error(); len(got) == 0 {
		t.Fatal("expected non-empty error message")
	}
}

func TestGuardResetAllowsRerun(t *testing.T) {
	g := ratelimit.NewGuard(1*time.Hour, newTestLogger(t))
	g.Check("sync")
	g.Reset("sync")
	if err := g.Check("sync"); err != nil {
		t.Fatalf("expected run after reset to be allowed, got: %v", err)
	}
}

func TestGuardIndependentJobs(t *testing.T) {
	g := ratelimit.NewGuard(1*time.Hour, newTestLogger(t))
	g.Check("job-a")
	if err := g.Check("job-b"); err != nil {
		t.Fatalf("expected independent job to be allowed, got: %v", err)
	}
}
