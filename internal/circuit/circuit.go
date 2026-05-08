// Package circuit implements a simple circuit breaker for cron job execution.
// It tracks consecutive failures and opens the circuit after a threshold,
// preventing further runs until a cooldown period has elapsed.
package circuit

import (
	"fmt"
	"sync"
	"time"
)

// State represents the circuit breaker state.
type State int

const (
	StateClosed   State = iota // normal operation
	StateOpen                  // failures exceeded threshold; runs blocked
	StateHalfOpen              // cooldown elapsed; next run is a probe
)

// Breaker is a thread-safe circuit breaker.
type Breaker struct {
	mu           sync.Mutex
	state        State
	failures      int
	threshold    int
	cooldown     time.Duration
	openedAt     time.Time
	lastFailure  time.Time
}

// New creates a Breaker that opens after threshold consecutive failures
// and attempts recovery after cooldown.
func New(threshold int, cooldown time.Duration) *Breaker {
	if threshold < 1 {
		threshold = 1
	}
	return &Breaker{
		threshold: threshold,
		cooldown:  cooldown,
	}
}

// Allow reports whether a run should be permitted and transitions
// an open circuit to half-open once the cooldown has elapsed.
func (b *Breaker) Allow() (bool, State) {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case StateClosed:
		return true, StateClosed
	case StateOpen:
		if time.Since(b.openedAt) >= b.cooldown {
			b.state = StateHalfOpen
			return true, StateHalfOpen
		}
		return false, StateOpen
	case StateHalfOpen:
		return true, StateHalfOpen
	}
	return false, b.state
}

// RecordSuccess resets the breaker to closed.
func (b *Breaker) RecordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures = 0
	b.state = StateClosed
}

// RecordFailure increments the failure counter and may open the circuit.
func (b *Breaker) RecordFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures++
	b.lastFailure = time.Now()
	if b.failures >= b.threshold && b.state != StateOpen {
		b.state = StateOpen
		b.openedAt = time.Now()
	}
}

// Snapshot returns a read-only view of the current breaker state.
func (b *Breaker) Snapshot() Snapshot {
	b.mu.Lock()
	defer b.mu.Unlock()
	return Snapshot{
		State:       b.state,
		Failures:    b.failures,
		Threshold:   b.threshold,
		OpenedAt:    b.openedAt,
		LastFailure: b.lastFailure,
	}
}

// Snapshot is an immutable view of breaker state.
type Snapshot struct {
	State       State
	Failures    int
	Threshold   int
	OpenedAt    time.Time
	LastFailure time.Time
}

// String returns a human-readable label for the state.
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return fmt.Sprintf("unknown(%d)", int(s))
	}
}
