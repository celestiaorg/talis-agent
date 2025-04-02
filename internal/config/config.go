package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v2"
)

var configPaths = []string{
	"/etc/talis-agent/config.yaml",
	"config.yaml",
}

// SetConfigPaths sets the paths to search for config files (for testing only)
func SetConfigPaths(paths []string) {
	configPaths = paths
}

// Config represents the application configuration
type Config struct {
	HTTP     HTTPConfig     `yaml:"http"`
	Logging  LoggingConfig  `yaml:"logging"`
	Metrics  MetricsConfig  `yaml:"metrics"`
	Security SecurityConfig `yaml:"security"`
}

// HTTPConfig contains HTTP server configuration
type HTTPConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// MetricsConfig contains metrics collection configuration
type MetricsConfig struct {
	CollectionInterval string `yaml:"collection_interval"`
	RetentionDays      int    `yaml:"retention_days"`
}

// SecurityConfig contains security-related configuration
type SecurityConfig struct {
	TLSEnabled bool   `yaml:"tls_enabled"`
	CertFile   string `yaml:"cert_file"`
	KeyFile    string `yaml:"key_file"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		HTTP: HTTPConfig{
			Port: 25550,
			Host: "0.0.0.0",
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
		Metrics: MetricsConfig{
			CollectionInterval: "15s",
			RetentionDays:      7,
		},
		Security: SecurityConfig{
			TLSEnabled: false,
		},
	}
}

// Load loads the configuration from the specified file
func Load() (*Config, error) {
	var configFile string
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			configFile = path
			break
		}
	}

	cfg := DefaultConfig()

	if configFile == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(configFile) // nolint: gosec
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

// Save saves the configuration to a YAML file
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil { // nolint: gosec
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil { // nolint: gosec
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate HTTP port
	if c.HTTP.Port < 1 || c.HTTP.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", c.HTTP.Port)
	}

	// Validate metrics collection interval
	if _, err := time.ParseDuration(c.Metrics.CollectionInterval); err != nil {
		return fmt.Errorf("invalid collection interval: %s", c.Metrics.CollectionInterval)
	}

	// Validate log level
	switch c.Logging.Level {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("invalid log level: %s", c.Logging.Level)
	}

	return nil
}
