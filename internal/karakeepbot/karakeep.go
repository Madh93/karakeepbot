package karakeepbot

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/Madh93/go-karakeep"
	"github.com/Madh93/karakeepbot/internal/config"
	"github.com/Madh93/karakeepbot/internal/kkprivate"
	"github.com/Madh93/karakeepbot/internal/logging"
)

// Karakeep embeds the Karakeep API Client to add high level functionality.
type Karakeep struct {
	*karakeep.ClientWithResponses
	Private *kkprivate.Client
}

// createKarakeep initializes the Karakeep API Client.
func createKarakeep(logger *logging.Logger, config *config.KarakeepConfig) *Karakeep {
	logger.Debug(fmt.Sprintf("Initializing Karakeep API Client at %s using %s token", config.URL, config.Token))

	publicURL, err := url.Parse(config.URL)
	if err != nil {
		logger.Fatal("Error parsing URL.", "error", err)
	}
	publicURL.Path, err = url.JoinPath(publicURL.Path, "/api/v1")
	if err != nil {
		logger.Fatal("Error joining path.", "error", err)
	}

	privateURL, err := url.Parse(config.URL)
	if err != nil {
		logger.Fatal("Error parsing URL.", "error", err)
	}
	privateURL.Path, err = url.JoinPath(privateURL.Path, "/api")
	if err != nil {
		logger.Fatal("Error joining path.", "error", err)
	}

	// Setup Bearer Token Authentication
	auth := func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.Token.Value()))
		return nil
	}

	karakeepClient, err := karakeep.NewClientWithResponses(
		publicURL.String(),
		karakeep.WithRequestEditorFn(auth),
	)
	if err != nil {
		logger.Fatal("Error creating Karakeep API client.", "error", err)
	}

	return &Karakeep{
		ClientWithResponses: karakeepClient,
		Private: &kkprivate.Client{
			URL:        privateURL.String(),
			Token:      config.Token.Value(),
			HTTPClient: &http.Client{},
		},
	}
}

// CreateBookmark creates a new bookmark in Karakeep.
func (k Karakeep) CreateBookmark(ctx context.Context, body io.Reader) (*KarakeepBookmark, error) {

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

// AttachAssetToBookmark attaches an existing asset to an existing bookmark.
func (k Karakeep) AttachAssetToBookmark(
	ctx context.Context,
	bookmarkId string,
	body karakeep.PostBookmarksBookmarkIdAssetsJSONRequestBody,
) error {
	response, err := k.PostBookmarksBookmarkIdAssetsWithResponse(
		ctx,
		bookmarkId,
		body,
	)
	if err != nil {
		return err
	}

	if response.StatusCode() != http.StatusCreated && response.StatusCode() != http.StatusOK {
		return fmt.Errorf("received HTTP status: %s", response.Status())
	}

	return nil
}
