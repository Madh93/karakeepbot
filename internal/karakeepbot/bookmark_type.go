package karakeepbot

import (
	"bytes"
	"encoding/json"
	"io"
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
