// Package config manages application configuration using the Koanf library.
//
// This package provides structures and functions for handling application
// configuration settings, including configurations for Telegram, Karakeep, and
// Logging. It uses the Koanf library to facilitate loading configuration from
// various sources, such as files (in TOML format) and environment variables.
//
// The main structures include:
//
//   - Config: Represents the overall configuration object, containing nested
//     configurations for Telegram, Karakeep, and Logging settings.
//
//   - KarakeepConfig: Represents the configuration settings specific to the Karakeep
//     server, including the URL and API token. It includes validation to ensure
//     that the provided settings are valid.
//
//   - TelegramConfig: Contains the configuration for Telegram, including the bot
//     token and an allowlist of user IDs. It also validates the token format.
//
//   - LoggingConfig: Holds logging configuration settings, including the log level,
//     format, output destination, and path for log files. It includes validation
//     to ensure the logging settings are correct and conform to allowed values.
//
// The package also provides a New function to create a new configuration
// instance, initializing it with default values, loading settings from a file,
// and processing command line parameters. It ensures that settings are
// validated before they are used, enhancing the reliability of the application.
package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Madh93/karakeepbot/internal/version"
	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Config represents a configuration object. This type is
// designed to hold server and database configurations.
type Config struct {
	Telegram TelegramConfig `koanf:"telegram"` // Telegram configuration
	Karakeep KarakeepConfig `koanf:"karakeep"` // Karakeep configuration
	Logging  LoggingConfig  `koanf:"logging"`  // Logging configuration
	Path     string         `koanf:"path"`     // Path to the configuration file
}

// AppName is the name of the bot.
const AppName = "karakeepbot"

// DefaultConfigFile is the name of the default configuration file.
const DefaultConfigFile = "config.default.toml"

// DefaultPath is the default path to the configuration file. It takes into
// account the [KO_DATA_PATH] environment variable for [ko] builds.
//
// [KO_DATA_PATH]: https://ko.build/features/static-assets/
// [ko]: https://ko.build/
var DefaultPath = filepath.Join(os.Getenv("KO_DATA_PATH"), DefaultConfigFile)

// DefaultConfig is the default configuration for the bot.
var DefaultConfig = Config{
	Telegram: TelegramConfig{
		Allowlist: []int64{-1}, // Enforce allowlist by default.
		Threads:   []int(nil),
	},
	Karakeep: KarakeepConfig{
		URL:      "http://localhost:3000",
		Interval: 5,
	},
	Logging: LoggingConfig{
		Level:   "info",
		Format:  "text",
		Output:  "stdout",
		Path:    AppName + ".log",
		Colored: true,
	},
	Path: DefaultPath,
}

// New returns a new config instance. This initializes the default configuration
// and loads configurations from files (.toml) and environment variables.
func New() *Config {
	// Default config
	config := DefaultConfig

	// Setup koanf
	k := koanf.New(".")

	// Parse command line flags
	parseCommandLineFlags(&config)

	// Load configuration from path
	if err := k.Load(file.Provider(config.Path), toml.Parser()); err != nil {
		if config.Path != DefaultPath {
			log.Fatalf("Error loading config file: %v", err)
		}
		config.Path = "" // Ignore configuration path if default path doesn't exist
	}

	// Configuration from environment variables (with KARAKEEPBOT_ prefix)
	// NOTE: This can't handle multi-word environment variables like TELEGRAM_SECRET_KEY
	// See: https://github.com/knadh/koanf/issues/295
	prefix := strings.ToUpper(AppName)
	if err := k.Load(env.ProviderWithValue(prefix, ".", func(s string, v string) (string, interface{}) {
		// Strip out the prefix, lowercase and using . as key delimiter.
		key := strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(s, prefix+"_")), "_", ".")
		// Split comma-separated values into a slice.
		if strings.Contains(v, ",") {
			return key, strings.Split(v, ",")
		}
		return key, v
	}), nil); err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
	}

	// Override the configuration
	if err := k.Unmarshal("", &config); err != nil {
		log.Fatalf("Error unmarshaling config: %v", err)
	}

	// Validate the configuration
	if err := validateConfig(&config); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	return &config
}

// parseCommandLineFlags parses the command line flags for the configuration.
func parseCommandLineFlags(config *Config) {
	f := flag.NewFlagSet(AppName, flag.ExitOnError)

	f.StringVar(&config.Path, "config", DefaultPath, "Custom configuration file")
	showVersion := f.Bool("version", false, "Show version information")
	showHelp := f.Bool("help", false, "Show help information")

	if err := f.Parse(os.Args[1:]); err != nil {
		log.Fatalf("Error parsing command line flags: %v", err)
	}

	if *showVersion {
		fmt.Println(AppName, version.Get())
		os.Exit(0)
	}

	if *showHelp {
		f.Usage()
		os.Exit(0)
	}
}

// validateConfig checks the validity of the configuration.
func validateConfig(config *Config) error {
	if err := config.Telegram.Validate(); err != nil {
		return err
	}
	if err := config.Karakeep.Validate(); err != nil {
		return err
	}
	if err := config.Logging.Validate(); err != nil {
		return err
	}
	return nil
}
