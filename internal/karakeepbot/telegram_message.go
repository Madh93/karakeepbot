package karakeepbot

import (
	"fmt"

	"github.com/go-telegram/bot/models"
)

// TelegramUpdate is an alias for models.Update.
type TelegramUpdate = models.Update

// TelegramUser represents a user in Telegram.
// Forward declaration, assuming it's compatible with models.User or defined elsewhere.
type TelegramUser models.User

// TelegramChat represents a chat in Telegram.
// Forward declaration, assuming it's compatible with models.Chat or defined elsewhere.
type TelegramChat models.Chat

// TelegramPhotoSize represents the size of a photo sent in a Telegram message.
type TelegramPhotoSize struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	FileSize     int    `json:"file_size,omitempty"`
}

// TelegramMessage represents a message received from the Telegram bot API.
// It's a custom struct to allow adding fields like Photo.
type TelegramMessage struct {
	ID              int                 `json:"message_id"`
	From            *TelegramUser       `json:"from,omitempty"` // models.User
	Date            int                 `json:"date"`
	Chat            *TelegramChat       `json:"chat"` // models.Chat
	MessageThreadID int                 `json:"message_thread_id,omitempty"`
	Text            string              `json:"text,omitempty"`
	Document        *models.Document    `json:"document,omitempty"`
	Photo           []TelegramPhotoSize `json:"photo,omitempty"` // New field for photos
	// Fields used by Attrs() that were part of original models.Message
	IsTopicMessage bool `json:"is_topic_message,omitempty"`
}

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

	// Accessing tm.Photo directly now as it's part of TelegramMessage struct
	if tm.Photo != nil && len(tm.Photo) > 0 {
		for i, photo := range tm.Photo {
			attrs = append(attrs, fmt.Sprintf("photo_id_%d", i), photo.FileID)
			attrs = append(attrs, fmt.Sprintf("photo_size_%d", i), photo.FileSize) // Assuming FileSize is the relevant size for logging
			attrs = append(attrs, fmt.Sprintf("photo_width_%d", i), photo.Width)
			attrs = append(attrs, fmt.Sprintf("photo_height_%d", i), photo.Height)
		}
	}

	return attrs
}

// AttrsWithError returns a slice of logging attributes for the message with an
// error.
func (tm TelegramMessage) AttrsWithError(err error) []any {
	return append(tm.Attrs(), "error", err)
}
