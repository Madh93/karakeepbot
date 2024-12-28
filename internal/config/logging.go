package config

import (
	"fmt"

	"github.com/Madh93/hoarderbot/internal/validation"
)

// LoggingConfig represents a configuration for logging.
type LoggingConfig struct {
	Level   string `koanf:"level"`   // Log level
	Format  string `koanf:"format"`  // Log format
	Output  string `koanf:"output"`  // Output destination
	Path    string `koanf:"path"`    // File path for logging output (if output is a file)
	Colored bool   `koanf:"colored"` // Whether to colorize log output (if format is text)
}

// Logging configuration values.
const (
	// Log levels.
	DebugLevel = "debug"
	InfoLevel  = "info"
	WarnLevel  = "warn"
	ErrorLevel = "error"
	FatalLevel = "fatal"

	// Log formats.
	TextFormat = "text"
	JSONFormat = "json"

	// Output destinations.
	StdoutOutput = "stdout"
	StderrOutput = "stderr"
	FileOutput   = "file"
)

// validLogging maps each logging field to a list of allowed values.
var validLogging = map[string][]string{
	"Level":  {DebugLevel, InfoLevel, WarnLevel, ErrorLevel, FatalLevel},
	"Format": {TextFormat, JSONFormat},
	"Output": {StdoutOutput, StderrOutput, FileOutput},
}

// Validate checks if the logging configuration is valid.
func (c LoggingConfig) Validate() error {
	if err := validation.Validate(c.Level, validLogging["Level"]); err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}
	if err := validation.Validate(c.Format, validLogging["Format"]); err != nil {
		return fmt.Errorf("invalid log format: %w", err)
	}
	if err := validation.Validate(c.Output, validLogging["Output"]); err != nil {
		return fmt.Errorf("invalid output destination: %w", err)
	}
	if err := validation.Validate(c.Colored, []bool{true, false}); err != nil {
		return fmt.Errorf("invalid colored setting: %w", err)
	}

	return nil
}
