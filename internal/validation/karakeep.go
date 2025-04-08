package validation

import (
	"fmt"
	"regexp"

	"github.com/Madh93/karakeepbot/internal/secret"
)

// ValidateKarakeepToken checks if the provided Karakeep API Key is valid.
func ValidateKarakeepToken(token secret.String) error {
	// Define the pattern for a valid Karakeep API Key
	// See: https://github.com/karakeep-app/karakeep/blob/v0.20.0/packages/trpc/auth.ts#L14
	pattern := `^ak1_[a-f0-9]{20}_[a-f0-9]{20}$`
	re := regexp.MustCompile(pattern)

	// Check if the token matches the defined pattern
	if !re.MatchString(token.Value()) {
		return fmt.Errorf("invalid Karakeep API Key: %s", token)
	}

	return nil
}
