package handlers

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/gofiber/fiber/v2"
)

type CommandRequest struct {
	Command string `json:"command"`
}

func CommandsHandler(c *fiber.Ctx) error {
	var req CommandRequest
	fmt.Println("Received command request:", req)
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Command == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No command provided",
		})
	}

	// Execute the command
	cmd := exec.Command("bash", "-c", req.Command)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  fmt.Sprintf("Command failed: %v", err),
			"stderr": stderr.String(),
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"stdout": stdout.String(),
		"stderr": stderr.String(),
	})
}
