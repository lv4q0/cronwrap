package metrics

import (
	"sync"
	"time"
)

// Summary holds aggregated execution statistics for a job.
type Summary struct {
	mu sync.Mutex

	TotalRuns    int
	SuccessRuns  int
	FailureRuns  int
	TotalRetries int
	LastRunAt    time.Time
	LastDuration time.Duration
	LastSuccess  bool
}

// Record adds a single run result to the summary.
func (s *Summary) Record(success bool, retries int, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.TotalRuns++
	s.TotalRetries += retries
	s.LastRunAt = time.Now()
	s.LastDuration = duration
	s.LastSuccess = success

	if success {
		s.SuccessRuns++
	} else {
		s.FailureRuns++
	}
}

// SuccessRate returns the ratio of successful runs to total runs.
// Returns 0 if no runs have been recorded.
func (s *Summary) SuccessRate() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.TotalRuns == 0 {
		return 0
	}
	return float64(s.SuccessRuns) / float64(s.TotalRuns)
}

// Snapshot returns a copy of the current summary (safe for external use).
func (s *Summary) Snapshot() Summary {
	s.mu.Lock()
	defer s.mu.Unlock()

	return Summary{
		TotalRuns:    s.TotalRuns,
		SuccessRuns:  s.SuccessRuns,
		FailureRuns:  s.FailureRuns,
		TotalRetries: s.TotalRetries,
		LastRunAt:    s.LastRunAt,
		LastDuration: s.LastDuration,
		LastSuccess:  s.LastSuccess,
	}
}

// Reset clears all recorded statistics.
func (s *Summary) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	*s = Summary{}
}
