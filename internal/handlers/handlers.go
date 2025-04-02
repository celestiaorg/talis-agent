package handlers

import (
	"net"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"

	"github.com/celestiaorg/talis-agent/internal/metrics"
)

// Handler handles HTTP requests
type Handler struct {
	collector *metrics.Collector
}

// NewHandler creates a new Handler
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

// Endpoints returns a list of available endpoints
func (h *Handler) Endpoints(c *fiber.Ctx) error {
	endpoints := []string{
		"/metrics",
		"/alive",
		"/ip",
	}

	return c.JSON(fiber.Map{
		"endpoints": endpoints,
	})
}
