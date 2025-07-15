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
			Name:                     "add-magic-tool",
			Description:              "Runs a script on reaction for user",
			DefaultMemberPermissions: &defaultRunCommandPermissions,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "Select user",
					Required:    true,
				},
			},
		},
		{
			Name:                     "remove-magic-tool",
			Description:              "Stops running a script on reaction for user",
			DefaultMemberPermissions: &defaultRunCommandPermissions,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "Select user",
					Required:    true,
				},
			},
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"show-top-emojis":   showTopEmojis,
		"show-top-users":    showTopUsers,
		"add-magic-tool":    addAutoScrubber,
		"remove-magic-tool": removeAutoScrubber,
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
			msg += fmt.Sprintf("%s %d", top[v].EmojiName, top[v].Count)
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
	for k := range top {
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
				users = append(users, fmt.Sprintf("%s %d", topUsers[sv].EmojiName, topUsers[sv].Count))
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

// addAutoScrubber - Scrubs emojis after a set period
func addAutoScrubber(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

	err := startScrubbingUser(i.GuildID, user.ID)
	if err != nil {
		slog.Error("Error starting auto scrubber", "err", err, "guild_id", i.GuildID, "user_id", user.ID)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content:         ":x:",
				AllowedMentions: &discordgo.MessageAllowedMentions{},
			},
		})

		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:         ":white_check_mark:",
			AllowedMentions: &discordgo.MessageAllowedMentions{},
		},
	})
}

// removeAutoScrubber - Stops scrubbing emojis
func removeAutoScrubber(s *discordgo.Session, i *discordgo.InteractionCreate) {
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

	err := stopScrubbingUser(i.GuildID, user.ID)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content:         ":x:",
				AllowedMentions: &discordgo.MessageAllowedMentions{},
			},
		})

		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:         ":white_check_mark:",
			AllowedMentions: &discordgo.MessageAllowedMentions{},
		},
	})
}
