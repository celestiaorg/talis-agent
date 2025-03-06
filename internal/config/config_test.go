package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Create a temporary config file
	configContent := `
api_server: "http://test-server:8080"
token: "test-token"
checkin_interval: "5s"
http_port: 25550
log_level: "debug"
metrics:
  collection_interval: "5s"
  endpoints:
    telemetry: "/v1/agent/telemetry"
    checkin: "/v1/agent/checkin"
payload:
  path: "/etc/talis-agent/payload"
`
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpfile.Name()); err != nil {
			t.Errorf("Failed to remove temporary file: %v", err)
		}
	}()

	if _, err := tmpfile.Write([]byte(configContent)); err != nil {
		t.Fatalf("Failed to write config content: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Test loading valid configuration
	config, err := Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify configuration values
	if config.APIServer != "http://test-server:8080" {
		t.Errorf("Expected APIServer to be 'http://test-server:8080', got '%s'", config.APIServer)
	}
	if config.Token != "test-token" {
		t.Errorf("Expected Token to be 'test-token', got '%s'", config.Token)
	}
	if config.HTTPPort != 25550 {
		t.Errorf("Expected HTTPPort to be 25550, got %d", config.HTTPPort)
	}
	if config.CheckinInterval != 5*time.Second {
		t.Errorf("Expected CheckinInterval to be 5s, got %v", config.CheckinInterval)
	}
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
				APIServer: "http://localhost:8080",
				Token:     "valid-token",
				HTTPPort:  25550,
			},
			wantErr: false,
		},
		{
			name: "missing api server",
			config: Config{
				Token:    "valid-token",
				HTTPPort: 25550,
			},
			wantErr: true,
		},
		{
			name: "missing token",
			config: Config{
				APIServer: "http://localhost:8080",
				HTTPPort:  25550,
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			config: Config{
				APIServer: "http://localhost:8080",
				Token:     "valid-token",
				HTTPPort:  70000,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(&tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
