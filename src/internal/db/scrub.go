package db

type Scrub struct {
	GuildID string `json:"guild_id"`
	UserID  string `json:"user_id"`
}

// AddScrub - Add an auto scrubber for a guild/user
func (db *Database) AddScrub(guildID, userID string) error {
	_, err := db.db.Exec(
		"INSERT INTO `scrub` (`guild_id`, `user_id`) VALUES (?,?)",
		guildID, userID,
	)

	return err
}

// RemoveScrub - Delete for guild/channel/message/user
func (db *Database) RemoveScrub(guildID, userID string) error {
	_, err := db.db.Exec(
		"DELETE FROM `scrub` WHERE `guild_id` = ? AND `user_id` = ?",
		guildID, userID,
	)

	return err
}

// GetAllScrubs - Get all scrubbers
func (db *Database) GetAllScrubs() ([]Scrub, error) {
	data := make([]Scrub, 0)
	row, err := db.db.Query("SELECT guild_id, user_id from `scrub`")
	if err != nil {
		return data, err
	}

	defer row.Close()
	for row.Next() {
		var guildID string
		var userID string
		row.Scan(&guildID, &userID)
		data = append(data, Scrub{GuildID: guildID, UserID: userID})
	}

	return data, nil
}
