package karakeepbot

import (
	"io"
	"strings"
	"testing"
)

// TestBookmarkConstructors verifies that each constructor function
// correctly initializes its respective bookmark struct.
func TestBookmarkConstructors(t *testing.T) {
	// Test NewLinkBookmark
	t.Run("creates a valid link bookmark", func(t *testing.T) {
		url := "https://example.com"
		bookmark := NewLinkBookmark(url)
		if bookmark.Type != "link" || bookmark.URL != url {
			t.Errorf("NewLinkBookmark(%q) created an invalid bookmark: %+v", url, *bookmark)
		}
	})

	// Test NewTextBookmark
	t.Run("creates a valid text bookmark", func(t *testing.T) {
		text := "A piece of text"
		bookmark := NewTextBookmark(text)
		if bookmark.Type != "text" || bookmark.Text != text {
			t.Errorf("NewTextBookmark(%q) created an invalid bookmark: %+v", text, *bookmark)
		}
	})

	// Test NewAssetBookmark
	t.Run("creates a valid asset bookmark", func(t *testing.T) {
		assetID := "asset-123"
		assetType := ImageAssetType
		bookmark := NewAssetBookmark(assetID, assetType)
		if bookmark.Type != "asset" || bookmark.AssetID != assetID || bookmark.AssetType != assetType {
			t.Errorf("NewAssetBookmark(%q, %q) created an invalid bookmark: %+v", assetID, assetType, *bookmark)
		}
	})
}

// TestBookmarkStringer verifies the String() method for all bookmark types.
func TestBookmarkStringer(t *testing.T) {
	tests := []struct {
		name     string
		bookmark BookmarkType
		expected string
	}{
		{
			name:     "LinkBookmark",
			bookmark: NewLinkBookmark("https://karakeep.app/"),
			expected: "LinkBookmark (URL: https://karakeep.app/)",
		},
		{
			name:     "TextBookmark with long text",
			bookmark: NewTextBookmark("This is a very long text that should definitely be truncated by the String method."),
			expected: "TextBookmark (Text: This is a very long text that ...)",
		},
		{
			name:     "AssetBookmark for image",
			bookmark: NewAssetBookmark("img-uuid-456", ImageAssetType),
			expected: "AssetBookmark (Type: image, ID: img-uuid-456)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.bookmark.String()
			if got != test.expected {
				t.Errorf("For bookmark type %s, expected string %q, but got %q", test.name, test.expected, got)
			}
		})
	}
}

// TestToJSONReader verifies that bookmarks are correctly marshalled to a JSON io.Reader.
func TestToJSONReader(t *testing.T) {
	tests := []struct {
		name     string
		bookmark BookmarkType
		expected string
	}{
		{
			name:     "LinkBookmark to JSON",
			bookmark: NewLinkBookmark("https://github.com"),
			expected: `{"type":"link","url":"https://github.com"}`,
		},
		{
			name:     "TextBookmark to JSON",
			bookmark: NewTextBookmark("Hello World"),
			expected: `{"type":"text","text":"Hello World"}`,
		},
		{
			name:     "AssetBookmark to JSON",
			bookmark: NewAssetBookmark("img-uuid-789", ImageAssetType),
			expected: `{"type":"asset","assetId":"img-uuid-789","assetType":"image"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reader, err := ToJSONReader(test.bookmark)
			if err != nil {
				t.Fatalf("ToJSONReader returned an unexpected error: %v", err)
			}

			jsonData, err := io.ReadAll(reader)
			if err != nil {
				t.Fatalf("Failed to read from reader: %v", err)
			}

			got := strings.TrimSpace(string(jsonData))
			if got != test.expected {
				t.Errorf("For input %s, expected JSON %q, but got %q", test.name, test.expected, got)
			}
		})
	}
}
