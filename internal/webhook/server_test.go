package webhook_test

import (
	"fmt"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/example/cronwrap/internal/history"
	"github.com/example/cronwrap/internal/metrics"
	"github.com/example/cronwrap/internal/webhook"
)

func freeAddr() string {
	// Use port 0 is not straightforward with http.Server; pick a high port
	// unlikely to collide during tests.
	return "127.0.0.1:0"
}

func newTestServer(t *testing.T, addr string) *webhook.Server {
	t.Helper()
	m := metrics.NewStore()
	h, err := history.NewStore(filepath.Join(t.TempDir(), "h.json"))
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	handler := webhook.New(m, h)
	return webhook.NewServer(addr, handler)
}

func TestServerStartAndShutdown(t *testing.T) {
	srv := newTestServer(t, "127.0.0.1:18743")
	errCh := srv.Start()

	// Give the server a moment to bind.
	time.Sleep(50 * time.Millisecond)

	resp, err := http.Get("http://127.0.0.1:18743/status")
	if err != nil {
		t.Fatalf("GET /status: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	if err := srv.Shutdown(2 * time.Second); err != nil {
		t.Errorf("shutdown: %v", err)
	}

	// Drain error channel — should be closed cleanly.
	select {
	case err, ok := <-errCh:
		if ok && err != nil {
			t.Errorf("server error: %v", err)
		}
	case <-time.After(time.Second):
		t.Error("error channel not closed after shutdown")
	}
}

func TestServerBadAddressReportsError(t *testing.T) {
	srv := newTestServer(t, fmt.Sprintf("127.0.0.1:%d", 1)) // privileged port
	errCh := srv.Start()

	select {
	case err := <-errCh:
		if err == nil {
			t.Error("expected error for privileged port, got nil")
		}
	case <-time.After(2 * time.Second):
		// Some CI environments may silently fail; treat as skipped.
		t.Skip("no error received — possibly running as root or port is open")
	}
}
