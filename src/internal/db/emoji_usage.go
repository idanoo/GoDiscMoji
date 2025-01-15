package db

import "fmt"

// LogEmojiUsage - Log usage
func (db *Database) LogEmojiUsage(guildID, channelID, userID, emojiID string) error {
	_, err := db.db.Exec(
		"INSERT INTO `emoji_usage` (`guild_id`, `channel_id`, `user_id`, `emoji_id`, `timestamp`) VALUES (?,?,?,?, datetime())",
		guildID, channelID, userID, emojiID,
	)

	return err
}

// GetTopUsersForGuild - Report usage
func (db *Database) GetTopUsersForGuild(guildID string, num int) ([]string, error) {
	var data []string
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
		data = append(data, fmt.Sprintf("<@%s>: %d", user, count))
	}

	return data, nil
}

// GetTopEmojisForGuild - Report usage
func (db *Database) GetTopEmojisForGuild(guildID string, num int) ([]string, error) {
	var data []string
	row, err := db.db.Query(
		"SELECT emoji_id, count(*) FROM `emoji_usage` WHERE `guild_id` = ? GROUP BY emoji_id ORDER BY count(*) DESC LIMIT ?",
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
		data = append(data, fmt.Sprintf("%s: %d", user, count))
	}

	return data, nil
}
