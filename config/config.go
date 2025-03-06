package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ConfigPath is the path to the configuration file
var ConfigPath = "config.yaml"

type Config struct {
	HTTP struct {
		Port int `yaml:"port"`
	} `yaml:"http"`
	Logging struct {
		Level string `yaml:"level"`
	} `yaml:"logging"`
}

func Load() (*Config, error) {
	config := &Config{}

	// Try to read from current directory
	data, err := os.ReadFile(ConfigPath)
	if err != nil {
		// If not found, try the system path
		systemPath := "/etc/talis-agent/config.yaml"
		data, err = os.ReadFile(systemPath)
		if err != nil {
			// If both paths fail, return default config
			config.HTTP.Port = 25550
			config.Logging.Level = "info"
			return config, nil
		}
	}

	// If we got data, parse it
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults if not specified
	if config.HTTP.Port == 0 {
		config.HTTP.Port = 25550
	}
	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}

	return config, nil
}
