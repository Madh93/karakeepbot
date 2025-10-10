package karakeepbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// AssetType defines the type of an asset bookmark.
type AssetType string

const (
	// ImageAssetType represents an image asset.
	ImageAssetType AssetType = "image"
)

// BookmarkType is an interface that represents any type of bookmark.
// It is useful for handling different bookmark types polymorphically.
type BookmarkType interface {
	String() string
}

// Bookmark is a base struct embedded in other bookmark types
// to share common fields.
type Bookmark struct {
	Type string `json:"type"`
}

// LinkBookmark represents a bookmark with a URL.
type LinkBookmark struct {
	Bookmark
	URL string `json:"url"`
}

// NewLinkBookmark creates a new LinkBookmark with the given URL.
func NewLinkBookmark(url string) *LinkBookmark {
	return &LinkBookmark{
		Bookmark: Bookmark{Type: "link"},
		URL:      url,
	}
}

// String returns a human-readable representation of the LinkBookmark.
func (lb LinkBookmark) String() string {
	return fmt.Sprintf("LinkBookmark (URL: %s)", lb.URL)
}

// TextBookmark represents a bookmark with text content.
type TextBookmark struct {
	Bookmark
	Text string `json:"text"`
}

// NewTextBookmark creates a new TextBookmark with the given text content.
func NewTextBookmark(text string) *TextBookmark {
	return &TextBookmark{
		Bookmark: Bookmark{Type: "text"},
		Text:     text,
	}
}

// String returns a human-readable representation of the TextBookmark.
func (tb TextBookmark) String() string {
	return fmt.Sprintf("TextBookmark (Text: %.30s...)", tb.Text)
}

// AssetBookmark represents a bookmark with an asset.
type AssetBookmark struct {
	Bookmark
	AssetID   string    `json:"assetId"`
	AssetType AssetType `json:"assetType"`
	Title     string    `json:"title,omitempty"`
	Note      string    `json:"note,omitempty"`
}

// NewAssetBookmark creates a new AssetBookmark for a given asset.
func NewAssetBookmark(assetID string, assetType AssetType, note string) *AssetBookmark {
	title := note

	// Truncate title to avoid max length (https://github.com/karakeep-app/karakeep/commit/aecbe6ae8b3dbc7bcdcf33f1c8c086dafb77eb24#diff-18ac11cb95d04713123a9df4fb60e90f9ee5ecada941c2aa1b6f72eb1c8bb674R6-R136)
	runes := []rune(note)
	if len(runes) > 150 {
		title = string(runes[:150]) + "..."
	}

	return &AssetBookmark{
		Bookmark:  Bookmark{Type: "asset"},
		AssetID:   assetID,
		AssetType: assetType,
		Title:     title,
		Note:      note,
	}
}

// String returns a human-readable representation of the AssetBookmark.
func (ab AssetBookmark) String() string {
	return fmt.Sprintf("AssetBookmark (Type: %s, ID: %s)", ab.AssetType, ab.AssetID)
}

// ToJSONReader converts any BookmarkType into an io.Reader containing its
// JSON representation.
func ToJSONReader(b BookmarkType) (io.Reader, error) {
	data, err := json.Marshal(b)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal bookmark to JSON: %w", err)
	}
	return bytes.NewReader(data), nil
}
