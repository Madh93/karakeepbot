package filevalidator

import (
	"os"
	"testing"
)

func TestImageValidator(t *testing.T) {
	// Magic numbers for various file types.
	var (
		// PNG: starts with \x89PNG\r\n\x1a\n
		pngHeader = []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
		// JPEG: starts with 0xFF, 0xD8, 0xFF
		jpegHeader = []byte{0xff, 0xd8, 0xff}
		// WebP: starts with RIFF, then file size, then WEBP
		webpHeader = []byte{'R', 'I', 'F', 'F', 0x00, 0x00, 0x00, 0x00, 'W', 'E', 'B', 'P'}
		// A simple text file.
		textFile = []byte("this is not an image")
	)

	tests := []struct {
		name                string
		fileContent         []byte
		expectError         bool
		expectedContentType string
	}{
		{
			name:                "Valid PNG file",
			fileContent:         pngHeader,
			expectError:         false,
			expectedContentType: "image/png",
		},
		{
			name:                "Valid JPEG file",
			fileContent:         jpegHeader,
			expectError:         false,
			expectedContentType: "image/jpeg",
		},
		{
			name:                "Valid WebP file",
			fileContent:         webpHeader,
			expectError:         false,
			expectedContentType: "image/webp",
		},
		{
			name:                "Invalid file (text)",
			fileContent:         textFile,
			expectError:         true,
			expectedContentType: "",
		},
		{
			name:                "Empty file",
			fileContent:         []byte{},
			expectError:         true,
			expectedContentType: "",
		},
		{
			name:                "File too small for WebP check",
			fileContent:         []byte{'R', 'I', 'F', 'F'},
			expectError:         true,
			expectedContentType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file for the test.
			tmpFile, err := os.CreateTemp("", "test-image-*.tmp")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer func() {
				if err := os.Remove(tmpFile.Name()); err != nil {
					t.Errorf("Failed to remove temporary file: %v", err)
				}
			}()

			// Write the test content to the file.
			if _, err := tmpFile.Write(tt.fileContent); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			if err := tmpFile.Close(); err != nil { // Close the file to ensure content is flushed.
				t.Fatalf("Failed to close temp file: %v", err)
			}

			// Re-open the file for reading, as the validator expects a readable file.
			fileToValidate, err := os.Open(tmpFile.Name())
			if err != nil {
				t.Fatalf("Failed to open temp file for validation: %v", err)
			}
			defer func() {
				if err := fileToValidate.Close(); err != nil {
					t.Errorf("Failed to close file to validate: %v", err)
				}
			}()

			// Run the validator.
			contentType, err := ImageValidator(fileToValidate)

			// Check for errors.
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
			}

			// Check the content type.
			if contentType != tt.expectedContentType {
				t.Errorf("Expected content type '%s', but got '%s'", tt.expectedContentType, contentType)
			}
		})
	}
}
