package config

import (
	"errors"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all cronwrap settings loaded from a YAML file.
type Config struct {
	Command     string        `yaml:"command"`
	Args        []string      `yaml:"args"`
	Schedule    string        `yaml:"schedule"`
	Timeout     time.Duration `yaml:"timeout"`
	Retries     int           `yaml:"retries"`
	RetryDelay  time.Duration `yaml:"retry_delay"`
	WebhookURL  string        `yaml:"webhook_url"`
	LogLevel    string        `yaml:"log_level"`
	HistoryFile string        `yaml:"history_file"`
}

func defaults() Config {
	return Config{
		Timeout:     30 * time.Second,
		Retries:     0,
		RetryDelay:  5 * time.Second,
		LogLevel:    "info",
		HistoryFile: "/var/log/cronwrap/history.jsonl",
	}
}

// LoadFile reads and parses a YAML config file, applying defaults for omitted fields.
func LoadFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := defaults()
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Command == "" {
		return nil, errors.New("config: command is required")
	}
	return &cfg, nil
}
