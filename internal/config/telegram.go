package config

import (
	"github.com/Madh93/karakeepbot/internal/secret"
	"github.com/Madh93/karakeepbot/internal/validation"
)

// TelegramConfig represents a configuration for Telegram.
type TelegramConfig struct {
	Token     secret.String `koanf:"token"`     // Telegram bot token.
	Allowlist []int64       `koanf:"allowlist"` // Allowed chat IDs for the bot to interact with.
	Threads   []int         `koanf:"threads"`   // Allowed thread IDs (a.k.a topics) for the bot to interact with.
}

// Validate checks if the Telegram configuration is valid.
func (c TelegramConfig) Validate() error {
	if err := validation.ValidateTelegramToken(c.Token); err != nil {
		return err
	}

	return nil
}
