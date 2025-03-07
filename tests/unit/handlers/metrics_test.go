package handlers_test

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/celestiaorg/talis-agent/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestMetricsHandler(t *testing.T) {
	// Create a new metrics handler
	metricsHandler := handlers.NewMetricsHandler()

	// Create a new Fiber app
	app := fiber.New()

	// Register the metrics handler
	app.Get("/metrics", metricsHandler.Handle)

	// Create a test request
	req := httptest.NewRequest("GET", "/metrics", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Read response body
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	// Parse response
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	assert.NoError(t, err)

	// Verify response contains metrics
	assert.Contains(t, result, "metrics")
	metrics := result["metrics"].([]interface{})
	assert.Greater(t, len(metrics), 0)

	// Verify metric names
	metricNames := make(map[string]bool)
	for _, metric := range metrics {
		metricMap := metric.(map[string]interface{})
		name := metricMap["name"].(string)
		metricNames[name] = true
	}

	assert.True(t, metricNames["system_cpu_usage_percent"])
	assert.True(t, metricNames["system_memory_usage_bytes"])
	assert.True(t, metricNames["go_goroutines"])
}
