package metrics

import (
	"testing"
	"time"
)

func TestRecordSuccess(t *testing.T) {
	var s Summary
	s.Record(true, 0, 100*time.Millisecond)

	if s.TotalRuns != 1 {
		t.Fatalf("expected TotalRuns=1, got %d", s.TotalRuns)
	}
	if s.SuccessRuns != 1 {
		t.Fatalf("expected SuccessRuns=1, got %d", s.SuccessRuns)
	}
	if s.FailureRuns != 0 {
		t.Fatalf("expected FailureRuns=0, got %d", s.FailureRuns)
	}
	if !s.LastSuccess {
		t.Fatal("expected LastSuccess=true")
	}
}

func TestRecordFailure(t *testing.T) {
	var s Summary
	s.Record(false, 2, 200*time.Millisecond)

	if s.FailureRuns != 1 {
		t.Fatalf("expected FailureRuns=1, got %d", s.FailureRuns)
	}
	if s.TotalRetries != 2 {
		t.Fatalf("expected TotalRetries=2, got %d", s.TotalRetries)
	}
	if s.LastSuccess {
		t.Fatal("expected LastSuccess=false")
	}
}

func TestSuccessRate(t *testing.T) {
	var s Summary

	if rate := s.SuccessRate(); rate != 0 {
		t.Fatalf("expected 0 for empty summary, got %f", rate)
	}

	s.Record(true, 0, time.Millisecond)
	s.Record(true, 0, time.Millisecond)
	s.Record(false, 1, time.Millisecond)

	expected := 2.0 / 3.0
	if rate := s.SuccessRate(); rate != expected {
		t.Fatalf("expected rate=%f, got %f", expected, rate)
	}
}

func TestSnapshot(t *testing.T) {
	var s Summary
	s.Record(true, 1, 50*time.Millisecond)

	snap := s.Snapshot()
	s.Record(false, 0, 10*time.Millisecond)

	// Snapshot should not reflect the second record.
	if snap.TotalRuns != 1 {
		t.Fatalf("snapshot TotalRuns should be 1, got %d", snap.TotalRuns)
	}
}

func TestReset(t *testing.T) {
	var s Summary
	s.Record(true, 0, time.Millisecond)
	s.Reset()

	if s.TotalRuns != 0 {
		t.Fatalf("expected TotalRuns=0 after reset, got %d", s.TotalRuns)
	}
	if s.SuccessRate() != 0 {
		t.Fatal("expected SuccessRate=0 after reset")
	}
}
