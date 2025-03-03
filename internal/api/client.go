package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/celestiaorg/talis-agent/internal/logging"
)

// CircuitBreakerState represents the state of the circuit breaker
type CircuitBreakerState int

const (
	// CircuitClosed means the circuit is closed and requests can flow
	CircuitClosed CircuitBreakerState = iota
	// CircuitOpen means the circuit is open and requests are blocked
	CircuitOpen
	// CircuitHalfOpen means the circuit is testing if it can close
	CircuitHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	state            CircuitBreakerState
	failureCount     int
	lastFailure      time.Time
	failureThreshold int
	resetTimeout     time.Duration
	mutex            sync.RWMutex
}

// Client represents the API client with circuit breaker and rate limiting
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	limiter    *rate.Limiter
	breaker    *CircuitBreaker
	maxRetries int
	retryDelay time.Duration
}

// ClientConfig holds the configuration for the API client
type ClientConfig struct {
	BaseURL          string
	Token            string
	RequestTimeout   time.Duration
	MaxRetries       int
	RetryDelay       time.Duration
	RateLimit        rate.Limit
	BurstLimit       int
	FailureThreshold int
	ResetTimeout     time.Duration
}

// NewClient creates a new API client with the given configuration
func NewClient(cfg ClientConfig) *Client {
	return &Client{
		baseURL: cfg.BaseURL,
		token:   cfg.Token,
		httpClient: &http.Client{
			Timeout: cfg.RequestTimeout,
		},
		limiter: rate.NewLimiter(cfg.RateLimit, cfg.BurstLimit),
		breaker: &CircuitBreaker{
			state:            CircuitClosed,
			failureThreshold: cfg.FailureThreshold,
			resetTimeout:     cfg.ResetTimeout,
		},
		maxRetries: cfg.MaxRetries,
		retryDelay: cfg.RetryDelay,
	}
}

// Request makes an HTTP request with circuit breaker, retries, and rate limiting
func (c *Client) Request(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	// Check circuit breaker
	if !c.breaker.AllowRequest() {
		return nil, fmt.Errorf("circuit breaker is open")
	}

	// Wait for rate limiter
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.retryDelay):
			}
		}

		resp, err := c.doRequest(ctx, method, path, body)
		if err != nil {
			lastErr = err
			c.breaker.RecordFailure()
			logging.Warn().
				Err(err).
				Int("attempt", attempt+1).
				Int("max_retries", c.maxRetries).
				Msg("Request failed, will retry")
			continue
		}

		// Record success and return response
		c.breaker.RecordSuccess()
		return resp, nil
	}

	return nil, fmt.Errorf("request failed after %d retries: %w", c.maxRetries, lastErr)
}

// doRequest performs the actual HTTP request
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	url := fmt.Sprintf("%s%s", c.baseURL, path)
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			logging.Error().Err(cerr).Msg("error closing response body")
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, respBody)
	}

	return respBody, nil
}

// AllowRequest checks if a request can be made
func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.mutex.RUnlock()
			cb.mutex.Lock()
			cb.state = CircuitHalfOpen
			cb.mutex.Unlock()
			cb.mutex.RLock()
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	default:
		return false
	}
}

// RecordSuccess records a successful request
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failureCount = 0
	cb.state = CircuitClosed
}

// RecordFailure records a failed request
func (cb *CircuitBreaker) RecordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failureCount++
	cb.lastFailure = time.Now()

	if cb.state == CircuitHalfOpen || cb.failureCount >= cb.failureThreshold {
		cb.state = CircuitOpen
	}
}
