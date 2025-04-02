package http

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/talis-agent/internal/config"
)

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
