package validation

import (
	"testing"
)

func TestValidateURL(t *testing.T) {
	// Define test cases
	tests := []struct {
		rawURL   string
		expected bool
	}{
		{"http://example.com", true},                       // Valid URL
		{"https://example.com", true},                      // Valid URL
		{"http://localhost:8080", true},                    // Valid URL
		{"ftp://example.com", false},                       // Invalid scheme
		{"http://", false},                                 // Missing host
		{"https://", false},                                // Missing host
		{"invalid-url", false},                             // Completely invalid
		{"http://example.com/path?query=1#fragment", true}, // Valid URL with path, query, and fragment
	}

	for _, test := range tests {
		err := ValidateURL(test.rawURL)
		got := err == nil
		if got != test.expected {
			t.Errorf("For URL: %q, expected %v, but got: %v", test.rawURL, test.expected, got)
		}
	}
}
