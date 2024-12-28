package secret

import (
	"testing"
)

func TestNew(t *testing.T) {
	// Test creating a new secret.String
	input := New("mySecret")
	expected := "mySecret"
	got := string(input)

	if got != expected {
		t.Errorf("For input %q to New(), expected %q, but got %q", input, expected, got)
	}
}

func TestString(t *testing.T) {
	// Test the String() method for different cases
	tests := []struct {
		input    String
		expected string
	}{
		{New(""), ""},
		{New("123"), "123"},
		{New("secret"), "secr**"},
		{New("verylongsecret"), "very********"},
		{New("12345678"), "1234****"},
	}

	for _, test := range tests {
		got := test.input.String()
		if got != test.expected {
			t.Errorf("For input %q to String(), expected %q, but got %q", test.input, test.expected, got)
		}
	}
}

func TestValue(t *testing.T) {
	// Test the Value() method
	input := "anotherSecret"
	expected := "anotherSecret"
	got := New(input).Value()

	if got != expected {
		t.Errorf("For input %q to Value(), expected %q, but got %q", input, expected, got)
	}
}
