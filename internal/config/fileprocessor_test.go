package config

import (
	"testing"
)

func TestFileProcessorConfig_Validate(t *testing.T) {
	// Define test cases for the Validate() method
	tests := []struct {
		name     string
		config   FileProcessorConfig
		expected bool
	}{
		{
			name:     "Valid config",
			config:   FileProcessorConfig{Maxsize: 1024, Timeout: 30, Mimetypes: []string{"image/png", "image/jpeg"}},
			expected: true,
		},
		{
			name:     "Valid config with empty MIME types list",
			config:   FileProcessorConfig{Maxsize: 1024, Timeout: 30, Mimetypes: []string{}},
			expected: true,
		},
		{
			name:     "Invalid Maxsize (zero)",
			config:   FileProcessorConfig{Maxsize: 0, Timeout: 30},
			expected: false,
		},
		{
			name:     "Invalid Maxsize (negative)",
			config:   FileProcessorConfig{Maxsize: -100, Timeout: 30},
			expected: false,
		},
		{
			name:     "Invalid Timeout (zero)",
			config:   FileProcessorConfig{Maxsize: 1024, Timeout: 0},
			expected: false,
		},
		{
			name:     "Invalid Timeout (negative)",
			config:   FileProcessorConfig{Maxsize: 1024, Timeout: -5},
			expected: false,
		},
		{
			name:     "Invalid with empty MIME type",
			config:   FileProcessorConfig{Maxsize: 1024, Timeout: 30, Mimetypes: []string{"image/png", ""}},
			expected: false,
		},
		{
			name:     "Invalid with duplicate MIME type",
			config:   FileProcessorConfig{Maxsize: 1024, Timeout: 30, Mimetypes: []string{"image/png", "image/jpeg", "image/png"}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			got := err == nil
			if got != tt.expected {
				t.Errorf("For config %+v, expected valid: %v, but got error: %v", tt.config, tt.expected, err)
			}
		})
	}
}
