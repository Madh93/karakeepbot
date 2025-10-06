package validation

import (
	"testing"

	"github.com/Madh93/karakeepbot/internal/secret"
)

func TestValidateKarakeepToken(t *testing.T) {
	// Define test cases
	tests := []struct {
		token    secret.String
		expected bool
	}{
		{secret.New("ak1_1fa4507e4b58b5850672_13cb03dc5372fbe200d5"), true},           // Valid token
		{secret.New("ak1_cc2109516bfc2282a2bb_d7ad539acb2401b3f4b8"), true},           // Valid token
		{secret.New("ak2_628336a9531f92a42f55_08835688ea5708f0b83b"), true},           // Valid token
		{secret.New("ak3_628336a9531f92a42f55_08835688ea5708f0b83b"), true},         // Invalid prefix
		{secret.New("ak1_12345_67890"), false},                                        // Invalid format (too short)
		{secret.New("ak1_4d22ace06b233b9d2d22_c96e719"), false},                       // Invalid (too short for second segment)
		{secret.New("ak1_eca9dd7db9a4d7585f61b3063337d_054ef423745a72c0a7a8"), false}, // Invalid (too long for first segment)
		{secret.New("invalidtoken"), false},                                           // Completely invalid
	}

	// Iterate over the test cases
	for _, test := range tests {
		err := ValidateKarakeepToken(test.token)
		got := err == nil
		if got != test.expected {
			t.Errorf("For token: %q, expected error: %v, but got: %v", test.token.Value(), test.expected, got)
		}
	}
}
