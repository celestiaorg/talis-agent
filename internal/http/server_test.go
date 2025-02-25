package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/celestiaorg/talis-agent/internal/config"
)

func TestHandlePayload(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "talis-agent-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	payloadPath := filepath.Join(tmpDir, "payload")

	cfg := &config.Config{
		Token: "test-token",
		Payload: config.PayloadConfig{
			Path: payloadPath,
		},
	}

	server := NewServer(cfg)

	tests := []struct {
		name          string
		method        string
		token         string
		body          string
		expectedCode  int
		expectedError string
	}{
		{
			name:         "Valid request",
			method:       http.MethodPost,
			token:        "Bearer test-token",
			body:         "test payload",
			expectedCode: http.StatusOK,
		},
		{
			name:          "Invalid method",
			method:        http.MethodGet,
			token:         "Bearer test-token",
			expectedCode:  http.StatusMethodNotAllowed,
			expectedError: "Method not allowed\n",
		},
		{
			name:          "Invalid token",
			method:        http.MethodPost,
			token:         "Bearer invalid-token",
			expectedCode:  http.StatusUnauthorized,
			expectedError: "Unauthorized\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/payload", strings.NewReader(tt.body))
			req.Header.Set("Authorization", tt.token)
			w := httptest.NewRecorder()

			server.handlePayload(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedCode, resp.StatusCode)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			if tt.expectedError != "" {
				if string(body) != tt.expectedError {
					t.Errorf("Expected error %q, got %q", tt.expectedError, string(body))
				}
			} else if tt.expectedCode == http.StatusOK {
				// Verify payload was written correctly
				content, err := os.ReadFile(payloadPath)
				if err != nil {
					t.Fatalf("Failed to read payload file: %v", err)
				}
				if string(content) != tt.body {
					t.Errorf("Expected payload %q, got %q", tt.body, string(content))
				}
			}
		})
	}
}

func TestHandleCommands(t *testing.T) {
	cfg := &config.Config{
		Token: "test-token",
	}

	server := NewServer(cfg)

	tests := []struct {
		name           string
		method         string
		token          string
		command        string
		expectedCode   int
		expectedOutput string
		expectError    bool
	}{
		{
			name:           "Valid command",
			method:         http.MethodPost,
			token:          "Bearer test-token",
			command:        "echo 'hello world'",
			expectedCode:   http.StatusOK,
			expectedOutput: "hello world\n",
			expectError:    false,
		},
		{
			name:         "Invalid command",
			method:       http.MethodPost,
			token:        "Bearer test-token",
			command:      "invalid_command",
			expectedCode: http.StatusOK,
			expectError:  true,
		},
		{
			name:         "Invalid method",
			method:       http.MethodGet,
			token:        "Bearer test-token",
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "Invalid token",
			method:       http.MethodPost,
			token:        "Bearer invalid-token",
			command:      "echo 'test'",
			expectedCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody []byte
			if tt.command != "" {
				reqBody, _ = json.Marshal(CommandRequest{Command: tt.command})
			}

			req := httptest.NewRequest(tt.method, "/commands", bytes.NewBuffer(reqBody))
			req.Header.Set("Authorization", tt.token)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.handleCommands(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedCode, resp.StatusCode)
			}

			if tt.expectedCode == http.StatusOK {
				var cmdResp CommandResponse
				if err := json.NewDecoder(resp.Body).Decode(&cmdResp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if tt.expectError {
					if cmdResp.Error == "" {
						t.Error("Expected error in response, got none")
					}
				} else {
					if cmdResp.Output != tt.expectedOutput {
						t.Errorf("Expected output %q, got %q", tt.expectedOutput, cmdResp.Output)
					}
					if cmdResp.Error != "" {
						t.Errorf("Unexpected error: %s", cmdResp.Error)
					}
				}
			}
		})
	}
}

func TestNewServer(t *testing.T) {
	cfg := &config.Config{
		HTTPPort: 25550,
		Token:    "test-token",
	}

	server := NewServer(cfg)

	if server == nil {
		t.Fatal("Expected non-nil server")
	}

	if server.config != cfg {
		t.Error("Config not properly set")
	}
}
