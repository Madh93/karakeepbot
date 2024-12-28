package hoarderbot

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Madh93/go-hoarder"
	"github.com/Madh93/hoarderbot/internal/config"
	"github.com/Madh93/hoarderbot/internal/logging"
)

// Hoarder embeds the Hoarder API Client to add high level functionality.
type Hoarder struct {
	*hoarder.ClientWithResponses
}

// createHoarder initializes the Hoarder API Client.
func createHoarder(logger *logging.Logger, config *config.HoarderConfig) *Hoarder {
	logger.Debug(fmt.Sprintf("Initializing Hoarder API Client at %s using %s token", config.URL, config.Token))

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

	hoarderClient, err := hoarder.NewClientWithResponses(parsedURL.String(), hoarder.WithRequestEditorFn(auth))
	if err != nil {
		logger.Fatal("Error creating Hoarder API client.", "error", err)
	}

	return &Hoarder{ClientWithResponses: hoarderClient}
}

// CreateBookmark creates a new bookmark in Hoarder.
func (h Hoarder) CreateBookmark(ctx context.Context, b BookmarkType) (*HoarderBookmark, error) {
	// Parse the JSON body of the request
	body, err := b.ToJSONReader()
	if err != nil {
		return nil, err
	}

	// Create bookmark
	response, err := h.PostBookmarksWithBodyWithResponse(ctx, "application/json", body)
	if err != nil {
		return nil, err
	}

	// Check if the bookmark was created successfully
	if response.StatusCode() != http.StatusCreated {
		return nil, fmt.Errorf("received HTTP status: %s", response.Status())
	}

	// Return bookmark
	bookmark := HoarderBookmark(*response.JSON201)
	return &bookmark, nil
}

// RetrieveBookmarkById retrieves a bookmark by its ID.
func (h Hoarder) RetrieveBookmarkById(ctx context.Context, id string) (*HoarderBookmark, error) {
	// Retrieve bookmark
	response, err := h.GetBookmarksBookmarkIdWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if the bookmark was created successfully
	if response.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("received HTTP status: %s", response.Status())
	}

	// Return bookmark
	bookmark := HoarderBookmark(*response.JSON200)
	return &bookmark, nil
}
