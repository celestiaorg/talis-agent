package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/celestiaorg/talis-agent/internal/config"
	"github.com/celestiaorg/talis-agent/internal/logging"
)

// ErrServerClosed is returned by the Server's Start method after a call to Shutdown
var ErrServerClosed = errors.New("http: Server closed")

// Server represents the HTTP server
type Server struct {
	config *config.Config
	srv    *http.Server
}

// NewServer creates a new HTTP server
func NewServer(config *config.Config) *Server {
	return &Server{
		config: config,
	}
}

// Address returns the server's address
func (s *Server) Address() string {
	return fmt.Sprintf("%s:%d", s.config.HTTP.Host, s.config.HTTP.Port)
}

// Start starts the HTTP server
func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("/alive", s.handleHealthCheck)
	mux.HandleFunc("/metrics", s.handleMetrics)
	mux.HandleFunc("/ip", s.handleIP)

	s.srv = &http.Server{
		Addr:              s.Address(),
		Handler:           mux,
		ReadHeaderTimeout: 20 * time.Second, // Prevent slow HTTP header attacks
		ReadTimeout:       1 * time.Minute,  // Maximum duration for reading entire request
		WriteTimeout:      2 * time.Minute,  // Maximum duration for writing response
		IdleTimeout:       2 * time.Minute,  // Maximum duration to wait for the next request
		MaxHeaderBytes:    1 << 20,          // Maximum size of request headers (1MB)
	}

	logging.Info().Str("address", s.Address()).Msg("Starting HTTP server")

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
		// Create a timeout context for shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Attempt graceful shutdown
		if err := s.srv.Shutdown(shutdownCtx); err != nil {
			// If shutdown times out, force close
			logging.Error().Err(err).Msg("Server shutdown timed out, forcing close")
			return s.srv.Close()
		}
		return nil
	}
}

func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		logging.Error().Err(err).Msg("Failed to encode response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	promhttp.Handler().ServeHTTP(w, r)
}

func (s *Server) handleIP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var ips []string
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string][]string{"ips": ips}); err != nil {
		logging.Error().Err(err).Msg("Failed to encode response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
