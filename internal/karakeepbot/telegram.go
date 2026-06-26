package karakeepbot

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Madh93/karakeepbot/internal/config"
	"github.com/Madh93/karakeepbot/internal/logging"
	"github.com/Madh93/karakeepbot/internal/secret"
	tgbotapi "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// Bot is an alias for tgbotapi.Bot.
type Bot = tgbotapi.Bot

// Telegram embeds the Telegram bot API client to add high level functionality.
type Telegram struct {
	*Bot
	token secret.String
}

// createTelegram initializes the Telegram Bot API client.
func createTelegram(logger *logging.Logger, config *config.TelegramConfig) *Telegram {
	logger.Debug(fmt.Sprintf("Initializing Telegram Bot API using %s token", config.Token))

	var opts []tgbotapi.Option

	if config.ProxyEnabled {
		proxyURL, err := url.Parse(config.ProxyURL)
		if err != nil {
			logger.Fatal("Error parsing proxy URL.", "error", err)
		}

		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
		client := &http.Client{
			Transport: transport,
		}
		opts = append(opts, tgbotapi.WithHTTPClient(30*time.Second, client))
		opts = append(opts, tgbotapi.WithCheckInitTimeout(30*time.Second))
	}

	telegramBot, err := tgbotapi.New(config.Token.Value(), opts...)
	if err != nil {
		logger.Fatal("Error creating Telegram Bot API.", "error", err)
	}

	return &Telegram{Bot: telegramBot, token: config.Token}
}

// SendNewMessage sends a new message to the user's chat.
func (t Telegram) SendNewMessage(ctx context.Context, msg *TelegramMessage) error {
	params := &tgbotapi.SendMessageParams{
		ChatID:          msg.Chat.ID,
		MessageThreadID: msg.MessageThreadID,
		Text:            msg.Text,
	}

	if _, err := t.SendMessage(ctx, params); err != nil {
		return err
	}

	return nil
}

// SendPhotoWithCaption sends a photo with a caption.
func (t *Telegram) SendPhotoWithCaption(ctx context.Context, msg *TelegramMessage, photoID string, caption string) error {
	params := &tgbotapi.SendPhotoParams{
		ChatID:          msg.Chat.ID,
		MessageThreadID: msg.MessageThreadID,
		Photo:           &models.InputFileString{Data: photoID},
		Caption:         caption,
	}

	if _, err := t.SendPhoto(ctx, params); err != nil {
		return err
	}

	return nil
}

// SendReply sends a reply to a specific message.
func (t Telegram) SendReply(ctx context.Context, msg *TelegramMessage, text string) error {
	params := &tgbotapi.SendMessageParams{
		ChatID:          msg.Chat.ID,
		MessageThreadID: msg.MessageThreadID,
		ReplyParameters: &models.ReplyParameters{MessageID: msg.ID},
		Text:            text,
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

// GetFileURL returns the download URL for a given file ID.
func (t Telegram) GetFileURL(ctx context.Context, fileID string) (string, error) {
	file, err := t.GetFile(ctx, &tgbotapi.GetFileParams{FileID: fileID})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", t.token.Value(), file.FilePath), nil
}
