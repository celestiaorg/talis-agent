package metrics

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/celestiaorg/talis-agent/internal/config"
)

func TestNewTelemetryClient(t *testing.T) {
	cfg := &config.Config{
		APIServer:       "http://localhost:8080",
		Token:           "test-token",
		CheckinInterval: 5 * time.Second,
		Metrics: config.MetricsConfig{
			CollectionInterval: 5 * time.Second,
			Endpoints: struct {
				Telemetry string `yaml:"telemetry"`
				Checkin   string `yaml:"checkin"`
			}{
				Telemetry: "/v1/agent/telemetry",
				Checkin:   "/v1/agent/checkin",
			},
		},
	}

	client := NewTelemetryClient(cfg)

	if client == nil {
		t.Fatal("Expected non-nil telemetry client")
	}

	if client.config != cfg {
		t.Error("Config not properly set")
	}

	if client.collector == nil {
		t.Error("Collector not properly initialized")
	}

	if client.apiClient == nil {
		t.Error("API client not properly initialized")
	}
}

func TestSendMetrics(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify authorization header
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected Authorization header to be Bearer test-token")
		}

		// Verify metrics payload
		var metrics SystemMetrics
		if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
			t.Errorf("Failed to decode metrics: %v", err)
		}
		if metrics.CPU.UsagePercent < 0 || metrics.CPU.UsagePercent > 100 {
			t.Errorf("Invalid CPU usage: %v", metrics.CPU.UsagePercent)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create a test config
	cfg := &config.Config{
		APIServer:       server.URL,
		Token:           "test-token",
		CheckinInterval: 5 * time.Second,
		Metrics: config.MetricsConfig{
			CollectionInterval: 5 * time.Second,
			Endpoints: struct {
				Telemetry string `yaml:"telemetry"`
				Checkin   string `yaml:"checkin"`
			}{
				Telemetry: "/v1/agent/telemetry",
				Checkin:   "/v1/agent/checkin",
			},
		},
	}

	client := NewTelemetryClient(cfg)

	// Collect metrics
	metrics, err := client.collector.Collect()
	if err != nil {
		t.Fatalf("Failed to collect metrics: %v", err)
	}

	// Send metrics
	ctx := context.Background()
	if err := client.sendMetrics(ctx, metrics); err != nil {
		t.Fatalf("Failed to send metrics: %v", err)
	}
}

func TestSendCheckin(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify authorization header
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected Authorization header to be Bearer test-token")
		}

		// Verify checkin payload
		var checkin CheckinPayload
		if err := json.NewDecoder(r.Body).Decode(&checkin); err != nil {
			t.Errorf("Failed to decode checkin: %v", err)
		}
		if checkin.Token != "test-token" {
			t.Errorf("Expected token test-token, got %s", checkin.Token)
		}
		if checkin.Status != "alive" {
			t.Errorf("Expected status alive, got %s", checkin.Status)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create a test config
	cfg := &config.Config{
		APIServer:       server.URL,
		Token:           "test-token",
		CheckinInterval: 5 * time.Second,
		Metrics: config.MetricsConfig{
			CollectionInterval: 5 * time.Second,
			Endpoints: struct {
				Telemetry string `yaml:"telemetry"`
				Checkin   string `yaml:"checkin"`
			}{
				Telemetry: "/v1/agent/telemetry",
				Checkin:   "/v1/agent/checkin",
			},
		},
	}

	client := NewTelemetryClient(cfg)

	// Send checkin
	ctx := context.Background()
	if err := client.sendCheckin(ctx); err != nil {
		t.Fatalf("Failed to send checkin: %v", err)
	}
}

func TestTelemetryClientStart(t *testing.T) {
	var (
		metricsCount     = 0
		checkinCount     = 0
		expectedMetrics  = 1
		expectedCheckins = 1
		mu               sync.Mutex
		requestWg        sync.WaitGroup
	)

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify authorization header
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected Authorization header to be Bearer test-token")
		}

		switch r.URL.Path {
		case "/v1/agent/telemetry":
			mu.Lock()
			if metricsCount < expectedMetrics {
				metricsCount++
				requestWg.Done()
			}
			mu.Unlock()
			// Verify metrics payload
			var metrics SystemMetrics
			if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
				t.Errorf("Failed to decode metrics: %v", err)
			}
			if metrics.CPU.UsagePercent < 0 || metrics.CPU.UsagePercent > 100 {
				t.Errorf("Invalid CPU usage: %v", metrics.CPU.UsagePercent)
			}
		case "/v1/agent/checkin":
			mu.Lock()
			if checkinCount < expectedCheckins {
				checkinCount++
				requestWg.Done()
			}
			mu.Unlock()
			// Verify checkin payload
			var checkin CheckinPayload
			if err := json.NewDecoder(r.Body).Decode(&checkin); err != nil {
				t.Errorf("Failed to decode checkin: %v", err)
			}
			if checkin.Token != "test-token" {
				t.Errorf("Expected token test-token, got %s", checkin.Token)
			}
			if checkin.Status != "alive" {
				t.Errorf("Expected status alive, got %s", checkin.Status)
			}
		default:
			t.Errorf("Unexpected request to %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create a test config with reasonable intervals for testing
	cfg := &config.Config{
		APIServer:       server.URL,
		Token:           "test-token",
		CheckinInterval: 100 * time.Millisecond, // Keep interval short but not too short
		Metrics: config.MetricsConfig{
			CollectionInterval: 100 * time.Millisecond, // Keep interval short but not too short
			Endpoints: struct {
				Telemetry string `yaml:"telemetry"`
				Checkin   string `yaml:"checkin"`
			}{
				Telemetry: "/v1/agent/telemetry",
				Checkin:   "/v1/agent/checkin",
			},
		},
	}

	client := NewTelemetryClient(cfg)

	// Set up WaitGroup for expected requests
	requestWg.Add(expectedMetrics + expectedCheckins)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start the client in a goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- client.Start(ctx)
	}()

	// Wait for all expected requests or timeout
	waitCh := make(chan struct{})
	go func() {
		requestWg.Wait()
		close(waitCh)
	}()

	// Wait for completion or timeout
	select {
	case <-waitCh:
		// Verify we got at least the expected number of each type of request
		mu.Lock()
		if metricsCount < expectedMetrics {
			t.Errorf("Expected at least %d metrics, got %d", expectedMetrics, metricsCount)
		}
		if checkinCount < expectedCheckins {
			t.Errorf("Expected at least %d checkins, got %d", expectedCheckins, checkinCount)
		}
		mu.Unlock()
		cancel() // Stop the client
		if err := <-errCh; err != nil && !errors.Is(err, context.Canceled) {
			t.Errorf("Expected context.Canceled error, got: %v", err)
		}
	case err := <-errCh:
		t.Fatalf("Client stopped unexpectedly: %v", err)
	case <-ctx.Done():
		mu.Lock()
		t.Fatalf("Test timed out waiting for requests. Got %d/%d metrics and %d/%d checkins",
			metricsCount, expectedMetrics, checkinCount, expectedCheckins)
		mu.Unlock()
	}
}
