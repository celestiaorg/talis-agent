package logging

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestLoggerInitialization(t *testing.T) {
	// Create temporary directory for test logs
	tmpDir, err := os.MkdirTemp("", "talis-agent-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	logPath := filepath.Join(tmpDir, "test.log")

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "console only",
			config: Config{
				Level:   "info",
				Console: true,
			},
			wantErr: false,
		},
		{
			name: "file only",
			config: Config{
				Level: "info",
				File: &FileConfig{
					Path:       logPath,
					MaxSize:    1,
					MaxBackups: 1,
					MaxAge:     1,
					Compress:   false,
				},
			},
			wantErr: false,
		},
		{
			name: "both console and file",
			config: Config{
				Level:   "info",
				Console: true,
				File: &FileConfig{
					Path:       logPath,
					MaxSize:    1,
					MaxBackups: 1,
					MaxAge:     1,
					Compress:   false,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid directory",
			config: Config{
				Level: "info",
				File: &FileConfig{
					Path:       "/nonexistent/directory/test.log",
					MaxSize:    1,
					MaxBackups: 1,
					MaxAge:     1,
					Compress:   false,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InitLogger(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("InitLogger() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLogLevels(t *testing.T) {
	// Create temporary directory for test logs
	tmpDir, err := os.MkdirTemp("", "talis-agent-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	logPath := filepath.Join(tmpDir, "test.log")

	// Initialize logger with file output only
	config := Config{
		Level:   "debug",
		Console: false,
		File: &FileConfig{
			Path:       logPath,
			MaxSize:    1,
			MaxBackups: 1,
			MaxAge:     1,
			Compress:   false,
		},
	}

	if err := InitLogger(config); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Test different log levels
	Debug().Msg("debug message")
	Info().Msg("info message")
	Warn().Msg("warn message")
	Error().Msg("error message")

	// Read and verify log file
	time.Sleep(100 * time.Millisecond) // Wait for logs to be written
	logs, err := readLogFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Skip initialization message
	if len(logs) < 5 { // 1 init message + 4 test messages
		t.Errorf("Expected at least 5 log entries, got %d", len(logs))
		return
	}

	// Find test messages (skip initialization message)
	var testLogs []logEntry
	for _, entry := range logs {
		if strings.Contains(entry.Message, "message") {
			testLogs = append(testLogs, entry)
		}
	}

	expectedLogs := []struct {
		level   string
		message string
	}{
		{"debug", "debug message"},
		{"info", "info message"},
		{"warn", "warn message"},
		{"error", "error message"},
	}

	if len(testLogs) != len(expectedLogs) {
		t.Errorf("Expected %d test log entries, got %d", len(expectedLogs), len(testLogs))
		return
	}

	for i, expected := range expectedLogs {
		if testLogs[i].Level != expected.level {
			t.Errorf("Log entry %d: expected level %s, got %s", i, expected.level, testLogs[i].Level)
		}
		if testLogs[i].Message != expected.message {
			t.Errorf("Log entry %d: expected message %q, got %q", i, expected.message, testLogs[i].Message)
		}
	}
}

func TestLogRotation(t *testing.T) {
	// Create temporary directory for test logs
	tmpDir, err := os.MkdirTemp("", "talis-agent-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	logPath := filepath.Join(tmpDir, "test.log")

	// Initialize logger with small max size to test rotation
	config := Config{
		Level:   "info",
		Console: false,
		File: &FileConfig{
			Path:       logPath,
			MaxSize:    1, // 1 MB
			MaxBackups: 2,
			MaxAge:     1,
			Compress:   false,
		},
	}

	if err := InitLogger(config); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Write enough logs to trigger rotation
	message := strings.Repeat("a", 512) // 512 bytes
	for i := 0; i < 2048; i++ {         // Should write ~1MB of logs
		Info().Msg(message)
	}

	// Wait for logs to be written and rotated
	time.Sleep(100 * time.Millisecond)

	// Check if log files exist
	files, err := filepath.Glob(logPath + "*")
	if err != nil {
		t.Fatalf("Failed to list log files: %v", err)
	}

	// Should have main log file and up to MaxBackups backup files
	expectedFiles := config.File.MaxBackups + 1
	if len(files) > expectedFiles {
		t.Errorf("Expected at most %d log files, got %d", expectedFiles, len(files))
	}
}

func TestLogLevelFiltering(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "talis-agent-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name          string
		configLevel   string
		logMessages   map[string]string // level -> message
		expectedCount int
		unexpectedMsg []string
	}{
		{
			name:        "info level filtering",
			configLevel: "info",
			logMessages: map[string]string{
				"debug": "debug test",
				"info":  "info test",
				"warn":  "warn test",
				"error": "error test",
			},
			expectedCount: 3, // info, warn, error (excluding initialization message)
			unexpectedMsg: []string{"debug test"},
		},
		{
			name:        "warn level filtering",
			configLevel: "warn",
			logMessages: map[string]string{
				"debug": "debug test",
				"info":  "info test",
				"warn":  "warn test",
				"error": "error test",
			},
			expectedCount: 2, // warn, error (excluding initialization message)
			unexpectedMsg: []string{"debug test", "info test"},
		},
		{
			name:        "error level filtering",
			configLevel: "error",
			logMessages: map[string]string{
				"debug": "debug test",
				"info":  "info test",
				"warn":  "warn test",
				"error": "error test",
			},
			expectedCount: 1, // error only (excluding initialization message)
			unexpectedMsg: []string{"debug test", "info test", "warn test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logPath := filepath.Join(tmpDir, tt.name+".log")
			config := Config{
				Level:   tt.configLevel,
				Console: false,
				File: &FileConfig{
					Path:       logPath,
					MaxSize:    1,
					MaxBackups: 1,
					MaxAge:     1,
					Compress:   false,
				},
			}

			if err := InitLogger(config); err != nil {
				t.Fatalf("Failed to initialize logger: %v", err)
			}

			// Write logs at different levels
			for level, msg := range tt.logMessages {
				switch level {
				case "debug":
					Debug().Msg(msg)
				case "info":
					Info().Msg(msg)
				case "warn":
					Warn().Msg(msg)
				case "error":
					Error().Msg(msg)
				}
			}

			time.Sleep(100 * time.Millisecond)
			logs, err := readLogFile(logPath)
			if err != nil {
				t.Fatalf("Failed to read log file: %v", err)
			}

			// Count non-initialization messages
			var count int
			for _, entry := range logs {
				if !strings.Contains(entry.Message, "Logger initialized") {
					count++
					// Verify that unexpected messages are not present
					for _, unexpected := range tt.unexpectedMsg {
						if entry.Message == unexpected {
							t.Errorf("Found unexpected message %q in log file", unexpected)
						}
					}
				}
			}

			if count != tt.expectedCount {
				t.Errorf("Expected %d log entries, got %d", tt.expectedCount, count)
			}
		})
	}
}

func TestStructuredLogging(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "talis-agent-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	logPath := filepath.Join(tmpDir, "structured.log")
	config := Config{
		Level:   "debug",
		Console: false,
		File: &FileConfig{
			Path:       logPath,
			MaxSize:    1,
			MaxBackups: 1,
			MaxAge:     1,
			Compress:   false,
		},
	}

	if err := InitLogger(config); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Test structured logging with different field types
	Info().
		Str("string", "test").
		Int("integer", 42).
		Float64("float", 3.14).
		Bool("boolean", true).
		Msg("structured log test")

	time.Sleep(100 * time.Millisecond)
	logs, err := readStructuredLogFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Find the test log entry
	var testEntry structuredLogEntry
	for _, entry := range logs {
		if entry.Message == "structured log test" {
			testEntry = entry
			break
		}
	}

	// Verify structured fields
	if testEntry.String != "test" {
		t.Errorf("Expected string field 'test', got %q", testEntry.String)
	}
	if testEntry.Integer != 42 {
		t.Errorf("Expected integer field 42, got %d", testEntry.Integer)
	}
	if testEntry.Float != 3.14 {
		t.Errorf("Expected float field 3.14, got %f", testEntry.Float)
	}
	if !testEntry.Boolean {
		t.Error("Expected boolean field to be true")
	}
}

func TestConcurrentLogging(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "talis-agent-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	logPath := filepath.Join(tmpDir, "concurrent.log")
	config := Config{
		Level:   "info",
		Console: false,
		File: &FileConfig{
			Path:       logPath,
			MaxSize:    1,
			MaxBackups: 1,
			MaxAge:     1,
			Compress:   false,
		},
	}

	if err := InitLogger(config); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Test concurrent logging
	var wg sync.WaitGroup
	numGoroutines := 10
	messagesPerGoroutine := 100

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				Info().
					Int("goroutine", id).
					Int("iteration", j).
					Msgf("concurrent test from goroutine %d message %d", id, j)
			}
		}(i)
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	logs, err := readLogFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Count concurrent test messages
	var count int
	for _, entry := range logs {
		if strings.Contains(entry.Message, "concurrent test from goroutine") {
			count++
		}
	}

	expected := numGoroutines * messagesPerGoroutine
	if count != expected {
		t.Errorf("Expected %d concurrent log entries, got %d", expected, count)
	}
}

func TestLogFilePermissions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "talis-agent-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	logPath := filepath.Join(tmpDir, "permissions.log")
	config := Config{
		Level:   "info",
		Console: false,
		File: &FileConfig{
			Path:       logPath,
			MaxSize:    1,
			MaxBackups: 1,
			MaxAge:     1,
			Compress:   false,
		},
	}

	if err := InitLogger(config); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	Info().Msg("test message")
	time.Sleep(100 * time.Millisecond)

	// Check file permissions
	info, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("Failed to stat log file: %v", err)
	}

	mode := info.Mode()
	if mode&0077 != 0 {
		t.Errorf("Log file has incorrect permissions: %v", mode)
	}
}

type logEntry struct {
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"time"`
}

func readLogFile(path string) ([]logEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []logEntry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var entry logEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, scanner.Err()
}

type structuredLogEntry struct {
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"time"`
	String    string    `json:"string"`
	Integer   int       `json:"integer"`
	Float     float64   `json:"float"`
	Boolean   bool      `json:"boolean"`
}

func readStructuredLogFile(path string) ([]structuredLogEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []structuredLogEntry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var entry structuredLogEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, scanner.Err()
}

func BenchmarkLogging(b *testing.B) {
	// Create temporary directory for benchmark logs
	tmpDir, err := os.MkdirTemp("", "talis-agent-bench-*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	benchmarks := []struct {
		name    string
		config  Config
		logFunc func()
	}{
		{
			name: "console only simple message",
			config: Config{
				Level:   "info",
				Console: true,
			},
			logFunc: func() {
				Info().Msg("simple log message")
			},
		},
		{
			name: "file only simple message",
			config: Config{
				Level: "info",
				File: &FileConfig{
					Path:       filepath.Join(tmpDir, "file_simple.log"),
					MaxSize:    100,
					MaxBackups: 1,
					MaxAge:     1,
					Compress:   false,
				},
			},
			logFunc: func() {
				Info().Msg("simple log message")
			},
		},
		{
			name: "both outputs simple message",
			config: Config{
				Level:   "info",
				Console: true,
				File: &FileConfig{
					Path:       filepath.Join(tmpDir, "both_simple.log"),
					MaxSize:    100,
					MaxBackups: 1,
					MaxAge:     1,
					Compress:   false,
				},
			},
			logFunc: func() {
				Info().Msg("simple log message")
			},
		},
		{
			name: "structured logging with multiple fields",
			config: Config{
				Level: "info",
				File: &FileConfig{
					Path:       filepath.Join(tmpDir, "structured.log"),
					MaxSize:    100,
					MaxBackups: 1,
					MaxAge:     1,
					Compress:   false,
				},
			},
			logFunc: func() {
				Info().
					Str("string", "test").
					Int("integer", 42).
					Float64("float", 3.14).
					Bool("boolean", true).
					Msg("structured log message")
			},
		},
		{
			name: "debug level (disabled) logging",
			config: Config{
				Level: "info",
				File: &FileConfig{
					Path:       filepath.Join(tmpDir, "disabled.log"),
					MaxSize:    100,
					MaxBackups: 1,
					MaxAge:     1,
					Compress:   false,
				},
			},
			logFunc: func() {
				Debug().Msg("debug message that should be filtered")
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			if err := InitLogger(bm.config); err != nil {
				b.Fatalf("Failed to initialize logger: %v", err)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				bm.logFunc()
			}
		})
	}
}

func BenchmarkConcurrentLogging(b *testing.B) {
	// Create temporary directory for benchmark logs
	tmpDir, err := os.MkdirTemp("", "talis-agent-bench-*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	config := Config{
		Level: "info",
		File: &FileConfig{
			Path:       filepath.Join(tmpDir, "concurrent.log"),
			MaxSize:    100,
			MaxBackups: 1,
			MaxAge:     1,
			Compress:   false,
		},
	}

	if err := InitLogger(config); err != nil {
		b.Fatalf("Failed to initialize logger: %v", err)
	}

	benchmarks := []struct {
		name         string
		numGoroutine int
	}{
		{"2 goroutines", 2},
		{"4 goroutines", 4},
		{"8 goroutines", 8},
		{"16 goroutines", 16},
		{"32 goroutines", 32},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				var wg sync.WaitGroup
				wg.Add(bm.numGoroutine)

				for g := 0; g < bm.numGoroutine; g++ {
					go func(id int) {
						defer wg.Done()
						Info().
							Int("goroutine", id).
							Msg("concurrent benchmark message")
					}(g)
				}

				wg.Wait()
			}
		})
	}
}

func BenchmarkLogLevels(b *testing.B) {
	// Create temporary directory for benchmark logs
	tmpDir, err := os.MkdirTemp("", "talis-agent-bench-*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	levels := []string{"debug", "info", "warn", "error"}
	logFuncs := map[string]func() *zerolog.Event{
		"debug": Debug,
		"info":  Info,
		"warn":  Warn,
		"error": Error,
	}

	for _, configLevel := range levels {
		config := Config{
			Level: configLevel,
			File: &FileConfig{
				Path:       filepath.Join(tmpDir, configLevel+".log"),
				MaxSize:    100,
				MaxBackups: 1,
				MaxAge:     1,
				Compress:   false,
			},
		}

		if err := InitLogger(config); err != nil {
			b.Fatalf("Failed to initialize logger: %v", err)
		}

		for _, logLevel := range levels {
			name := fmt.Sprintf("config=%s/log=%s", configLevel, logLevel)
			b.Run(name, func(b *testing.B) {
				logFunc := logFuncs[logLevel]
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					logFunc().Msg("benchmark message")
				}
			})
		}
	}
}
