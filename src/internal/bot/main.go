package bot

import (
	"log/slog"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/idanoo/GoDiscMoji/internal/db"
)

var b *Bot

type Bot struct {
	DiscordSession *discordgo.Session
	Token          string

	registeredCommands []*discordgo.ApplicationCommand
	Db                 *db.Database
}

// New - Return new instance of *Bot
func New(token string) *Bot {
	return &Bot{
		Token:              token,
		registeredCommands: make([]*discordgo.ApplicationCommand, len(commands)),
	}
}

// Start - Boots the bot!
func (bot *Bot) Start() error {
	// Boot db
	db, err := db.InitDb()
	if err != nil {
		return err
	}
	defer db.CloseDbConn()
	bot.Db = db

	// Boot discord
	discord, err := discordgo.New("Bot " + bot.Token)
	if err != nil {
		return err
	}
	bot.DiscordSession = discord

	// Add command handler
	bot.DiscordSession.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	// Add handlers
	// bot.DiscordSession.AddHandler(bot.HandleMessage)
	bot.DiscordSession.AddHandler(bot.HandleReaction)

	// Load session
	err = discord.Open()
	if err != nil {
		return err
	}
	defer discord.Close()

	// Register commands
	bot.RegisterCommands()
	b = bot

	// Keep running untill there is NO os interruption (ctrl + C)
	slog.Info("Bot is now running. Press CTRL-C to exit.")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	// Deregister any commands we created
	bot.DeregisterCommands()

	return nil
}

// func (bot *Bot) HandleMessage(discord *discordgo.Session, message *discordgo.MessageCreate) {
// 	// Don't reply to self
// 	if message.Author.ID == discord.State.User.ID {
// 		return
// 	}

// 	slog.Info("Message received", "message", message.Content)
// 	// respond to user message if it contains `!help` or `!bye`
// 	switch {
// 	case strings.HasPrefix(message.Content, "!help"):
// 		_, err := discord.ChannelMessageSend(message.ChannelID, "Hello World😃")
// 		if err != nil {
// 			slog.Error("Failed to send message", "err", err)
// 		}
// 	case strings.Contains(message.Content, "!bye"):
// 		discord.ChannelMessageSend(message.ChannelID, "Good Bye👋")
// 		// add more cases if required
// 	}

// }

// HandleReaction - Simply log it
func (bot *Bot) HandleReaction(discord *discordgo.Session, reaction *discordgo.MessageReactionAdd) {
	err := bot.Db.LogEmojiUsage(reaction.GuildID, reaction.ChannelID, reaction.UserID, reaction.Emoji.Name)
	if err != nil {
		slog.Error("Failed to log emoji usage", "err", err)
	}
}
