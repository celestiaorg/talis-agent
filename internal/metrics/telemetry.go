package metrics

import (
	"context"
	"fmt"
	"net"
	"time"

	"golang.org/x/time/rate"

	"github.com/celestiaorg/talis-agent/internal/api"
	"github.com/celestiaorg/talis-agent/internal/config"
	"github.com/celestiaorg/talis-agent/internal/logging"
)

// TelemetryClient handles sending metrics to the API server
type TelemetryClient struct {
	config    *config.Config
	collector *Collector
	apiClient *api.Client
	startTime time.Time
}

// NewTelemetryClient creates a new telemetry client
func NewTelemetryClient(cfg *config.Config) *TelemetryClient {
	// Create API client with circuit breaker and rate limiting
	apiClient := api.NewClient(api.ClientConfig{
		BaseURL:          cfg.APIServer,
		Token:            cfg.Token,
		RequestTimeout:   10 * time.Second,
		MaxRetries:       3,
		RetryDelay:       time.Second,
		RateLimit:        rate.Limit(20), // 20 requests per second
		BurstLimit:       5,              // Allow bursts of 5 requests
		FailureThreshold: 5,              // Open circuit after 5 failures
		ResetTimeout:     30 * time.Second,
	})

	return &TelemetryClient{
		config:    cfg,
		collector: NewCollector(cfg.Metrics.CollectionInterval),
		apiClient: apiClient,
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

	logging.Info().
		Str("metrics_interval", t.config.Metrics.CollectionInterval.String()).
		Str("checkin_interval", t.config.CheckinInterval.String()).
		Msg("Starting telemetry collection")

	for {
		select {
		case <-ctx.Done():
			logging.Info().Msg("Stopping telemetry collection")
			return ctx.Err()
		case <-metricsTicker.C:
			metrics, err := t.collector.Collect()
			if err != nil {
				logging.Error().Err(err).Msg("Failed to collect metrics")
				continue
			}

			// Update Prometheus metrics
			promMetrics.UpdateSystemMetrics(metrics)

			logging.Debug().
				Float64("cpu_usage", metrics.CPU.UsagePercent).
				Float64("memory_usage", metrics.Memory.UsedPercent).
				Float64("disk_usage", metrics.Disk.UsedPercent).
				Msg("System metrics collected")

			if err := t.sendMetrics(ctx, metrics); err != nil {
				logging.Error().Err(err).Msg("Failed to send metrics")
			}
		case <-checkinTicker.C:
			if err := t.sendCheckin(ctx); err != nil {
				logging.Error().Err(err).Msg("Failed to send check-in")
			} else {
				promMetrics.RecordCheckin(float64(time.Now().Unix()))
				logging.Debug().Msg("Check-in sent successfully")
			}
		case <-uptimeTicker.C:
			promMetrics.RecordUptime(1)
		}
	}
}

// sendMetrics sends metrics to the API server
func (t *TelemetryClient) sendMetrics(ctx context.Context, metrics *SystemMetrics) error {
	_, err := t.apiClient.Request(ctx, "POST", t.config.Metrics.Endpoints.Telemetry, metrics)
	if err != nil {
		return fmt.Errorf("failed to send metrics: %w", err)
	}
	return nil
}

// sendCheckin sends a check-in request to the API server
func (t *TelemetryClient) sendCheckin(ctx context.Context) error {
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

	_, err = t.apiClient.Request(ctx, "POST", t.config.Metrics.Endpoints.Checkin, payload)
	if err != nil {
		return fmt.Errorf("failed to send check-in: %w", err)
	}

	return nil
}

// getOutboundIP gets the preferred outbound IP address
func (t *TelemetryClient) getOutboundIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			logging.Error().Err(closeErr).Msg("failed to close connection")
		}
	}()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP, nil
}
