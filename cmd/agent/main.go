package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/celestiaorg/talis-agent/internal/config"
	"github.com/celestiaorg/talis-agent/internal/http"
	"github.com/celestiaorg/talis-agent/internal/metrics"
)

var (
	configPath = flag.String("config", "config.yaml", "path to configuration file")
)

func main() {
	flag.Parse()

	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting Talis Agent with configuration from: %s", *configPath)
	log.Printf("API Server: %s", cfg.APIServer)
	log.Printf("HTTP Port: %d", cfg.HTTPPort)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create wait group for goroutines
	var wg sync.WaitGroup

	// Create and start telemetry client
	telemetryClient := metrics.NewTelemetryClient(cfg)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := telemetryClient.Start(ctx); err != nil && err != context.Canceled {
			log.Printf("Telemetry client error: %v", err)
		}
	}()

	// Create and start HTTP server
	server := http.NewServer(cfg)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := server.Start(); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for termination signal
	sig := <-sigChan
	log.Printf("Received signal %v, shutting down...", sig)

	// Cancel context to stop telemetry client
	cancel()

	// Wait for goroutines to finish
	wg.Wait()
	log.Println("Shutdown complete")
}
