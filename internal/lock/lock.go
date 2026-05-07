// Package lock provides file-based locking to prevent concurrent cron job execution.
package lock

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Lock represents a file-based process lock.
type Lock struct {
	path string
}

// New creates a new Lock for the given job name, stored in dir.
func New(dir, jobName string) *Lock {
	fileName := fmt.Sprintf("cronwrap-%s.lock", sanitize(jobName))
	return &Lock{path: filepath.Join(dir, fileName)}
}

// Acquire attempts to acquire the lock. Returns an error if already locked.
func (l *Lock) Acquire() error {
	if _, err := os.Stat(l.path); err == nil {
		data, readErr := os.ReadFile(l.path)
		if readErr == nil {
			parts := strings.SplitN(strings.TrimSpace(string(data)), "\n", 2)
			if len(parts) >= 1 {
				if pid, err := strconv.Atoi(parts[0]); err == nil {
					if isRunning(pid) {
						return fmt.Errorf("job already running with pid %d (lock: %s)", pid, l.path)
					}
				}
			}
		}
		// Stale lock — remove it
		_ = os.Remove(l.path)
	}

	content := fmt.Sprintf("%d\n%s\n", os.Getpid(), time.Now().UTC().Format(time.RFC3339))
	if err := os.WriteFile(l.path, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write lock file: %w", err)
	}
	return nil
}

// Release removes the lock file.
func (l *Lock) Release() error {
	if err := os.Remove(l.path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove lock file: %w", err)
	}
	return nil
}

// Path returns the lock file path.
func (l *Lock) Path() string {
	return l.path
}

func sanitize(name string) string {
	replacer := strings.NewReplacer("/", "_", " ", "_", ":", "_")
	return replacer.Replace(name)
}

func isRunning(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Unix, Signal(0) checks existence without sending a signal.
	err = proc.Signal(os.Signal(nil))
	return err == nil
}
