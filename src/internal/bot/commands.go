package bot

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "show-top-emojis",
			Description: "Show top emojis",
		},
		{
			Name:        "show-top-users",
			Description: "Show top users",
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"show-top-emojis": showTopEmojis,
		"show-top-users":  showTopUsers,
	}
)

// RegisterCommands
func (bot *Bot) RegisterCommands() {
	bot.registeredCommands = make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := bot.DiscordSession.ApplicationCommandCreate(bot.DiscordSession.State.User.ID, "", v)
		if err != nil {
			slog.Error("Error creating command", "err", err)
		}

		bot.registeredCommands[i] = cmd
	}
}

// DeregisterCommands - Deregister all commands
func (bot *Bot) DeregisterCommands() {
	for _, v := range bot.registeredCommands {
		err := bot.DiscordSession.ApplicationCommandDelete(bot.DiscordSession.State.User.ID, "", v.ID)
		if err != nil {
			slog.Error("Error deleting command", "err", err)
		}
	}
}

// showTopEmojis - Show top emojis with users
func showTopEmojis(s *discordgo.Session, i *discordgo.InteractionCreate) {
	top, err := b.Db.GetTopEmojisForGuild(i.GuildID, 5)
	if err != nil {
		slog.Error("Error getting top emojis", "err", err)
		return
	}

	msg := "Most used emojis:\n"
	for k, v := range top {
		topUsers, err := b.Db.GetTopUsersForGuildEmoji(i.GuildID, k, 3)
		if err != nil {
			slog.Error("Error getting top users for guild emoji", "err", err)
			continue
		}

		users := []string{}
		msg += fmt.Sprintf("%s: %d", k, v)
		for sk, sv := range topUsers {
			users = append(users, fmt.Sprintf("`<@%s>`: %d", sk, sv))
		}
		msg += "  (" + strings.Join(users, ", ") + ")\n"
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
		},
	})
}

// showTopUsers - Show top users with emojis
func showTopUsers(s *discordgo.Session, i *discordgo.InteractionCreate) {
	top, err := b.Db.GetTopUsersForGuild(i.GuildID, 5)
	if err != nil {
		slog.Error("Error getting top users", "err", err)
		return
	}

	msg := "Users who use the most emojis:\n"
	for k, v := range top {
		topUsers, err := b.Db.GetTopEmojisForGuildUser(i.GuildID, k, 3)
		if err != nil {
			slog.Error("Error getting top emojis for guild user", "err", err)
			continue
		}

		users := []string{}
		msg += fmt.Sprintf("`<@%s>`: %d", k, v)
		for sk, sv := range topUsers {
			users = append(users, fmt.Sprintf("%s: %d", sk, sv))
		}
		msg += "  (" + strings.Join(users, ", ") + ")\n"
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
		},
	})
}
