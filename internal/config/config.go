package config

import (
	"errors"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all cronwrap settings loaded from a YAML file.
type Config struct {
	JobName     string        `yaml:"job_name"`
	Command     string        `yaml:"command"`
	Timeout     time.Duration `yaml:"timeout"`
	Retries     int           `yaml:"retries"`
	WebhookURL  string        `yaml:"webhook_url"`
	LogLevel    string        `yaml:"log_level"`
	HistoryFile string        `yaml:"history_file"`
}

func defaults() Config {
	return Config{
		JobName:     "unnamed",
		Timeout:     30 * time.Second,
		Retries:     0,
		LogLevel:    "info",
		HistoryFile: "",
	}
}

// LoadFile reads and parses a YAML config file, applying defaults for
// any fields that are not explicitly set.
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
