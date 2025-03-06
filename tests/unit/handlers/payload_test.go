package handlers_test

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"talis-agent/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestPayloadHandler(t *testing.T) {
	// Create a temporary directory for payloads
	tmpDir, err := os.MkdirTemp("", "payload-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Temporarily override the payload directory
	originalPayloadDir := handlers.PayloadDir
	handlers.PayloadDir = tmpDir
	defer func() {
		handlers.PayloadDir = originalPayloadDir
	}()

	// Create a new Fiber app
	app := fiber.New()

	// Register the payload handler
	app.Post("/payload", handlers.PayloadHandler)

	// Test cases
	tests := []struct {
		name           string
		payload        string
		expectedStatus int
		checkFile      bool
	}{
		{
			name:           "Valid payload",
			payload:        "test payload",
			expectedStatus: 200,
			checkFile:      true,
		},
		{
			name:           "Empty payload",
			payload:        "",
			expectedStatus: 400,
			checkFile:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("POST", "/payload", strings.NewReader(tt.payload))
			req.Header.Set("Content-Type", "text/plain")

			// Send request
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.checkFile {
				// Read response
				body, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				// Parse response
				var result map[string]string
				err = json.Unmarshal(body, &result)
				assert.NoError(t, err)

				// Verify response structure
				assert.Contains(t, result, "status")
				assert.Contains(t, result, "file")
				assert.Equal(t, "success", result["status"])

				// Verify file was created
				filePath := filepath.Join(tmpDir, result["file"])
				_, err = os.Stat(filePath)
				assert.NoError(t, err)

				// Verify file contents
				content, err := os.ReadFile(filePath)
				assert.NoError(t, err)
				assert.Equal(t, tt.payload, string(content))
			}
		})
	}
}
