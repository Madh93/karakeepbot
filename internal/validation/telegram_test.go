package validation

import (
	"testing"

	"github.com/Madh93/karakeepbot/internal/secret"
)

func TestValidateTelegramToken(t *testing.T) {
	// Define test cases
	tests := []struct {
		token    secret.String
		expected bool
	}{
		{secret.New("123456789:ABCdefGhIjKlmNOPQRstUVWXYZ123456789"), true},        // Valid token
		{secret.New("12345678:abcdefgHIJKLMNopqrstuVWXYZ012345678"), true},         // Valid token
		{secret.New("12345:AaZz0AaZz9"), false},                                    // Invalid (too short)
		{secret.New("123456789:AaZz0A"), false},                                    // Invalid (too short after ":")
		{secret.New("1234567890:ABCdefGhIjKlmNOPQRstUVWXYZ123456789extra"), false}, // Invalid (too long)
		{secret.New("ftp://bot:abcdefghij1234567890123456789012345"), false},       // Invalid scheme
		{secret.New("invalid-token"), false},                                       // Completely invalid
	}

	// Iterate over the test cases
	for _, test := range tests {
		err := ValidateTelegramToken(test.token)
		got := err == nil
		if got != test.expected {
			t.Errorf("For token: %q, expected error: %v, but got: %v", test.token.Value(), test.expected, got)
		}
	}
}
