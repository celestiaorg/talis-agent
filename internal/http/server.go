package http

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/celestiaorg/talis-agent/internal/config"
	"github.com/celestiaorg/talis-agent/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server represents the HTTP server
type Server struct {
	config *config.Config
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
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.HTTPPort),
		Handler: mux,
	}

	log.Printf("Starting HTTP server on port %d", s.config.HTTPPort)
	return server.ListenAndServe()
}

// handlePayload handles POST requests to /payload
func (s *Server) handlePayload(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify token
	if !s.verifyToken(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(s.config.Payload.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("Failed to create directory %s: %v", dir, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Open file for writing
	file, err := os.Create(s.config.Payload.Path)
	if err != nil {
		log.Printf("Failed to create file: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Copy request body to file and count bytes
	written, err := io.Copy(file, r.Body)
	if err != nil {
		log.Printf("Failed to write payload: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Record metrics
	metrics.GetPrometheusMetrics().RecordPayloadReceived(written)

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
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify token
	if !s.verifyToken(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req CommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Execute command
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
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode response: %v", err)
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
