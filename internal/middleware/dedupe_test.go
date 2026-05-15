package middleware_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/yourorg/cronwrap/internal/lock"
	"github.com/yourorg/cronwrap/internal/middleware"
)

func tempLockPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "test.lock")
}

func TestDedupeAllowsFirstRun(t *testing.T) {
	lk := lock.New(tempLockPath(t))
	mw := middleware.DedupeMiddleware(lk)

	ran := false
	job := mw(func(ctx context.Context) error {
		ran = true
		return nil
	})

	if err := job(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ran {
		t.Fatal("expected job to run")
	}
}

func TestDedupeBlocksSecondConcurrentRun(t *testing.T) {
	path := tempLockPath(t)
	lk1 := lock.New(path)
	lk2 := lock.New(path)

	// Acquire the lock manually to simulate a running job.
	if ok, err := lk1.TryAcquire(); err != nil || !ok {
		t.Fatalf("failed to acquire lock: %v", err)
	}
	defer lk1.Release() //nolint:errcheck

	mw := middleware.DedupeMiddleware(lk2)
	job := mw(func(ctx context.Context) error { return nil })

	err := job(context.Background())
	if !errors.Is(err, middleware.ErrJobAlreadyRunning) {
		t.Fatalf("expected ErrJobAlreadyRunning, got %v", err)
	}
}

func TestDedupeReleasesLockAfterRun(t *testing.T) {
	path := tempLockPath(t)
	lk := lock.New(path)
	mw := middleware.DedupeMiddleware(lk)

	job := mw(func(ctx context.Context) error { return nil })
	if err := job(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Lock file should be gone after the job finishes.
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("expected lock file to be removed after job completion")
	}
}

func TestDedupeReleasesLockOnJobError(t *testing.T) {
	path := tempLockPath(t)
	lk := lock.New(path)
	mw := middleware.DedupeMiddleware(lk)

	sentinel := errors.New("job failed")
	job := mw(func(ctx context.Context) error { return sentinel })

	if err := job(context.Background()); !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}

	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("expected lock file to be removed after job error")
	}
}

func TestDedupeAllowsRerunAfterCompletion(t *testing.T) {
	lk := lock.New(tempLockPath(t))
	mw := middleware.DedupeMiddleware(lk)

	var mu sync.Mutex
	runs := 0
	job := mw(func(ctx context.Context) error {
		mu.Lock()
		runs++
		mu.Unlock()
		return nil
	})

	for i := 0; i < 3; i++ {
		if err := job(context.Background()); err != nil {
			t.Fatalf("run %d: unexpected error: %v", i, err)
		}
	}
	if runs != 3 {
		t.Fatalf("expected 3 runs, got %d", runs)
	}
}
