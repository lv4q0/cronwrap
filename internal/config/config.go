package config

import (
	"errors"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the full cronwrap job configuration loaded from a YAML file.
type Config struct {
	Command  string            `yaml:"command"`
	Schedule string            `yaml:"schedule"`
	Timeout  time.Duration     `yaml:"timeout"`
	Retries  int               `yaml:"retries"`
	Env      map[string]string `yaml:"env"`
	LogLevel string            `yaml:"log_level"`
	Webhook  string            `yaml:"webhook"`
	LockFile string            `yaml:"lock_file"`
	HistFile string            `yaml:"history_file"`
	AuditFile string           `yaml:"audit_file"`
	Tags     map[string]string `yaml:"tags"`
}

func defaults() Config {
	return Config{
		Timeout:  30 * time.Second,
		Retries:  0,
		LogLevel: "info",
	}
}

// LoadFile reads and parses a YAML config file, applying defaults for missing
// fields. Returns an error if the file is missing, unparseable, or the
// required `command` field is absent.
func LoadFile(path string) (Config, error) {
	cfg := defaults()

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	if cfg.Command == "" {
		return cfg, errors.New("config: command is required")
	}

	return cfg, nil
}
