package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/example/cronwrap/internal/config"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString(content)
	_ = f.Close()
	return f.Name()
}

func TestLoadFileValid(t *testing.T) {
	path := writeTemp(t, `
command: echo hello
schedule: "@hourly"
timeout: 10s
retries: 3
env:
  FOO: bar
tags:
  team: platform
`)
	cfg, err := config.LoadFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Command != "echo hello" {
		t.Errorf("command mismatch: %q", cfg.Command)
	}
	if cfg.Timeout != 10*time.Second {
		t.Errorf("timeout mismatch: %v", cfg.Timeout)
	}
	if cfg.Retries != 3 {
		t.Errorf("retries mismatch: %d", cfg.Retries)
	}
	if cfg.Env["FOO"] != "bar" {
		t.Errorf("env mismatch: %v", cfg.Env)
	}
	if cfg.Tags["team"] != "platform" {
		t.Errorf("tags mismatch: %v", cfg.Tags)
	}
}

func TestLoadFileDefaults(t *testing.T) {
	path := writeTemp(t, "command: ls\n")
	cfg, err := config.LoadFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("expected default timeout 30s, got %v", cfg.Timeout)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("expected default log level info, got %q", cfg.LogLevel)
	}
	if cfg.Retries != 0 {
		t.Errorf("expected default retries 0, got %d", cfg.Retries)
	}
}

func TestLoadFileMissingCommand(t *testing.T) {
	path := writeTemp(t, "schedule: \"@daily\"\n")
	_, err := config.LoadFile(path)
	if err == nil {
		t.Fatal("expected error for missing command")
	}
}

func TestLoadFileNotFound(t *testing.T) {
	_, err := config.LoadFile(filepath.Join(t.TempDir(), "missing.yaml"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadFileInvalidYAML(t *testing.T) {
	path := writeTemp(t, ": : : invalid yaml {{{")
	_, err := config.LoadFile(path)
	if err == nil {
		t.Fatal("expected parse error")
	}
}
