package webhook_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/example/cronwrap/internal/history"
	"github.com/example/cronwrap/internal/metrics"
	"github.com/example/cronwrap/internal/webhook"
)

func newTestHandler(t *testing.T) *webhook.Handler {
	t.Helper()
	m := metrics.NewStore()
	h, err := history.NewStore(filepath.Join(t.TempDir(), "history.json"))
	if err != nil {
		t.Fatalf("history.NewStore: %v", err)
	}
	return webhook.New(m, h)
}

func TestStatusEndpointReturns200(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	mux := http.NewServeMux()
	newTestHandler(t).RegisterRoutes(mux)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", body["status"])
	}
}

func TestMetricsEndpointReturnsSnapshot(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	mux := http.NewServeMux()
	newTestHandler(t).RegisterRoutes(mux)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("unexpected Content-Type: %s", ct)
	}
}

func TestHistoryEndpointEmptyStore(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/history", nil)
	mux := http.NewServeMux()
	newTestHandler(t).RegisterRoutes(mux)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var entries []any
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty history, got %d entries", len(entries))
	}
}

func TestHistoryEndpointUnreadablePath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")
	if err := os.WriteFile(path, []byte("not-json"), 0o000); err != nil {
		t.Skip("cannot create unreadable file")
	}
	h, _ := history.NewStore(path)
	m := metrics.NewStore()
	handler := webhook.New(m, h)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/history", nil)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	mux.ServeHTTP(rec, req)
	// Either 200 (empty) or 500 depending on OS permissions; just ensure no panic.
	if rec.Code != http.StatusOK && rec.Code != http.StatusInternalServerError {
		t.Errorf("unexpected status %d", rec.Code)
	}
}
