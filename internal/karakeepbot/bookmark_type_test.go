package karakeepbot

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageBookmark_ToJSONReader(t *testing.T) {
	tests := []struct {
		name               string
		bookmark           ImageBookmark
		expectedJSONOutput map[string]interface{} // Using map for easier comparison of JSON structure
		expectError        bool
	}{
		{
			name: "ImageBookmark with caption and KaraKeepImageURL",
			bookmark: ImageBookmark{
				Text:             "A beautiful sunset",
				KaraKeepImageURL: "http://karakeep.example.com/image/123.jpg",
				BestPhotoFileID:  "telegram_file_id_best",
				Photos:           []TelegramPhotoSize{{FileID: "telegram_file_id_best", FileSize: 1024}},
			},
			expectedJSONOutput: map[string]interface{}{
				"type":      "image",
				"text":      "A beautiful sunset",
				"image_url": "http://karakeep.example.com/image/123.jpg",
			},
			expectError: false,
		},
		{
			name: "ImageBookmark with no caption",
			bookmark: ImageBookmark{
				KaraKeepImageURL: "http://karakeep.example.com/image/456.png",
				BestPhotoFileID:  "telegram_file_id_another",
				Photos:           []TelegramPhotoSize{{FileID: "telegram_file_id_another", FileSize: 2048}},
			},
			expectedJSONOutput: map[string]interface{}{
				"type":      "image",
				// "text" field should be omitted by omitempty
				"image_url": "http://karakeep.example.com/image/456.png",
			},
			expectError: false,
		},
		{
			name: "ImageBookmark with no KaraKeepImageURL (e.g., before upload)",
			bookmark: ImageBookmark{
				Text:            "A cat picture",
				BestPhotoFileID: "telegram_file_id_cat",
				Photos:          []TelegramPhotoSize{{FileID: "telegram_file_id_cat", FileSize: 512}},
			},
			expectedJSONOutput: map[string]interface{}{
				"type": "image",
				"text": "A cat picture",
				// "image_url" field should be omitted by omitempty
			},
			expectError: false,
		},
		{
			name:     "Empty ImageBookmark (should still produce valid JSON with type)",
			bookmark: ImageBookmark{}, // BestPhotoFileID and Photos are not part of JSON output
			expectedJSONOutput: map[string]interface{}{
				"type": "image",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := tt.bookmark.ToJSONReader()

			if tt.expectError {
				require.Error(t, err, "ToJSONReader should have returned an error")
			} else {
				require.NoError(t, err, "ToJSONReader should not have returned an error")
				require.NotNil(t, reader, "ToJSONReader should have returned a non-nil reader")

				jsonData, readErr := io.ReadAll(reader)
				require.NoError(t, readErr, "Failed to read from io.Reader")

				var actualJSONOutput map[string]interface{}
				unmarshalErr := json.Unmarshal(jsonData, &actualJSONOutput)
				require.NoError(t, unmarshalErr, "Failed to unmarshal JSON output")

				// Check for expected keys and values
				assert.Equal(t, tt.expectedJSONOutput["type"], actualJSONOutput["type"], "JSON 'type' field mismatch")

				if expectedText, ok := tt.expectedJSONOutput["text"]; ok {
					assert.Equal(t, expectedText, actualJSONOutput["text"], "JSON 'text' field mismatch")
				} else {
					_, actualTextOk := actualJSONOutput["text"]
					assert.False(t, actualTextOk, "JSON 'text' field should be omitted but was present")
				}

				if expectedImageURL, ok := tt.expectedJSONOutput["image_url"]; ok {
					assert.Equal(t, expectedImageURL, actualJSONOutput["image_url"], "JSON 'image_url' field mismatch")
				} else {
					_, actualImageURLOk := actualJSONOutput["image_url"]
					assert.False(t, actualImageURLOk, "JSON 'image_url' field should be omitted but was present")
				}
			}
		})
	}
}

func TestLinkBookmark_ToJSONReader(t *testing.T) {
	lb := NewLinkBookmark("http://example.com/link")
	reader, err := lb.ToJSONReader()
	require.NoError(t, err)
	data, err := io.ReadAll(reader)
	require.NoError(t, err)

	expectedJSON := `{"type":"link","url":"http://example.com/link"}`
	assert.JSONEq(t, expectedJSON, string(data))
}

func TestTextBookmark_ToJSONReader(t *testing.T) {
	tb := NewTextBookmark("This is a text bookmark.")
	reader, err := tb.ToJSONReader()
	require.NoError(t, err)
	data, err := io.ReadAll(reader)
	require.NoError(t, err)

	expectedJSON := `{"type":"text","text":"This is a text bookmark."}`
	assert.JSONEq(t, expectedJSON, string(data))
}
