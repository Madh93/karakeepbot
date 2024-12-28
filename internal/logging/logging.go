// Package logging manages application logging using the log/slog package.
//
// This package provides a structured logging facility for applications. It
// allows the creation of a Logger instance that can log messages at different
// severity levels such as Debug, Info, Warn, Error, and Fatal. The logging
// configuration is flexible and supports different output destinations (such as
// standard output or files) and formats (such as JSON or colored text).
//
// The Logger uses the slog package for structured logging and can be configured
// to determine the logging output and format based on user-defined settings.
//
// Use the New function to create a Logger instance with specified logging
// configuration. Various methods are provided to log messages at different
// severity levels with additional context.
package logging

import (
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/Madh93/hoarderbot/internal/config"
	"github.com/lmittmann/tint"
)

// Logger represents an instance of the logging system.
type Logger struct {
	slogger *slog.Logger
}

// New creates a new Logger instance with the specified logging configuration.
func New(config *config.LoggingConfig) *Logger {
	// Set the output based on the configuration
	output, err := parseOutput(config)
	if err != nil {
		log.Fatalf("Couldn't parse logging output: %v", err)
	}

	// Setup the handler based on the format
	handler, err := parseFormat(output, config)
	if err != nil {
		log.Fatalf("Couldn't setup logging handler: %v", err)
	}

	return &Logger{slogger: slog.New(handler)}
}

// Debug logs a message at the debug level.
func (l *Logger) Debug(msg string, args ...any) {
	l.slogger.Debug(msg, args...)
}

// Info logs a message at the info level.
func (l *Logger) Info(msg string, args ...any) {
	l.slogger.Info(msg, args...)
}

// Warn logs a message at the warn level.
func (l *Logger) Warn(msg string, args ...any) {
	l.slogger.Warn(msg, args...)
}

// Error logs a message at the error level.
func (l *Logger) Error(msg string, args ...any) {
	l.slogger.Error(msg, args...)
}

// Fatal simulates the behavior of a fatal log level.
func (l *Logger) Fatal(msg string, args ...any) {
	l.slogger.Error(msg, args...)
	os.Exit(1)
}

// parseOutput takes a logging configuration and returns an *os.File.
func parseOutput(config *config.LoggingConfig) (*os.File, error) {
	var output *os.File
	var err error

	switch config.Output {
	case "stderr":
		output = os.Stderr
	case "file":
		if config.Path == "" {
			output = os.Stdout // In case a path is not provided, use stdout
		} else {
			output, err = os.OpenFile(config.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
			if err != nil {
				return nil, err
			}
		}
	default:
		output = os.Stdout
	}

	return output, nil
}

// parseLevel takes a string representation of a log level and attempts to
// unmarshal it into a slog.Level.
func parseLevel(s string) slog.Level {
	var level slog.Level
	if err := level.UnmarshalText([]byte(s)); err != nil {
		return slog.LevelInfo
	}
	return level
}

// parseFormat takes a logging configuration and returns an slog.Handler.
func parseFormat(output *os.File, config *config.LoggingConfig) (slog.Handler, error) {
	var handler slog.Handler

	switch config.Format {
	case "json":
		handler = slog.NewJSONHandler(output, &slog.HandlerOptions{
			Level: parseLevel(config.Level),
		})
	default:
		handler = tint.NewHandler(output, &tint.Options{
			Level:      parseLevel(config.Level),
			TimeFormat: time.DateTime,
			NoColor:    !config.Colored,
		})
	}

	return handler, nil
}
