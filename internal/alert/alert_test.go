package alert_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourorg/cronwrap/internal/alert"
)

func makeEvent(name string, code int) alert.Event {
	return alert.Event{
		JobName:   name,
		Message:   "job failed",
		ExitCode:  code,
		Duration:  2 * time.Second,
		Timestamp: time.Now(),
	}
}

func TestWebhookNotifierSuccess(t *testing.T) {
	var received alert.Event
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := alert.NewWebhookNotifier(ts.URL, 5*time.Second)
	ev := makeEvent("backup-job", 1)
	if err := n.Notify(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.JobName != "backup-job" {
		t.Errorf("expected job_name 'backup-job', got '%s'", received.JobName)
	}
	if received.ExitCode != 1 {
		t.Errorf("expected exit_code 1, got %d", received.ExitCode)
	}
}

func TestWebhookNotifierNon2xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	n := alert.NewWebhookNotifier(ts.URL, 5*time.Second)
	if err := n.Notify(makeEvent("job", 2)); err == nil {
		t.Fatal("expected error for non-2xx response, got nil")
	}
}

func TestWebhookNotifierBadURL(t *testing.T) {
	n := alert.NewWebhookNotifier("http://127.0.0.1:0/no-server", 1*time.Second)
	if err := n.Notify(makeEvent("job", 1)); err == nil {
		t.Fatal("expected error for unreachable URL, got nil")
	}
}

func TestNoopNotifier(t *testing.T) {
	n := &alert.NoopNotifier{}
	if err := n.Notify(makeEvent("job", 0)); err != nil {
		t.Fatalf("noop notifier should never error, got: %v", err)
	}
}
