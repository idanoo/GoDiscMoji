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

	defaultRunCommandPermissions int64 = discordgo.PermissionKickMembers

	commands = []*discordgo.ApplicationCommand{
		{
			Name:                     "show-top-emojis",
			Description:              "Show top emojis",
			DefaultMemberPermissions: &defaultRunCommandPermissions,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "amount",
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
					Name:        "amount",
					Description: "Amount to show",
					MinValue:    &integerOptionMinValue,
					MaxValue:    20,
					Required:    false,
				},
			},
		},
		{
			Name:                     "purge-recent-emojis",
			Description:              "Purges recent emojis",
			DefaultMemberPermissions: &defaultRunCommandPermissions,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "Select user",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "hours",
					Description: "Hours to purge",
					Required:    true,
				},
			},
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"show-top-emojis":     showTopEmojis,
		"show-top-users":      showTopUsers,
		"purge-recent-emojis": purgeRecentEmojis,
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
	if opt, ok := optionMap["amount"]; ok {
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
		if top[v].EmojiID == top[v].EmojiName {
			// Handle bad data with stock emojis
			msg += fmt.Sprintf(":%s: %d", top[v].EmojiName, top[v].Count)
		} else {
			msg += fmt.Sprintf("<:%s:%s> %d", top[v].EmojiName, top[v].EmojiID, top[v].Count)
		}
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
	if opt, ok := optionMap["amount"]; ok {
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
			if topUsers[sv].EmojiID == topUsers[sv].EmojiName {
				// Handle bad data with stock emojis
				users = append(users, fmt.Sprintf(":%s: %d", topUsers[sv].EmojiName, topUsers[sv].Count))
			} else {
				users = append(users, fmt.Sprintf("<:%s:%s> %d", topUsers[sv].EmojiName, topUsers[sv].EmojiID, topUsers[sv].Count))
			}
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

// purgeRecentEmojis - Purges recent emojis for a user
func purgeRecentEmojis(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Access options in the order provided by the user.
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	user := &discordgo.User{}
	if opt, ok := optionMap["user"]; ok {
		user = opt.UserValue(s)
	} else {
		slog.Error("Invalid user option provided")
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content:         "No user specified",
				AllowedMentions: &discordgo.MessageAllowedMentions{},
			},
		})
		return
	}

	hours := int64(24)
	if opt, ok := optionMap["hours"]; ok {
		hours = opt.IntValue()
	} else {
		slog.Error("Invalid hours option provided")
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content:         "No hours specified",
				AllowedMentions: &discordgo.MessageAllowedMentions{},
			},
		})
		return
	}

	emojis, err := b.Db.GetRecentEmojisForUser(i.GuildID, user.ID, hours)
	if err != nil {
		slog.Error("Error getting recent emojis for user", "err", err)
		return
	}

	x := 0
	for _, emoji := range emojis {
		emojiID := emoji.EmojiID
		if emojiID == "" {
			emojiID = emoji.EmojiName
		}

		err := s.MessageReactionRemove(emoji.ChannelID, emoji.MessageID, emojiID, emoji.UserID)
		if err != nil {
			slog.Error("Error removing emoji reaction", "err", err, "emoji", emoji.EmojiID, "user", user.ID)
			continue
		}

		err = b.Db.DeleteEmojiUsage(i.GuildID, emoji.ChannelID, emoji.MessageID, emoji.UserID, emoji.EmojiID)
		if err != nil {
			slog.Error("Error deleting emoji usage", "err", err)
			continue
		}
		x++
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:         fmt.Sprintf("Purged %d emojis for user %s", x, user.Username),
			AllowedMentions: &discordgo.MessageAllowedMentions{},
		},
	})
}
