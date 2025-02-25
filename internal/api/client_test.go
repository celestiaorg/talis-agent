package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestNewClient(t *testing.T) {
	cfg := ClientConfig{
		BaseURL:          "http://test-api",
		Token:            "test-token",
		RequestTimeout:   10 * time.Second,
		MaxRetries:       3,
		RetryDelay:       time.Second,
		RateLimit:        rate.Limit(10),
		BurstLimit:       5,
		FailureThreshold: 5,
		ResetTimeout:     30 * time.Second,
	}

	client := NewClient(cfg)

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	if client.baseURL != cfg.BaseURL {
		t.Errorf("Expected baseURL %s, got %s", cfg.BaseURL, client.baseURL)
	}

	if client.token != cfg.Token {
		t.Errorf("Expected token %s, got %s", cfg.Token, client.token)
	}

	if client.maxRetries != cfg.MaxRetries {
		t.Errorf("Expected maxRetries %d, got %d", cfg.MaxRetries, client.maxRetries)
	}

	if client.retryDelay != cfg.RetryDelay {
		t.Errorf("Expected retryDelay %v, got %v", cfg.RetryDelay, client.retryDelay)
	}
}

func TestRequest(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type header to be application/json")
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected Authorization header to be Bearer test-token")
		}

		// Return success response
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}))
	defer server.Close()

	// Create client
	cfg := ClientConfig{
		BaseURL:          server.URL,
		Token:            "test-token",
		RequestTimeout:   10 * time.Second,
		MaxRetries:       3,
		RetryDelay:       time.Second,
		RateLimit:        rate.Limit(10),
		BurstLimit:       5,
		FailureThreshold: 5,
		ResetTimeout:     30 * time.Second,
	}
	client := NewClient(cfg)

	// Test successful request
	resp, err := client.Request(context.Background(), http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(resp, &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result["status"] != "success" {
		t.Errorf("Expected status success, got %s", result["status"])
	}
}

func TestCircuitBreaker(t *testing.T) {
	// Create test server that fails initially then recovers
	failureCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if failureCount < 5 {
			failureCount++
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}))
	defer server.Close()

	// Create client with low failure threshold
	cfg := ClientConfig{
		BaseURL:          server.URL,
		Token:            "test-token",
		RequestTimeout:   10 * time.Second,
		MaxRetries:       1,
		RetryDelay:       100 * time.Millisecond,
		RateLimit:        rate.Limit(10),
		BurstLimit:       5,
		FailureThreshold: 3,
		ResetTimeout:     500 * time.Millisecond,
	}
	client := NewClient(cfg)

	// Make requests until circuit breaker opens
	ctx := context.Background()
	var lastErr error
	for i := 0; i < 4; i++ {
		_, err := client.Request(ctx, http.MethodGet, "/test", nil)
		lastErr = err
	}

	if lastErr == nil || lastErr.Error() != "circuit breaker is open" {
		t.Errorf("Expected circuit breaker to be open, got error: %v", lastErr)
	}

	// Wait for circuit breaker to reset
	time.Sleep(600 * time.Millisecond)

	// Make request after recovery
	resp, err := client.Request(ctx, http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatalf("Expected successful request after circuit breaker reset, got error: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(resp, &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result["status"] != "success" {
		t.Errorf("Expected status success, got %s", result["status"])
	}
}

func TestRateLimiting(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}))
	defer server.Close()

	// Create client with strict rate limit
	cfg := ClientConfig{
		BaseURL:          server.URL,
		Token:            "test-token",
		RequestTimeout:   10 * time.Second,
		MaxRetries:       1,
		RetryDelay:       100 * time.Millisecond,
		RateLimit:        rate.Limit(2), // 2 requests per second
		BurstLimit:       1,
		FailureThreshold: 3,
		ResetTimeout:     500 * time.Millisecond,
	}
	client := NewClient(cfg)

	// Make concurrent requests
	ctx := context.Background()
	start := time.Now()

	// Make 5 requests
	for i := 0; i < 5; i++ {
		_, err := client.Request(ctx, http.MethodGet, "/test", nil)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i, err)
		}
	}

	duration := time.Since(start)

	// With rate limit of 2 per second, 5 requests should take at least 2 seconds
	if duration < 2*time.Second {
		t.Errorf("Requests completed too quickly. Expected > 2s, got %v", duration)
	}
}
