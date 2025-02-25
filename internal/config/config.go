package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure
type Config struct {
	APIServer       string        `yaml:"api_server"`
	Token           string        `yaml:"token"`
	CheckinInterval time.Duration `yaml:"checkin_interval"`
	HTTPPort        int           `yaml:"http_port"`
	LogLevel        string        `yaml:"log_level"`
	Metrics         MetricsConfig `yaml:"metrics"`
	Payload         PayloadConfig `yaml:"payload"`
}

// MetricsConfig holds the metrics-related configuration
type MetricsConfig struct {
	CollectionInterval time.Duration `yaml:"collection_interval"`
	Endpoints          struct {
		Telemetry string `yaml:"telemetry"`
		Checkin   string `yaml:"checkin"`
	} `yaml:"endpoints"`
}

// PayloadConfig holds the payload-related configuration
type PayloadConfig struct {
	Path string `yaml:"path"`
}

// Load reads and parses the configuration file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// validateConfig performs basic validation of the configuration
func validateConfig(config *Config) error {
	if config.APIServer == "" {
		return fmt.Errorf("api_server is required")
	}
	if config.Token == "" {
		return fmt.Errorf("token is required")
	}
	if config.HTTPPort <= 0 || config.HTTPPort > 65535 {
		return fmt.Errorf("invalid http_port: %d", config.HTTPPort)
	}
	return nil
}
