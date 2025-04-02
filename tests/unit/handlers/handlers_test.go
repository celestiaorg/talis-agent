package handlers_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"

	"github.com/celestiaorg/talis-agent/internal/handlers"
)

func setupTestApp() (*fiber.App, *handlers.Handler) {
	app := fiber.New()
	collector := prometheus.NewRegistry()
	h := handlers.NewHandler(collector)
	return app, h
}

func TestEndpoints(t *testing.T) {
	app, h := setupTestApp()
	app.Get("/", h.Endpoints)

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestHealthCheck(t *testing.T) {
	app, h := setupTestApp()
	app.Get("/alive", h.HealthCheck)

	req := httptest.NewRequest("GET", "/alive", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestGetMetrics(t *testing.T) {
	app, h := setupTestApp()
	app.Get("/metrics", h.GetMetrics)

	req := httptest.NewRequest("GET", "/metrics", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestGetIP(t *testing.T) {
	app, h := setupTestApp()
	app.Get("/ip", h.GetIP)

	req := httptest.NewRequest("GET", "/ip", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestHandlePayload(t *testing.T) {
	app, h := setupTestApp()
	app.Post("/payload", h.HandlePayload)

	// Test empty payload
	req := httptest.NewRequest("POST", "/payload", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	// Test valid payload
	req = httptest.NewRequest("POST", "/payload", strings.NewReader("test payload"))
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestExecuteCommand(t *testing.T) {
	app, h := setupTestApp()
	app.Post("/commands", h.ExecuteCommand)

	// Test empty command
	req := httptest.NewRequest("POST", "/commands", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	// Test valid command
	req = httptest.NewRequest("POST", "/commands", strings.NewReader("echo test"))
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}
