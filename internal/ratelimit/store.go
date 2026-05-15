package ratelimit

import (
	"sync"
	"time"
)

// Store holds per-key last-run timestamps used by rate-limit checks.
// It is safe for concurrent use.
type Store struct {
	mu      sync.Mutex
	entries map[string]time.Time
}

// New creates an empty, ready-to-use Store.
func New() *Store {
	return &Store{entries: make(map[string]time.Time)}
}

// Check returns whether the key is allowed to run given the interval, and the
// time at which it will next be allowed. It does NOT record the run.
func (s *Store) Check(key string, interval time.Duration) (allowed bool, nextAllowed time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	last, ok := s.entries[key]
	if !ok {
		return true, time.Time{}
	}
	next := last.Add(interval)
	if time.Now().Before(next) {
		return false, next
	}
	return true, time.Time{}
}

// Record marks key as having run at the current time.
func (s *Store) Record(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[key] = time.Now()
}

// Reset removes the stored timestamp for key, allowing an immediate next run.
func (s *Store) Reset(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, key)
}
