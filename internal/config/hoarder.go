package config

import (
	"fmt"

	"github.com/Madh93/hoarderbot/internal/secret"
	"github.com/Madh93/hoarderbot/internal/validation"
)

// HoarderConfig represents a configuration for the Hoarder server.
type HoarderConfig struct {
	URL      string        `koanf:"url"`      // Base URL of the Hoarder server
	Token    secret.String `koanf:"token"`    // Hoarder API key
	Interval int           `koanf:"interval"` // Interval (in seconds) before retrying tagging status
}

// Validate checks if the Hoarder configuration is valid.
func (c HoarderConfig) Validate() error {
	if err := validation.ValidateURL(c.URL); err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if err := validation.ValidateHoarderToken(c.Token); err != nil {
		return err
	}

	return nil
}
