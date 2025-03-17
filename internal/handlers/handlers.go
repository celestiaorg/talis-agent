package handlers

import (
	"io/ioutil"
	"net"
	"os/exec"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"

	"github.com/celestiaorg/talis-agent/internal/metrics"
)

// Handler contains dependencies for HTTP handlers
type Handler struct {
	collector *metrics.Collector
}

// NewHandler creates a new Handler instance
func NewHandler(collector *metrics.Collector) *Handler {
	return &Handler{
		collector: collector,
	}
}

// HealthCheck handles the /alive endpoint
func (h *Handler) HealthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
	})
}

// GetMetrics handles the /metrics endpoint
func (h *Handler) GetMetrics(c *fiber.Ctx) error {
	// Convert promhttp.Handler to fasthttp handler
	handler := fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())
	handler(c.Context())
	return nil
}

// GetIP handles the /ip endpoint
func (h *Handler) GetIP(c *fiber.Ctx) error {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var ips []string
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}

	return c.JSON(fiber.Map{
		"ips": ips,
	})
}

// HandlePayload handles the /payload endpoint
func (h *Handler) HandlePayload(c *fiber.Ctx) error {
	payload := c.Body()
	if len(payload) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "empty payload",
		})
	}

	payloadPath := filepath.Join("/etc/talis-agent", "payload")
	if err := ioutil.WriteFile(payloadPath, payload, 0644); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status": "payload stored successfully",
	})
}

// ExecuteCommand handles the /commands endpoint
func (h *Handler) ExecuteCommand(c *fiber.Ctx) error {
	command := string(c.Body())
	if command == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "empty command",
		})
	}

	cmd := exec.Command("bash", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  err.Error(),
			"output": string(output),
		})
	}

	return c.JSON(fiber.Map{
		"output": string(output),
	})
}
