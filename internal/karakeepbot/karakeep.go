package karakeepbot

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"

	"github.com/Madh93/go-karakeep"
	"github.com/Madh93/karakeepbot/internal/config"
	"github.com/Madh93/karakeepbot/internal/logging"
)

// Karakeep embeds the Karakeep API Client to add high level functionality.
type Karakeep struct {
	*karakeep.ClientWithResponses
}

// createKarakeep initializes the Karakeep API Client.
func createKarakeep(logger *logging.Logger, config *config.KarakeepConfig) *Karakeep {
	logger.Debug(fmt.Sprintf("Initializing Karakeep API Client at %s using %s token", config.URL, config.Token))

	// Setup API Endpoint
	parsedURL, err := url.Parse(config.URL)
	if err != nil {
		logger.Fatal("Error parsing URL.", "error", err)
	}
	parsedURL.Path, err = url.JoinPath(parsedURL.Path, "/api/v1")
	if err != nil {
		logger.Fatal("Error joining path.", "error", err)
	}

	// Setup Bearer Token Authentication
	auth := func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.Token.Value()))
		return nil
	}

	karakeepClient, err := karakeep.NewClientWithResponses(parsedURL.String(), karakeep.WithRequestEditorFn(auth))
	if err != nil {
		logger.Fatal("Error creating Karakeep API client.", "error", err)
	}

	return &Karakeep{ClientWithResponses: karakeepClient}
}

// CreateBookmark creates a new bookmark in Karakeep.
func (k Karakeep) CreateBookmark(ctx context.Context, b BookmarkType) (*KarakeepBookmark, error) {
	// Parse the JSON body of the request
	body, err := ToJSONReader(b)
	if err != nil {
		return nil, err
	}

	// Create bookmark
	response, err := k.PostBookmarksWithBodyWithResponse(ctx, "application/json", body)
	if err != nil {
		return nil, err
	}

	// Check if the bookmark was created successfully
	if response.StatusCode() != http.StatusCreated {
		return nil, fmt.Errorf("received HTTP status: %s", response.Status())
	}

	// Return bookmark
	bookmark := KarakeepBookmark(*response.JSON201)
	return &bookmark, nil
}

// RetrieveBookmarkById retrieves a bookmark by its ID.
func (k Karakeep) RetrieveBookmarkById(ctx context.Context, id string) (*KarakeepBookmark, error) {
	// Retrieve bookmark
	response, err := k.GetBookmarksBookmarkIdWithResponse(ctx, id, nil)
	if err != nil {
		return nil, err
	}

	// Check if the bookmark was created successfully
	if response.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("received HTTP status: %s", response.Status())
	}

	// Return bookmark
	bookmark := KarakeepBookmark(*response.JSON200)
	return &bookmark, nil
}

// CreateAsset uploads a file to Karakeep and returns the asset details.
// This version streams the request body, which is more memory-efficient and robust.
func (k Karakeep) CreateAsset(ctx context.Context, filePath string, mimeType string) (*KarakeepAsset, error) {
	pipeReader, pipeWriter := io.Pipe()
	writer := multipart.NewWriter(pipeWriter)
	errChan := make(chan error, 1)

	go func() {
		// Use a single error variable to capture the first failure.
		var err error

		// This defer ensures we always close the channel, signaling completion to the main function.
		defer close(errChan)

		// Defer the closing of the pipe and multipart writers.
		// The function will capture the 'err' variable by reference.
		// They will run in LIFO order: writer.Close() first, then pipeWriter.Close().
		defer func() {
			if closeErr := pipeWriter.Close(); closeErr != nil && err == nil {
				err = fmt.Errorf("failed to close pipe writer: %w", closeErr)
			}
		}()
		defer func() {
			if closeErr := writer.Close(); closeErr != nil && err == nil {
				err = fmt.Errorf("failed to close multipart writer: %w", closeErr)
			}
		}()

		// The rest of the logic is wrapped in a function to make error handling clean.
		// If this function returns an error, it gets sent to the channel.
		err = func() error {
			file, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("failed to open file: %w", err)
			}
			defer func() {
				if closeErr := file.Close(); closeErr != nil && err == nil {
					err = fmt.Errorf("failed to close file: %w", closeErr)
				}
			}()

			fileName := filepath.Base(filePath)
			part, err := writer.CreatePart(textproto.MIMEHeader{
				"Content-Disposition": []string{fmt.Sprintf(`form-data; name="file"; filename="%s"`, fileName)},
				"Content-Type":        []string{mimeType},
			})
			if err != nil {
				return fmt.Errorf("failed to create form part: %w", err)
			}

			if _, err := io.Copy(part, file); err != nil {
				return fmt.Errorf("failed to copy file to form: %w", err)
			}

			return nil // Success
		}()

		// If any error occurred in the logic above, send it to the channel.
		if err != nil {
			errChan <- err
		}
	}()

	// Make the HTTP request. This will now correctly unblock when the pipeWriter is closed.
	response, err := k.PostAssetsWithBodyWithResponse(ctx, writer.FormDataContentType(), pipeReader)

	// Check for errors from the goroutine. This also waits for the goroutine to finish its cleanup.
	if goroutineErr := <-errChan; goroutineErr != nil {
		return nil, goroutineErr
	}

	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("request cancelled: %w", ctx.Err())
		}
		return nil, fmt.Errorf("failed to upload asset: %w", err)
	}

	if response.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to upload asset, received HTTP status: %s, body: %s", response.Status(), string(response.Body))
	}

	asset := KarakeepAsset(*response.JSON200)
	return &asset, nil
}
