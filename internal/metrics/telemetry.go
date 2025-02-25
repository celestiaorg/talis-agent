package metrics

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/celestiaorg/talis-agent/internal/config"
)

// TelemetryClient handles sending metrics to the API server
type TelemetryClient struct {
	config     *config.Config
	collector  *Collector
	httpClient *http.Client
	startTime  time.Time
}

// NewTelemetryClient creates a new telemetry client
func NewTelemetryClient(cfg *config.Config) *TelemetryClient {
	return &TelemetryClient{
		config:    cfg,
		collector: NewCollector(cfg.Metrics.CollectionInterval),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   5 * time.Second,
				ResponseHeaderTimeout: 5 * time.Second,
			},
		},
		startTime: time.Now(),
	}
}

// CheckinPayload represents the payload for agent check-ins
type CheckinPayload struct {
	Token     string `json:"token"`
	IP        string `json:"ip"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// Start begins the telemetry collection and transmission loop
func (t *TelemetryClient) Start(ctx context.Context) error {
	// Start the metrics collection loop
	metricsTicker := time.NewTicker(t.config.Metrics.CollectionInterval)
	defer metricsTicker.Stop()

	// Start the check-in loop
	checkinTicker := time.NewTicker(t.config.CheckinInterval)
	defer checkinTicker.Stop()

	// Start the uptime recording loop
	uptimeTicker := time.NewTicker(time.Second)
	defer uptimeTicker.Stop()

	promMetrics := GetPrometheusMetrics()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-metricsTicker.C:
			metrics, err := t.collector.Collect()
			if err != nil {
				fmt.Printf("Failed to collect metrics: %v\n", err)
				continue
			}

			// Update Prometheus metrics
			promMetrics.UpdateSystemMetrics(metrics)

			if err := t.sendMetrics(); err != nil {
				fmt.Printf("Failed to send metrics: %v\n", err)
			}
		case <-checkinTicker.C:
			if err := t.sendCheckin(); err != nil {
				fmt.Printf("Failed to send check-in: %v\n", err)
			} else {
				promMetrics.RecordCheckin(float64(time.Now().Unix()))
			}
		case <-uptimeTicker.C:
			promMetrics.RecordUptime(1) // Record 1 second of uptime
		}
	}
}

// sendMetrics collects and sends metrics to the API server
func (t *TelemetryClient) sendMetrics() error {
	metrics, err := t.collector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect metrics: %w", err)
	}

	data, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	url := fmt.Sprintf("%s%s", t.config.APIServer, t.config.Metrics.Endpoints.Telemetry)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.config.Token))

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send metrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// sendCheckin sends a check-in request to the API server
func (t *TelemetryClient) sendCheckin() error {
	// Get the IP address
	ip, err := t.getOutboundIP()
	if err != nil {
		return fmt.Errorf("failed to get IP address: %w", err)
	}

	payload := CheckinPayload{
		Token:     t.config.Token,
		IP:        ip.String(),
		Status:    "alive",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal check-in payload: %w", err)
	}

	url := fmt.Sprintf("%s%s", t.config.APIServer, t.config.Metrics.Endpoints.Checkin)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.config.Token))

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send check-in: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// getOutboundIP gets the preferred outbound IP address
func (t *TelemetryClient) getOutboundIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP, nil
}
