package circuit

import (
	"testing"
	"time"
)

func newBreaker(threshold int) *Breaker {
	return New(threshold, 50*time.Millisecond)
}

func TestInitialStateIsClosed(t *testing.T) {
	b := newBreaker(3)
	ok, state := b.Allow()
	if !ok || state != StateClosed {
		t.Fatalf("expected closed/allowed, got %v/%v", state, ok)
	}
}

func TestOpensAfterThreshold(t *testing.T) {
	b := newBreaker(2)
	b.RecordFailure()
	b.RecordFailure()

	ok, state := b.Allow()
	if ok || state != StateOpen {
		t.Fatalf("expected open/blocked, got %v/%v", state, ok)
	}
}

func TestDoesNotOpenBeforeThreshold(t *testing.T) {
	b := newBreaker(3)
	b.RecordFailure()
	b.RecordFailure()

	ok, state := b.Allow()
	if !ok || state != StateClosed {
		t.Fatalf("expected still closed, got %v/%v", state, ok)
	}
}

func TestTransitionsToHalfOpenAfterCooldown(t *testing.T) {
	b := newBreaker(1)
	b.RecordFailure()

	// Still open immediately
	ok, _ := b.Allow()
	if ok {
		t.Fatal("expected blocked while open")
	}

	time.Sleep(60 * time.Millisecond)

	ok, state := b.Allow()
	if !ok || state != StateHalfOpen {
		t.Fatalf("expected half-open after cooldown, got %v/%v", state, ok)
	}
}

func TestRecordSuccessCloses(t *testing.T) {
	b := newBreaker(1)
	b.RecordFailure()
	time.Sleep(60 * time.Millisecond)
	b.Allow() // transition to half-open
	b.RecordSuccess()

	ok, state := b.Allow()
	if !ok || state != StateClosed {
		t.Fatalf("expected closed after success, got %v/%v", state, ok)
	}
}

func TestSnapshotValues(t *testing.T) {
	b := newBreaker(2)
	b.RecordFailure()

	snap := b.Snapshot()
	if snap.Failures != 1 {
		t.Errorf("expected 1 failure, got %d", snap.Failures)
	}
	if snap.Threshold != 2 {
		t.Errorf("expected threshold 2, got %d", snap.Threshold)
	}
	if snap.State != StateClosed {
		t.Errorf("expected closed state, got %v", snap.State)
	}
}

func TestStateString(t *testing.T) {
	cases := map[State]string{
		StateClosed:   "closed",
		StateOpen:     "open",
		StateHalfOpen: "half-open",
	}
	for s, want := range cases {
		if got := s.String(); got != want {
			t.Errorf("State(%d).String() = %q, want %q", int(s), got, want)
		}
	}
}
