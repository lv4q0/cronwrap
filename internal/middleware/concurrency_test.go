package middleware_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cronwrap/internal/middleware"
)

func TestConcurrencyAllowsUpToLimit(t *testing.T) {
	m := middleware.NewConcurrencyMiddleware(3, false)
	var wg sync.WaitGroup
	start := make(chan struct{})
	results := make([]error, 3)

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			wrapped := m.Wrap(func(ctx context.Context) error {
				<-start
				return nil
			})
			results[idx] = wrapped(context.Background())
		}(i)
	}

	time.Sleep(20 * time.Millisecond)
	close(start)
	wg.Wait()

	for i, err := range results {
		if err != nil {
			t.Errorf("job %d: unexpected error: %v", i, err)
		}
	}
}

func TestConcurrencyRejectsWhenLimitExceeded(t *testing.T) {
	m := middleware.NewConcurrencyMiddleware(1, false)
	hold := make(chan struct{})
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		wrapped := m.Wrap(func(ctx context.Context) error {
			<-hold
			return nil
		})
		wrapped(context.Background()) //nolint:errcheck
	}()

	time.Sleep(20 * time.Millisecond)

	wrapped := m.Wrap(func(ctx context.Context) error { return nil })
	err := wrapped(context.Background())
	if err == nil || !strings.Contains(err.Error(), "concurrency limit reached") {
		t.Errorf("expected concurrency limit error, got: %v", err)
	}

	close(hold)
	wg.Wait()
}

func TestConcurrencyBlockingWaitsForSlot(t *testing.T) {
	m := middleware.NewConcurrencyMiddleware(1, true)
	hold := make(chan struct{})
	var wg sync.WaitGroup
	var order []int
	var mu sync.Mutex

	wg.Add(1)
	go func() {
		defer wg.Done()
		wrapped := m.Wrap(func(ctx context.Context) error {
			<-hold
			mu.Lock()
			order = append(order, 1)
			mu.Unlock()
			return nil
		})
		wrapped(context.Background()) //nolint:errcheck
	}()

	time.Sleep(20 * time.Millisecond)
	wg.Add(1)
	go func() {
		defer wg.Done()
		wrapped := m.Wrap(func(ctx context.Context) error {
			mu.Lock()
			order = append(order, 2)
			mu.Unlock()
			return nil
		})
		wrapped(context.Background()) //nolint:errcheck
	}()

	time.Sleep(20 * time.Millisecond)
	close(hold)
	wg.Wait()

	if len(order) != 2 || order[0] != 1 || order[1] != 2 {
		t.Errorf("expected order [1 2], got %v", order)
	}
}

func TestConcurrencyBlockingRespectsContextCancel(t *testing.T) {
	m := middleware.NewConcurrencyMiddleware(1, true)
	hold := make(chan struct{})

	go func() {
		wrapped := m.Wrap(func(ctx context.Context) error {
			<-hold
			return nil
		})
		wrapped(context.Background()) //nolint:errcheck
	}()

	time.Sleep(20 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	wrapped := m.Wrap(func(ctx context.Context) error { return nil })
	err := wrapped(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded, got: %v", err)
	}

	close(hold)
}

func TestConcurrencyRunningCounter(t *testing.T) {
	m := middleware.NewConcurrencyMiddleware(5, false)
	if m.Running() != 0 {
		t.Fatalf("expected 0 running, got %d", m.Running())
	}
}
