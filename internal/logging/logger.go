package logging

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	// defaultLogger is the default logger instance
	defaultLogger zerolog.Logger
)

// Config represents logger configuration
type Config struct {
	Level      string
	TimeFormat string
	Console    bool
	File       *FileConfig
}

// FileConfig represents file-based logging configuration
type FileConfig struct {
	Path       string
	MaxSize    int  // Maximum size in megabytes before rotation
	MaxBackups int  // Maximum number of old log files to retain
	MaxAge     int  // Maximum number of days to retain old log files
	Compress   bool // Compress rotated files
}

// DefaultFileConfig returns the default file configuration
func DefaultFileConfig() *FileConfig {
	return &FileConfig{
		Path:       "/var/log/talis-agent/agent.log",
		MaxSize:    100,  // 100 MB
		MaxBackups: 5,    // Keep 5 backups
		MaxAge:     30,   // Keep logs for 30 days
		Compress:   true, // Compress old logs
	}
}

// InitLogger initializes the global logger with the given configuration
func InitLogger(cfg Config) error {
	// Set default time format if not specified
	if cfg.TimeFormat == "" {
		cfg.TimeFormat = time.RFC3339
	}

	// Configure time field format
	zerolog.TimeFieldFormat = cfg.TimeFormat

	// Set log level
	level := parseLevel(cfg.Level)
	zerolog.SetGlobalLevel(level)

	var writers []io.Writer

	// Configure console output if requested
	if cfg.Console {
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: cfg.TimeFormat,
		}
		writers = append(writers, consoleWriter)
	}

	// Configure file output if requested
	if cfg.File != nil {
		// Create log directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(cfg.File.Path), 0750); err != nil {
			return err
		}

		fileWriter := &lumberjack.Logger{
			Filename:   cfg.File.Path,
			MaxSize:    cfg.File.MaxSize,
			MaxBackups: cfg.File.MaxBackups,
			MaxAge:     cfg.File.MaxAge,
			Compress:   cfg.File.Compress,
		}
		writers = append(writers, fileWriter)
	}

	// Create multi-writer if we have multiple outputs
	var output io.Writer
	if len(writers) > 1 {
		output = zerolog.MultiLevelWriter(writers...)
	} else if len(writers) == 1 {
		output = writers[0]
	} else {
		output = os.Stdout // Default to stdout if no writers specified
	}

	// Create logger
	defaultLogger = zerolog.New(output).With().Timestamp().Logger()

	// Log initial message
	Info().
		Str("level", level.String()).
		Bool("console", cfg.Console).
		Bool("file", cfg.File != nil).
		Msg("Logger initialized")

	return nil
}

// parseLevel converts a string level to zerolog.Level
func parseLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}

// Debug returns a debug level event handler
func Debug() *zerolog.Event {
	return defaultLogger.Debug()
}

// Info returns an info level event handler
func Info() *zerolog.Event {
	return defaultLogger.Info()
}

// Warn returns a warn level event handler
func Warn() *zerolog.Event {
	return defaultLogger.Warn()
}

// Error returns an error level event handler
func Error() *zerolog.Event {
	return defaultLogger.Error()
}

// Fatal returns a fatal level event handler
func Fatal() *zerolog.Event {
	return defaultLogger.Fatal()
}

// With returns a new logger with the given fields
func With() zerolog.Context {
	return defaultLogger.With()
}
