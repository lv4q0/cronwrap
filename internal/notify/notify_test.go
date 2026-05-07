package notify

import (
	"strings"
	"testing"
	"time"
)

func makeEvent() Event {
	return Event{
		JobName:    "backup",
		Level:      LevelError,
		Message:    "command exited with code 1",
		OccurredAt: time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
		Duration:   2*time.Second + 345*time.Millisecond,
		ExitCode:   1,
		Stdout:     "backing up...",
		Stderr:     "disk full",
	}
}

func TestTextFormatterContainsFields(t *testing.T) {
	f := TextFormatter{}
	out := f.Format(makeEvent())

	for _, want := range []string{"[ERROR]", "backup", "command exited", "2.345s", "disk full", "backing up"} {
		if !strings.Contains(out, want) {
			t.Errorf("TextFormatter output missing %q\ngot:\n%s", want, out)
		}
	}
}

func TestTextFormatterNoStdoutStderrWhenEmpty(t *testing.T) {
	e := makeEvent()
	e.Stdout = ""
	e.Stderr = ""
	out := TextFormatter{}.Format(e)
	if strings.Contains(out, "stdout") || strings.Contains(out, "stderr") {
		t.Errorf("expected no stdout/stderr lines when empty, got:\n%s", out)
	}
}

func TestJSONFormatterContainsFields(t *testing.T) {
	f := JSONFormatter{}
	out := f.Format(makeEvent())

	for _, want := range []string{`"job":"backup"`, `"level":"error"`, `"exit_code":1`, `"duration_ms":2345`} {
		if !strings.Contains(out, want) {
			t.Errorf("JSONFormatter output missing %q\ngot: %s", want, out)
		}
	}
}

func TestJSONFormatterInfoLevel(t *testing.T) {
	e := makeEvent()
	e.Level = LevelInfo
	out := JSONFormatter{}.Format(e)
	if !strings.Contains(out, `"level":"info"`) {
		t.Errorf("expected level info, got: %s", out)
	}
}
