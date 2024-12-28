package config

import (
	"testing"
)

func TestLoggingConfigValidate(t *testing.T) {
	// Test the Validate() method for different cases
	tests := []struct {
		config   LoggingConfig
		expected bool
	}{
		{LoggingConfig{Level: "info", Format: "text", Output: "file", Colored: false}, true},
		{LoggingConfig{Level: "debug", Format: "json", Output: "stdout", Colored: false}, true},
		{LoggingConfig{Level: "error", Format: "text", Output: "stderr", Colored: false}, true},
		{LoggingConfig{Level: "warn", Format: "json", Output: "stderr", Colored: true}, true},
		{LoggingConfig{Level: "invalid", Format: "text", Output: "file", Colored: false}, false}, // Invalid level
		{LoggingConfig{Level: "info", Format: "invalid", Output: "file", Colored: false}, false}, // Invalid format
		{LoggingConfig{Level: "info", Format: "text", Output: "invalid", Colored: false}, false}, // Invalid output destination
		{LoggingConfig{}, false}, // Empty configuration
	}

	for _, test := range tests {
		got := test.config.Validate()
		if (got == nil) != test.expected {
			t.Errorf("For logging config: %+v, expected %v, but got: %v", test.config, test.expected, got)
		}
	}
}
