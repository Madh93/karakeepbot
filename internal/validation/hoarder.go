package validation

import (
	"fmt"
	"regexp"

	"github.com/Madh93/hoarderbot/internal/secret"
)

// ValidateHoarderToken checks if the provided Hoarder API Key is valid.
func ValidateHoarderToken(token secret.String) error {
	// Define the pattern for a valid Hoarder API Key
	// See: https://github.com/hoarder-app/hoarder/blob/v0.20.0/packages/trpc/auth.ts#L14
	pattern := `^ak1_[a-f0-9]{20}_[a-f0-9]{20}$`
	re := regexp.MustCompile(pattern)

	// Check if the token matches the defined pattern
	if !re.MatchString(token.Value()) {
		return fmt.Errorf("invalid Hoarder API Key: %s", token)
	}

	return nil
}
