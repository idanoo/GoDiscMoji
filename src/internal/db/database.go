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
		"`user_id` TEXT, " +
		"`emoji_id` TEXT, " +
		"`timestamp` DATETIME, `viewed` INT DEFAULT 0" +
		")")

	return db, err
}

// CloseDbConn - Closes DB connection
func (db *Database) CloseDbConn() {
	db.db.Close()
}
