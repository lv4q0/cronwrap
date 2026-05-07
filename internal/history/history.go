package history

import (
	"encoding/json"
	"os"
	"time"
)

// Entry represents a single cron job execution record.
type Entry struct {
	JobName   string        `json:"job_name"`
	StartedAt time.Time     `json:"started_at"`
	Duration  time.Duration `json:"duration_ns"`
	ExitCode  int           `json:"exit_code"`
	Success   bool          `json:"success"`
	Attempts  int           `json:"attempts"`
	Output    string        `json:"output,omitempty"`
}

// Store persists job execution history to a JSON file.
type Store struct {
	path string
}

// NewStore creates a Store backed by the given file path.
func NewStore(path string) *Store {
	return &Store{path: path}
}

// Append adds a new entry to the history file, creating it if necessary.
func (s *Store) Append(e Entry) error {
	f, err := os.OpenFile(s.path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	var entries []Entry
	dec := json.NewDecoder(f)
	if err := dec.Decode(&entries); err != nil && err.Error() != "EOF" {
		return err
	}

	entries = append(entries, e)

	if err := f.Truncate(0); err != nil {
		return err
	}
	if _, err := f.Seek(0, 0); err != nil {
		return err
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(entries)
}

// ReadAll returns all entries stored in the history file.
func (s *Store) ReadAll() ([]Entry, error) {
	f, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var entries []Entry
	if err := json.NewDecoder(f).Decode(&entries); err != nil && err.Error() != "EOF" {
		return nil, err
	}
	return entries, nil
}
