// Package secret provides functionality for managing secrets.
//
// This package defines a custom type `String` that represents a secret string.
// A secret string can be created using a specific constructor and has methods
// to represent it in a masked format. The masked format displays only the first
// four characters of the string, obscuring the remaining characters with
// asterisks. Additionally, there is a method to retrieve the original value of
// the secret string as a regular string.
//
// The intended use of this package is to encapsulate sensitive information
// while allowing safe display without revealing the entire content of the
// string.
package secret

import "strings"

// String is a type that represents a secret string.
type String string

// New creates a new secret string.
func New(s string) String {
	return String(s)
}

// String returns the string representation, showing the first 4 characters and
// obscuring the rest with asterisks.
func (s String) String() string {
	str := string(s)

	if len(str) > 4 {
		if len(str) > 12 {
			str = str[:12] // Limit the length of the string to 12 characters
		}
		return str[:4] + strings.Repeat("*", len(str)-4) // Show the first 4 chars and replace the rest
	}

	return str // For short strings, return the string as is
}

// Value converts the secret.String type to a regular string.
func (s String) Value() string {
	return string(s)
}
