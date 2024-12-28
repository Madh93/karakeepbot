package hoarderbot

import (
	"context"
	"fmt"

	"github.com/Madh93/hoarderbot/internal/config"
	"github.com/Madh93/hoarderbot/internal/logging"
	tgbotapi "github.com/go-telegram/bot"
)

// Bot is an alias for tgbotapi.Bot.
type Bot = tgbotapi.Bot

// Telegram embeds the Telegram bot API client to add high level functionality.
type Telegram struct {
	*Bot
}

// createTelegram initializes the Telegram Bot API client.
func createTelegram(logger *logging.Logger, config *config.TelegramConfig) *Telegram {
	logger.Debug(fmt.Sprintf("Initializing Telegram Bot API using %s token", config.Token))

	telegramBot, err := tgbotapi.New(config.Token.Value())
	if err != nil {
		logger.Fatal("Error creating Telegram Bot API.", "error", err)
	}

	return &Telegram{telegramBot}
}

// SendNewMessage sends a new message to the user's chat.
func (t Telegram) SendNewMessage(ctx context.Context, msg *TelegramMessage) error {
	params := &tgbotapi.SendMessageParams{
		ChatID: msg.Chat.ID,
		Text:   msg.Text,
	}

	if _, err := t.SendMessage(ctx, params); err != nil {
		return err
	}

	return nil
}

// DeleteOriginalMessage deletes the original message from the user's chat.
func (t Telegram) DeleteOriginalMessage(ctx context.Context, msg *TelegramMessage) error {
	params := &tgbotapi.DeleteMessageParams{
		ChatID:    msg.Chat.ID,
		MessageID: msg.ID,
	}

	if _, err := t.DeleteMessage(ctx, params); err != nil {
		return err
	}

	return nil
}
