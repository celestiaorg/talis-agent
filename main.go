package main

import (
	"fmt"
	"log"
	"time"

	"talis-agent/config"
	"talis-agent/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "Talis Agent",
	})

	// Add middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path}\n",
	}))

	// Create handlers
	metricsHandler := handlers.NewMetricsHandler()

	// Register routes
	app.Get("/alive", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	app.Get("/metrics", metricsHandler.Handle)
	app.Get("/ip", handlers.IPHandler)
	app.Post("/payload", handlers.PayloadHandler)
	app.Post("/commands", handlers.CommandsHandler)

	// Start server
	addr := fmt.Sprintf(":%d", cfg.HTTP.Port)
	log.Printf("Starting server on %s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
