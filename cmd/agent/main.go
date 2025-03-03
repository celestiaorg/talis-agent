package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/celestiaorg/talis-agent/internal/config"
	"github.com/celestiaorg/talis-agent/internal/http"
	"github.com/celestiaorg/talis-agent/internal/logging"
	"github.com/celestiaorg/talis-agent/internal/metrics"
)

var (
	configPath = flag.String("config", "config.yaml", "path to configuration file")
	logPath    = flag.String("log", "", "path to log file (default: /var/log/talis-agent/agent.log)")
	version    = "0.1.0" // This should be set during build
)

func main() {
	flag.Parse()

	// Initialize logging with both console and file output
	logConfig := logging.Config{
		Level:      "info", // Default to info level
		TimeFormat: time.RFC3339,
		Console:    true, // Use console format for better readability during development
		File:       logging.DefaultFileConfig(),
	}

	// Override log file path if provided via flag
	if *logPath != "" {
		logConfig.File.Path = *logPath
	}

	// Initialize logging
	if err := logging.InitLogger(logConfig); err != nil {
		// If we can't initialize logging, write to stderr and exit
		fmt.Fprintf(os.Stderr, "Failed to initialize logging: %v\n", err)
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		logging.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Update log level from configuration
	if cfg.LogLevel != "" {
		logConfig.Level = cfg.LogLevel
		if err := logging.InitLogger(logConfig); err != nil {
			logging.Error().Err(err).Msg("Failed to update log level")
		}
	}

	// Print startup information
	logging.Info().
		Str("version", version).
		Str("config_path", *configPath).
		Str("log_path", logConfig.File.Path).
		Str("api_server", cfg.APIServer).
		Int("http_port", cfg.HTTPPort).
		Str("metrics_interval", cfg.Metrics.CollectionInterval.String()).
		Msg("Starting Talis Agent")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create wait group for goroutines
	var wg sync.WaitGroup

	// Initialize Prometheus metrics
	promMetrics := metrics.GetPrometheusMetrics()

	// Create and start telemetry client
	telemetryClient := metrics.NewTelemetryClient(cfg)
	wg.Add(1)
	go func() {
		defer wg.Done()
		logging.Info().
			Str("interval", cfg.Metrics.CollectionInterval.String()).
			Msg("Starting telemetry client")
		if err := telemetryClient.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
			logging.Error().Err(err).Msg("Telemetry client error")
		}
	}()

	// Create and start HTTP server
	server := http.NewServer(cfg)
	wg.Add(1)
	go func() {
		defer wg.Done()
		logging.Info().
			Int("port", cfg.HTTPPort).
			Msg("Starting HTTP server")
		if err := server.Start(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				logging.Error().Err(err).Msg("HTTP server error")
			}
		}
	}()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start recording uptime
	uptimeTicker := time.NewTicker(time.Second)
	defer uptimeTicker.Stop()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-uptimeTicker.C:
				promMetrics.RecordUptime(1)
			}
		}
	}()

	// Wait for termination signal
	sig := <-sigChan
	logging.Info().
		Str("signal", sig.String()).
		Msg("Initiating graceful shutdown")

	// Cancel context to stop all components
	cancel()

	// Create a timeout context for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server gracefully
	if err := server.Shutdown(shutdownCtx); err != nil {
		logging.Error().Err(err).Msg("Error during HTTP server shutdown")
	}

	// Wait for all goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logging.Info().Msg("All components shut down successfully")
	case <-shutdownCtx.Done():
		logging.Warn().Msg("Shutdown timed out, forcing exit")
	}

	logging.Info().Msg("Talis Agent stopped")
}
