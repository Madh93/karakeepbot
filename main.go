// Package main is the entry point of the application. It initializes the
// configuration, sets up logging, and starts the Hoarderbot.
package main

import (
	"fmt"

	"github.com/Madh93/hoarderbot/internal/config"
	"github.com/Madh93/hoarderbot/internal/hoarderbot"
	"github.com/Madh93/hoarderbot/internal/logging"
)

// main initializes the configuration, sets up logging, and starts the
// Hoarderbot.
func main() {
	// Load configuration
	config := config.New()

	// Setup logger
	logger := logging.New(&config.Logging)
	if config.Path != "" {
		logger.Debug(fmt.Sprintf("Loaded configuration from %s", config.Path))
	}

	// Setup hoarderbot
	hoarderbot := hoarderbot.New(logger, &hoarderbot.Config{
		Hoarder:  &config.Hoarder,
		Telegram: &config.Telegram,
	})

	// Let's go
	logger.Info("3, 2, 1...  Launching Hoarderbot... 🚀")
	if err := hoarderbot.Run(); err != nil {
		logger.Fatal("💥 Something went wrong.", "error", err)
	}
}
