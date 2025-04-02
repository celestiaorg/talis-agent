package handlers_test

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/talis-agent/internal/handlers"
	"github.com/celestiaorg/talis-agent/internal/metrics"
)

func setupTestApp(t *testing.T) (*fiber.App, *handlers.Handler) {
	app := fiber.New()
	collector := metrics.NewCollector(15 * time.Second)
	prometheus.MustRegister(collector)

	h := handlers.NewHandler(collector)

	return app, h
}

func TestEndpoints(t *testing.T) {
	app, h := setupTestApp(t)
	app.Get("/", h.Endpoints)

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err, "Failed to execute request")
	require.Equal(t, 200, resp.StatusCode, "Expected status code 200")

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")

	var result map[string][]string
	require.NoError(t, json.Unmarshal(body, &result), "Failed to unmarshal response")
	require.Contains(t, result, "endpoints", "Response missing endpoints key")
	require.Contains(t, result["endpoints"], "/metrics", "Response missing /metrics endpoint")
}

func TestHealthCheck(t *testing.T) {
	app, h := setupTestApp(t)
	app.Get("/alive", h.HealthCheck)

	req := httptest.NewRequest("GET", "/alive", nil)
	resp, err := app.Test(req)
	require.NoError(t, err, "Failed to execute request")
	require.Equal(t, 200, resp.StatusCode, "Expected status code 200")

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")

	var result map[string]string
	require.NoError(t, json.Unmarshal(body, &result), "Failed to unmarshal response")
	require.Equal(t, "ok", result["status"], "Expected status to be 'ok'")
}

func TestGetMetrics(t *testing.T) {
	app, h := setupTestApp(t)
	app.Get("/metrics", h.GetMetrics)

	req := httptest.NewRequest("GET", "/metrics", nil)
	resp, err := app.Test(req)
	require.NoError(t, err, "Failed to execute request")
	require.Equal(t, 200, resp.StatusCode, "Expected status code 200")

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")
	require.Contains(t, string(body), "system_", "Response missing system metrics")
}

func TestGetIP(t *testing.T) {
	app, h := setupTestApp(t)
	app.Get("/ip", h.GetIP)

	req := httptest.NewRequest("GET", "/ip", nil)
	resp, err := app.Test(req)
	require.NoError(t, err, "Failed to execute request")
	require.Equal(t, 200, resp.StatusCode, "Expected status code 200")

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")

	var result map[string][]string
	require.NoError(t, json.Unmarshal(body, &result), "Failed to unmarshal response")
	require.Contains(t, result, "ips", "Response missing ips key")
	require.NotEmpty(t, result["ips"], "Expected non-empty IPs list")
}
