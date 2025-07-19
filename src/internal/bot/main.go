package bot

import (
	"log/slog"
	"os"
	"os/signal"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/idanoo/GoDiscMoji/internal/db"
)

const dynoUserID = "155149108183695360"

var b *Bot

type Bot struct {
	DiscordSession *discordgo.Session
	Token          string

	registeredCommands []*discordgo.ApplicationCommand
	Db                 *db.Database

	// Scrub map[GuildID][UserID]bool
	scrubs      *map[string]map[string]bool
	scrubsMutex sync.RWMutex
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
	bot.DiscordSession.AddHandler(bot.HandleAddReaction)
	bot.DiscordSession.AddHandler(bot.HandleRemoveReaction)
	bot.DiscordSession.AddHandler(bot.HandleRemoveAllReaction)

	// Load session
	err = discord.Open()
	if err != nil {
		return err
	}
	defer discord.Close()

	// Register commands
	bot.RegisterCommands()
	b = bot

	// Add scrubs
	initScrub()

	// Keep running untill there is NO os interruption (ctrl + C)
	slog.Info("Bot is now running. Press CTRL-C to exit.")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	// Deregister any commands we created
	bot.DeregisterCommands()

	return nil
}

// HandleReaction - Simply log it
func (bot *Bot) HandleAddReaction(discord *discordgo.Session, reaction *discordgo.MessageReactionAdd) {
	// Ignore Dyno user
	if reaction.UserID == dynoUserID {
		return
	}

	if scrub.shouldScrub(reaction.GuildID, reaction.UserID) {
		err := b.DiscordSession.MessageReactionRemove(reaction.ChannelID, reaction.MessageID, reaction.Emoji.APIName(), reaction.UserID)
		if err == nil {
			return
		}

		slog.Error("Failed to remove emoji reaction", "err", err, "reaction", reaction)
	}

	err := bot.Db.LogEmojiUsage(reaction.GuildID, reaction.ChannelID, reaction.MessageID, reaction.UserID, reaction.Emoji.ID, reaction.Emoji.Name)
	if err != nil {
		slog.Error("Failed to log emoji usage", "err", err)
	}
}

// HandleRemoveReaction - Remove for user/message/emoji
func (bot *Bot) HandleRemoveReaction(discord *discordgo.Session, reaction *discordgo.MessageReactionRemove) {
	// Ignore Dyno user
	if reaction.UserID == dynoUserID {
		return
	}

	err := bot.Db.DeleteEmojiUsage(reaction.GuildID, reaction.ChannelID, reaction.MessageID, reaction.UserID, reaction.Emoji.ID)
	if err != nil {
		slog.Error("Failed to delete single emoji usage", "err", err)
	}
}

// HandleRemoveAllReaction - Remove all for message
func (bot *Bot) HandleRemoveAllReaction(discord *discordgo.Session, reaction *discordgo.MessageReactionRemoveAll) {
	err := bot.Db.DeleteEmojiAll(reaction.GuildID, reaction.ChannelID, reaction.MessageID)
	if err != nil {
		slog.Error("Failed to delete all emoji usage for message", "err", err)
	}
}
