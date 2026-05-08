package ratelimit_test

import (
	"testing"
	"time"

	"github.com/yourorg/cronwrap/internal/ratelimit"
)

func TestAllowFirstRun(t *testing.T) {
	l := ratelimit.New(5 * time.Minute)
	if !l.Allow("job1") {
		t.Fatal("expected first run to be allowed")
	}
}

func TestAllowBlocksWithinInterval(t *testing.T) {
	l := ratelimit.New(1 * time.Hour)
	l.Allow("job1") // record first run
	if l.Allow("job1") {
		t.Fatal("expected second run within interval to be blocked")
	}
}

func TestAllowPermitsAfterInterval(t *testing.T) {
	l := ratelimit.New(10 * time.Millisecond)
	l.Allow("job1")
	time.Sleep(20 * time.Millisecond)
	if !l.Allow("job1") {
		t.Fatal("expected run to be allowed after interval elapsed")
	}
}

func TestAllowIndependentKeys(t *testing.T) {
	l := ratelimit.New(1 * time.Hour)
	l.Allow("job1")
	if !l.Allow("job2") {
		t.Fatal("expected different key to be allowed independently")
	}
}

func TestReset(t *testing.T) {
	l := ratelimit.New(1 * time.Hour)
	l.Allow("job1")
	l.Reset("job1")
	if !l.Allow("job1") {
		t.Fatal("expected run to be allowed after reset")
	}
}

func TestNextAllowedZeroForUnknown(t *testing.T) {
	l := ratelimit.New(5 * time.Minute)
	if !l.NextAllowed("unknown").IsZero() {
		t.Fatal("expected zero time for unknown key")
	}
}

func TestNextAllowedAfterRun(t *testing.T) {
	interval := 5 * time.Minute
	l := ratelimit.New(interval)
	before := time.Now()
	l.Allow("job1")
	next := l.NextAllowed("job1")
	if next.Before(before.Add(interval)) {
		t.Fatalf("expected next allowed >= %v, got %v", before.Add(interval), next)
	}
}
