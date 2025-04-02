package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/celestiaorg/talis-agent/internal/config"
	"github.com/celestiaorg/talis-agent/internal/logging"
	"github.com/celestiaorg/talis-agent/internal/metrics"
)

// ErrServerClosed is returned by the Server's Start method after a call to Shutdown
var ErrServerClosed = errors.New("http: Server closed")

// Server represents the HTTP server
type Server struct {
	config *config.Config
	srv    *http.Server
}

// NewServer creates a new HTTP server
func NewServer(cfg *config.Config) *Server {
	return &Server{
		config: cfg,
	}
}

// Start starts the HTTP server
func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.config.HTTP.Host, s.config.HTTP.Port)
	s.srv = &http.Server{
		Addr: addr,
	}

	logging.Info().Str("address", addr).Msg("Starting HTTP server")

	errCh := make(chan error, 1)
	go func() {
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("server error: %w", err)
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return s.srv.Shutdown(context.Background())
	}
}

// Address returns the server's address
func (s *Server) Address() string {
	return fmt.Sprintf("%s:%d", s.config.HTTP.Host, s.config.HTTP.Port)
}

// handlePayload handles POST requests to /payload
func (s *Server) handlePayload(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		logging.Warn().
			Str("method", r.Method).
			Str("path", "/payload").
			Msg("Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get payload path from environment variable or use default
	payloadPath := os.Getenv("TALIS_PAYLOAD_PATH")
	if payloadPath == "" {
		payloadPath = "/etc/talis-agent/payload"
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(payloadPath)
	// #nosec G301 -- This directory needs to be readable by other processes
	if err := os.MkdirAll(dir, 0750); err != nil {
		logging.Error().
			Err(err).
			Str("directory", dir).
			Msg("Failed to create directory")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Open file for writing
	file, err := os.Create(payloadPath)
	if err != nil {
		logging.Error().
			Err(err).
			Str("path", payloadPath).
			Msg("Failed to create file")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			logging.Error().
				Err(err).
				Msg("Failed to close file")
		}
	}()

	// Copy request body to file and count bytes
	written, err := io.Copy(file, r.Body)
	if err != nil {
		logging.Error().
			Err(err).
			Str("path", payloadPath).
			Msg("Failed to write payload")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Record metrics
	metrics.GetPrometheusMetrics().RecordPayloadReceived(written)

	logging.Info().
		Int64("bytes", written).
		Str("path", payloadPath).
		Msg("Payload received and written")

	w.WriteHeader(http.StatusOK)
}

// CommandRequest represents a command execution request
type CommandRequest struct {
	Command string `json:"command"`
}

// CommandResponse represents a command execution response
type CommandResponse struct {
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

// handleCommands handles POST requests to /commands
func (s *Server) handleCommands(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		logging.Warn().
			Str("method", r.Method).
			Str("path", "/commands").
			Msg("Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req CommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logging.Error().
			Err(err).
			Str("path", "/commands").
			Msg("Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Execute command
	logging.Debug().
		Str("command", req.Command).
		Msg("Executing command")

	// #nosec G204 -- Command execution is a core feature of this endpoint
	cmd := exec.Command("bash", "-c", req.Command)
	output, err := cmd.CombinedOutput()

	// Record metrics
	metrics.GetPrometheusMetrics().RecordCommandExecution(err == nil)

	// Prepare response
	resp := CommandResponse{
		Output: string(output),
	}
	if err != nil {
		resp.Error = err.Error()
		logging.Error().
			Err(err).
			Str("command", req.Command).
			Msg("Command execution failed")
	} else {
		logging.Info().
			Str("command", req.Command).
			Msg("Command executed successfully")
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logging.Error().
			Err(err).
			Msg("Failed to encode response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
