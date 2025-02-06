// Package hoarderbot implements a Telegram bot that allows users to create
// bookmarks through messages. The bot interacts with the Hoarder API to manage
// bookmarks and handles incoming messages by checking if the chat ID is
// allowed, creating bookmarks, and sending back updated messages with tags.
package hoarderbot

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"time"

	"github.com/Madh93/go-hoarder"
	"github.com/Madh93/hoarderbot/internal/config"
	"github.com/Madh93/hoarderbot/internal/logging"
	"github.com/Madh93/hoarderbot/internal/validation"
)

// Config holds the configuration for the HoarderBot.
type Config struct {
	Hoarder  *config.HoarderConfig
	Telegram *config.TelegramConfig
}

// HoarderBot represents the bot with its dependencies, including the Hoarder
// client, Telegram bot, logger and other options.
type HoarderBot struct {
	hoarder      *Hoarder
	telegram     *Telegram
	logger       *logging.Logger
	allowlist    []int64
	threads      []int
	waitInterval int
}

// New creates a new HoarderBot instance, initializing the Hoarder and Telegram
// clients.
func New(logger *logging.Logger, config *Config) *HoarderBot {
	return &HoarderBot{
		hoarder:      createHoarder(logger, config.Hoarder),
		telegram:     createTelegram(logger, config.Telegram),
		allowlist:    config.Telegram.Allowlist,
		threads:      config.Telegram.Threads,
		waitInterval: config.Hoarder.Interval,
		logger:       logger,
	}
}

// Run starts the bot and handles incoming messages.
func (hb *HoarderBot) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Set default handler
	hb.telegram.RegisterHandlerMatchFunc(func(*TelegramUpdate) bool { return true }, hb.handler)

	// Start the bot
	hb.telegram.Start(ctx)

	return nil
}

// handler is the main handler for incoming messages. It processes the message
// and sends a response back to the user.
func (hb HoarderBot) handler(ctx context.Context, _ *Bot, update *TelegramUpdate) {
	if update.Message == nil {
		return
	}

	msg := TelegramMessage(*update.Message)

	// Check if the chat ID is allowed
	if !hb.isChatIdAllowed(msg.Chat.ID) {
		hb.logger.Warn(fmt.Sprintf("Received message from not allowed chat ID. Allowed chats IDs: %v", hb.allowlist), msg.Attrs()...)
		return
	}

	// Check if the thread ID is allowed
	if !hb.isThreadIdAllowed(msg.MessageThreadID) {
		hb.logger.Warn(fmt.Sprintf("Received message from not allowed thread ID. Allowed thread IDs: %v", hb.threads), msg.Attrs()...)
		return
	}

	hb.logger.Debug("Received message from allowed chat ID and allowed thread ID", msg.Attrs()...)

	// Parse the message to get corresponding bookmark type
	hb.logger.Debug("Parsing message to get corresponding bookmark type", msg.Attrs()...)
	b, err := parseMessage(msg)
	if err != nil {
		hb.logger.Error("Failed to parse message", msg.AttrsWithError(err)...)
		return
	}

	// Create the bookmark
	hb.logger.Debug(fmt.Sprintf("Creating bookmark of type %s", b))
	bookmark, err := hb.hoarder.CreateBookmark(ctx, b)
	if err != nil {
		hb.logger.Error("Failed to create bookmark", "error", err)
		return
	}
	hb.logger.Info("Created bookmark", bookmark.Attrs()...)

	// Wait until bookmark tags are updated
	hb.logger.Debug("Waiting for bookmark tags to be updated", bookmark.Attrs()...)
	for {
		bookmark, err = hb.hoarder.RetrieveBookmarkById(ctx, bookmark.Id)
		if err != nil {
			hb.logger.Error("Failed to retrieve bookmark", "error", err)
			return
		}
		if *bookmark.TaggingStatus == hoarder.Success {
			break
		} else if *bookmark.TaggingStatus == hoarder.Failure {
			hb.logger.Error("Failed to update bookmark tags", bookmark.AttrsWithError(err)...)
			return
		}
		hb.logger.Debug(fmt.Sprintf("Bookmark is still pending, waiting %d seconds before retrying", hb.waitInterval), bookmark.Attrs()...)
		time.Sleep(time.Duration(hb.waitInterval) * time.Second)
	}

	// Add tags
	msg.Text = msg.Text + "\n\n" + bookmark.Hashtags()

	// Send back new message with tags
	hb.logger.Debug("Sending updated message with tags", msg.Attrs()...)
	if err := hb.telegram.SendNewMessage(ctx, &msg); err != nil {
		hb.logger.Error("Failed to send new message", msg.AttrsWithError(err)...)
		return
	}

	// Delete original message
	hb.logger.Debug("Deleting original message", msg.Attrs()...)
	if err := hb.telegram.DeleteOriginalMessage(ctx, &msg); err != nil {
		hb.logger.Error("Failed to delete original message", msg.AttrsWithError(err)...)
		return
	}

	hb.logger.Info("Updated message", msg.Attrs()...)
}

// isChatIdAllowed checks if the chat ID is allowed to receive messages.
func (hb HoarderBot) isChatIdAllowed(chatId int64) bool {
	return len(hb.allowlist) == 0 || slices.Contains(hb.allowlist, chatId)
}

// isThreadIdAllowed checks if the thread ID is allowed to receive messages.
func (hb HoarderBot) isThreadIdAllowed(threadId int) bool {
	return len(hb.threads) == 0 || slices.Contains(hb.threads, threadId)
}

// parseMessage parses the incoming Telegram message and returns the corresponding Bookmark type.
func parseMessage(msg TelegramMessage) (BookmarkType, error) {
	switch {
	case validation.ValidateURL(msg.Text) == nil:
		return NewLinkBookmark(msg.Text), nil
	case msg.Text != "":
		return NewTextBookmark(msg.Text), nil
	default:
		return nil, errors.New("unsupported bookmark type")
	}
}
