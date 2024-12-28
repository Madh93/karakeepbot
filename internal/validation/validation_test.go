package validation

import (
	"testing"
)

// TestValidate checks the Validate function for different cases.
func TestValidate(t *testing.T) {
	// Define test cases for integers
	testsInt := []struct {
		value        int
		validOptions []int
		expected     bool
	}{
		{1, []int{1, 2, 3}, false}, // Valid case
		{4, []int{1, 2, 3}, true},  // Invalid case
	}

	for _, test := range testsInt {
		err := Validate(test.value, test.validOptions)
		got := err != nil
		if got != test.expected {
			t.Errorf("For value: %v, expected error: %v, but got: %v", test.value, test.expected, got)
		}
	}

	// Define test cases for strings
	testsStr := []struct {
		value        string
		validOptions []string
		expected     bool
	}{
		{"apple", []string{"apple", "banana", "cherry"}, false}, // Valid case
		{"grape", []string{"apple", "banana", "cherry"}, true},  // Invalid case
	}

	for _, test := range testsStr {
		err := Validate(test.value, test.validOptions)
		got := err != nil
		if got != test.expected {
			t.Errorf("For value: %q, expected error: %v, but got: %v", test.value, test.expected, got)
		}
	}
}
