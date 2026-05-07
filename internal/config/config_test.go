package config

import (
	"os"
	"testing"
	"time"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "cronwrap-*.yaml")
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
	path := writeTemp(t, `
command: /usr/bin/backup
args: ["-v", "--all"]
schedule: "0 2 * * *"
timeout: 2m
retries: 3
webhook_url: https://hooks.example.com/alert
log_level: debug
`)
	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Command != "/usr/bin/backup" {
		t.Errorf("Command = %q, want /usr/bin/backup", cfg.Command)
	}
	if cfg.Timeout != 2*time.Minute {
		t.Errorf("Timeout = %v, want 2m", cfg.Timeout)
	}
	if cfg.Retries != 3 {
		t.Errorf("Retries = %d, want 3", cfg.Retries)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want debug", cfg.LogLevel)
	}
}

func TestLoadFileDefaults(t *testing.T) {
	path := writeTemp(t, "command: /bin/true\n")
	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("default Timeout = %v, want 30s", cfg.Timeout)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("default LogLevel = %q, want info", cfg.LogLevel)
	}
	if cfg.RetryDelay != 5*time.Second {
		t.Errorf("default RetryDelay = %v, want 5s", cfg.RetryDelay)
	}
}

func TestLoadFileMissingCommand(t *testing.T) {
	path := writeTemp(t, "schedule: \"@hourly\"\n")
	_, err := LoadFile(path)
	if err == nil {
		t.Fatal("expected error for missing command, got nil")
	}
}

func TestLoadFileNotFound(t *testing.T) {
	_, err := LoadFile("/nonexistent/path/cronwrap.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadFileInvalidYAML(t *testing.T) {
	path := writeTemp(t, "command: [unclosed")
	_, err := LoadFile(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}
