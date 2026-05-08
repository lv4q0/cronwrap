package backoff_test

import (
	"testing"
	"time"

	"github.com/cronwrap/cronwrap/internal/backoff"
)

func TestConstantStrategy(t *testing.T) {
	s := backoff.ConstantStrategy{Delay: 5 * time.Second}
	for _, attempt := range []int{0, 1, 5, 100} {
		if got := s.Next(attempt); got != 5*time.Second {
			t.Errorf("attempt %d: expected 5s, got %v", attempt, got)
		}
	}
}

func TestExponentialStrategyDoubles(t *testing.T) {
	s := backoff.ExponentialStrategy{Base: time.Second, Max: 60 * time.Second}

	expected := []time.Duration{
		1 * time.Second,  // 2^0
		2 * time.Second,  // 2^1
		4 * time.Second,  // 2^2
		8 * time.Second,  // 2^3
		16 * time.Second, // 2^4
	}

	for i, want := range expected {
		if got := s.Next(i); got != want {
			t.Errorf("attempt %d: expected %v, got %v", i, want, got)
		}
	}
}

func TestExponentialStrategyCapsAtMax(t *testing.T) {
	s := backoff.ExponentialStrategy{Base: time.Second, Max: 10 * time.Second}

	for _, attempt := range []int{4, 5, 10, 20} {
		if got := s.Next(attempt); got > 10*time.Second {
			t.Errorf("attempt %d: expected <= 10s, got %v", attempt, got)
		}
	}
}

func TestExponentialStrategyDefaultBase(t *testing.T) {
	// Zero Base should fall back to 1 second.
	s := backoff.ExponentialStrategy{Max: 30 * time.Second}
	if got := s.Next(0); got != time.Second {
		t.Errorf("expected 1s default base, got %v", got)
	}
}

func TestLinearStrategy(t *testing.T) {
	s := backoff.LinearStrategy{Base: 2 * time.Second, Step: 3 * time.Second}

	cases := []struct {
		attempt int
		want    time.Duration
	}{
		{0, 2 * time.Second},
		{1, 5 * time.Second},
		{2, 8 * time.Second},
		{3, 11 * time.Second},
	}

	for _, tc := range cases {
		if got := s.Next(tc.attempt); got != tc.want {
			t.Errorf("attempt %d: expected %v, got %v", tc.attempt, tc.want, got)
		}
	}
}

func TestDefaultReturnsExponential(t *testing.T) {
	s := backoff.Default()
	// First attempt should be 1 second.
	if got := s.Next(0); got != time.Second {
		t.Errorf("expected 1s for first attempt, got %v", got)
	}
	// Should cap at 30 seconds.
	if got := s.Next(100); got > 30*time.Second {
		t.Errorf("expected <= 30s cap, got %v", got)
	}
}
