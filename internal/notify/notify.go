package notify

import (
	"fmt"
	"strings"
	"time"
)

// Level represents the severity of a notification.
type Level string

const (
	LevelInfo    Level = "info"
	LevelWarning Level = "warning"
	LevelError   Level = "error"
)

// Event holds the data for a notification event.
type Event struct {
	JobName   string
	Level     Level
	Message   string
	OccurredAt time.Time
	Duration  time.Duration
	ExitCode  int
	Stdout    string
	Stderr    string
}

// Formatter converts an Event into a human-readable string.
type Formatter interface {
	Format(e Event) string
}

// TextFormatter renders events as plain text.
type TextFormatter struct{}

func (f TextFormatter) Format(e Event) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "[%s] %s — %s\n", strings.ToUpper(string(e.Level)), e.JobName, e.Message)
	fmt.Fprintf(&sb, "  occurred : %s\n", e.OccurredAt.Format(time.RFC3339))
	fmt.Fprintf(&sb, "  duration : %s\n", e.Duration.Round(time.Millisecond))
	fmt.Fprintf(&sb, "  exit_code: %d\n", e.ExitCode)
	if e.Stdout != "" {
		fmt.Fprintf(&sb, "  stdout   : %s\n", e.Stdout)
	}
	if e.Stderr != "" {
		fmt.Fprintf(&sb, "  stderr   : %s\n", e.Stderr)
	}
	return sb.String()
}

// JSONFormatter renders events as a JSON-like string (no external deps).
type JSONFormatter struct{}

func (f JSONFormatter) Format(e Event) string {
	return fmt.Sprintf(
		`{"job":%q,"level":%q,"message":%q,"occurred_at":%q,"duration_ms":%d,"exit_code":%d}`,
		e.JobName,
		string(e.Level),
		e.Message,
		e.OccurredAt.Format(time.RFC3339),
		e.Duration.Milliseconds(),
		e.ExitCode,
	)
}
