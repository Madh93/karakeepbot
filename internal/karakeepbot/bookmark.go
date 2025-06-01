package karakeepbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

type BookmarkType string

type Bookmark struct {
	Type BookmarkType `json:"type"`

	URL  string `json:"url,omitempty"`
	Text string `json:"text,omitempty"`
}

const (
	BookmarkTypeText BookmarkType = "text"
	BookmarkTypeLink BookmarkType = "link"
)

func (b *Bookmark) String() string {

	// This behavior is legacy. I assume that this is used in
	// logging.
	return string(b.Type)
}

func (b *Bookmark) JSONReader() (io.Reader, error) {
	bb, err := json.Marshal(b)
	if err != nil {
		return nil, fmt.Errorf("bookmark failed to marshal: %w", err)
	}

	return bytes.NewReader(bb), nil
}
