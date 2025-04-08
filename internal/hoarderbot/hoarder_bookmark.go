package hoarderbot

import (
	"strings"

	"github.com/Madh93/go-karakeep"
)

// HoarderBookmark represents a bookmark received from the hoarder API.
type HoarderBookmark karakeep.Bookmark

// Attrs returns a slice of logging attributes for the bookmark.
func (hb HoarderBookmark) Attrs() []any {
	return []any{
		"scope", "hoarder",
		"bookmark_id", hb.Id,
		"tagging_status", *hb.TaggingStatus,
	}
}

// AttrsWithError returns a slice of logging attributes for the bookmark with an
// error.
func (hb HoarderBookmark) AttrsWithError(err error) []any {
	return append(hb.Attrs(), "error", err)
}

// Hashtags returns a string of hashtags associated with the bookmark.
func (hb HoarderBookmark) Hashtags() string {
	var tags []string
	for _, tag := range hb.Tags {
		name := sanitizeTag(tag.Name)
		tags = append(tags, "#"+name)
	}
	return strings.Join(tags, " ")
}

// sanitizeTag removes any spaces or hyphens from the tag name.
func sanitizeTag(tag string) string {
	tag = strings.ReplaceAll(tag, " ", "")
	return strings.ReplaceAll(tag, "-", "")
}
