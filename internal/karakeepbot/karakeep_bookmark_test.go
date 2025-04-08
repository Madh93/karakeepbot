package karakeepbot

import (
	"testing"
)

func TestSanitizeTag(t *testing.T) {
	// Test the sanitizeTag() method for different cases
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"golang", "golang"},
		{"go programming", "goprogramming"},
		{"go-programming", "goprogramming"},
		{" spaces ", "spaces"},
		{"-hyphens-", "hyphens"},
		{"  multiple   spaces  ", "multiplespaces"},
		{"--multiple---hyphens--", "multiplehyphens"},
		{"--spaces and hyphens--", "spacesandhyphens"},
	}

	for _, test := range tests {
		got := sanitizeTag(test.input)
		if got != test.expected {
			t.Errorf("For input %q to sanitizeTag(), expected %q, but got %q", test.input, test.expected, got)
		}
	}
}
