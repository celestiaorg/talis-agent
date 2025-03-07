package handlers

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"syscall"

	"github.com/gofiber/fiber/v2"
)

// CommandRequest represents a command execution request
type CommandRequest struct {
	Command string `json:"command"`
}

// CommandsHandler handles command execution requests
func CommandsHandler(c *fiber.Ctx) error {
	var req CommandRequest
	fmt.Println("Received command request:", req)
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate command input
	if req.Command == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Command cannot be empty",
		})
	}

	// Sanitize command input
	sanitizedCmd := strings.TrimSpace(req.Command)
	if len(sanitizedCmd) > 1000 { // Limit command length
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Command too long",
		})
	}

	// Check for potentially dangerous commands
	if strings.ContainsAny(sanitizedCmd, "&|;`$<>") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Command contains potentially dangerous characters",
		})
	}

	// Execute command based on input
	var cmd *exec.Cmd
	switch sanitizedCmd {
	case "ls":
		cmd = exec.Command("ls")
	case "ps":
		cmd = exec.Command("ps")
	case "df":
		cmd = exec.Command("df")
	case "free":
		cmd = exec.Command("free")
	case "uptime":
		cmd = exec.Command("uptime")
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Command not allowed",
		})
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
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
