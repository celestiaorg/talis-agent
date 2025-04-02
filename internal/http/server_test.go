package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/celestiaorg/talis-agent/internal/config"
	"github.com/stretchr/testify/require"
)

func TestHandlePayload(t *testing.T) {
	// Create a temporary directory for the payload
	tmpDir, err := os.MkdirTemp("", "talis-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	payloadPath := filepath.Join(tmpDir, "payload")
	t.Setenv("TALIS_PAYLOAD_PATH", payloadPath)

	cfg := &config.Config{
		HTTP: config.HTTPConfig{
			Host: "localhost",
			Port: 25550,
		},
	}

	server := NewServer(cfg)
	require.NotNil(t, server)

	// Create a test request
	payload := []byte("test payload")
	req := httptest.NewRequest(http.MethodPost, "/payload", bytes.NewReader(payload))
	w := httptest.NewRecorder()

	// Handle the request
	server.handlePayload(w, req)

	// Check response
	require.Equal(t, http.StatusOK, w.Code)

	// Verify the file was written
	content, err := os.ReadFile(payloadPath)
	require.NoError(t, err)
	require.Equal(t, payload, content)
}

func TestHandleCommands(t *testing.T) {
	cfg := &config.Config{
		HTTP: config.HTTPConfig{
			Host: "localhost",
			Port: 25550,
		},
	}

	server := NewServer(cfg)
	require.NotNil(t, server)

	// Create a test request
	cmdReq := CommandRequest{Command: "echo 'test'"}
	body, err := json.Marshal(cmdReq)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/commands", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// Handle the request
	server.handleCommands(w, req)

	// Check response
	require.Equal(t, http.StatusOK, w.Code)

	var resp CommandResponse
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	require.Equal(t, "test\n", resp.Output)
	require.Empty(t, resp.Error)
}

func TestNewServer(t *testing.T) {
	cfg := &config.Config{
		HTTP: config.HTTPConfig{
			Host: "localhost",
			Port: 25550,
		},
		Metrics: config.MetricsConfig{
			CollectionInterval: "15s",
			RetentionDays:      7,
		},
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "json",
		},
		Security: config.SecurityConfig{
			TLSEnabled: false,
			CertFile:   "",
			KeyFile:    "",
		},
	}

	server := NewServer(cfg)
	require.NotNil(t, server)
	require.Equal(t, cfg, server.config)
}

func TestServerAddress(t *testing.T) {
	cfg := &config.Config{
		HTTP: config.HTTPConfig{
			Host: "localhost",
			Port: 25550,
		},
		Metrics: config.MetricsConfig{
			CollectionInterval: "15s",
			RetentionDays:      7,
		},
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "json",
		},
		Security: config.SecurityConfig{
			TLSEnabled: false,
			CertFile:   "",
			KeyFile:    "",
		},
	}

	server := NewServer(cfg)
	require.Equal(t, "localhost:25550", server.Address())
}

func TestServerStart(t *testing.T) {
	cfg := &config.Config{
		HTTP: config.HTTPConfig{
			Host: "localhost",
			Port: 0, // Use port 0 to let the OS assign a random port
		},
		Metrics: config.MetricsConfig{
			CollectionInterval: "15s",
			RetentionDays:      7,
		},
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "json",
		},
		Security: config.SecurityConfig{
			TLSEnabled: false,
			CertFile:   "",
			KeyFile:    "",
		},
	}

	server := NewServer(cfg)
	require.NotNil(t, server)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start server in background
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Start(ctx)
	}()

	// Wait for context to be done
	select {
	case err := <-errCh:
		require.NoError(t, err)
	case <-ctx.Done():
		// Expected timeout
	}
}
