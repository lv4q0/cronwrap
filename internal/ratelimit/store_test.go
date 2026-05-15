package ratelimit_test

import (
	"testing"
	"time"

	"github.com/cronwrap/internal/ratelimit"
)

func TestCheckAllowsUnseenKey(t *testing.T) {
	s := ratelimit.New()
	allowed, next := s.Check("job", time.Minute)
	if !allowed {
		t.Fatal("expected allowed for unseen key")
	}
	if !next.IsZero() {
		t.Errorf("expected zero nextAllowed, got %v", next)
	}
}

func TestCheckBlocksAfterRecord(t *testing.T) {
	s := ratelimit.New()
	s.Record("job")
	allowed, next := s.Check("job", time.Minute)
	if allowed {
		t.Fatal("expected blocked after record within interval")
	}
	if next.IsZero() {
		t.Error("expected a non-zero nextAllowed time")
	}
	if time.Until(next) > time.Minute {
		t.Errorf("nextAllowed too far in the future: %v", next)
	}
}

func TestCheckAllowsAfterInterval(t *testing.T) {
	s := ratelimit.New()
	// Simulate a very short interval that has already elapsed.
	s.Record("job")
	time.Sleep(5 * time.Millisecond)
	allowed, _ := s.Check("job", time.Millisecond)
	if !allowed {
		t.Fatal("expected allowed after interval elapsed")
	}
}

func TestResetAllowsImmediateRun(t *testing.T) {
	s := ratelimit.New()
	s.Record("job")
	s.Reset("job")
	allowed, _ := s.Check("job", time.Minute)
	if !allowed {
		t.Fatal("expected allowed after reset")
	}
}

func TestCheckIndependentKeys(t *testing.T) {
	s := ratelimit.New()
	s.Record("jobA")
	allowedA, _ := s.Check("jobA", time.Minute)
	allowedB, _ := s.Check("jobB", time.Minute)
	if allowedA {
		t.Fatal("jobA should be blocked")
	}
	if !allowedB {
		t.Fatal("jobB should be allowed (independent key)")
	}
}
