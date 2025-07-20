# Karakeepbot

[![Latest release](https://img.shields.io/github/v/tag/Madh93/karakeepbot?label=Release)](https://github.com/Madh93/karakeepbot/releases)
[![Go Version](https://img.shields.io/badge/Go-1.24-blue)](https://go.dev/doc/install)
[![Go Reference](https://pkg.go.dev/badge/github.com/Madh93/karakeepbot.svg)](https://pkg.go.dev/github.com/Madh93/karakeepbot)
[![License](https://img.shields.io/badge/License-MIT-brightgreen)](LICENSE)

`Karakeepbot` is a simple [Telegram Bot](https://core.telegram.org/bots) written in [Go](https://go.dev/) that enables users to effortlessly save bookmarks to [Karakeep](https://karakeep.app) (previously Hoarder), a self-hostable bookmark-everything app, directly through Telegram.

![Karakeepbot Demo](./docs/gif/demo.gif)

<p align="center">
  <a href="#features">Features</a> •
  <a href="#requirements">Requirements</a> •
  <a href="#installation">Installation</a> •
  <a href="#configuration">Configuration</a> •
  <a href="#contributing">Contributing</a> •
  <a href="#license">License</a>
</p>

## Features

- 📄 Add **text** and **URL bookmarks** into your Karakeep instance (tested on [v0.25.0](https://github.com/karakeep-app/karakeep/releases/tag/v0.25.0)).
- 🤖 Obtain **AI-generated tags** in **hashtag format** for easy searching on Telegram.
- 🔒 **Mandatory chat ID allowlist** to prevent abuse (**thread ID allowlist** is also supported).
- 🐳 **Production-ready Docker image** for easy **deployment**.

## Requirements

- A [Telegram bot token](https://core.telegram.org/bots/features#botfather) (you can get one by talking to [@BotFather](https://t.me/BotFather) on Telegram)
- Your Telegram Chat ID (see [How to Find Your Chat ID](#how-to-find-your-chat-id)).
- A valid API key from [Karakeep](https://docs.karakeep.app/screenshots#settings).

## Installation

### Docker

#### Using `docker run`

Use the `docker run` command to start `Karakeepbot`. Make sure to set the required environment variables:

```sh
docker run --name karakeepbot \
  -e KARAKEEPBOT_TELEGRAM_TOKEN=your-telegram-bot-token \
  -e KARAKEEPBOT_TELEGRAM_ALLOWLIST=your-telegram-chat-id \
  -e KARAKEEPBOT_KARAKEEP_TOKEN=your-karakeep-api-key \
  -e KARAKEEPBOT_KARAKEEP_URL=https://your-karakeep-instance.tld \
  ghcr.io/madh93/karakeepbot:latest
```

#### Using `docker compose`

Create a `docker-compose.yml` file with the following content:

```yml
services:
  karakeepbot:
    image: ghcr.io/madh93/karakeepbot:latest
    restart: unless-stopped
    # volumes:
    #   - ./custom.config.toml:/var/run/ko/config.default.toml # Optional: specify a custom configuration file instead of the default one
    environment:
      - KARAKEEPBOT_TELEGRAM_TOKEN=your-telegram-bot-token
      - KARAKEEPBOT_TELEGRAM_ALLOWLIST=your-telegram-chat-id
      - KARAKEEPBOT_KARAKEEP_TOKEN=your-karakeep-api-key
      - KARAKEEPBOT_KARAKEEP_URL=https://your-karakeep-instance.tld
```

Use the `docker compose up` command to start `Karakeepbot`:

```sh
docker compose up
```

### From releases

Download the latest binary from [the releases page](https://github.com/Madh93/karakeepbot/releases):

```sh
curl -L https://github.com/Madh93/karakeepbot/releases/latest/download/karakeepbot_$(uname -s)_$(uname -m).tar.gz | tar -xz -O karakeepbot > /usr/local/bin/karakeepbot
chmod +x /usr/local/bin/karakeepbot
```

### From source

If you have Go installed:

```sh
go install github.com/Madh93/karakeepbot@latest
```

## Configuration

`Karakeepbot` comes with a [default configuration file](config.default.toml) that you can modify to suit your needs.

### Loading a custom configuration file

You can load a different configuration file by using the `-config path/to/config/file` flag when starting the application:

```sh
karakeepbot -config custom.config.toml
```

### Overriding with environment variables

Additionally, you can override the configuration values using environment variables that begin with the prefix `KARAKEEPBOT_`. This allows you to customize your setup without needing to modify any configuration file:

```sh
KARAKEEPBOT_LOGGING_LEVEL=debug KARAKEEPBOT_TELEGRAM_ALLOWLIST=chat_id_1,chat_id_2 karakeepbot
```

### Security Concerns

To protect your bot from abuse and spam from unauthorized users, `Karakeepbot` implements a **mandatory Chat ID allowlist**.

By default, the bot is configured with a placeholder value and **it will not start** until you explicitly configure the `telegram.allowlist` parameter. You have two options:

1. **Recommended:** Provide a list of one or more specific Chat IDs. This ensures only you and other authorized users can interact with the bot.
2. **Not Recommended:** Provide an empty list (`[]`). This will allow any Telegram user to interact with your bot, exposing it to potential abuse.

#### How to Find Your Chat ID

1. Open your Telegram client.
2. Search for the bot `@userinfobot` and start a chat with it.
3. The bot will immediately reply with your user information, including your **Id**. This is your Chat ID.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any bug fixes or enhancements.

1. Fork the repository.
2. Create a new branch (`git checkout -b feature-branch`).
3. Commit your changes (`git commit -am 'Add new feature'`).
4. Push to the branch (`git push origin feature-branch`).
5. Open a Pull Request.

## License

This project is licensed under the [MIT license](LICENSE).
