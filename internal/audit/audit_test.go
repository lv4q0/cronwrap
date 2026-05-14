package audit_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cronwrap/cronwrap/internal/audit"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "audit.jsonl")
}

func makeEntry(job string, success bool, trigger audit.TriggerKind) audit.Entry {
	return audit.Entry{
		Timestamp: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
		JobName:   job,
		Trigger:   trigger,
		Success:   success,
		Duration:  1.23,
		Message:   "ok",
	}
}

func TestRecordAndReadAll(t *testing.T) {
	log := audit.NewLog(tempPath(t))

	e1 := makeEntry("backup", true, audit.TriggerScheduled)
	e2 := makeEntry("cleanup", false, audit.TriggerManual)

	if err := log.Record(e1); err != nil {
		t.Fatalf("Record e1: %v", err)
	}
	if err := log.Record(e2); err != nil {
		t.Fatalf("Record e2: %v", err)
	}

	entries, err := log.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].JobName != "backup" {
		t.Errorf("expected job backup, got %s", entries[0].JobName)
	}
	if entries[1].Success {
		t.Errorf("expected cleanup to be failed")
	}
}

func TestReadAllMissingFile(t *testing.T) {
	log := audit.NewLog("/tmp/cronwrap_no_such_audit_file.jsonl")
	entries, err := log.ReadAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entries != nil {
		t.Errorf("expected nil entries for missing file")
	}
}

func TestRecordCreatesFile(t *testing.T) {
	path := tempPath(t)
	log := audit.NewLog(path)

	if err := log.Record(makeEntry("job", true, audit.TriggerRetry)); err != nil {
		t.Fatalf("Record: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected file to exist: %v", err)
	}
}

func TestRecordSetsTimestampWhenZero(t *testing.T) {
	log := audit.NewLog(tempPath(t))
	e := audit.Entry{JobName: "job", Trigger: audit.TriggerScheduled, Success: true}
	if err := log.Record(e); err != nil {
		t.Fatalf("Record: %v", err)
	}
	entries, _ := log.ReadAll()
	if entries[0].Timestamp.IsZero() {
		t.Error("expected non-zero timestamp to be set automatically")
	}
}

func TestRecordBadPath(t *testing.T) {
	log := audit.NewLog("/no_such_dir/audit.jsonl")
	err := log.Record(makeEntry("job", true, audit.TriggerScheduled))
	if err == nil {
		t.Error("expected error for bad path")
	}
}
