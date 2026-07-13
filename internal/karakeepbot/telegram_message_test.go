package karakeepbot

import (
	"slices"
	"testing"
)

func TestHashtags(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		caption  string
		expected []string
	}{
		{
			name:     "no hashtags",
			text:     "just some text",
			expected: nil,
		},
		{
			name:     "hashtags in text",
			text:     "check this out #golang #programming",
			expected: []string{"golang", "programming"},
		},
		{
			name:     "hashtags in photo caption",
			caption:  "sunset at the beach #travel #photography",
			expected: []string{"travel", "photography"},
		},
		{
			name:     "duplicate hashtags are deduplicated",
			text:     "#golang stuff #golang",
			caption:  "#golang",
			expected: []string{"golang"},
		},
		{
			name:     "unicode hashtags",
			caption:  "#путешествия and #日本語",
			expected: []string{"путешествия", "日本語"},
		},
		{
			name:     "hashtag with underscore and digits",
			text:     "#web_dev #tag123",
			expected: []string{"web_dev", "tag123"},
		},
		{
			name:     "hash without tag name is ignored",
			text:     "# not a tag, nor is #",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := TelegramMessage{Text: tt.text, Caption: tt.caption}
			got := msg.Hashtags()
			if !slices.Equal(got, tt.expected) {
				t.Errorf("Hashtags() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
