package history_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourorg/cronwrap/internal/history"
)

func tempPath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "history.json")
}

func makeEntry(name string, success bool) history.Entry {
	return history.Entry{
		JobName:   name,
		StartedAt: time.Now().UTC(),
		Duration:  2 * time.Second,
		ExitCode:  0,
		Success:   success,
		Attempts:  1,
		Output:    "ok",
	}
}

func TestAppendAndReadAll(t *testing.T) {
	store := history.NewStore(tempPath(t))

	if err := store.Append(makeEntry("job-a", true)); err != nil {
		t.Fatalf("first append: %v", err)
	}
	if err := store.Append(makeEntry("job-b", false)); err != nil {
		t.Fatalf("second append: %v", err)
	}

	entries, err := store.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].JobName != "job-a" || entries[1].JobName != "job-b" {
		t.Errorf("unexpected entries: %+v", entries)
	}
}

func TestReadAllMissingFile(t *testing.T) {
	store := history.NewStore("/tmp/cronwrap_nonexistent_history.json")
	entries, err := store.ReadAll()
	if err != nil {
		t.Fatalf("expected nil error for missing file, got %v", err)
	}
	if entries != nil {
		t.Errorf("expected nil entries, got %v", entries)
	}
}

func TestAppendCreatesFile(t *testing.T) {
	p := tempPath(t)
	store := history.NewStore(p)

	if err := store.Append(makeEntry("create-test", true)); err != nil {
		t.Fatalf("append: %v", err)
	}
	if _, err := os.Stat(p); err != nil {
		t.Errorf("file not created: %v", err)
	}
}
