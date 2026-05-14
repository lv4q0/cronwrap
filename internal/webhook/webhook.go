// Package webhook provides an HTTP handler that exposes job status,
// metrics snapshots, and run history over a simple JSON API.
package webhook

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/example/cronwrap/internal/history"
	"github.com/example/cronwrap/internal/metrics"
)

// Handler serves the cronwrap status API.
type Handler struct {
	metrics *metrics.Store
	history *history.Store
	started time.Time
}

// New creates a Handler backed by the provided metrics store and history store.
func New(m *metrics.Store, h *history.Store) *Handler {
	return &Handler{metrics: m, history: h, started: time.Now()}
}

// RegisterRoutes attaches the handler routes to mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/status", h.handleStatus)
	mux.HandleFunc("/metrics", h.handleMetrics)
	mux.HandleFunc("/history", h.handleHistory)
}

type statusResponse struct {
	Status  string        `json:"status"`
	Uptime  string        `json:"uptime"`
}

func (h *Handler) handleStatus(w http.ResponseWriter, r *http.Request) {
	resp := statusResponse{
		Status: "ok",
		Uptime: time.Since(h.started).Round(time.Second).String(),
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleMetrics(w http.ResponseWriter, r *http.Request) {
	snap := h.metrics.Snapshot()
	writeJSON(w, http.StatusOK, snap)
}

func (h *Handler) handleHistory(w http.ResponseWriter, r *http.Request) {
	entries, err := h.history.ReadAll()
	if err != nil {
		http.Error(w, "failed to read history", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, entries)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
