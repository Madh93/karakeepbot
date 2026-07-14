package karakeepbot

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/go-telegram/bot/models"
)

// hashtagRegexp matches Telegram-style hashtags, capturing the tag name
// without the leading '#'. Supports unicode letters and digits.
var hashtagRegexp = regexp.MustCompile(`#([\p{L}\p{N}_]+)`)

// TelegramUpdate is an alias for models.Update.
type TelegramUpdate = models.Update

// TelegramMessage represents a message received from the Telegram bot API.
type TelegramMessage models.Message

// Attrs returns a slice of logging attributes for the message.
func (tm TelegramMessage) Attrs() []any {
	attrs := []any{
		"scope", "telegram",
		"chat_id", tm.Chat.ID,
		"user_id", tm.From.ID,
		"username", tm.From.Username,
		"message_id", tm.ID,
	}

	if tm.IsTopicMessage {
		attrs = append(attrs, "message_thread_id", tm.MessageThreadID)
	}

	if tm.Text != "" {
		attrs = append(attrs, "message_text", tm.Text)
	}

	if tm.Document != nil {
		attrs = append(attrs, "document_id", tm.Document.FileID)
		attrs = append(attrs, "document_name", tm.Document.FileName)
		attrs = append(attrs, "document_size", tm.Document.FileSize)
		attrs = append(attrs, "document_type", tm.Document.MimeType)
	}

	if tm.Photo != nil {
		for i, photo := range tm.Photo {
			attrs = append(attrs, fmt.Sprintf("photo_id_%d", i), photo.FileID)
			attrs = append(attrs, fmt.Sprintf("photo_size_%d", i), photo.FileSize)
		}
	}

	if tm.ForwardOrigin != nil {
		if sourceType, displayName := tm.authorInfo(); sourceType != "" {
			attrs = append(attrs, "forward_origin", fmt.Sprintf("%s %s", sourceType, displayName))
		}
	}

	return attrs
}

// AttrsWithError returns a slice of logging attributes for the message with an
// error.
func (tm TelegramMessage) AttrsWithError(err error) []any {
	return append(tm.Attrs(), "error", err)
}

// ExtractURL returns the first URL found in message entities of type text_link.
// Returns empty string if no text_link entity exists.
func (tm TelegramMessage) ExtractURL() string {
	for _, entity := range tm.Entities {
		if entity.Type == models.MessageEntityTypeTextLink && entity.URL != "" {
			return entity.URL
		}
	}
	return ""
}

// EntityURLs returns all unique URLs found in message entities of type text_link.
func (tm TelegramMessage) EntityURLs() []string {
	seen := make(map[string]struct{})
	var urls []string
	for _, entity := range tm.Entities {
		if entity.Type == models.MessageEntityTypeTextLink && entity.URL != "" {
			if _, ok := seen[entity.URL]; !ok {
				seen[entity.URL] = struct{}{}
				urls = append(urls, entity.URL)
			}
		}
	}
	return urls
}

// Hashtags returns the unique hashtags found in the message text and photo
// caption, without the leading '#'.
func (tm TelegramMessage) Hashtags() []string {
	seen := make(map[string]struct{})
	var tags []string
	for _, source := range []string{tm.Text, tm.Caption} {
		for _, match := range hashtagRegexp.FindAllStringSubmatch(source, -1) {
			tag := match[1]
			if _, ok := seen[tag]; !ok {
				seen[tag] = struct{}{}
				tags = append(tags, tag)
			}
		}
	}
	return tags
}

// authorInfo returns the origin type and display name for the message author.
// For forwarded messages it uses the forward origin; for direct messages it
// falls back to the sender. Returns empty strings when no author info exists.
func (tm TelegramMessage) authorInfo() (sourceType, displayName string) {
	switch {
	case tm.ForwardOrigin != nil && tm.ForwardOrigin.MessageOriginChannel != nil:
		ch := tm.ForwardOrigin.MessageOriginChannel
		if ch.Chat.Username != "" {
			return "channel", "@" + ch.Chat.Username
		}
		return "channel", ch.Chat.Title
	case tm.ForwardOrigin != nil && tm.ForwardOrigin.MessageOriginUser != nil:
		u := tm.ForwardOrigin.MessageOriginUser.SenderUser
		if u.Username != "" {
			return "user", "@" + u.Username
		}
		return "user", strings.TrimSpace(u.FirstName + " " + u.LastName)
	case tm.ForwardOrigin != nil && tm.ForwardOrigin.MessageOriginHiddenUser != nil:
		return "hidden_user", tm.ForwardOrigin.MessageOriginHiddenUser.SenderUserName
	case tm.ForwardOrigin != nil && tm.ForwardOrigin.MessageOriginChat != nil:
		sc := tm.ForwardOrigin.MessageOriginChat.SenderChat
		if sc.Username != "" {
			return "chat", "@" + sc.Username
		}
		return "chat", sc.Title
	case tm.From != nil:
		if tm.From.Username != "" {
			return "direct", "@" + tm.From.Username
		}
		return "direct", strings.TrimSpace(tm.From.FirstName + " " + tm.From.LastName)
	}
	return "", ""
}

// ChannelPostLink constructs the t.me link for a forwarded channel post.
// Returns "https://t.me/{username}/{messageID}" or empty string if the
// message is not forwarded from a channel with a username.
func (tm TelegramMessage) ChannelPostLink() string {
	if tm.ForwardOrigin == nil || tm.ForwardOrigin.MessageOriginChannel == nil {
		return ""
	}
	ch := tm.ForwardOrigin.MessageOriginChannel
	if ch.Chat.Username == "" {
		return ""
	}
	return fmt.Sprintf("https://t.me/%s/%d", ch.Chat.Username, ch.MessageID)
}

// ContextNote builds a short note describing the Telegram origin of the
// bookmark.
func (tm TelegramMessage) ContextNote() string {
	var b strings.Builder

	b.WriteString("📎 From Telegram\n")

	if _, displayName := tm.authorInfo(); displayName != "" {
		fmt.Fprintf(&b, "✍️ %s\n", displayName)
	}

	if tm.Chat.Title != "" {
		fmt.Fprintf(&b, "💬 %s\n", tm.Chat.Title)
	}

	if tm.Date > 0 {
		t := time.Unix(int64(tm.Date), 0)
		fmt.Fprintf(&b, "📅 %s\n", t.Format("2006-01-02 15:04"))
	}

	return strings.TrimRight(b.String(), "\n")
}

// MessageTime returns the message timestamp as time.Time.
func (tm TelegramMessage) MessageTime() time.Time {
	return time.Unix(int64(tm.Date), 0)
}
