package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	// Create a temporary config file
	content := `
http:
  host: "localhost"
  port: 8080
metrics:
  collection_interval: "15s"
  retention_days: 7
logging:
  level: "info"
  format: "json"
security:
  tls_enabled: false
  cert_file: ""
  key_file: ""
`
	tmpfile, err := os.CreateTemp("", "config.*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(content))
	require.NoError(t, err)
	err = tmpfile.Close()
	require.NoError(t, err)

	// Test loading the config
	cfg, err := Load()
	require.NoError(t, err)

	// Verify the loaded values
	require.Equal(t, "localhost", cfg.HTTP.Host)
	require.Equal(t, 8080, cfg.HTTP.Port)
	require.Equal(t, "15s", cfg.Metrics.CollectionInterval)
	require.Equal(t, 7, cfg.Metrics.RetentionDays)
	require.Equal(t, "info", cfg.Logging.Level)
	require.Equal(t, "json", cfg.Logging.Format)
	require.False(t, cfg.Security.TLSEnabled)
	require.Empty(t, cfg.Security.CertFile)
	require.Empty(t, cfg.Security.KeyFile)
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
