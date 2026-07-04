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
	"strings"
	"time"

	"github.com/Madh93/go-karakeep"
	"github.com/Madh93/karakeepbot/internal/config"
	"github.com/Madh93/karakeepbot/internal/fileprocessor"
	"github.com/Madh93/karakeepbot/internal/filevalidator"
	"github.com/Madh93/karakeepbot/internal/logging"
	"github.com/Madh93/karakeepbot/internal/validation"
)

// maxTagRetries is the number of times to wait for Karakeep AI tagging to
// complete before proceeding. At 5s per retry, 6 retries = ~30s timeout.
const maxTagRetries = 6

// KarakeepBot represents the bot with its dependencies, including the Karakeep
// client, Telegram bot, logger and other options.
type KarakeepBot struct {
	karakeep       *Karakeep
	telegram       *Telegram
	logger         *logging.Logger
	fileProcessor  *fileprocessor.Processor
	fileValidators map[string]fileprocessor.Validator
	allowlist      []int64
	threads        []int
	waitInterval   int
}

// New creates a new KarakeepBot instance, initializing the Karakeep and Telegram
// clients.
func New(logger *logging.Logger, config *config.Config) *KarakeepBot {
	// Initialize FileProcessor
	fileProcessor, err := fileprocessor.New(&config.FileProcessor, config.Telegram.ProxyEnabled, config.Telegram.ProxyURL)
	if err != nil {
		logger.Fatal("Failed to create file processor", "error", err)
	}

	// Setup Supported File Validators
	fileValidators := make(map[string]fileprocessor.Validator)
	fileValidators["image/jpeg"] = filevalidator.ImageValidator
	fileValidators["image/png"] = filevalidator.ImageValidator
	fileValidators["image/webp"] = filevalidator.ImageValidator

	// Check if the validators passed in the configuration are supported
	if len(config.FileProcessor.Mimetypes) > 0 {
		for _, mimetype := range config.FileProcessor.Mimetypes {
			if _, supported := fileValidators[mimetype]; !supported {
				logger.Fatal("Configuration error: unsupported MIME type configured", "mime_type", mimetype)
			}
		}
	}

	return &KarakeepBot{
		karakeep:       createKarakeep(logger, &config.Karakeep),
		telegram:       createTelegram(logger, &config.Telegram),
		allowlist:      config.Telegram.Allowlist,
		threads:        config.Telegram.Threads,
		waitInterval:   config.Karakeep.Interval,
		fileProcessor:  fileProcessor,
		fileValidators: fileValidators,
		logger:         logger,
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

	msg := TelegramMessage(*update.Message)

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
	b, err := kb.parseMessage(ctx, msg)
	if err != nil {
		kb.logger.Error("Failed to parse message", msg.AttrsWithError(err)...)
		return
	}

	// Create the bookmark
	kb.logger.Debug(fmt.Sprintf("Creating bookmark of type %s", b))
	bookmark, err := kb.karakeep.CreateBookmark(ctx, b)
	if err != nil {
		kb.logger.Error("Failed to create bookmark", "error", err)
		return
	}
	kb.logger.Info("Created bookmark", bookmark.Attrs()...)

	// Enrich bookmark with Telegram origin metadata
	kb.logger.Debug("Enriching bookmark with Telegram origin metadata", bookmark.Attrs()...)
	kb.enrichBookmark(ctx, msg, bookmark)

	// Wait until bookmark tags are updated (with a timeout to avoid hanging on
	// uncrawlable URLs)
	kb.logger.Debug("Waiting for bookmark tags to be updated", bookmark.Attrs()...)
	bookmark, err = kb.waitForTagCompletion(ctx, bookmark)
	if err != nil {
		kb.logger.Error("Failed to wait for bookmark tagging", "error", err)
		return
	}

	// Get hashtags
	hashtags := bookmark.Hashtags()

	// Send back with hashtags
	if msg.Photo != nil {
		// Add hashtags
		caption := msg.Caption + "\n\n" + hashtags

		// Send back the original photo with hashtags as caption
		kb.logger.Debug("Sending updated message with photo and hashtags", msg.Attrs()...)
		if err := kb.telegram.SendPhotoWithCaption(ctx, &msg, msg.Photo[len(msg.Photo)-1].FileID, caption); err != nil {
			kb.logger.Error("Failed to send photo with caption", msg.AttrsWithError(err)...)
			return
		}
	} else {
		// Add hashtags
		msg.Text = msg.Text + "\n\n" + hashtags

		// Send back new message with hashtags
		kb.logger.Debug("Sending updated message with hashtags", msg.Attrs()...)
		if err := kb.telegram.SendNewMessage(ctx, &msg); err != nil {
			kb.logger.Error("Failed to send new message", msg.AttrsWithError(err)...)
			return
		}
	}

	// Delete original message
	kb.logger.Debug("Deleting original message", msg.Attrs()...)
	if err := kb.telegram.DeleteOriginalMessage(ctx, &msg); err != nil {
		kb.logger.Error("Failed to delete original message", msg.AttrsWithError(err)...)
		return
	}

	kb.logger.Info("Updated message", msg.Attrs()...)
}

// isChatIdAllowed checks if the chat ID is allowed to receive messages.
func (kb KarakeepBot) isChatIdAllowed(chatId int64) bool {
	// When no allowlist is provided, all chat IDs are allowed
	if len(kb.allowlist) == 0 {
		return true
	}

	// When the allowlist provided by environment variable is empty, it contains
	// a single element with value 0.
	if len(kb.allowlist) == 1 && kb.allowlist[0] == 0 {
		return true
	}

	return slices.Contains(kb.allowlist, chatId)
}

// isThreadIdAllowed checks if the thread ID is allowed to receive messages.
func (kb KarakeepBot) isThreadIdAllowed(threadId int) bool {
	return len(kb.threads) == 0 || slices.Contains(kb.threads, threadId)
}

// waitForTagCompletion polls the bookmark tagging status until it succeeds,
// fails, or the retry timeout is reached. Returns the updated bookmark.
func (kb *KarakeepBot) waitForTagCompletion(ctx context.Context, bookmark *KarakeepBookmark) (*KarakeepBookmark, error) {
	retries := 0
	for {
		var err error
		bookmark, err = kb.karakeep.RetrieveBookmarkById(ctx, bookmark.Id)
		if err != nil {
			return nil, err
		}
		if *bookmark.TaggingStatus == karakeep.BookmarkTaggingStatusSuccess {
			return bookmark, nil
		}
		if *bookmark.TaggingStatus == karakeep.BookmarkTaggingStatusFailure {
			return nil, fmt.Errorf("bookmark tagging failed")
		}
		retries++
		if retries >= maxTagRetries {
			kb.logger.Warn("Bookmark tagging did not complete within timeout, proceeding anyway", bookmark.Attrs()...)
			return bookmark, nil
		}
		kb.logger.Debug(fmt.Sprintf("Bookmark is still pending, waiting %d seconds before retrying", kb.waitInterval), bookmark.Attrs()...)
		time.Sleep(time.Duration(kb.waitInterval) * time.Second)
	}
}

// parseMessage parses the incoming Telegram message and returns the
// corresponding Bookmark type.
func (kb KarakeepBot) parseMessage(ctx context.Context, msg TelegramMessage) (BookmarkType, error) {
	if msg.Photo != nil {
		return kb.handlePhotoMessage(ctx, msg)
	}

	if url := msg.ChannelPostLink(); url != "" {
		lb := NewLinkBookmark(url)
		lb.Title = extractTitle(msg.Text)

		var parts []string
		if text := strings.TrimSpace(msg.Text); text != "" {
			parts = append(parts, text)
		}
		if entityURLs := msg.EntityURLs(); len(entityURLs) > 0 {
			parts = append(parts, "Links:\n"+strings.Join(entityURLs, "\n"))
		}
		if ctxNote := msg.ContextNote(); ctxNote != "" {
			parts = append(parts, ctxNote)
		}
		lb.Note = strings.Join(parts, "\n\n")

		return lb, nil
	}

	if url := msg.ExtractURL(); url != "" {
		return newSimpleLinkBookmark(url, msg), nil
	}

	if validation.ValidateURL(msg.Text) == nil {
		return newSimpleLinkBookmark(msg.Text, msg), nil
	}

	if url := extractEmbeddedURL(msg.Text); url != "" {
		return newSimpleLinkBookmark(url, msg), nil
	}

	if msg.Text != "" {
		tb := NewTextBookmark(msg.Text)
		if ctxNote := msg.ContextNote(); ctxNote != "" {
			tb.Note = ctxNote
		}
		return tb, nil
	}

	return nil, errors.New("unsupported bookmark type")
}

// newSimpleLinkBookmark creates a LinkBookmark from a URL and message, with
// title extracted from the first line and a note combining the message text
// with Telegram origin context.
func newSimpleLinkBookmark(url string, msg TelegramMessage) *LinkBookmark {
	lb := NewLinkBookmark(url)
	lb.Title = extractTitle(msg.Text)
	lb.Note = buildNote(msg.Text, msg.ContextNote())
	return lb
}

// extractTitle returns the first line of text as a title, truncated to 150
// characters to avoid Karakeep limits.
func extractTitle(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	lines := strings.SplitN(text, "\n", 2)
	title := strings.TrimSpace(lines[0])
	runes := []rune(title)
	if len(runes) > 150 {
		title = string(runes[:150]) + "..."
	}
	return title
}

// buildNote combines the original message text with a context note.
func buildNote(text, contextNote string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return contextNote
	}
	if contextNote == "" {
		return text
	}
	return text + "\n\n" + contextNote
}

// extractEmbeddedURL finds the first http or https URL within arbitrary text.
// Returns empty string if no URL is found.
func extractEmbeddedURL(text string) string {
	words := strings.Fields(text)
	for _, word := range words {
		word = strings.TrimRight(word, ",.!?:;)]}")
		if err := validation.ValidateURL(word); err == nil {
			return word
		}
	}
	return ""
}

// handlePhotoMessage processes a message containing a photo.
func (kb *KarakeepBot) handlePhotoMessage(ctx context.Context, msg TelegramMessage) (bookmark BookmarkType, err error) {
	// Select the largest photo
	photo := msg.Photo[len(msg.Photo)-1]
	kb.logger.Debug("Handling Telegram image", "file_id", photo.FileID, "file_size", photo.FileSize)

	// Get file URL
	fileURL, err := kb.telegram.GetFileURL(ctx, photo.FileID)
	if err != nil {
		kb.logger.Error("Failed to get file URL", msg.AttrsWithError(err)...)
		if replyErr := kb.telegram.SendReply(ctx, &msg, "⚠️ Failed to process image from Telegram servers, try again later"); replyErr != nil {
			kb.logger.Error("Failed to send reply to user", "reply_error", replyErr)
		}
		return nil, errors.New("couldn't get file URL")
	}

	// Download file. NOTE: Telegram Photo does not have mime type info. We can't use any validator.
	filePath, mimeType, err := kb.fileProcessor.Process(fileURL, nil)
	if err != nil {
		kb.logger.Error("Failed to process image", msg.AttrsWithError(err)...)
		if replyErr := kb.telegram.SendReply(ctx, &msg, "⚠️ Failed to process image"); replyErr != nil {
			kb.logger.Error("Failed to send reply to user", "reply_error", replyErr)
		}
		return nil, errors.New("couldn't process image")
	}
	defer func() {
		if cleanupErr := kb.fileProcessor.Cleanup(filePath); cleanupErr != nil {
			kb.logger.Error("Failed to cleanup temporary file", "path", filePath, "error", cleanupErr)
			if err == nil {
				err = cleanupErr
			}
		}
	}()

	kb.logger.Debug("Detected MIME type", "mime_type", mimeType)

	// Upload asset to Karakeep
	asset, err := kb.karakeep.CreateAsset(ctx, filePath, mimeType)
	if err != nil {
		kb.logger.Error("Failed to upload asset", msg.AttrsWithError(err)...)
		if replyErr := kb.telegram.SendReply(ctx, &msg, "⚠️ Failed to upload asset to Karakeep"); replyErr != nil {
			kb.logger.Error("Failed to send reply to user", "reply_error", replyErr)
		}
		return nil, errors.New("couldn't upload asset")
	}

	kb.logger.Debug("Asset uploaded successfully", "asset_id", asset.AssetId)

	// Get note from caption and append Telegram origin context
	note := strings.TrimSpace(msg.Caption)
	ab := NewAssetBookmark(asset.AssetId, ImageAssetType, note)
	if ctxNote := msg.ContextNote(); ctxNote != "" {
		if ab.Note != "" {
			ab.Note += "\n\n" + ctxNote
		} else {
			ab.Note = ctxNote
		}
	}
	return ab, nil
}
