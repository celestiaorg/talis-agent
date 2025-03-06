package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
)

// PayloadDir is the directory where payloads are stored
var PayloadDir = "/etc/talis-agent/payload"

func PayloadHandler(c *fiber.Ctx) error {
	// Get the payload data
	payload := c.Body()
	if len(payload) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No payload provided",
		})
	}

	// Create a filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("payload_%s.txt", timestamp)
	filepath := filepath.Join(PayloadDir, filename)

	// Write the payload to file
	if err := os.WriteFile(filepath, payload, 0644); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to write payload: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"file":   filename,
	})
}
