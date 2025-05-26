package karakeepbot

import (
	"context"
	"fmt"

	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Madh93/karakeepbot/internal/config"
	"github.com/Madh93/karakeepbot/internal/logging"
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
	}

	if _, err := t.SendMessage(ctx, params); err != nil {
		return err
	}

	return nil
}

// DownloadFile downloads a file from Telegram given its fileID and saves it to destinationPath.
// The destinationPath should be the full path including the desired filename.
func (t *Telegram) DownloadFile(ctx context.Context, fileID string, destinationPath string) error {
	// Get file metadata from Telegram
	params := &tgbotapi.GetFileParams{
		FileID: fileID,
	}
	file, err := t.GetFile(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Construct the download URL
	downloadURL := t.FileDownloadLink(file)
	if downloadURL == "" {
		// The library FileDownloadLink needs *models.File, and GetFile returns *models.File.
		// It also needs t.APIEndpoint which is part of the Bot structure.
		// If still empty, it means the FilePath on models.File was empty.
		return fmt.Errorf("failed to get download URL for fileID %s (file_path: '%s')", fileID, file.FilePath)
	}

	// Fetch the file via HTTP GET
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download file from %s: %w", downloadURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file: received status %s from %s", resp.Status, downloadURL)
	}

	// Ensure the destination directory exists
	dir := filepath.Dir(destinationPath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create destination directory %s: %w", dir, err)
	}

	// Create the destination file
	out, err := os.Create(destinationPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", destinationPath, err)
	}
	defer out.Close()

	// Copy the downloaded content to the file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file to %s: %w", destinationPath, err)
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
