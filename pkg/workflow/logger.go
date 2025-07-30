package workflow

import (
	"context"
	"log/slog"
	"os"

	"github.com/ratlabs-io/go-agent-kit/pkg/constants"
)

// Logger defines the interface for logging within the workflow system.
// This allows users to provide their own logger implementation or use the default.
type Logger interface {
	// Debug logs a debug message with optional key-value pairs
	Debug(msg string, keysAndValues ...interface{})
	
	// Info logs an info message with optional key-value pairs
	Info(msg string, keysAndValues ...interface{})
	
	// Warn logs a warning message with optional key-value pairs
	Warn(msg string, keysAndValues ...interface{})
	
	// Error logs an error message with optional key-value pairs
	Error(msg string, keysAndValues ...interface{})
	
	// With returns a new logger with the given key-value pairs added as context
	With(keysAndValues ...interface{}) Logger
}

// SlogLogger wraps Go's standard slog.Logger to implement our Logger interface
type SlogLogger struct {
	logger *slog.Logger
}

// NewSlogLogger creates a new SlogLogger wrapping the provided slog.Logger
func NewSlogLogger(logger *slog.Logger) *SlogLogger {
	return &SlogLogger{logger: logger}
}

// Debug logs a debug message
func (l *SlogLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.logger.Debug(msg, keysAndValues...)
}

// Info logs an info message
func (l *SlogLogger) Info(msg string, keysAndValues ...interface{}) {
	l.logger.Info(msg, keysAndValues...)
}

// Warn logs a warning message
func (l *SlogLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.logger.Warn(msg, keysAndValues...)
}

// Error logs an error message
func (l *SlogLogger) Error(msg string, keysAndValues ...interface{}) {
	l.logger.Error(msg, keysAndValues...)
}

// With returns a new logger with additional context
func (l *SlogLogger) With(keysAndValues ...interface{}) Logger {
	return &SlogLogger{logger: l.logger.With(keysAndValues...)}
}

// NoOpLogger is a logger that does nothing - useful for disabling logging
type NoOpLogger struct{}

// NewNoOpLogger creates a new no-op logger
func NewNoOpLogger() *NoOpLogger {
	return &NoOpLogger{}
}

// Debug does nothing
func (l *NoOpLogger) Debug(msg string, keysAndValues ...interface{}) {}

// Info does nothing
func (l *NoOpLogger) Info(msg string, keysAndValues ...interface{}) {}

// Warn does nothing
func (l *NoOpLogger) Warn(msg string, keysAndValues ...interface{}) {}

// Error does nothing
func (l *NoOpLogger) Error(msg string, keysAndValues ...interface{}) {}

// With returns itself (no-op)
func (l *NoOpLogger) With(keysAndValues ...interface{}) Logger {
	return l
}

// defaultLogger provides a sensible default logger using slog
var defaultLogger Logger = NewSlogLogger(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
	Level: slog.LevelInfo,
})))

// SetDefaultLogger sets the default logger used by all workflow components
// when no specific logger is provided
func SetDefaultLogger(logger Logger) {
	defaultLogger = logger
}

// GetDefaultLogger returns the current default logger
func GetDefaultLogger() Logger {
	return defaultLogger
}

// LoggerFromContext extracts a logger from context, falling back to default if not present
func LoggerFromContext(ctx context.Context) Logger {
	if logger, ok := ctx.Value(constants.KeyLogger).(Logger); ok {
		return logger
	}
	return defaultLogger
}

// WithLogger adds a logger to the context
func WithLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, constants.KeyLogger, logger)
}