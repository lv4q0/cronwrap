package lock

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// Lock is a file-based advisory lock. It stores the owning process PID inside
// the lock file so that stale locks left by crashed processes can be detected
// and reclaimed.
type Lock struct {
	path string
}

// New returns a Lock that uses path as the lock file.
func New(path string) *Lock {
	return &Lock{path: path}
}

// TryAcquire attempts to acquire the lock without blocking. It returns true if
// the lock was acquired, false if another live process holds it, and an error
// if the underlying file operations fail.
func (l *Lock) TryAcquire() (bool, error) {
	if err := os.MkdirAll(filepath.Dir(l.path), 0o755); err != nil {
		return false, fmt.Errorf("lock: create dir: %w", err)
	}

	// Check for an existing lock file.
	if data, err := os.ReadFile(l.path); err == nil {
		pid, parseErr := strconv.Atoi(strings.TrimSpace(string(data)))
		if parseErr == nil && isAlive(pid) {
			return false, nil
		}
		// Stale lock — remove it and proceed.
		_ = os.Remove(l.path)
	}

	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return false, nil
		}
		return false, fmt.Errorf("lock: open: %w", err)
	}
	defer f.Close()

	if _, err := fmt.Fprintf(f, "%d\n", os.Getpid()); err != nil {
		_ = os.Remove(l.path)
		return false, fmt.Errorf("lock: write pid: %w", err)
	}
	return true, nil
}

// Release removes the lock file. It is a no-op if the file does not exist.
func (l *Lock) Release() error {
	if err := os.Remove(l.path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("lock: release: %w", err)
	}
	return nil
}

// Path returns the path of the lock file.
func (l *Lock) Path() string { return l.path }

// isAlive returns true if the process with the given PID is running.
func isAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Unix, signal 0 tests whether the process exists without sending a
	// real signal. On Windows FindProcess always succeeds, so this may
	// return true for dead processes — acceptable for our use-case.
	return proc.Signal(syscall.Signal(0)) == nil
}
