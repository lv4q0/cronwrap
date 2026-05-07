package notify

import (
	"log"
)

// Sender is the interface for delivering a formatted notification.
type Sender interface {
	Send(subject, body string) error
}

// Dispatcher routes Events to one or more Senders using a Formatter.
type Dispatcher struct {
	formatter Formatter
	senders   []Sender
	minLevel  Level
}

var levelRank = map[Level]int{
	LevelInfo:    0,
	LevelWarning: 1,
	LevelError:   2,
}

// NewDispatcher creates a Dispatcher with the given formatter, minimum level, and senders.
func NewDispatcher(f Formatter, minLevel Level, senders ...Sender) *Dispatcher {
	if f == nil {
		f = TextFormatter{}
	}
	return &Dispatcher{formatter: f, minLevel: minLevel, senders: senders}
}

// Dispatch formats and sends the event if its level meets the minimum threshold.
func (d *Dispatcher) Dispatch(e Event) {
	if levelRank[e.Level] < levelRank[d.minLevel] {
		return
	}
	body := d.formatter.Format(e)
	subject := "[cronwrap] " + e.JobName + " — " + string(e.Level)
	for _, s := range d.senders {
		if err := s.Send(subject, body); err != nil {
			log.Printf("notify: sender error: %v", err)
		}
	}
}

// NoopSender silently discards notifications (useful in tests / dry-run).
type NoopSender struct{}

func (n NoopSender) Send(_, _ string) error { return nil }
