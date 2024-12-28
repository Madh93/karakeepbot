package hoarderbot

import (
	"fmt"

	"github.com/go-telegram/bot/models"
)

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

	return attrs
}

// AttrsWithError returns a slice of logging attributes for the message with an
// error.
func (tm TelegramMessage) AttrsWithError(err error) []any {
	return append(tm.Attrs(), "error", err)
}
