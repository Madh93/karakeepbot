package version

import (
	"testing"
)

func TestGet(t *testing.T) {
	// Test getting the version information.
	vi := Get()

	if vi.AppVersion != "unknown" {
		t.Errorf("expected AppVersion to be 'unknown', but got '%s'", vi.AppVersion)
	}

	if vi.CommitHash != "unknown" {
		t.Errorf("expected CommitHash to be 'unknown', but got '%s'", vi.CommitHash)
	}
}

func TestString(t *testing.T) {
	// Test the String() method for different version cases
	tests := []struct {
		input    VersionInfo
		expected string
	}{
		{VersionInfo{}, "version  ()"},
		{VersionInfo{AppVersion: "1.0.0", CommitHash: "abc123"}, "version 1.0.0 (abc123)"},
		{VersionInfo{AppVersion: "2.1.3", CommitHash: "def456"}, "version 2.1.3 (def456)"},
		{VersionInfo{AppVersion: "v1.0.0-beta", CommitHash: "ghi789"}, "version v1.0.0-beta (ghi789)"},
	}

	for _, test := range tests {
		got := test.input.String()
		if got != test.expected {
			t.Errorf("For input %q to String(), expected %q, but got %q", test.input, test.expected, got)
		}
	}
}
