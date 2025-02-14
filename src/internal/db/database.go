package db

import (
	"database/sql"

	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

// InitDb - Initialize DB connection
func InitDb() (*Database, error) {
	ddb := Database{}

	db, err := sql.Open("sqlite3", "file:/data/db.sqlite?loc=auto")
	if err != nil {
		return &ddb, err
	}

	db.SetConnMaxLifetime(0)
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)

	err = db.Ping()
	if err != nil {
		return &ddb, err
	}

	ddb.db = db

	return ddb.runMigrations()
}

// runMigrations - Run migrations for connection
func (db *Database) runMigrations() (*Database, error) {
	// Hacked up af - Rerunnable
	_, err := db.db.Exec("CREATE TABLE IF NOT EXISTS `emoji_usage` (" +
		"`id` INTEGER PRIMARY KEY AUTOINCREMENT, " +
		"`guild_id` TEXT, " +
		"`channel_id` TEXT, " +
		"`message_id` TEXT, " +
		"`user_id` TEXT, " +
		"`emoji_id` TEXT, " +
		"`timestamp` DATETIME" +
		")")
	if err != nil {
		return db, err
	}

	_, err = db.db.Exec("CREATE INDEX IF NOT EXISTS `idx_emoji_usage_guild_id_user_id` ON `emoji_usage` (`guild_id`, `user_id`, `emoji_id`)")
	if err != nil {
		return db, err
	}

	_, err = db.db.Exec("CREATE INDEX IF NOT EXISTS `idx_emoji_usage_message_id_user_id_emoji_id` ON `emoji_usage` (`message_id`, `user_id`, `guild_id`, `emoji_id`)")
	if err != nil {
		return db, err
	}

	_, err = db.db.Exec("CREATE INDEX IF NOT EXISTS `idx_emoji_usage_emoji_id,guild_id` ON `emoji_usage` (`guild_id`, `emoji_id`)")
	if err != nil {
		return db, err
	}

	_, err = db.db.Exec("DELETE FROM emoji_usage WHERE user_id = 155149108183695360")
	if err != nil {
		return db, err
	}
	return db, err
}

// CloseDbConn - Closes DB connection
func (db *Database) CloseDbConn() {
	db.db.Close()
}
