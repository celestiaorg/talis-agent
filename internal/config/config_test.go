package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	// Save the original config paths and restore them after the test
	origPaths := configPaths
	defer func() { configPaths = origPaths }()

	// Set empty config paths to ensure we get the default config
	SetConfigPaths([]string{})

	// Test loading the config (should return default config when no file exists)
	cfg, err := Load()
	require.NoError(t, err)

	// Verify default values
	require.Equal(t, "0.0.0.0", cfg.HTTP.Host)
	require.Equal(t, 25550, cfg.HTTP.Port)
	require.Equal(t, "15s", cfg.Metrics.CollectionInterval)
	require.Equal(t, 7, cfg.Metrics.RetentionDays)
	require.Equal(t, "info", cfg.Logging.Level)
	require.Equal(t, "json", cfg.Logging.Format)
	require.False(t, cfg.Security.TLSEnabled)
	require.Empty(t, cfg.Security.CertFile)
	require.Empty(t, cfg.Security.KeyFile)
}

func TestLoadCustomConfig(t *testing.T) {
	// Save the original config paths and restore them after the test
	origPaths := configPaths
	defer func() { configPaths = origPaths }()

	// Create a test config
	testCfg := &Config{
		HTTP: HTTPConfig{
			Host: "localhost",
			Port: 8080,
		},
		Metrics: MetricsConfig{
			CollectionInterval: "30s",
			RetentionDays:      14,
		},
		Logging: LoggingConfig{
			Level:  "debug",
			Format: "text",
		},
		Security: SecurityConfig{
			TLSEnabled: true,
			CertFile:   "cert.pem",
			KeyFile:    "key.pem",
		},
	}

	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "talis-agent-test")
	require.NoError(t, err)
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("failed to remove temporary directory: %v", err)
		}
	}()

	// Save the test config
	configPath := filepath.Join(tmpDir, "config.yaml")
	err = testCfg.Save(configPath)
	require.NoError(t, err)

	// Set the config path to our test config
	SetConfigPaths([]string{configPath})

	// Test loading the config
	cfg, err := Load()
	require.NoError(t, err)

	// Verify custom values
	require.Equal(t, "localhost", cfg.HTTP.Host)
	require.Equal(t, 8080, cfg.HTTP.Port)
	require.Equal(t, "30s", cfg.Metrics.CollectionInterval)
	require.Equal(t, 14, cfg.Metrics.RetentionDays)
	require.Equal(t, "debug", cfg.Logging.Level)
	require.Equal(t, "text", cfg.Logging.Format)
	require.True(t, cfg.Security.TLSEnabled)
	require.Equal(t, "cert.pem", cfg.Security.CertFile)
	require.Equal(t, "key.pem", cfg.Security.KeyFile)
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	require.Equal(t, "0.0.0.0", cfg.HTTP.Host)
	require.Equal(t, 25550, cfg.HTTP.Port)
	require.Equal(t, "15s", cfg.Metrics.CollectionInterval)
	require.Equal(t, 7, cfg.Metrics.RetentionDays)
	require.Equal(t, "info", cfg.Logging.Level)
	require.Equal(t, "json", cfg.Logging.Format)
	require.False(t, cfg.Security.TLSEnabled)
	require.Empty(t, cfg.Security.CertFile)
	require.Empty(t, cfg.Security.KeyFile)
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				HTTP: HTTPConfig{
					Host: "localhost",
					Port: 25550,
				},
				Metrics: MetricsConfig{
					CollectionInterval: "15s",
					RetentionDays:      7,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				Security: SecurityConfig{
					TLSEnabled: false,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid port",
			config: Config{
				HTTP: HTTPConfig{
					Host: "localhost",
					Port: 70000,
				},
				Metrics: MetricsConfig{
					CollectionInterval: "15s",
					RetentionDays:      7,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				Security: SecurityConfig{
					TLSEnabled: false,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid collection interval",
			config: Config{
				HTTP: HTTPConfig{
					Host: "localhost",
					Port: 25550,
				},
				Metrics: MetricsConfig{
					CollectionInterval: "invalid",
					RetentionDays:      7,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				Security: SecurityConfig{
					TLSEnabled: false,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid log level",
			config: Config{
				HTTP: HTTPConfig{
					Host: "localhost",
					Port: 25550,
				},
				Metrics: MetricsConfig{
					CollectionInterval: "15s",
					RetentionDays:      7,
				},
				Logging: LoggingConfig{
					Level:  "invalid",
					Format: "json",
				},
				Security: SecurityConfig{
					TLSEnabled: false,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
