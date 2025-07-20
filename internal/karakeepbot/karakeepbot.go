// Package karakeepbot implements a Telegram bot that allows users to create
// bookmarks through messages. The bot interacts with the Karakeep API to manage
// bookmarks and handles incoming messages by checking if the chat ID is
// allowed, creating bookmarks, and sending back updated messages with tags.
package karakeepbot

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"time"

	"github.com/Madh93/go-karakeep"
	"github.com/Madh93/karakeepbot/internal/config"
	"github.com/Madh93/karakeepbot/internal/kkprivate"
	"github.com/Madh93/karakeepbot/internal/logging"
	"github.com/Madh93/karakeepbot/internal/markdown"
	"github.com/Madh93/karakeepbot/internal/validation"
	"github.com/go-telegram/bot/models"
)

// Config holds the configuration for the KarakeepBot.
type Config struct {
	Karakeep *config.KarakeepConfig
	Telegram *config.TelegramConfig
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
}

// New creates a new KarakeepBot instance, initializing the Karakeep and Telegram
// clients.
func New(logger *logging.Logger, config *Config) *KarakeepBot {
	return &KarakeepBot{
		karakeep:     createKarakeep(logger, config.Karakeep),
		telegram:     createTelegram(logger, config.Telegram),
		allowlist:    config.Telegram.Allowlist,
		threads:      config.Telegram.Threads,
		waitInterval: config.Karakeep.Interval,
		logger:       logger,
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

	// Look into the message to build the corresponding bookmark
	kb.logger.Debug("Parsing message to get corresponding bookmark type", msg.Attrs()...)
	b, err := makeBookmark(msg)
	if err != nil {
		kb.logger.Error("Failed to make bookmark", msg.AttrsWithError(err)...)
		return
	}

	createRequest, err := b.JSONReader()
	if err != nil {
		kb.logger.Error("Failed to make bookmark create request", msg.AttrsWithError(err)...)
		return
	}

	// Create the bookmark
	kb.logger.Debug(fmt.Sprintf("Creating bookmark of type %s", b))
	bookmark, err := kb.karakeep.CreateBookmark(ctx, createRequest)
	if err != nil {
		kb.logger.Error("Failed to create bookmark", "error", err)
		return
	}
	kb.logger.Info("Created bookmark", bookmark.Attrs()...)

	// Upload asset, if any.
	if len(msg.Photo) != 0 {
		asset, err := kb.uploadPhoto(ctx, msg.Photo)
		if err != nil {
			kb.logger.Error("Failed to create an asset", "error", err)
			return
		}
		if err := kb.karakeep.AttachAssetToBookmark(
			ctx,
			bookmark.Id,
			karakeep.PostBookmarksBookmarkIdAssetsJSONRequestBody{
				AssetType: karakeep.BannerImage,
				Id:        asset.AssetID,
			},
		); err != nil {
			kb.logger.Error("Failed to attach an asset to bookmark", "error", err)
			return
		}
	}

	// Wait until bookmark tags are updated
	kb.logger.Debug("Waiting for bookmark tags to be updated", bookmark.Attrs()...)
	for {
		bookmark, err = kb.karakeep.RetrieveBookmarkById(ctx, bookmark.Id)
		if err != nil {
			kb.logger.Error("Failed to retrieve bookmark", "error", err)
			return
		}
		if *bookmark.TaggingStatus == karakeep.BookmarkTaggingStatusSuccess {
			break
		} else if *bookmark.TaggingStatus == karakeep.BookmarkTaggingStatusFailure {
			kb.logger.Error("Failed to update bookmark tags", bookmark.AttrsWithError(err)...)
			return
		}
		kb.logger.Debug(fmt.Sprintf("Bookmark is still pending, waiting %d seconds before retrying", kb.waitInterval), bookmark.Attrs()...)
		time.Sleep(time.Duration(kb.waitInterval) * time.Second)
	}

	// Add tags
	switch {
	case msg.Text != "":
		msg.Text = msg.Text + "\n\n" + bookmark.Hashtags()
	case msg.Caption != "":
		msg.Caption = msg.Caption + "\n\n" + bookmark.Hashtags()
	}

	// Send back new message with tags
	kb.logger.Debug("Sending updated message with tags", msg.Attrs()...)
	if len(msg.Photo) == 0 {
		err = kb.telegram.SendNewMessage(ctx, &msg)
	} else {
		err = kb.telegram.SendNewPhoto(ctx, &msg)
	}
	if err != nil {
		kb.logger.Error("Failed to send new message", msg.AttrsWithError(err)...)
		return
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
	return len(kb.allowlist) == 0 || slices.Contains(kb.allowlist, chatId)
}

// isThreadIdAllowed checks if the thread ID is allowed to receive messages.
func (kb KarakeepBot) isThreadIdAllowed(threadId int) bool {
	return len(kb.threads) == 0 || slices.Contains(kb.threads, threadId)
}

func (kb KarakeepBot) uploadPhoto(
	ctx context.Context,
	photo []models.PhotoSize,
) (*kkprivate.Asset, error) {
	photo = slices.Clone(photo)
	slices.SortFunc(
		photo,
		func(s1, s2 models.PhotoSize) int {
			return cmp.Compare(s2.Width*s2.Height, s1.Width*s1.Height)
		},
	)

	file, err := kb.telegram.DownloadFile(ctx, photo[0].FileID)
	if err != nil {
		return nil, fmt.Errorf("failed to make a request to download a file: %w", err)
	}
	defer file.Close()

	asset, err := kb.karakeep.Private.CreateAsset(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create an asset: %w", err)
	}

	return asset, nil
}

// makeBookmark looks into the incoming Telegram message and builds the
// corresponding Bookmark.
func makeBookmark(msg TelegramMessage) (*Bookmark, error) {
	switch {
	case validation.ValidateURL(msg.Text) == nil:
		return &Bookmark{
			Type: BookmarkTypeLink,
			URL:  msg.Text,
		}, nil
	case msg.Text != "":
		b := &Bookmark{
			Type: BookmarkTypeText,
			Text: msg.Text,
		}
		b.Text = markdown.EncodeURLs(b.Text, msg.Entities)
		return b, nil
	case msg.Caption != "" && len(msg.Photo) != 0:
		b := &Bookmark{
			Type: BookmarkTypeText,
			Text: msg.Caption,
		}
		b.Text = markdown.EncodeURLs(b.Text, msg.CaptionEntities)
		return b, nil
	default:
		return nil, errors.New("unsupported message")
	}
}
