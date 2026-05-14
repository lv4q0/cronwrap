// Package webhook exposes a lightweight HTTP status server for cronwrap.
//
// Three endpoints are available:
//
//	 GET /status  – liveness probe; returns uptime and a static "ok" status.
//	 GET /metrics – JSON snapshot of the in-memory metrics store (success
//	               count, failure count, success rate, last durations).
//	 GET /history – JSON array of all persisted run-history entries read
//	               from the history store.
//
// Usage:
//
//	h := webhook.New(metricsStore, historyStore)
//	srv := webhook.NewServer(":9090", h)
//	errCh := srv.Start()
//	// … later …
//	_ = srv.Shutdown(5 * time.Second)
package webhook
