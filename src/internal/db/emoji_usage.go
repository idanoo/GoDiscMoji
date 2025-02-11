package db

import "log/slog"

type EmojiMap struct {
	EmojiID string
	Count   int64
}

// LogEmojiUsage - Log usage
func (db *Database) LogEmojiUsage(guildID, channelID, userID, emojiID string) error {
	_, err := db.db.Exec(
		"INSERT INTO `emoji_usage` (`guild_id`, `channel_id`, `user_id`, `emoji_id`, `timestamp`) VALUES (?,?,?,?, datetime())",
		guildID, channelID, userID, emojiID,
	)

	return err
}

// GetTopUsersForGuild - Report usage
func (db *Database) GetTopUsersForGuild(guildID string, num int64) (map[string]int64, error) {
	data := make(map[string]int64)
	row, err := db.db.Query(
		"SELECT user_id, count(*) FROM `emoji_usage` WHERE `guild_id` = ? GROUP BY user_id ORDER BY count(*) DESC LIMIT ?",
		guildID,
		num,
	)

	if err != nil {
		return data, err
	}

	defer row.Close()
	for row.Next() {
		var user string
		var count int64
		row.Scan(&user, &count)
		slog.Error("Error getting top emojis for guild user", "user_id", user, "count", count)
		data[user] = count
	}

	return data, nil
}

// GetTopUsersForGuildEmoji - Report usage
func (db *Database) GetTopUsersForGuildEmoji(guildID string, emojiID string, num int) (map[string]int64, error) {
	data := make(map[string]int64)
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
	for row.Next() {
		var user string
		var count int64
		row.Scan(&user, &count)
		data[user] = count
	}

	return data, nil
}

// GetTopEmojisForGuild - Report usage
func (db *Database) GetTopEmojisForGuild(guildID string, num int64) (map[int]EmojiMap, error) {
	data := make(map[int]EmojiMap)
	row, err := db.db.Query(
		"SELECT emoji_id, count(*) FROM `emoji_usage` WHERE `guild_id` = ? GROUP BY emoji_id ORDER BY count(*) DESC LIMIT ?",
		guildID,
		num,
	)

	if err != nil {
		return data, err
	}

	defer row.Close()
	i := 0
	for row.Next() {
		var emoji string
		var count int64
		row.Scan(&emoji, &count)
		data[i] = EmojiMap{EmojiID: emoji, Count: count}
		i++
	}

	return data, nil
}

// GetTopEmojisForGuildUser - Report usage
func (db *Database) GetTopEmojisForGuildUser(guildID string, userID string, num int) (map[int]EmojiMap, error) {
	data := make(map[int]EmojiMap)
	row, err := db.db.Query(
		"SELECT emoji_id, count(*) FROM `emoji_usage` WHERE `guild_id` = ? AND `user_id` = ? GROUP BY emoji_id ORDER BY count(*) DESC LIMIT ?",
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
		var emoji string
		var count int64
		row.Scan(&emoji, &count)
		data[i] = EmojiMap{EmojiID: emoji, Count: count}
		i++
	}

	return data, nil
}
