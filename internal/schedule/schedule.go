package schedule

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

// Entry holds a parsed cron schedule and its next/previous run times.
type Entry struct {
	Expression string
	Schedule   cron.Schedule
}

// Parse validates and parses a cron expression string.
func Parse(expr string) (*Entry, error) {
	p := cron.NewParser(
		cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)
	sched, err := p.Parse(expr)
	if err != nil {
		return nil, fmt.Errorf("invalid cron expression %q: %w", expr, err)
	}
	return &Entry{Expression: expr, Schedule: sched}, nil
}

// Next returns the next activation time after t.
func (e *Entry) Next(t time.Time) time.Time {
	return e.Schedule.Next(t)
}

// IsDue reports whether the schedule is due within the given tolerance window
// centered on now. A typical tolerance is one minute for minute-granularity crons.
func (e *Entry) IsDue(now time.Time, tolerance time.Duration) bool {
	next := e.Schedule.Next(now.Add(-tolerance))
	return !next.After(now.Add(tolerance))
}

// UntilNext returns the duration from now until the next scheduled run.
func (e *Entry) UntilNext(now time.Time) time.Duration {
	return e.Schedule.Next(now).Sub(now)
}
