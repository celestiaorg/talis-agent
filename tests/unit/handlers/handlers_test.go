package handlers_test

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/talis-agent/internal/handlers"
	"github.com/celestiaorg/talis-agent/internal/metrics"
)

func setupTestApp(t *testing.T) (*fiber.App, *handlers.Handler, string) {
	app := fiber.New()
	collector := metrics.NewCollector(15 * time.Second)
	prometheus.MustRegister(collector)

	// Create temp dir for payload tests
	tmpDir, err := os.MkdirTemp("", "talis-test-*")
	require.NoError(t, err)

	// Set environment variable for payload path
	os.Setenv("TALIS_PAYLOAD_DIR", tmpDir)

	h := handlers.NewHandler(collector)

	// Cleanup after test
	t.Cleanup(func() {
		prometheus.Unregister(collector)
		os.RemoveAll(tmpDir)
		os.Unsetenv("TALIS_PAYLOAD_DIR")
	})

	return app, h, tmpDir
}

func TestEndpoints(t *testing.T) {
	app, h, _ := setupTestApp(t)
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
	app, h, _ := setupTestApp(t)
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
	app, h, _ := setupTestApp(t)
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
	app, h, _ := setupTestApp(t)
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

func TestHandlePayload(t *testing.T) {
	app, h, tmpDir := setupTestApp(t)
	app.Post("/payload", h.HandlePayload)

	// Test empty payload
	req := httptest.NewRequest("POST", "/payload", nil)
	resp, err := app.Test(req)
	require.NoError(t, err, "Failed to execute request")
	require.Equal(t, 400, resp.StatusCode, "Expected status code 400")

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")

	var errorResult map[string]string
	require.NoError(t, json.Unmarshal(body, &errorResult), "Failed to unmarshal error response")
	require.Equal(t, "empty payload", errorResult["error"], "Expected empty payload error")

	// Test valid payload
	testPayload := "test payload"
	req = httptest.NewRequest("POST", "/payload", strings.NewReader(testPayload))
	resp, err = app.Test(req)
	require.NoError(t, err, "Failed to execute request")
	require.Equal(t, 200, resp.StatusCode, "Expected status code 200")

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")

	var successResult map[string]string
	require.NoError(t, json.Unmarshal(body, &successResult), "Failed to unmarshal success response")
	require.Equal(t, "payload stored successfully", successResult["status"], "Expected success status")

	// Verify the payload was actually written
	payloadPath := filepath.Join(tmpDir, "payload")
	content, err := os.ReadFile(payloadPath)
	require.NoError(t, err, "Failed to read written payload file")
	require.Equal(t, testPayload, string(content), "Written payload doesn't match input")
}

func TestExecuteCommand(t *testing.T) {
	app, h, _ := setupTestApp(t)
	app.Post("/commands", h.ExecuteCommand)

	// Test empty command
	req := httptest.NewRequest("POST", "/commands", nil)
	resp, err := app.Test(req)
	require.NoError(t, err, "Failed to execute request")
	require.Equal(t, 400, resp.StatusCode, "Expected status code 400")

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")

	var errorResult map[string]string
	require.NoError(t, json.Unmarshal(body, &errorResult), "Failed to unmarshal error response")
	require.Equal(t, "empty command", errorResult["error"], "Expected empty command error")

	// Test valid command
	req = httptest.NewRequest("POST", "/commands", strings.NewReader("echo test"))
	resp, err = app.Test(req)
	require.NoError(t, err, "Failed to execute request")
	require.Equal(t, 200, resp.StatusCode, "Expected status code 200")

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")

	var successResult map[string]string
	require.NoError(t, json.Unmarshal(body, &successResult), "Failed to unmarshal success response")
	require.Equal(t, "test\n", successResult["output"], "Command output doesn't match expected")
}
