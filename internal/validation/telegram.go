package validation

import (
	"fmt"
	"regexp"

	"github.com/Madh93/hoarderbot/internal/secret"
)

// ValidateTelegramToken checks if the provided Telegram bot token is valid.
func ValidateTelegramToken(token secret.String) error {
	// Define the pattern for a valid Telegram bot token
	// See: https://stackoverflow.com/a/61888374
	pattern := `^[0-9]{8,10}:[a-zA-Z0-9_-]{35}$`
	re := regexp.MustCompile(pattern)

	// Check if the token matches the defined pattern
	if !re.MatchString(token.Value()) {
		return fmt.Errorf("invalid Telegram bot token: %s", token)
	}

	return nil
}
