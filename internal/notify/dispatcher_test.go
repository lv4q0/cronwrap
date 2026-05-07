package notify

import (
	"errors"
	"strings"
	"testing"
	"time"
)

type captureSender struct {
	subjects []string
	bodies   []string
	fail     bool
}

func (c *captureSender) Send(subject, body string) error {
	if c.fail {
		return errors.New("send failed")
	}
	c.subjects = append(c.subjects, subject)
	c.bodies = append(c.bodies, body)
	return nil
}

func infoEvent() Event {
	return Event{
		JobName:    "cleanup",
		Level:      LevelInfo,
		Message:    "completed successfully",
		OccurredAt: time.Now(),
		Duration:   500 * time.Millisecond,
	}
}

func TestDispatcherSendsAboveMinLevel(t *testing.T) {
	cs := &captureSender{}
	d := NewDispatcher(TextFormatter{}, LevelWarning, cs)

	d.Dispatch(Event{JobName: "j", Level: LevelInfo, OccurredAt: time.Now()})
	if len(cs.subjects) != 0 {
		t.Errorf("expected no send for info < warning threshold")
	}

	d.Dispatch(Event{JobName: "j", Level: LevelWarning, OccurredAt: time.Now()})
	if len(cs.subjects) != 1 {
		t.Errorf("expected 1 send for warning level")
	}
}

func TestDispatcherSubjectContainsJobName(t *testing.T) {
	cs := &captureSender{}
	d := NewDispatcher(nil, LevelInfo, cs)
	d.Dispatch(infoEvent())
	if !strings.Contains(cs.subjects[0], "cleanup") {
		t.Errorf("subject missing job name: %s", cs.subjects[0])
	}
}

func TestDispatcherMultipleSenders(t *testing.T) {
	a, b := &captureSender{}, &captureSender{}
	d := NewDispatcher(TextFormatter{}, LevelInfo, a, b)
	d.Dispatch(infoEvent())
	if len(a.subjects) != 1 || len(b.subjects) != 1 {
		t.Errorf("expected both senders to receive the event")
	}
}

func TestDispatcherSenderErrorDoesNotPanic(t *testing.T) {
	cs := &captureSender{fail: true}
	d := NewDispatcher(nil, LevelInfo, cs)
	// should not panic
	d.Dispatch(infoEvent())
}

func TestNoopSender(t *testing.T) {
	var n NoopSender
	if err := n.Send("s", "b"); err != nil {
		t.Errorf("NoopSender.Send should return nil, got %v", err)
	}
}
