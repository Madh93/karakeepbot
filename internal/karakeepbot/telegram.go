package karakeepbot

import (
	"cmp"
	"context"
	"fmt"
	"io"
	"net/http"
	"slices"

	"github.com/Madh93/karakeepbot/internal/config"
	"github.com/Madh93/karakeepbot/internal/logging"
	"github.com/go-telegram/bot"
	tgbotapi "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
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
		ChatID:          msg.Chat.ID,
		MessageThreadID: msg.MessageThreadID,
		Text:            msg.Text,
		ParseMode:       models.ParseModeMarkdown,
	}

	if _, err := t.SendMessage(ctx, params); err != nil {
		return err
	}

	return nil
}

// SendNewPhoto sends a new photo to the user's chat.
func (t Telegram) SendNewPhoto(ctx context.Context, msg *TelegramMessage) error {
	photo := slices.Clone(msg.Photo)
	slices.SortFunc(
		photo,
		func(s1, s2 models.PhotoSize) int {
			return cmp.Compare(s2.Width*s2.Height, s1.Width*s1.Height)
		},
	)

	if _, err := t.SendPhoto(ctx, &tgbotapi.SendPhotoParams{
		ChatID:          msg.Chat.ID,
		MessageThreadID: msg.MessageThreadID,
		Photo:           &models.InputFileString{Data: photo[0].FileID},
		Caption:         msg.Text,
		ParseMode:       models.ParseModeMarkdown,
	}); err != nil {
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

func (t Telegram) DownloadFile(ctx context.Context, fileID string) (io.ReadCloser, error) {
	file, err := t.GetFile(ctx, &bot.GetFileParams{
		FileID: fileID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call GetFile: %w", err)
	}

	resp, err := http.Get(
		fmt.Sprintf(
			"https://api.telegram.org/file/bot%s/%s",
			t.Token(),
			file.FilePath,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to make a request to download a file: %w", err)
	}

	return resp.Body, nil
}
