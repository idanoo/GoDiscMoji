package main

import (
	"log/slog"
	"os"

	"github.com/idanoo/GoDiscMoji/internal/bot"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	// Get required vars
	discordToken := os.Getenv("DISCORD_TOKEN")
	if discordToken == "" {
		slog.Error("DISCORD_TOKEN env var is required")
		os.Exit(1)
	}

	// Start the bot
	bot := bot.New(discordToken)
	err := bot.Start()
	if err != nil {
		slog.Error("Error starting bot", "err", err)
		os.Exit(1)
	}
}
