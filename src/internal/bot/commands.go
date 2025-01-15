package bot

import (
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
		"show-top-emojis": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			top, err := b.Db.GetTopEmojisForGuild(i.GuildID, 5)
			if err != nil {
				slog.Error("Error getting top emojis", "err", err)
				return
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Top Emojis:\n" + strings.Join(top, "\n"),
				},
			})
		},
		"show-top-users": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			top, err := b.Db.GetTopUsersForGuild(i.GuildID, 5)
			if err != nil {
				slog.Error("Error getting top users", "err", err)
				return
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Top Users:\n" + strings.Join(top, "\n"),
				},
			})
		},
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
