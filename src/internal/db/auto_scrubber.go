package db

import (
	"time"
)

type AutoScrubber struct {
	GuildID  string        `json:"guild_id"`
	UserID   string        `json:"user_id"`
	Duration time.Duration `json:"duration"`
}

// AddAutoScrubber - Add an auto scrubber for a guild/user
func (db *Database) AddAutoScrubber(guildID, userID string, duration time.Duration) error {
	_, err := db.db.Exec(
		"INSERT INTO `auto_scrubber` (`guild_id`, `user_id`, `duration`) VALUES (?,?,?)",
		guildID, userID, duration,
	)

	return err
}

// RemoveAutoScrubber - Delete for guild/channel/message/user
func (db *Database) RemoveAutoScrubber(guildID, userID string) error {
	_, err := db.db.Exec(
		"DELETE FROM `auto_scrubber` WHERE `guild_id` = ? AND `user_id` = ?",
		guildID, userID,
	)

	return err
}

// DeleteEmojiAll - Delete for whole message
func (db *Database) GetAllAutoScrubbers() ([]AutoScrubber, error) {
	data := make([]AutoScrubber, 0)
	row, err := db.db.Query("SELECT guild_id, user_id, duration from `auto_scrubber`")
	if err != nil {
		return data, err
	}

	defer row.Close()
	for row.Next() {
		var guildID string
		var userID string
		var duration time.Duration
		row.Scan(&guildID, &userID, &duration)
		data = append(data, AutoScrubber{GuildID: guildID, UserID: userID, Duration: duration})
	}

	return data, nil
}
