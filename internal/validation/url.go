package validation

import (
	"errors"
	"fmt"
	"net/url"
)

// ValidateURL checks if the given URL is valid based on valid HTTP/HTTPS schemes.
func ValidateURL(rawURL string) error {
	validSchemes := []string{"http", "https"}

	// Parse the URL using net/url
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return err // Return the error if URL parsing fails
	}

	// Check if the scheme is in the list of valid schemes
	isValidScheme := false
	for _, scheme := range validSchemes {
		if parsedURL.Scheme == scheme {
			isValidScheme = true
			break
		}
	}

	if !isValidScheme {
		return fmt.Errorf("URL scheme must be one of: %v", validSchemes)
	}

	// Check if the host is not empty
	if parsedURL.Host == "" {
		return errors.New("URL must have a valid host")
	}

	return nil // Return nil if the URL is valid
}
