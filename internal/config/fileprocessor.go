package config

import (
	"errors"
	"fmt"
	"strings"
)

// FileProcessorConfig represents a configuration for the file processor.
type FileProcessorConfig struct {
	Tempdir   string   `koanf:"tempdir"`   // Temporary directory for storing files.
	Maxsize   int64    `koanf:"maxsize"`   // Maximum allowed file size in bytes.
	Timeout   int      `koanf:"timeout"`   // Maximum time to wait for file download in seconds.
	Mimetypes []string `koanf:"mimetypes"` // Allowed MIME types. If empty, all types are allowed.
}

// Validate checks if the FileProcessor configuration is valid.
func (c FileProcessorConfig) Validate() error {
	if c.Maxsize <= 0 {
		return fmt.Errorf("invalid maxsize: must be a positive value, got %d", c.Maxsize)
	}

	if c.Timeout <= 0 {
		return fmt.Errorf("invalid timeout: must be a positive value, got %d", c.Timeout)
	}

	// Check for empty or duplicate entries in Mimetypes.
	seen := make(map[string]bool)
	for _, mimeType := range c.Mimetypes {
		if strings.TrimSpace(mimeType) == "" {
			return errors.New("invalid mimetypes: contains an empty entry")
		}
		if seen[mimeType] {
			return fmt.Errorf("invalid mimetypes: contains duplicate entry '%s'", mimeType)
		}
		seen[mimeType] = true
	}

	return nil
}
