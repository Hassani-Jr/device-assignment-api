package logger

import (
	"log/slog"
	"os"
)

// Logger defines the logging interface
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// SlogLogger implements Logger using the standard library slog
type SlogLogger struct {
	logger *slog.Logger
}

// New creates a new structured logger
func New() Logger {
	// Create a structured logger with JSON output for production
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	return &SlogLogger{logger: logger}
}

// NewWithLevel creates a new logger with the specified log level
func NewWithLevel(level slog.Level) Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))

	return &SlogLogger{logger: logger}
}

// Debug logs a debug message
func (l *SlogLogger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

// Info logs an info message
func (l *SlogLogger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

// Warn logs a warning message
func (l *SlogLogger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

// Error logs an error message
func (l *SlogLogger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}
