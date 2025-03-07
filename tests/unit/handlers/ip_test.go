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

func TestIPHandler(t *testing.T) {
	// Create a new Fiber app
	app := fiber.New()

	// Register the IP handler
	app.Get("/ip", handlers.IPHandler)

	// Create a test request
	req := httptest.NewRequest("GET", "/ip", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Read response body
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	// Parse response
	var result map[string][]string
	err = json.Unmarshal(body, &result)
	assert.NoError(t, err)

	// Verify response structure
	assert.Contains(t, result, "ips")
	assert.IsType(t, []string{}, result["ips"])
}
