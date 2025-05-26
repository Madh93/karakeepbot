package karakeepbot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
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
	body, err := b.ToJSONReader()
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

// KarakeepFileUploadResponse defines the expected JSON structure of the file upload response.
type KarakeepFileUploadResponse struct {
	URL    string `json:"url"`
	FileID string `json:"file_id"`
}

// UploadImageToKaraKeep uploads an image file to the Karakeep server.
// It returns the URL of the uploaded image or an error.
func (k *Karakeep) UploadImageToKaraKeep(ctx context.Context, localFilePath string) (string, error) {
	file, err := os.Open(localFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", localFilePath, err)
	}
	defer file.Close()

	// Prepare the multipart form body
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Create a form file field
	part, err := writer.CreateFormFile("file", filepath.Base(localFilePath))
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return "", fmt.Errorf("failed to copy file to form: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Construct the target URL for file upload.
	// k.ClientWithResponses.Server is assumed to be "http://<host>:<port>/api/v1".
	// The file upload endpoint is assumed to be "/api/v1/files".
	uploadURL := k.ClientWithResponses.Server + "/files" // Guess based on task description

	req, err := http.NewRequestWithContext(ctx, "POST", uploadURL, &requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to create new HTTP request: %w", err)
	}

	// Set the content type and the Authorization header using the client's request editor.
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if len(k.ClientWithResponses.RequestEditors) > 0 {
		if err := k.ClientWithResponses.RequestEditors[0](ctx, req); err != nil {
			return "", fmt.Errorf("failed to apply request editor (auth): %w", err)
		}
	} else {
		return "", fmt.Errorf("karakeep client request editor not found for auth")
	}

	// Perform the HTTP request using the client's HTTPClient for consistency.
	httpClient := k.ClientWithResponses.Client
	if httpClient == nil {
		httpClient = http.DefaultClient // Fallback
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to Karakeep: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("karakeep API request failed with status %s: %s", resp.Status, string(errorBody))
	}

	// Parse the response
	var uploadResponse KarakeepFileUploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadResponse); err != nil {
		// Try to read the body as plain text for debugging if JSON parsing fails
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr == nil && len(bodyBytes) > 0 { // if body was not consumed by json.NewDecoder
			return "", fmt.Errorf("failed to decode Karakeep API response as JSON: %w. Response body: %s", err, string(bodyBytes))
		}
		return "", fmt.Errorf("failed to decode Karakeep API response as JSON: %w", err)
	}

	if uploadResponse.URL != "" {
		return uploadResponse.URL, nil
	}
	if uploadResponse.FileID != "" {
		// This might need refinement based on actual API behavior (e.g., constructing full URL).
		return uploadResponse.FileID, nil
	}

	return "", fmt.Errorf("karakeep API response did not contain a URL or FileID")
}

// RetrieveBookmarkById retrieves a bookmark by its ID.
func (k Karakeep) RetrieveBookmarkById(ctx context.Context, id string) (*KarakeepBookmark, error) {
	// Retrieve bookmark
	response, err := k.GetBookmarksBookmarkIdWithResponse(ctx, id)
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
