package bot

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	integerOptionMinValue = 5.0
	amountKey             = "amount"

	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "show-top-emojis",
			Description: "Show top emojis",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        amountKey,
					Description: "Amount to show",
					MinValue:    &integerOptionMinValue,
					MaxValue:    20,
					Required:    false,
				},
			},
		},
		{
			Name:        "show-top-users",
			Description: "Show top users",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        amountKey,
					Description: "Amount to show",
					MinValue:    &integerOptionMinValue,
					MaxValue:    20,
					Required:    false,
				},
			},
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
	// Access options in the order provided by the user.
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	amount := int64(5)
	if opt, ok := optionMap[amountKey]; ok {
		amount = opt.IntValue()
	}

	top, err := b.Db.GetTopEmojisForGuild(i.GuildID, amount)
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
			users = append(users, fmt.Sprintf("<@%s>: %d", sk, sv))
		}
		msg += "  (" + strings.Join(users, ", ") + ")\n"
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:         msg,
			AllowedMentions: &discordgo.MessageAllowedMentions{},
		},
	})
}

// showTopUsers - Show top users with emojis
func showTopUsers(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Access options in the order provided by the user.
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	amount := int64(5)
	if opt, ok := optionMap[amountKey]; ok {
		amount = opt.IntValue()
	}

	top, err := b.Db.GetTopUsersForGuild(i.GuildID, amount)
	if err != nil {
		slog.Error("Error getting top users", "err", err)
		return
	}

	msg := "Users who use the most emojis:\n"
	for k, v := range top {
		slog.Error("Error getting top emojis for guild user", "emoji_id", k, "count", v)
		topUsers, err := b.Db.GetTopEmojisForGuildUser(i.GuildID, k, 3)
		if err != nil {
			slog.Error("Error getting top emojis for guild user", "err", err)
			continue
		}

		users := []string{}
		msg += fmt.Sprintf("<@%s>: %d", k, v)
		for sk, sv := range topUsers {
			users = append(users, fmt.Sprintf("%s: %d", sk, sv))
		}
		msg += "  (" + strings.Join(users, ", ") + ")\n"
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:         msg,
			AllowedMentions: &discordgo.MessageAllowedMentions{},
		},
	})
}
