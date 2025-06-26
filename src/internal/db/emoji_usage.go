package db

import (
	"fmt"
	"time"
)

type EmojiMap struct {
	EmojiID   string
	EmojiName string
	Count     int64
}

type EmojiUsage struct {
	ID        int64
	GuildID   string
	ChannelID string
	MessageID string
	UserID    string
	EmojiID   string
	EmojiName string
	Timestamp time.Time
}

// LogEmojiUsage - Log usage
func (db *Database) LogEmojiUsage(guildID, channelID, messageID, userID, emojiID, emojiName string) error {
	_, err := db.db.Exec(
		"INSERT INTO `emoji_usage` (`guild_id`, `channel_id`, `message_id`, `user_id`, `emoji_id`, `emoji_name`, `timestamp`) VALUES (?,?,?,?,?,?, datetime())",
		guildID, channelID, messageID, userID, emojiID, emojiName,
	)

	return err
}

// DeleteEmojiUsage - Delete for guild/channel/message/user
func (db *Database) DeleteEmojiUsage(guildID, channelID, messageID, userID, emojiID string) error {
	_, err := db.db.Exec(
		"DELETE FROM `emoji_usage` WHERE `guild_id` = ? AND `channel_id` = ? AND `message_id` = ? AND `user_id` = ? AND `emoji_id` = ?",
		guildID, channelID, messageID, userID, emojiID,
	)

	return err
}

// DeleteEmojiUsageById - Delete for guild/channel/message/user
func (db *Database) DeleteEmojiUsageById(id int64) error {
	_, err := db.db.Exec(
		"DELETE FROM `emoji_usage` WHERE `id` = ?",
		id,
	)

	return err
}

// DeleteEmojiAll - Delete for whole message
func (db *Database) DeleteEmojiAll(guildID, channelID, messageID string) error {
	_, err := db.db.Exec(
		"DELETE FROM `emoji_usage` WHERE `guild_id` = ? AND `channel_id` = ? AND `message_id` = ?",
		guildID, channelID, messageID,
	)

	return err
}

// GetTopUsersForGuild - Report usage
func (db *Database) GetTopUsersForGuild(guildID string, num int64) (map[int]EmojiMap, error) {
	data := make(map[int]EmojiMap)
	row, err := db.db.Query(
		"SELECT user_id, count(*) FROM `emoji_usage` WHERE `guild_id` = ? GROUP BY user_id ORDER BY count(*) DESC LIMIT ?",
		guildID,
		num,
	)

	if err != nil {
		return data, err
	}

	defer row.Close()
	i := 0
	for row.Next() {
		var name string
		var count int64
		row.Scan(&name, &count)
		data[i] = EmojiMap{EmojiID: name, Count: count}
		i++
	}
	return data, nil
}

// GetTopUsersForGuildEmoji - Report usage
func (db *Database) GetTopUsersForGuildEmoji(guildID string, emojiID string, num int) (map[int]EmojiMap, error) {
	data := make(map[int]EmojiMap)
	row, err := db.db.Query(
		"SELECT user_id, count(*) FROM `emoji_usage` WHERE `guild_id` = ? AND `emoji_id` = ? GROUP BY emoji_id, user_id ORDER BY count(*) DESC LIMIT ?",
		guildID,
		emojiID,
		num,
	)

	if err != nil {
		return data, err
	}

	defer row.Close()
	i := 0
	for row.Next() {
		var name string
		var count int64
		row.Scan(&name, &count)
		data[i] = EmojiMap{EmojiID: name, Count: count}
		i++
	}

	return data, nil
}

// GetTopEmojisForGuild - Report usage
func (db *Database) GetTopEmojisForGuild(guildID string, num int64) (map[int]EmojiMap, error) {
	data := make(map[int]EmojiMap)
	row, err := db.db.Query(
		"SELECT emoji_name, emoji_id, count(*) FROM `emoji_usage` WHERE `guild_id` = ? GROUP BY emoji_id ORDER BY count(*) DESC LIMIT ?",
		guildID,
		num,
	)

	if err != nil {
		return data, err
	}

	defer row.Close()
	i := 0
	for row.Next() {
		var emojiName string
		var emojiID string
		var count int64
		row.Scan(&emojiName, &emojiID, &count)
		data[i] = EmojiMap{EmojiID: emojiID, EmojiName: emojiName, Count: count}
		i++
	}

	return data, nil
}

// GetTopEmojisForGuildUser - Report usage
func (db *Database) GetTopEmojisForGuildUser(guildID string, userID string, num int) (map[int]EmojiMap, error) {
	data := make(map[int]EmojiMap)
	row, err := db.db.Query(
		"SELECT emoji_name, emoji_id, count(*) FROM `emoji_usage` WHERE `guild_id` = ? AND `user_id` = ? GROUP BY emoji_name ORDER BY count(*) DESC LIMIT ?",
		guildID,
		userID,
		num,
	)

	if err != nil {
		return data, err
	}

	defer row.Close()
	i := 0
	for row.Next() {
		var emojiName string
		var emojiID string
		var count int64
		row.Scan(&emojiName, &emojiID, &count)
		data[i] = EmojiMap{EmojiID: emojiID, EmojiName: emojiName, Count: count}
		i++
	}

	return data, nil
}

// GetRecentEmojisForUser - Get recent emojis used by user map[]
func (db *Database) GetRecentEmojisForUser(guildID string, userID string, hours int64) ([]EmojiUsage, error) {
	var data []EmojiUsage
	row, err := db.db.Query(
		"SELECT id, guild_id, channel_id, message_id, user_id, emoji_id, emoji_name, timestamp "+
			"FROM `emoji_usage` WHERE `guild_id` = ? AND `user_id` = ? AND timestamp >= datetime('now', '-"+fmt.Sprintf("%d", hours)+" hours') "+
			"ORDER BY timestamp DESC",
		guildID,
		userID,
	)

	if err != nil {
		return data, err
	}

	defer row.Close()
	for row.Next() {
		usage := EmojiUsage{}
		row.Scan(&usage.ID, &usage.GuildID, &usage.ChannelID, &usage.MessageID, &usage.UserID, &usage.EmojiID, &usage.EmojiName, &usage.Timestamp)
		data = append(data, usage)
	}

	return data, nil
}

// GetAllEmojisForUser - Get all emojis used by user map[]
func (db *Database) GetAllEmojisForUser(guildID string, userID string) ([]EmojiUsage, error) {
	var data []EmojiUsage
	row, err := db.db.Query(
		"SELECT id, guild_id, channel_id, message_id, user_id, emoji_id, emoji_name, timestamp "+
			"FROM `emoji_usage` WHERE `guild_id` = ? AND `user_id` = ?",
		guildID,
		userID,
	)

	if err != nil {
		return data, err
	}

	defer row.Close()
	for row.Next() {
		usage := EmojiUsage{}
		row.Scan(&usage.ID, &usage.GuildID, &usage.ChannelID, &usage.MessageID, &usage.UserID, &usage.EmojiID, &usage.EmojiName, &usage.Timestamp)
		data = append(data, usage)
	}

	return data, nil
}
