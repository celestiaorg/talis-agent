package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/celestiaorg/talis-agent/internal/config"
	"github.com/celestiaorg/talis-agent/internal/handlers"
	"github.com/celestiaorg/talis-agent/internal/metrics"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Parse metrics collection interval
	interval, err := time.ParseDuration(cfg.Metrics.CollectionInterval)
	if err != nil {
		interval = 15 * time.Second // Default interval
		log.Printf("Using default metrics collection interval: %v", interval)
	}

	// Initialize metrics collector
	collector := metrics.NewCollector(interval)

	// Register collector with Prometheus
	prometheus.MustRegister(collector)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "Talis Agent",
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	})

	// Add middleware
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New())

	// Initialize handlers
	h := handlers.NewHandler(collector)

	// Setup routes
	setupRoutes(app, h)

	// Start server in a goroutine
	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
		log.Printf("Starting server on %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Unregister metrics collector
	prometheus.Unregister(collector)

	if err := app.Shutdown(); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

	log.Println("Server gracefully stopped")
}

func setupRoutes(app *fiber.App, h *handlers.Handler) {
	// Get the commands info
	app.Get("/", h.Endpoints)

	// Health check endpoint
	app.Get("/alive", h.HealthCheck)

	// Metrics endpoint
	app.Get("/metrics", h.GetMetrics)

	// IP endpoint
	app.Get("/ip", h.GetIP)

	// Payload endpoint
	app.Post("/payload", h.HandlePayload)

	// Commands endpoint
	app.Post("/commands", h.ExecuteCommand)
}
