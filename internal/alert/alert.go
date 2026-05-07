package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Notifier defines the interface for sending alerts.
type Notifier interface {
	Notify(event Event) error
}

// Event represents an alert event triggered by a cron job.
type Event struct {
	JobName   string        `json:"job_name"`
	Message   string        `json:"message"`
	ExitCode  int           `json:"exit_code"`
	Duration  time.Duration `json:"duration_ms"`
	Timestamp time.Time     `json:"timestamp"`
}

// WebhookNotifier sends alerts to an HTTP webhook endpoint.
type WebhookNotifier struct {
	URL     string
	Timeout time.Duration
	client  *http.Client
}

// NewWebhookNotifier creates a WebhookNotifier with the given URL.
func NewWebhookNotifier(url string, timeout time.Duration) *WebhookNotifier {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &WebhookNotifier{
		URL:     url,
		Timeout: timeout,
		client:  &http.Client{Timeout: timeout},
	}
}

// Notify sends the event payload as JSON to the configured webhook URL.
func (w *WebhookNotifier) Notify(event Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("alert: failed to marshal event: %w", err)
	}

	resp, err := w.client.Post(w.URL, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("alert: webhook request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("alert: webhook returned non-2xx status: %d", resp.StatusCode)
	}
	return nil
}

// NoopNotifier is a no-op implementation used when alerting is disabled.
type NoopNotifier struct{}

func (n *NoopNotifier) Notify(_ Event) error { return nil }
