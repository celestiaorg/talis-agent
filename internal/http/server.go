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
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

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
func (s *Server) Start() error {
	// Set up routes
	mux := http.NewServeMux()
	mux.HandleFunc("/payload", s.handlePayload)
	mux.HandleFunc("/commands", s.handleCommands)
	mux.Handle("/metrics", promhttp.Handler()) // Add Prometheus metrics endpoint

	// Create server with configured port
	s.srv = &http.Server{
		Addr:              fmt.Sprintf(":%d", s.config.HTTPPort),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second, // Prevent Slowloris attacks
	}

	logging.Info().Int("port", s.config.HTTPPort).Msg("Starting HTTP server")
	if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server error: %w", err)
	}
	return ErrServerClosed
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	if s.srv != nil {
		logging.Info().Msg("Shutting down HTTP server")
		return s.srv.Shutdown(ctx)
	}
	return nil
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

	// Verify token
	if !s.verifyToken(r) {
		logging.Warn().
			Str("path", "/payload").
			Msg("Unauthorized request")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(s.config.Payload.Path)
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
	file, err := os.Create(s.config.Payload.Path)
	if err != nil {
		logging.Error().
			Err(err).
			Str("path", s.config.Payload.Path).
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
			Str("path", s.config.Payload.Path).
			Msg("Failed to write payload")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Record metrics
	metrics.GetPrometheusMetrics().RecordPayloadReceived(written)

	logging.Info().
		Int64("bytes", written).
		Str("path", s.config.Payload.Path).
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

	// Verify token
	if !s.verifyToken(r) {
		logging.Warn().
			Str("path", "/commands").
			Msg("Unauthorized request")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
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
			Str("output", string(output)).
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

// verifyToken verifies the authorization token in the request
func (s *Server) verifyToken(r *http.Request) bool {
	auth := r.Header.Get("Authorization")
	expectedToken := fmt.Sprintf("Bearer %s", s.config.Token)
	return auth == expectedToken
}
