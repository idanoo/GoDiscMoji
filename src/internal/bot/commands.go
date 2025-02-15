package bot

import (
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	integerOptionMinValue = 1.0
	amountKey             = "amount"

	defaultRunCommandPermissions int64 = discordgo.PermissionKickMembers
	commands                           = []*discordgo.ApplicationCommand{
		{
			Name:                     "show-top-emojis",
			Description:              "Show top emojis",
			DefaultMemberPermissions: &defaultRunCommandPermissions,
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
			Name:                     "show-top-users",
			Description:              "Show top users",
			DefaultMemberPermissions: &defaultRunCommandPermissions,
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

	// Sort keys
	keys := make([]int, 0)
	for k, _ := range top {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	msg := "Most used emojis:\n"
	for _, v := range keys {
		topUsers, err := b.Db.GetTopUsersForGuildEmoji(i.GuildID, top[v].EmojiID, 3)
		if err != nil {
			slog.Error("Error getting top users for guild emoji", "err", err)
			continue
		}

		subkeys := make([]int, 0)
		for k, _ := range topUsers {
			subkeys = append(subkeys, k)
		}
		sort.Ints(subkeys)

		users := []string{}
		msg += fmt.Sprintf("%s: %d", top[v].EmojiID, top[v].Count)
		for _, sv := range subkeys {
			users = append(users, fmt.Sprintf("<@%s>: %d", topUsers[sv].EmojiID, topUsers[sv].Count))
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

	// Sort keys
	keys := make([]int, 0)
	for k, _ := range top {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	msg := "Users who use the most emojis:\n"
	for _, v := range keys {
		topUsers, err := b.Db.GetTopEmojisForGuildUser(i.GuildID, top[v].EmojiID, 3)
		if err != nil {
			slog.Error("Error getting top emojis for guild user", "err", err)
			continue
		}

		subkeys := make([]int, 0)
		for k, _ := range topUsers {
			subkeys = append(subkeys, k)
		}
		sort.Ints(subkeys)

		users := []string{}
		msg += fmt.Sprintf("<@%s>: %d", top[v].EmojiID, top[v].Count)
		for _, sv := range subkeys {
			users = append(users, fmt.Sprintf("%s: %d", topUsers[sv].EmojiID, topUsers[sv].Count))
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
