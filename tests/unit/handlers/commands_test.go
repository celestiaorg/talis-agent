package handlers_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/celestiaorg/talis-agent/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestCommandsHandler(t *testing.T) {
	// Create a new Fiber app
	app := fiber.New()

	// Register the commands handler
	app.Post("/commands", handlers.CommandsHandler)

	// Test cases
	tests := []struct {
		name           string
		command        string
		expectedStatus int
		checkOutput    bool
	}{
		{
			name:           "Valid command",
			command:        "echo 'test'",
			expectedStatus: 200,
			checkOutput:    true,
		},
		{
			name:           "Empty command",
			command:        "",
			expectedStatus: 400,
			checkOutput:    false,
		},
		{
			name:           "Invalid command",
			command:        "nonexistent_command",
			expectedStatus: 500,
			checkOutput:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request body
			reqBody := map[string]string{
				"command": tt.command,
			}
			jsonBody, err := json.Marshal(reqBody)
			assert.NoError(t, err)

			// Create request
			req := httptest.NewRequest("POST", "/commands", bytes.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			// Send request
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.checkOutput {
				// Read response
				body, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				// Parse response
				var result map[string]string
				err = json.Unmarshal(body, &result)
				assert.NoError(t, err)

				// Verify response structure
				if tt.expectedStatus == 200 {
					assert.Contains(t, result, "status")
					assert.Contains(t, result, "stdout")
					assert.Contains(t, result, "stderr")
					assert.Equal(t, "success", result["status"])
					assert.Equal(t, "test\n", result["stdout"])
				} else {
					assert.Contains(t, result, "error")
					assert.Contains(t, result, "stderr")
				}
			}
		})
	}
}
