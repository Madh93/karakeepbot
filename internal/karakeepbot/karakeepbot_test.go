package karakeepbot

import (
	"testing"

	"github.com/Madh93/karakeepbot/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMessage_ImageBookmark(t *testing.T) {
	tests := []struct {
		name             string
		message          TelegramMessage
		expectedCaption  string
		expectImage      bool
		expectedBestFileID string // Only if expectImage is true
	}{
		{
			name: "message with photo and caption",
			message: TelegramMessage{
				Text: "Test caption",
				Photo: []TelegramPhotoSize{
					{FileID: "file1", FileSize: 1000, Width: 100, Height: 100},
					{FileID: "file2", FileSize: 2000, Width: 200, Height: 200}, // Best
				},
			},
			expectedCaption:  "Test caption",
			expectImage:      true,
			expectedBestFileID: "file2",
		},
		{
			name: "message with photo and no caption",
			message: TelegramMessage{
				Photo: []TelegramPhotoSize{
					{FileID: "file_A", FileSize: 500, Width: 50, Height: 50},
				},
			},
			expectedCaption:  "",
			expectImage:      true,
			expectedBestFileID: "file_A",
		},
		{
			name: "message with multiple photos, largest is first",
			message: TelegramMessage{
				Text: "Another caption",
				Photo: []TelegramPhotoSize{
					{FileID: "best_file", FileSize: 3000, Width: 300, Height: 300}, // Best
					{FileID: "not_best_file", FileSize: 100, Width: 10, Height: 10},
				},
			},
			expectedCaption:  "Another caption",
			expectImage:      true,
			expectedBestFileID: "best_file",
		},
		{
			name: "message with photos, some with zero size",
			message: TelegramMessage{
				Photo: []TelegramPhotoSize{
					{FileID: "zero_size1", FileSize: 0, Width: 300, Height: 300},
					{FileID: "actual_best", FileSize: 150, Width: 10, Height: 10}, // Best
					{FileID: "zero_size2", FileSize: 0, Width: 400, Height: 400},
				},
			},
			expectedCaption:  "",
			expectImage:      true,
			expectedBestFileID: "actual_best",
		},
		{
			name: "message with photos, all zero size (fallback to last)",
			message: TelegramMessage{
				Photo: []TelegramPhotoSize{
					{FileID: "zero1", FileSize: 0, Width: 300, Height: 300},
					{FileID: "zero2", FileSize: 0, Width: 10, Height: 10},
					{FileID: "zero3_last", FileSize: 0, Width: 400, Height: 400}, // Fallback
				},
			},
			expectedCaption:  "",
			expectImage:      true,
			expectedBestFileID: "zero3_last",
		},
		{
			name: "message with empty photo slice",
			message: TelegramMessage{
				Text:  "Just text",
				Photo: []TelegramPhotoSize{},
			},
			expectImage: false, // Should be TextBookmark
		},
		{
			name: "message with nil photo slice",
			message: TelegramMessage{
				Text:  "More text",
				Photo: nil,
			},
			expectImage: false, // Should be TextBookmark
		},
		{
			name: "message with URL",
			message: TelegramMessage{
				Text: "http://example.com",
			},
			expectImage: false, // Should be LinkBookmark
		},
		{
			name: "message with plain text",
			message: TelegramMessage{
				Text: "Hello world",
			},
			expectImage: false, // Should be TextBookmark
		},
		{
			name: "empty message (no text, no photo, no document)",
			message: TelegramMessage{},
			expectImage: false, // Should be error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bookmark, err := parseMessage(tt.message)

			if tt.expectImage {
				require.NoError(t, err, "parseMessage should not return an error for image case")
				imgBookmark, ok := bookmark.(*ImageBookmark)
				require.True(t, ok, "Bookmark should be of type *ImageBookmark")
				assert.Equal(t, tt.expectedCaption, imgBookmark.Text, "ImageBookmark caption mismatch")
				assert.Equal(t, tt.expectedBestFileID, imgBookmark.BestPhotoFileID, "ImageBookmark BestPhotoFileID mismatch")
				// Check that Photos field is also populated
				assert.Equal(t, len(tt.message.Photo), len(imgBookmark.Photos), "ImageBookmark Photos slice length mismatch")
				if len(tt.message.Photo) > 0 {
					assert.EqualValues(t, tt.message.Photo, imgBookmark.Photos, "ImageBookmark Photos content mismatch")
				}

			} else {
				// Handle other types or errors
				if tt.message.Text == "" && tt.message.Photo == nil && tt.message.Document == nil { // Empty message case
					require.Error(t, err, "parseMessage should return an error for an empty message")
				} else if validation.ValidateURL(tt.message.Text) == nil { // Link case
					require.NoError(t, err)
					_, ok := bookmark.(*LinkBookmark)
					assert.True(t, ok, "Bookmark should be of type *LinkBookmark for URL")
				} else if tt.message.Text != "" { // Text case (includes empty photo slice or nil photo)
					require.NoError(t, err)
					textBookmark, ok := bookmark.(*TextBookmark)
					assert.True(t, ok, "Bookmark should be of type *TextBookmark")
					assert.Equal(t, tt.message.Text, textBookmark.Text)
				} else {
					// This case should ideally not be reached if logic is correct,
					// but it's a fallback for unexpected non-image, non-error scenarios.
					require.NoError(t, err, "parseMessage returned an unexpected error")
					assert.NotNil(t, bookmark, "Bookmark should not be nil for non-error, non-image cases")
				}
			}
		})
	}
}

// Minimal test for NewImageBookmark directly, focusing on BestPhotoFileID selection.
// More detailed tests are implicitly covered by TestParseMessage_ImageBookmark
// as parseMessage calls NewImageBookmark.
func TestNewImageBookmark_BestPhotoSelection(t *testing.T) {
	photos := []TelegramPhotoSize{
		{FileID: "small", FileSize: 100, Width: 10, Height: 10},
		{FileID: "large_but_zero_explicit_size", FileSize: 0, Width: 1000, Height: 1000},
		{FileID: "medium", FileSize: 500, Width: 50, Height: 50}, // Actual best by FileSize
	}
	caption := "test"
	bookmark := NewImageBookmark(photos, caption)

	assert.Equal(t, "medium", bookmark.BestPhotoFileID)
	assert.Equal(t, caption, bookmark.Text)
	assert.EqualValues(t, photos, bookmark.Photos)

	photosOnlyZero := []TelegramPhotoSize{
		{FileID: "zero1", FileSize: 0, Width: 10, Height: 10},
		{FileID: "zero2_last", FileSize: 0, Width: 1000, Height: 1000},
	}
	bookmarkZero := NewImageBookmark(photosOnlyZero, "caption_zero")
	assert.Equal(t, "zero2_last", bookmarkZero.BestPhotoFileID, "Fallback for all zero FileSize should pick last")

	var nilPhotos []TelegramPhotoSize
	bookmarkNil := NewImageBookmark(nilPhotos, "nil_photos")
	assert.Equal(t, "", bookmarkNil.BestPhotoFileID, "BestPhotoFileID should be empty for nil photos")

	emptyPhotos := []TelegramPhotoSize{}
	bookmarkEmpty := NewImageBookmark(emptyPhotos, "empty_photos")
	assert.Equal(t, "", bookmarkEmpty.BestPhotoFileID, "BestPhotoFileID should be empty for empty photo slice")
}
