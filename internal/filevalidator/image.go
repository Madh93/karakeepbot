// Package filevalidator provides a collection of validation functions that can
// be used with the fileprocessor. Each validator inspects a file and checks if
// it conforms to a specific set of rules, such as matching a MIME type.
package filevalidator

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
)

// ImageValidator checks if a file is a valid JPEG, PNG, or WebP image. It uses
// custom logic for WebP detection as it's not fully supported by the standard
// http.DetectContentType function.
func ImageValidator(file *os.File) (string, error) {
	// Ensure we read the file from the beginning.
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	// Read the first 512 bytes to detect the content type.
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}

	// Slice the buffer to the actual number of bytes read.
	buffer = buffer[:n]

	// Custom WebP detection logic. A WebP file starts with "RIFF" and has "WEBP"
	// at offset 8.
	if len(buffer) >= 12 && bytes.Equal(buffer[0:4], []byte("RIFF")) && bytes.Equal(buffer[8:12], []byte("WEBP")) {
		return "image/webp", nil
	}

	// Fallback to standard library detection for other formats.
	contentType := http.DetectContentType(buffer)
	switch contentType {
	case "image/jpeg", "image/png":
		return contentType, nil
	default:
		return "", errors.New("unsupported image format")
	}
}
