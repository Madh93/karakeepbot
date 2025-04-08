package config

import (
	"fmt"

	"github.com/Madh93/karakeepbot/internal/secret"
	"github.com/Madh93/karakeepbot/internal/validation"
)

// KarakeepConfig represents a configuration for the Karakeep server.
type KarakeepConfig struct {
	URL      string        `koanf:"url"`      // Base URL of the Karakeep server
	Token    secret.String `koanf:"token"`    // Karakeep API key
	Interval int           `koanf:"interval"` // Interval (in seconds) before retrying tagging status
}

// Validate checks if the Karakeep configuration is valid.
func (c KarakeepConfig) Validate() error {
	if err := validation.ValidateURL(c.URL); err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if err := validation.ValidateKarakeepToken(c.Token); err != nil {
		return err
	}

	return nil
}
