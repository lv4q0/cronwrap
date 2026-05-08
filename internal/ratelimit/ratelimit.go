package ratelimit

import (
	"fmt"
	"sync"
	"time"
)

// Limiter enforces a minimum interval between successive job executions.
type Limiter struct {
	mu       sync.Mutex
	lastRun  map[string]time.Time
	interval time.Duration
}

// New creates a Limiter that enforces the given minimum interval between runs.
func New(interval time.Duration) *Limiter {
	return &Limiter{
		lastRun:  make(map[string]time.Time),
		interval: interval,
	}
}

// Allow returns true if the job identified by key is permitted to run.
// It records the current time as the last run time when allowed.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	if last, ok := l.lastRun[key]; ok {
		if now.Sub(last) < l.interval {
			return false
		}
	}
	l.lastRun[key] = now
	return true
}

// Reset clears the recorded last-run time for the given key.
func (l *Limiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.lastRun, key)
}

// NextAllowed returns the earliest time the key will be permitted to run again.
// If the key has never run, it returns the zero time.
func (l *Limiter) NextAllowed(key string) time.Time {
	l.mu.Lock()
	defer l.mu.Unlock()

	last, ok := l.lastRun[key]
	if !ok {
		return time.Time{}
	}
	return last.Add(l.interval)
}

// String returns a human-readable description of the limiter.
func (l *Limiter) String() string {
	return fmt.Sprintf("Limiter(interval=%s)", l.interval)
}
