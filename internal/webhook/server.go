package webhook

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const (
	defaultReadTimeout  = 5 * time.Second
	defaultWriteTimeout = 10 * time.Second
	defaultIdleTimeout  = 60 * time.Second
)

// Server wraps an http.Server configured for the status API.
type Server struct {
	httpServer *http.Server
}

// NewServer creates a Server that listens on the given address and serves
// routes registered on handler.
func NewServer(addr string, h *Handler) *Server {
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	return &Server{
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      mux,
			ReadTimeout:  defaultReadTimeout,
			WriteTimeout: defaultWriteTimeout,
			IdleTimeout:  defaultIdleTimeout,
		},
	}
}

// Start begins listening in a background goroutine and returns immediately.
// Errors from ListenAndServe (other than http.ErrServerClosed) are sent on
// the returned channel.
func (s *Server) Start() <-chan error {
	errCh := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("webhook server: %w", err)
		}
		close(errCh)
	}()
	return errCh
}

// Shutdown gracefully stops the server within the given timeout.
func (s *Server) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}
