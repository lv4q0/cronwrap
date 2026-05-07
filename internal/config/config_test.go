package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "cronwrap-*.json")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoadFileValid(t *testing.T) {
	path := writeTemp(t, `{
		"command": "echo",
		"args": ["hello"],
		"timeout": 10000000000,
		"retries": 2,
		"retry_delay": 1000000000,
		"log_level": "debug",
		"webhook_url": "http://example.com/hook",
		"job_name": "test-job"
	}`)

	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Command != "echo" {
		t.Errorf("command: got %q, want %q", cfg.Command, "echo")
	}
	if cfg.Retries != 2 {
		t.Errorf("retries: got %d, want 2", cfg.Retries)
	}
	if cfg.Timeout != 10*time.Second {
		t.Errorf("timeout: got %v, want 10s", cfg.Timeout)
	}
}

func TestLoadFileDefaults(t *testing.T) {
	path := writeTemp(t, `{"command": "backup.sh"}`)

	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("log_level default: got %q, want %q", cfg.LogLevel, "info")
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("timeout default: got %v, want 30s", cfg.Timeout)
	}
}

func TestLoadFileMissingCommand(t *testing.T) {
	path := writeTemp(t, `{"retries": 1}`)
	_, err := LoadFile(path)
	if err == nil {
		t.Fatal("expected error for missing command, got nil")
	}
}

func TestLoadFileNotFound(t *testing.T) {
	_, err := LoadFile(filepath.Join(t.TempDir(), "nonexistent.json"))
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestValidateNegativeRetries(t *testing.T) {
	cfg := defaults()
	cfg.Command = "echo"
	cfg.Retries = -1
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for negative retries, got nil")
	}
}
