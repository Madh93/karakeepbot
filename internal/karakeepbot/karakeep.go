package karakeepbot

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

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
