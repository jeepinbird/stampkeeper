package database

import (
	"database/sql"
	"fmt"
)

// Migrate creates all the necessary tables for the application
func Migrate(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS storage_boxes (
			id TEXT PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			date_created TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS tags (
			id TEXT PRIMARY KEY,
			name TEXT UNIQUE NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS stamps (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			scott_number TEXT UNIQUE,
			issue_date TEXT,
			series TEXT,
			notes TEXT,
			image_url TEXT,
			is_owned BOOLEAN DEFAULT false,
			date_added TEXT NOT NULL,
			date_modified TEXT NOT NULL,
			date_deleted TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS stamp_instances (
			id TEXT PRIMARY KEY,
			stamp_id TEXT NOT NULL,
			condition TEXT,
			box_id TEXT,
			quantity INTEGER DEFAULT 1,
			date_added TEXT NOT NULL,
			date_modified TEXT NOT NULL,
			date_deleted TEXT,
			FOREIGN KEY (stamp_id) REFERENCES stamps(id) ON DELETE CASCADE,
			FOREIGN KEY (box_id) REFERENCES storage_boxes(id) ON DELETE SET NULL,
			UNIQUE(stamp_id, condition, box_id) -- Prevent duplicate condition/box combinations
		)`,
		`CREATE TABLE IF NOT EXISTS stamp_tags (
			stamp_id TEXT,
			tag_id TEXT,
			PRIMARY KEY (stamp_id, tag_id),
			FOREIGN KEY (stamp_id) REFERENCES stamps(id) ON DELETE CASCADE,
			FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
		)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute migration query: %v", err)
		}
	}

	return nil
}