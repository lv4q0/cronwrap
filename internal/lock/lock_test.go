package lock

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func tempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "cronwrap-lock-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func TestAcquireCreatesLockFile(t *testing.T) {
	dir := tempDir(t)
	l := New(dir, "my-job")

	if err := l.Acquire(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer l.Release()

	if _, err := os.Stat(l.Path()); os.IsNotExist(err) {
		t.Error("lock file should exist after Acquire")
	}
}

func TestAcquireBlocksSecondInstance(t *testing.T) {
	dir := tempDir(t)
	l1 := New(dir, "my-job")
	l2 := New(dir, "my-job")

	if err := l1.Acquire(); err != nil {
		t.Fatalf("first acquire failed: %v", err)
	}
	defer l1.Release()

	if err := l2.Acquire(); err == nil {
		defer l2.Release()
		t.Error("expected error on second acquire, got nil")
	}
}

func TestReleaseRemovesLockFile(t *testing.T) {
	dir := tempDir(t)
	l := New(dir, "cleanup-job")

	_ = l.Acquire()
	if err := l.Release(); err != nil {
		t.Fatalf("Release failed: %v", err)
	}
	if _, err := os.Stat(l.Path()); !os.IsNotExist(err) {
		t.Error("lock file should be removed after Release")
	}
}

func TestReleaseMissingFileIsNoop(t *testing.T) {
	dir := tempDir(t)
	l := New(dir, "ghost-job")
	if err := l.Release(); err != nil {
		t.Errorf("Release on missing file should not error, got: %v", err)
	}
}

func TestStaleLockIsOverwritten(t *testing.T) {
	dir := tempDir(t)
	l := New(dir, "stale-job")

	// Write a stale lock with a non-existent PID.
	stalePID := 999999999
	content := fmt.Sprintf("%d\n2000-01-01T00:00:00Z\n", stalePID)
	_ = os.WriteFile(l.Path(), []byte(content), 0600)

	if err := l.Acquire(); err != nil {
		t.Fatalf("expected stale lock to be overwritten, got: %v", err)
	}
	defer l.Release()
}

func TestSanitizeName(t *testing.T) {
	dir := tempDir(t)
	l := New(dir, "my job/task:1")
	expected := filepath.Join(dir, "cronwrap-my_job_task_1.lock")
	if l.Path() != expected {
		t.Errorf("expected path %q, got %q", expected, l.Path())
	}
}
