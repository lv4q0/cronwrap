package middleware

import (
	"context"
	"fmt"
	"sync"
)

// ConcurrencyMiddleware limits the number of jobs that can run simultaneously.
// Jobs that exceed the limit are either rejected or queued depending on the
// block parameter. When block is false, excess jobs return an error immediately.
type ConcurrencyMiddleware struct {
	mu      sync.Mutex
	running int
	max     int
	block   bool
	wait    chan struct{}
}

// NewConcurrencyMiddleware creates a middleware that allows at most max
// concurrent executions. If block is true, callers wait for a slot; otherwise
// they receive an error immediately when the limit is reached.
func NewConcurrencyMiddleware(max int, block bool) *ConcurrencyMiddleware {
	if max <= 0 {
		max = 1
	}
	return &ConcurrencyMiddleware{
		max:   max,
		block: block,
		wait:  make(chan struct{}, max),
	}
}

// Wrap implements the Middleware interface.
func (c *ConcurrencyMiddleware) Wrap(next func(ctx context.Context) error) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		if c.block {
			// Acquire a slot, respecting context cancellation.
			select {
			case c.wait <- struct{}{}:
			case <-ctx.Done():
				return ctx.Err()
			}
		} else {
			select {
			case c.wait <- struct{}{}:
			default:
				c.mu.Lock()
				current := c.running
				c.mu.Unlock()
				return fmt.Errorf("concurrency limit reached (%d/%d active)", current, c.max)
			}
		}

		c.mu.Lock()
		c.running++
		c.mu.Unlock()

		defer func() {
			<-c.wait
			c.mu.Lock()
			c.running--
			c.mu.Unlock()
		}()

		return next(ctx)
	}
}

// Running returns the current number of active executions.
func (c *ConcurrencyMiddleware) Running() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.running
}
