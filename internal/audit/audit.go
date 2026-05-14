// Package audit provides a structured audit trail for cronwrap job executions,
// recording who triggered a job, when, and what the outcome was.
package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// TriggerKind describes how a job was initiated.
type TriggerKind string

const (
	TriggerScheduled TriggerKind = "scheduled"
	TriggerManual    TriggerKind = "manual"
	TriggerRetry     TriggerKind = "retry"
)

// Entry is a single audit log record.
type Entry struct {
	Timestamp time.Time   `json:"timestamp"`
	JobName   string      `json:"job_name"`
	Trigger   TriggerKind `json:"trigger"`
	Success   bool        `json:"success"`
	Duration  float64     `json:"duration_seconds"`
	Message   string      `json:"message,omitempty"`
}

// Log is a thread-safe, file-backed audit log.
type Log struct {
	mu   sync.Mutex
	path string
}

// NewLog returns a Log that appends entries to the given file path.
func NewLog(path string) *Log {
	return &Log{path: path}
}

// Record appends a single entry to the audit log file.
func (l *Log) Record(e Entry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("audit: open %s: %w", l.path, err)
	}
	defer f.Close()

	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now().UTC()
	}

	line, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("audit: marshal entry: %w", err)
	}
	_, err = fmt.Fprintf(f, "%s\n", line)
	return err
}

// ReadAll returns all entries stored in the audit log file.
func (l *Log) ReadAll() ([]Entry, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	data, err := os.ReadFile(l.path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("audit: read %s: %w", l.path, err)
	}

	var entries []Entry
	for _, raw := range splitLines(data) {
		if len(raw) == 0 {
			continue
		}
		var e Entry
		if err := json.Unmarshal(raw, &e); err != nil {
			return nil, fmt.Errorf("audit: unmarshal line: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func splitLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i, b := range data {
		if b == '\n' {
			lines = append(lines, data[start:i])
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}
