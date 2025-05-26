package karakeepbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	// "strings" // Not used in the provided snippet for ImageBookmark
)

// BookmarkType is an interface that represents a bookmark.
type BookmarkType interface {
	String() string
	ToJSONReader() (io.Reader, error)
}

// LinkBookmark represents a bookmark with a URL.
type LinkBookmark struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

// NewLinkBookmark creates a new LinkBookmark with the given URL.
func NewLinkBookmark(url string) *LinkBookmark {
	return &LinkBookmark{
		Type: "link",
		URL:  url,
	}
}

// String returns the string representation of the bookmark.
func (lb LinkBookmark) String() string {
	return "LinkBookmark"
}

// ToJSONReader returns a JSON reader for the bookmark.
func (lb LinkBookmark) ToJSONReader() (io.Reader, error) {
	return toJSONReader(lb)
}

// TextBookmark represents a bookmark with text content.
type TextBookmark struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// NewTextBookmark creates a new TextBookmark with the given text content.
func NewTextBookmark(text string) *TextBookmark {
	return &TextBookmark{
		Type: "text",
		Text: text,
	}
}

// String returns the string representation of the bookmark.
func (tb TextBookmark) String() string {
	return "TextBookmark"
}

// ToJSONReader returns a JSON reader for the bookmark.
func (tb TextBookmark) ToJSONReader() (io.Reader, error) {
	return toJSONReader(tb)
}

// toJSONReader converts a bookmark to JSON and returns an io.Reader.
func toJSONReader(b BookmarkType) (io.Reader, error) {
	data, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}

// ImageBookmark represents a bookmark created from an image.
type ImageBookmark struct {
	Text            string              // Caption from Telegram
	BestPhotoFileID string              `json:"-"` // Internal use for downloading
	Photos          []TelegramPhotoSize `json:"-"` // Internal use for photo selection
	// No KaraKeepImageURL here at initial creation
}

// NewImageBookmark creates a new ImageBookmark.
// It takes the array of photo sizes and the caption.
func NewImageBookmark(photos []TelegramPhotoSize, caption string) *ImageBookmark {
	bestFileID := ""
	// Logic to select the best photo
	if len(photos) > 0 {
		selectedPhoto := photos[0] // Default to the first one
		maxSize := -1              // Initialize maxSize to handle cases where FileSize might be 0 or negative.
		maxPixels := -1            // Initialize maxPixels similarly.

		// First pass: check FileSize
		hasPositiveFileSize := false
		for _, p := range photos {
			if p.FileSize > 0 {
				hasPositiveFileSize = true
				break
			}
		}

		if hasPositiveFileSize {
			// If any photo has a positive FileSize, prioritize by FileSize.
			// Initialize maxSize with a value that ensures any positive FileSize is greater.
			maxSize = 0 // Smallest possible positive FileSize to start comparison.
			// Corrected logic: iterate and find the photo with the largest FileSize.
			// If multiple photos have the same largest FileSize, the first one encountered is chosen.
			// This also handles if selectedPhoto's FileSize is initially 0 but others are positive.
			
			// Find the first photo with positive FileSize to initialize selectedPhoto and maxSize
			// to ensure a valid starting point if photos[0].FileSize is not positive.
			initialIndex := -1
			for i, p := range photos {
				if p.FileSize > 0 {
					initialIndex = i
					break
				}
			}

			if initialIndex != -1 {
				selectedPhoto = photos[initialIndex]
				maxSize = selectedPhoto.FileSize
				for i := initialIndex + 1; i < len(photos); i++ {
					p := photos[i]
					if p.FileSize > maxSize {
						maxSize = p.FileSize
						selectedPhoto = p
					}
				}
			} else { 
				// This else should ideally not be reached if hasPositiveFileSize is true.
				// If somehow all FileSizes are <=0 despite hasPositiveFileSize being true (e.g. data error),
				// it will fall through to pixel logic, or stick with photos[0] if pixel logic also yields nothing.
				// To be safe, if hasPositiveFileSize is true but we didn't find one, re-evaluate.
				// However, the problem implies a fallback, so the current structure is okay.
			}

		}
		
		// Fallback or primary logic if no photo has a positive FileSize.
		// Or, if all FileSizes are 0, this logic will run.
		if !hasPositiveFileSize || maxSize <= 0 { // Check maxSize too in case initial photo had FileSize 0 and no other positive was found
			maxPixels = selectedPhoto.Width * selectedPhoto.Height // Initialize with the first photo's pixels
			for _, p := range photos {
				pixels := p.Width * p.Height
				if pixels > maxPixels {
					maxPixels = pixels
					selectedPhoto = p
				}
			}
		}
		bestFileID = selectedPhoto.FileID
	}

	return &ImageBookmark{
		Text:            caption, // Store caption
		BestPhotoFileID: bestFileID,
		Photos:          photos, // Store all photo variants
	}
}

// ToJSONReader converts the ImageBookmark to an io.Reader for the Karakeep API for initial creation.
func (ib *ImageBookmark) ToJSONReader() (io.Reader, error) {
	// Default text if caption is empty, to ensure the bookmark has some content if required by API
	bookmarkText := ib.Text
	if bookmarkText == "" {
		bookmarkText = "Image from Telegram" // Default text for image bookmark without caption
	}

	payload := struct {
		Text string `json:"text"`
		Type string `json:"type"`
	}{
		Text: bookmarkText,
		Type: "image", // Indicate this is an image type bookmark
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ImageBookmark for initial creation: %w", err)
	}
	return bytes.NewReader(jsonData), nil
}

// String representation for logging
func (ib *ImageBookmark) String() string {
	// KaraKeepImageURL is not known at this stage
	return fmt.Sprintf("ImageBookmark (Caption: '%s', BestPhotoFileID: '%s')", ib.Text, ib.BestPhotoFileID)
}
