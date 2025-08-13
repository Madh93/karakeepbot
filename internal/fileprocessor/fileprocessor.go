// Package fileprocessor provides a robust and configurable utility for
// downloading files from a URL, processing them, and saving them to a temporary
// location.
package fileprocessor

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/Madh93/karakeepbot/internal/config"
)

// Custom error types for better error handling.
var (
	ErrValidationFailed = errors.New("file validation failed")
	ErrDownloadFailed   = errors.New("failed to download file")
	ErrProcessingFailed = errors.New("failed to process file")
)

// Validator is a function type that defines a validation strategy for a file.
// It returns the detected content type and an error if validation fails.
type Validator func(file *os.File) (contentType string, err error)

// Processor is responsible for downloading and processing files.
type Processor struct {
	tempdir string
	maxsize int64
	timeout int
}

// New creates a new Processor using the provided configuration.
func New(config *config.FileProcessorConfig) (*Processor, error) {
	// Create a temporary directory for storing files.
	tempdir := config.Tempdir
	if tempdir == "" {
		tempdir = os.TempDir()
	}
	if err := os.MkdirAll(tempdir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	return &Processor{
		tempdir: tempdir,
		maxsize: config.Maxsize,
		timeout: config.Timeout,
	}, nil
}

// Process downloads a file from a URL, optionally validates it, and saves it to
// a temporary location. The caller is responsible for cleaning up the file
// using the Cleanup method.
func (p *Processor) Process(fileURL string, validator Validator) (path string, contentType string, err error) {
	// Create a temporary file.
	tmpFile, err := os.CreateTemp(p.tempdir, "karakeepbot-*.tmp")
	if err != nil {
		return "", "", fmt.Errorf("%w: %w", ErrProcessingFailed, err)
	}
	defer tmpFile.Close()

	// Configure an HTTP client with the specified timeout.
	client := http.Client{
		Timeout: time.Duration(p.timeout) * time.Second,
	}

	// Download the file.
	resp, err := client.Get(fileURL)
	if err != nil {
		p.Cleanup(tmpFile.Name())
		return "", "", fmt.Errorf("%w: %w", ErrDownloadFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		p.Cleanup(tmpFile.Name())
		return "", "", fmt.Errorf("%w: status code %d", ErrDownloadFailed, resp.StatusCode)
	}

	// Limit the download to prevent resource exhaustion and copy data
	limitedReader := &io.LimitedReader{R: resp.Body, N: p.maxsize + 1}
	bytesWritten, err := io.Copy(tmpFile, limitedReader)
	if err != nil {
		p.Cleanup(tmpFile.Name())
		return "", "", fmt.Errorf("%w: %w", ErrProcessingFailed, err)
	}

	// Check if the file was larger than the limit
	if bytesWritten > p.maxsize {
		p.Cleanup(tmpFile.Name())
		return "", "", fmt.Errorf("%w: exceeds %d bytes", ErrValidationFailed, p.maxsize)
	}

	// Run validation if a validator function is provided.
	if validator != nil {
		detectedContentType, err := validator(tmpFile)
		if err != nil {
			p.Cleanup(tmpFile.Name())
			return "", "", fmt.Errorf("%w: %w", ErrValidationFailed, err)
		}
		contentType = detectedContentType
	} else {
		// If no validator is provided, still attempt to detect the content type
		// without enforcing any rules.
		if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
			buffer := make([]byte, 512)
			n, _ := tmpFile.Read(buffer)
			contentType = http.DetectContentType(buffer[:n])
		}
	}

	return tmpFile.Name(), contentType, nil
}

// Cleanup removes the temporary file.
func (p *Processor) Cleanup(filePath string) error {
	return os.Remove(filePath)
}
