package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Config holds the full cronwrap configuration.
type Config struct {
	Command    string        `json:"command"`
	Args       []string      `json:"args"`
	Timeout    time.Duration `json:"timeout"`
	Retries    int           `json:"retries"`
	RetryDelay time.Duration `json:"retry_delay"`
	LogLevel   string        `json:"log_level"`
	WebhookURL string        `json:"webhook_url"`
	JobName    string        `json:"job_name"`
}

// defaults returns a Config populated with sensible defaults.
func defaults() Config {
	return Config{
		Timeout:    30 * time.Second,
		Retries:    0,
		RetryDelay: 5 * time.Second,
		LogLevel:   "info",
		JobName:    "cronwrap-job",
	}
}

// LoadFile reads a JSON config file from the given path and returns a Config.
// Missing fields fall back to defaults.
func LoadFile(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("config: open %q: %w", path, err)
	}
	defer f.Close()

	cfg := defaults()
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("config: decode %q: %w", path, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Validate checks that the config contains the required fields and
// that numeric values are within acceptable ranges.
func (c *Config) Validate() error {
	if c.Command == "" {
		return fmt.Errorf("config: command must not be empty")
	}
	if c.Retries < 0 {
		return fmt.Errorf("config: retries must be >= 0, got %d", c.Retries)
	}
	if c.Timeout <= 0 {
		return fmt.Errorf("config: timeout must be > 0")
	}
	return nil
}
