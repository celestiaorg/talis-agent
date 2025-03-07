package config_test

import (
	"os"
	"testing"

	"github.com/celestiaorg/talis-agent/config"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Remove(tmpfile.Name()); err != nil {
			t.Logf("Error removing temporary file: %v", err)
		}
	}()

	// Write test configuration
	configContent := `
http:
  port: 8080
logging:
  level: debug
`
	if _, err := tmpfile.Write([]byte(configContent)); err != nil {
		t.Fatal(err)
	}

	// Temporarily override the config path
	originalPath := config.ConfigPath
	config.ConfigPath = tmpfile.Name()
	defer func() {
		config.ConfigPath = originalPath
	}()

	// Test loading configuration
	cfg, err := config.Load()
	assert.NoError(t, err)

	// Verify configuration values
	assert.Equal(t, 8080, cfg.HTTP.Port)
	assert.Equal(t, "debug", cfg.Logging.Level)
}

func TestLoadConfigDefaults(t *testing.T) {
	// Create a temporary config file with minimal content
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Remove(tmpfile.Name()); err != nil {
			t.Logf("Error removing temporary file: %v", err)
		}
	}()

	// Write minimal configuration
	configContent := `{}`
	if _, err := tmpfile.Write([]byte(configContent)); err != nil {
		t.Fatal(err)
	}

	// Temporarily override the config path
	originalPath := config.ConfigPath
	config.ConfigPath = tmpfile.Name()
	defer func() {
		config.ConfigPath = originalPath
	}()

	// Test loading configuration
	cfg, err := config.Load()
	assert.NoError(t, err)

	// Verify default values
	assert.Equal(t, 25550, cfg.HTTP.Port)
	assert.Equal(t, "info", cfg.Logging.Level)
}

func TestLoadConfigMissingFile(t *testing.T) {
	// Temporarily override the config path to a non-existent file
	originalPath := config.ConfigPath
	config.ConfigPath = "/non/existent/path/config.yaml"
	defer func() {
		config.ConfigPath = originalPath
	}()

	// Test loading configuration
	cfg, err := config.Load()
	assert.NoError(t, err)

	// Verify default values are set
	assert.Equal(t, 25550, cfg.HTTP.Port)
	assert.Equal(t, "info", cfg.Logging.Level)
}
