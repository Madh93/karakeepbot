package karakeepbot

import (
	"strings"

	"github.com/Madh93/go-karakeep"
)

// KarakeepBookmark represents a bookmark received from the karakeep API.
type KarakeepBookmark karakeep.Bookmark

// Attrs returns a slice of logging attributes for the bookmark.
func (kb KarakeepBookmark) Attrs() []any {
	return []any{
		"scope", "karakeep",
		"bookmark_id", kb.Id,
		"tagging_status", *kb.TaggingStatus,
	}
}

// AttrsWithError returns a slice of logging attributes for the bookmark with an
// error.
func (kb KarakeepBookmark) AttrsWithError(err error) []any {
	return append(kb.Attrs(), "error", err)
}

// Hashtags returns a string of hashtags associated with the bookmark.
func (kb KarakeepBookmark) Hashtags() string {
	var tags []string
	for _, tag := range kb.Tags {
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
