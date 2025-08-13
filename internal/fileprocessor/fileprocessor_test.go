package fileprocessor

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Madh93/karakeepbot/internal/config"
	"github.com/Madh93/karakeepbot/internal/filevalidator"
)

func TestProcessor_Process(t *testing.T) {
	// Test data for various file types.
	var (
		pngData  = []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4, 0x89, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00, 0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82}
		webpData = []byte{'R', 'I', 'F', 'F', 0x00, 0x00, 0x00, 0x00, 'W', 'E', 'B', 'P'}
		textData = []byte("this is not a valid image file")
	)

	// Setup a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/valid.png":
			w.Write(pngData)
		case "/valid.webp":
			w.Write(webpData)
		case "/invalid.txt":
			w.Write(textData)
		case "/too-large":
			w.Write(make([]byte, 101)) // 101 bytes, larger than our test limit of 100
		case "/not-found":
			http.NotFound(w, r)
		case "/slow":
			time.Sleep(2 * time.Second)
			w.Write([]byte("slow response"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	tests := []struct {
		name          string
		config        config.FileProcessorConfig
		urlPath       string
		validator     Validator
		expectError   bool
		expectedError error
	}{
		{
			name:          "Success with valid PNG and validator",
			config:        config.FileProcessorConfig{Maxsize: 100, Timeout: 5},
			urlPath:       "/valid.png",
			validator:     filevalidator.ImageValidator,
			expectError:   false,
			expectedError: nil,
		},
		{
			name:          "Success with valid WebP and validator",
			config:        config.FileProcessorConfig{Maxsize: 100, Timeout: 5},
			urlPath:       "/valid.webp",
			validator:     filevalidator.ImageValidator,
			expectError:   false,
			expectedError: nil,
		},
		{
			name:          "Success without validator",
			config:        config.FileProcessorConfig{Maxsize: 100, Timeout: 5},
			urlPath:       "/valid.png",
			validator:     nil,
			expectError:   false,
			expectedError: nil,
		},
		{
			name:          "Fail on file too large",
			config:        config.FileProcessorConfig{Maxsize: 100, Timeout: 5},
			urlPath:       "/too-large",
			validator:     nil,
			expectError:   true,
			expectedError: ErrValidationFailed,
		},
		{
			name:          "Fail on validation error with text file",
			config:        config.FileProcessorConfig{Maxsize: 200, Timeout: 5},
			urlPath:       "/invalid.txt",
			validator:     filevalidator.ImageValidator,
			expectError:   true,
			expectedError: ErrValidationFailed,
		},
		{
			name:          "Fail on server 404",
			config:        config.FileProcessorConfig{Maxsize: 100, Timeout: 5},
			urlPath:       "/not-found",
			validator:     nil,
			expectError:   true,
			expectedError: ErrDownloadFailed,
		},
		{
			name:          "Fail on download timeout",
			config:        config.FileProcessorConfig{Maxsize: 100, Timeout: 1},
			urlPath:       "/slow",
			validator:     nil,
			expectError:   true,
			expectedError: ErrDownloadFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor, err := New(&tt.config)
			if err != nil {
				t.Fatalf("Failed to create processor: %v", err)
			}

			filePath, _, err := processor.Process(server.URL+tt.urlPath, tt.validator)
			if filePath != "" {
				defer processor.Cleanup(filePath)
			}

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got nil")
				}
				if tt.expectedError != nil && !errors.Is(err, tt.expectedError) {
					t.Errorf("Expected error type %v, but got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
			}
		})
	}
}
