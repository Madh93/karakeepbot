package karakeepbot

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Madh93/karakeepbot/internal/secret"
	tgbotapi "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTelegram_DownloadFile_Success(t *testing.T) {
	ctx := context.Background()
	expectedFileContent := "This is the mock file content."
	mockFileID := "testFileID123"

	// Mock server to serve the actual file content
	fileContentServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Path in the download URL will be like /bot<token>/<file_path_from_getFile_response>
		// We just serve the content regardless of the exact path for this test.
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(expectedFileContent))
	}))
	defer fileContentServer.Close()

	// Mock server to handle the getFile API call
	// It needs to return a models.File JSON response
	getFileAPIServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method) // getFile is a POST request
		bodyBytes, _ := io.ReadAll(r.Body)
		bodyString := string(bodyBytes)
		require.Contains(t, bodyString, fmt.Sprintf(`"file_id":"%s"`, mockFileID))

		// The actual download URL will be constructed by FileDownloadLink using this FilePath
		// and the fileContentServer.URL (which is the base for bot downloads).
		// So, FilePath should be relative to where the bot would serve files.
		// Example: if fileContentServer.URL is http://localhost:12345
		// and bot token is "test_token", and FilePath is "photos/file.jpg",
		// FileDownloadLink might produce http://localhost:12345/bot_test_token/photos/file.jpg
		// For simplicity, we'll make FileDownloadLink construct a URL that directly hits fileContentServer.
		// This requires FilePath to be the full path part of fileContentServer.URL.
		// However, FileDownloadLink prepends /bot<token>/.
		// A trick: the FileDownloadLink method in the library is `fmt.Sprintf("%s/bot%s/%s", b.apiEndpoint, b.token, file.FilePath)`
		// If we control b.apiEndpoint to be our fileContentServer.URL and make /bot%s/ part of the FilePath
		// then it might work. This is getting complex.

		// Simpler: Let's assume FileDownloadLink(file) will produce a URL that directly points
		// to our fileContentServer. The library's default FileDownloadLink does:
		// fmt.Sprintf("%s/bot%s/%s", b.apiEndpoint, b.token, file.FilePath)
		// We can override the apiEndpoint of the bot.
		// The models.File struct has FilePath.
		mockFilePath := "test_file_path.jpg" // This will be part of the final download URL

		fileResponse := models.Response[models.File]{
			Ok: true,
			Result: &models.File{
				FileID:   mockFileID,
				FilePath: mockFilePath, // This is the crucial part for FileDownloadLink
			},
		}
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(fileResponse)
		require.NoError(t, err)
	}))
	defer getFileAPIServer.Close()

	// Create a bot instance with a client that uses our getFileAPIServer
	// And set the APIEndpoint to our fileContentServer for the actual download part
	// This way, GetFile will hit getFileAPIServer, and FileDownloadLink will use fileContentServer.URL as base.
	// The FileDownloadLink will form a URL like: fileContentServer.URL + "/bot" + botToken + "/" + mockFilePath
	// So, the fileContentServer handler needs to expect this.
	// For simplicity, we'll make the fileContentServer more lenient for now.

	mockBotToken := "test_token_123"
	opts := []tgbotapi.Option{
		tgbotapi.WithClient(getFileAPIServer.Client()), // For GetFile call
		tgbotapi.WithAPIEndpoint(getFileAPIServer.URL),  // So GetFile hits this server
	}
	bot, err := tgbotapi.New(mockBotToken, opts...)
	require.NoError(t, err)

	// IMPORTANT: Override the APIEndpoint for the FileDownloadLink method.
	// FileDownloadLink uses `b.apiEndpoint` which is private.
	// However, the `tgbotapi.Bot` struct has a public `APIEndpoint` field which is initialized.
	// Let's try to make FileDownloadLink use fileContentServer.URL as its base.
	// The library's FileDownloadLink is `bot.APIEndpoint + /bot<token>/ + filePath`.
	// So, if bot.APIEndpoint is fileContentServer.URL, the download will target it.
	// This means we need two different APIEndpoints for the bot.
	// 1. getFileAPIServer.URL for GetFile
	// 2. fileContentServer.URL for FileDownloadLink.
	// This is not directly possible by setting one APIEndpoint field.

	// Let's adjust. The DownloadFile method uses t.GetFile and t.FileDownloadLink.
	// We can make t.Bot.APIEndpoint point to getFileAPIServer for the GetFile call.
	// Then, for FileDownloadLink, it will use that same endpoint.
	// This means getFileAPIServer must also serve the file content, which is not ideal.

	// Alternative: The `FileDownloadLink(file *models.File)` is a method on the Bot.
	// We can't easily mock just that part if `t.Bot` is concrete.

	// Backtrack: simplify the test target. Assume GetFile works and returns a models.File.
	// Assume FileDownloadLink works and returns a URL.
	// Then test if, given that URL, the rest of DownloadFile works.
	// This means we don't need to mock GetFile itself, only the HTTP client for the final download.
	// But DownloadFile calls `t.GetFile` internally.

	// Let's stick to the plan of mocking `GetFile`'s HTTP response, and ensure `FileDownloadLink`
	// constructs a URL that our `fileContentServer` can handle.
	// If Bot.APIEndpoint = getFileAPIServer.URL, then FileDownloadLink will be getFileAPIServer.URL/bot<token>/<filepath>
	// We need fileContentServer to respond to this. This is feasible if fileContentServer's handler is flexible.

	// Let the fileContentServer's handler be very simple and just serve the content.
	// The actual URL formed by FileDownloadLink will be `getFileAPIServer.URL + /bot<TOKEN>/ + mockFilePath`.
	// So, our fileContentServer (which `getFileAPIServer.URL` points to in this adjusted setup)
	// needs to respond to `/bot<TOKEN>/mockFilePath`.

	// Reconfigure:
	// Server 1 (masterServer): Handles BOTH GetFile and the actual file download.
	masterServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/bot"+mockBotToken+"/") { // Actual file download request
			// This part serves the file content
			requestedFile := strings.TrimPrefix(r.URL.Path, "/bot"+mockBotToken+"/")
			t.Logf("Master server: Serving file content for: %s", requestedFile)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(expectedFileContent))
		} else if r.URL.Path == "/getFile" { // getFile API call
			t.Logf("Master server: Handling /getFile request")
			bodyParams, _ := io.ReadAll(r.Body)
			require.Contains(t, string(bodyParams), mockFileID)
			// The FilePath here will be used by FileDownloadLink.
			// It will be appended to masterServer.URL/bot<token>/
			// So, if FilePath is "my/image.jpg", final URL is masterServer.URL/bot<token>/my/image.jpg
			fileResponse := models.Response[models.File]{
				Ok: true,
				Result: &models.File{FileID: mockFileID, FilePath: "my/image.jpg"},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(fileResponse)
		} else {
			t.Logf("Master server: Received unexpected request to: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer masterServer.Close()

	// Create bot with client pointing to masterServer and APIEndpoint also masterServer.URL
	botOpts := []tgbotapi.Option{
		tgbotapi.WithClient(masterServer.Client()),
		tgbotapi.WithAPIEndpoint(masterServer.URL), // Base for GetFile and FileDownloadLink
	}
	testBot, err := tgbotapi.New(mockBotToken, botOpts...)
	require.NoError(t, err)

	tgInstance := &Telegram{Bot: testBot}

	// Create a temporary directory for the download
	tempDir, err := os.MkdirTemp("", "telegram_download_test_")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	destinationPath := filepath.Join(tempDir, "downloaded_file.jpg")

	// Execute DownloadFile
	err = tgInstance.DownloadFile(ctx, mockFileID, destinationPath)
	require.NoError(t, err, "DownloadFile failed")

	// Verify the file was created and content is correct
	assert.FileExists(t, destinationPath)
	actualFileContent, readErr := os.ReadFile(destinationPath)
	require.NoError(t, readErr, "Failed to read downloaded file")
	assert.Equal(t, expectedFileContent, string(actualFileContent), "Downloaded file content mismatch")
}

func TestTelegram_DownloadFile_GetFile_Error(t *testing.T) {
	ctx := context.Background()
	mockFileID := "errorFileID"

	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate an error from Telegram's getFile API
		fileResponse := models.Response[models.File]{Ok: false, Description: "file not found"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound) // Or any error status
		_ = json.NewEncoder(w).Encode(fileResponse)
	}))
	defer errorServer.Close()

	botOpts := []tgbotapi.Option{
		tgbotapi.WithClient(errorServer.Client()),
		tgbotapi.WithAPIEndpoint(errorServer.URL),
	}
	testBot, err := tgbotapi.New("test_token", botOpts...)
	require.NoError(t, err)
	tgInstance := &Telegram{Bot: testBot}

	tempDir, _ := os.MkdirTemp("", "telegram_download_test_err_")
	defer os.RemoveAll(tempDir)
	destinationPath := filepath.Join(tempDir, "downloaded_file.jpg")

	err = tgInstance.DownloadFile(ctx, mockFileID, destinationPath)
	require.Error(t, err, "DownloadFile should have failed due to GetFile error")
	assert.Contains(t, err.Error(), "failed to get file info", "Error message mismatch")
	assert.FileExistsNotExist(t, destinationPath, "File should not have been created on GetFile error")
}


func TestTelegram_DownloadFile_Download_Error(t *testing.T) {
	ctx := context.Background()
	mockFileID := "downloadErrorFileID"

	// Server that successfully returns GetFile info, but fails the download
	masterServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/bottoken_for_dl_error/") { // Actual file download request
			w.WriteHeader(http.StatusInternalServerError) // Simulate download error
		} else if r.URL.Path == "/getFile" { // getFile API call
			fileResponse := models.Response[models.File]{
				Ok:     true,
				Result: &models.File{FileID: mockFileID, FilePath: "some/path.jpg"},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(fileResponse)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer masterServer.Close()

	botOpts := []tgbotapi.Option{
		tgbotapi.WithClient(masterServer.Client()),
		tgbotapi.WithAPIEndpoint(masterServer.URL),
	}
	testBot, err := tgbotapi.New("token_for_dl_error", botOpts...)
	require.NoError(t, err)
	tgInstance := &Telegram{Bot: testBot}

	tempDir, _ := os.MkdirTemp("", "telegram_download_test_dl_err_")
	defer os.RemoveAll(tempDir)
	destinationPath := filepath.Join(tempDir, "downloaded_file.jpg")

	err = tgInstance.DownloadFile(ctx, mockFileID, destinationPath)
	require.Error(t, err, "DownloadFile should have failed due to download error")
	// The error message comes from http.Get or resp.StatusCode check
	assert.Contains(t, err.Error(), "failed to download file", "Error message mismatch for download failure")
	assert.FileExistsNotExist(t, destinationPath, "File should not exist or be empty on download error")
}

func TestTelegram_DownloadFile_CreateDir_Error(t *testing.T) {
    // This test is hard to achieve without os-level mocking for MkdirAll.
    // A simpler check could be to ensure destinationPath is not empty or invalid,
    // but a direct MkdirAll failure is tricky.
    // For now, we'll skip a direct test for os.MkdirAll failure.
    // One could use a path that's known to be unwriteable, but that's platform-dependent.
	t.Skip("Skipping test for MkdirAll failure as it's hard to reliably induce without os-level mocking or platform-specific paths.")
}

func TestTelegram_DownloadFile_CreateFile_Error(t *testing.T) {
	ctx := context.Background()
	mockFileID := "createFileErrorID"
	expectedFileContent := "content"

	// Server that successfully returns GetFile info and allows download
	masterServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/bottoken_for_create_err/") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(expectedFileContent))
		} else if r.URL.Path == "/getFile" {
			fileResponse := models.Response[models.File]{
				Ok:     true,
				Result: &models.File{FileID: mockFileID, FilePath: "some/path.jpg"},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(fileResponse)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer masterServer.Close()

	botOpts := []tgbotapi.Option{
		tgbotapi.WithClient(masterServer.Client()),
		tgbotapi.WithAPIEndpoint(masterServer.URL),
	}
	testBot, err := tgbotapi.New("token_for_create_err", botOpts...)
	require.NoError(t, err)
	tgInstance := &Telegram{Bot: testBot}

	tempDir, _ := os.MkdirTemp("", "telegram_download_test_create_err_")
	defer os.RemoveAll(tempDir)
	
	// Make destinationPath invalid for file creation by pointing to the directory itself
	destinationPath := tempDir 

	err = tgInstance.DownloadFile(ctx, mockFileID, destinationPath)
	require.Error(t, err, "DownloadFile should have failed due to file creation error")
	assert.Contains(t, err.Error(), "failed to create destination file", "Error message mismatch for create file failure")
}


// Mock config.TelegramConfig if needed by other Telegram methods, not strictly by DownloadFile
func mockTelegramConfig() *secret.String {
	s, _ := secret.NewString("test_bot_token")
	return s
}

[end of internal/karakeepbot/telegram_test.go]
