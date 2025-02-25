package metrics

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

	if client.httpClient == nil {
		t.Error("HTTP client not properly initialized")
	}
}

func TestSendMetrics(t *testing.T) {
	// Create a test server
	var receivedToken string
	var receivedMetrics *SystemMetrics

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Expected Bearer test-token, got %s", auth)
		}
		receivedToken = auth

		// Decode the metrics
		var metrics SystemMetrics
		if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
			t.Errorf("Failed to decode metrics: %v", err)
		}
		receivedMetrics = &metrics

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
	err := client.sendMetrics()

	if err != nil {
		t.Fatalf("Failed to send metrics: %v", err)
	}

	if receivedToken != "Bearer test-token" {
		t.Errorf("Token not properly sent")
	}

	if receivedMetrics == nil {
		t.Fatal("No metrics received")
	}

	// Basic validation of received metrics
	if receivedMetrics.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}
}

func TestSendCheckin(t *testing.T) {
	// Create a test server
	var receivedPayload CheckinPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("Expected Bearer test-token, got %s", auth)
		}

		// Decode the payload
		if err := json.NewDecoder(r.Body).Decode(&receivedPayload); err != nil {
			t.Errorf("Failed to decode payload: %v", err)
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
	err := client.sendCheckin()

	if err != nil {
		t.Fatalf("Failed to send check-in: %v", err)
	}

	if receivedPayload.Token != "test-token" {
		t.Errorf("Expected token test-token, got %s", receivedPayload.Token)
	}

	if receivedPayload.Status != "alive" {
		t.Errorf("Expected status alive, got %s", receivedPayload.Status)
	}

	if receivedPayload.IP == "" {
		t.Error("Expected non-empty IP")
	}

	// Parse and validate timestamp
	_, err = time.Parse(time.RFC3339, receivedPayload.Timestamp)
	if err != nil {
		t.Errorf("Invalid timestamp format: %v", err)
	}
}

func TestTelemetryClientStart(t *testing.T) {
	metricsCount := 0
	checkinCount := 0

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/agent/telemetry":
			metricsCount++
		case "/v1/agent/checkin":
			checkinCount++
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create a test config with short intervals for testing
	cfg := &config.Config{
		APIServer:       server.URL,
		Token:           "test-token",
		CheckinInterval: 100 * time.Millisecond,
		Metrics: config.MetricsConfig{
			CollectionInterval: 100 * time.Millisecond,
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

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	// Start the client in a goroutine
	errCh := make(chan error)
	go func() {
		errCh <- client.Start(ctx)
	}()

	// Wait for context to be done
	select {
	case err := <-errCh:
		if err != context.DeadlineExceeded {
			t.Errorf("Expected DeadlineExceeded error, got: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Test timed out")
	}

	// Check that we received some metrics and check-ins
	if metricsCount == 0 {
		t.Error("No metrics were sent")
	}
	if checkinCount == 0 {
		t.Error("No check-ins were sent")
	}
}
