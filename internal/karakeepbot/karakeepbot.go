// Package karakeepbot implements a Telegram bot that allows users to create
// bookmarks through messages. The bot interacts with the Karakeep API to manage
// bookmarks and handles incoming messages by checking if the chat ID is
// allowed, creating bookmarks, and sending back updated messages with tags.
package karakeepbot

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"time"

	"github.com/Madh93/go-karakeep"
	"github.com/Madh93/karakeepbot/internal/config"
	"github.com/Madh93/karakeepbot/internal/logging"
	"github.com/Madh93/karakeepbot/internal/validation"
)

	"path/filepath"
)

// Config holds the configuration for the KarakeepBot.
type Config struct {
	Karakeep *config.KarakeepConfig
	Telegram *config.TelegramConfig
	TempDir  string // Add TempDir here
}

// KarakeepBot represents the bot with its dependencies, including the Karakeep
// client, Telegram bot, logger and other options.
type KarakeepBot struct {
	karakeep     *Karakeep
	telegram     *Telegram
	logger       *logging.Logger
	allowlist    []int64
	threads      []int
	waitInterval int
	tempDir      string // New field for temporary directory path
}

// New creates a new KarakeepBot instance, initializing the Karakeep and Telegram
// clients.
func New(logger *logging.Logger, config *Config) *KarakeepBot { // config here is karakeepbot.Config
	return &KarakeepBot{
		karakeep:     createKarakeep(logger, config.Karakeep),
		telegram:     createTelegram(logger, config.Telegram),
		allowlist:    config.Telegram.Allowlist,
		threads:      config.Telegram.Threads,
		waitInterval: config.Karakeep.Interval,
		logger:       logger,
		tempDir:      config.TempDir, // Initialize tempDir
	}
}

// Run starts the bot and handles incoming messages.
func (kb *KarakeepBot) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Set default handler
	kb.telegram.RegisterHandlerMatchFunc(func(*TelegramUpdate) bool { return true }, kb.handler)

	// Start the bot
	kb.telegram.Start(ctx)

	return nil
}

// handler is the main handler for incoming messages. It processes the message
// and sends a response back to the user.
func (kb KarakeepBot) handler(ctx context.Context, _ *Bot, update *TelegramUpdate) {
	if update.Message == nil {
		return
	}

	// Initial conversion (existing code)
	tgMsg := update.Message // *models.Message

	// Create our TelegramMessage
	// Ensure all necessary fields from tgMsg are copied to msg.
	// Based on the struct definition in telegram_message.go and its usage in Attrs().
	msg := TelegramMessage{
		ID:              tgMsg.ID,
		From:            (*TelegramUser)(tgMsg.From),
		Date:            tgMsg.Date,
		Chat:            (*TelegramChat)(tgMsg.Chat),
		MessageThreadID: tgMsg.MessageThreadID,
		Text:            tgMsg.Text,
		Document:        tgMsg.Document,       // Keep document if it was there
		IsTopicMessage:  tgMsg.IsTopicMessage, // Keep IsTopicMessage
		// Photo field will be populated next
	}

	// Populate the new Photo field
	if tgMsg.Photo != nil && len(tgMsg.Photo) > 0 {
		msg.Photo = make([]TelegramPhotoSize, len(tgMsg.Photo))
		for i, p := range tgMsg.Photo {
			msg.Photo[i] = TelegramPhotoSize{
				FileID:       p.FileID,
				FileUniqueID: p.FileUniqueID,
				Width:        p.Width,
				Height:       p.Height,
				FileSize:     p.FileSize,
			}
		}
	}
	// Now msg.Photo is populated.

	// Check if the chat ID is allowed
	if !kb.isChatIdAllowed(msg.Chat.ID) {
		kb.logger.Warn(fmt.Sprintf("Received message from not allowed chat ID. Allowed chats IDs: %v", kb.allowlist), msg.Attrs()...)
		return
	}

	// Check if the thread ID is allowed
	if !kb.isThreadIdAllowed(msg.MessageThreadID) {
		kb.logger.Warn(fmt.Sprintf("Received message from not allowed thread ID. Allowed thread IDs: %v", kb.threads), msg.Attrs()...)
		return
	}

	kb.logger.Debug("Received message from allowed chat ID and allowed thread ID", msg.Attrs()...)

	// Parse the message to get corresponding bookmark type
	kb.logger.Debug("Parsing message to get corresponding bookmark type", msg.Attrs()...)
	b, err := parseMessage(msg)
	if err != nil {
		kb.logger.Error("Failed to parse message", msg.AttrsWithError(err)...)
		return
	}

	// Type switch to handle different bookmark types
	switch bookmark := b.(type) {
	case *ImageBookmark:
		kb.logger.Debug("Processing ImageBookmark", msg.Attrs("caption", bookmark.Text)...)
		if bookmark.BestPhotoFileID == "" {
			kb.logger.Error("ImageBookmark has no FileID to download", msg.Attrs()...)
			// Optionally send an error message to the user via Telegram if desired
			return
		}

		fileName := bookmark.BestPhotoFileID + ".jpg" // Assuming .jpg, might need better logic
		localTempPath := filepath.Join(kb.tempDir, fileName)

		kb.logger.Debug("Downloading image from Telegram", msg.Attrs("file_id", bookmark.BestPhotoFileID, "temp_path", localTempPath)...)
		err = kb.telegram.DownloadFile(ctx, bookmark.BestPhotoFileID, localTempPath)
		if err != nil {
			kb.logger.Error("Failed to download image from Telegram", msg.AttrsWithError(err)...)
			return
		}
		kb.logger.Info("Image downloaded from Telegram", msg.Attrs("path", localTempPath)...)

		defer func() {
			kb.logger.Debug("Attempting to delete temporary image file", msg.Attrs("path", localTempPath)...)
			if err := os.Remove(localTempPath); err != nil {
				kb.logger.Warn("Failed to delete temporary image file", msg.AttrsWithError(err)...)
			} else {
				kb.logger.Info("Temporary image file deleted", msg.Attrs("path", localTempPath)...)
			}
		}()

		karaKeepImageURL, err := kb.karakeep.UploadImageToKaraKeep(ctx, localTempPath)
		if err != nil {
			kb.logger.Error("Failed to upload image to KaraKeep", msg.AttrsWithError(err)...)
			return
		}
		kb.logger.Info("Image uploaded to KaraKeep", msg.Attrs("karakeep_url", karaKeepImageURL)...)
		bookmark.KaraKeepImageURL = karaKeepImageURL

		createdKaraKeepBookmark, err := kb.karakeep.CreateBookmark(ctx, bookmark)
		if err != nil {
			kb.logger.Error("Failed to create image bookmark in KaraKeep", msg.AttrsWithError(err)...)
			return
		}
		kb.logger.Info("Created image bookmark in KaraKeep", createdKaraKeepBookmark.Attrs()...)

		err = kb.processKaraKeepBookmark(ctx, &msg, createdKaraKeepBookmark, bookmark.Text) // Pass original caption
		if err != nil {
			// Logged within processKaraKeepBookmark
			return
		}

	case *LinkBookmark, *TextBookmark:
		createdKaraKeepBookmark, err := kb.karakeep.CreateBookmark(ctx, b) // b is the original bookmark type
		if err != nil {
			kb.logger.Error("Failed to create bookmark", "error", err, msg.Attrs()...)
			return
		}
		kb.logger.Info("Created bookmark", createdKaraKeepBookmark.Attrs()...)

		originalText := ""
		if linkBookmark, ok := b.(*LinkBookmark); ok {
			originalText = linkBookmark.URL
		} else if textBookmark, ok := b.(*TextBookmark); ok {
			originalText = textBookmark.Text
		}

		err = kb.processKaraKeepBookmark(ctx, &msg, createdKaraKeepBookmark, originalText)
		if err != nil {
			// Logged within processKaraKeepBookmark
			return
		}

	default:
		kb.logger.Error(fmt.Sprintf("Unhandled bookmark type: %T", b), msg.Attrs()...)
		return
	}
}

// processKaraKeepBookmark handles the logic after a bookmark is successfully created in Karakeep.
// originalTextOrCaption is the text used for the reply message before tags. For images, this would be the caption.
func (kb KarakeepBot) processKaraKeepBookmark(ctx context.Context, msg *TelegramMessage, createdBookmark *KarakeepBookmark, originalTextOrCaption string) error {
	kb.logger.Debug("Waiting for bookmark tags to be updated", createdBookmark.Attrs()...)
	retrievedBookmark := createdBookmark // Start with the created one
	var err error
	for {
		retrievedBookmark, err = kb.karakeep.RetrieveBookmarkById(ctx, retrievedBookmark.Id) // Use retrievedBookmark.Id
		if err != nil {
			kb.logger.Error("Failed to retrieve bookmark during polling", "error", err, createdBookmark.Attrs()...)
			return err // Don't proceed if retrieval fails
		}
		if *retrievedBookmark.TaggingStatus == karakeep.Success {
			kb.logger.Info("Bookmark tags successfully updated", retrievedBookmark.Attrs()...)
			break
		} else if *retrievedBookmark.TaggingStatus == karakeep.Failure {
			kb.logger.Error("Failed to update bookmark tags (tagging_status: failure)", retrievedBookmark.Attrs()...)
			break // Exit polling loop on failure
		}
		kb.logger.Debug(fmt.Sprintf("Bookmark tagging still pending, waiting %d seconds", kb.waitInterval), retrievedBookmark.Attrs()...)
		time.Sleep(time.Duration(kb.waitInterval) * time.Second)
	}

	replyText := retrievedBookmark.Hashtags()
	if originalTextOrCaption != "" {
		replyText = fmt.Sprintf("%s\n\n%s", originalTextOrCaption, retrievedBookmark.Hashtags())
	}
	if *retrievedBookmark.TaggingStatus == karakeep.Failure {
		replyText = fmt.Sprintf("[Warning: Tagging failed for this item]\n%s", replyText)
	}

	msg.Text = replyText

	kb.logger.Debug("Sending updated message with tags", msg.Attrs()...)
	if err := kb.telegram.SendNewMessage(ctx, msg); err != nil {
		kb.logger.Error("Failed to send new message", msg.AttrsWithError(err)...)
		return err
	}

	kb.logger.Debug("Deleting original message", msg.Attrs()...)
	if err := kb.telegram.DeleteOriginalMessage(ctx, msg); err != nil {
		kb.logger.Error("Failed to delete original message", msg.AttrsWithError(err)...)
		// Log error but don't necessarily make the whole operation fail
	}

	kb.logger.Info("Successfully processed bookmark and updated message", msg.Attrs("final_text", msg.Text)...)
	return nil
}

// isChatIdAllowed checks if the chat ID is allowed to receive messages.
func (kb KarakeepBot) isChatIdAllowed(chatId int64) bool {
	return len(kb.allowlist) == 0 || slices.Contains(kb.allowlist, chatId)
}

// isThreadIdAllowed checks if the thread ID is allowed to receive messages.
func (kb KarakeepBot) isThreadIdAllowed(threadId int) bool {
	return len(kb.threads) == 0 || slices.Contains(kb.threads, threadId)
}

// parseMessage parses the incoming Telegram message and returns the corresponding Bookmark type.
func parseMessage(msg TelegramMessage) (BookmarkType, error) {
	switch {
	case msg.Photo != nil && len(msg.Photo) > 0:
		// Pass msg.Photo (which is []TelegramPhotoSize) and msg.Text (caption)
		return NewImageBookmark(msg.Photo, msg.Text), nil
	case validation.ValidateURL(msg.Text) == nil:
		return NewLinkBookmark(msg.Text), nil
	case msg.Text != "": // Should be after photo check, as photos can have captions (msg.Text)
		return NewTextBookmark(msg.Text), nil
	default:
		return nil, errors.New("unsupported bookmark type or empty message")
	}
}
