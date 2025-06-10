package database

import (
	"database/sql"
	"fmt"
)

// Migrate creates all the necessary tables for the application
func Migrate(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS storage_boxes (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(255) UNIQUE NOT NULL,
			date_created TIMESTAMP NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS tags (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(255) UNIQUE NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS stamps (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			scott_number VARCHAR(255) UNIQUE,
			issue_date VARCHAR(255),
			series VARCHAR(255),
			notes TEXT,
			image_url VARCHAR(512),
			is_owned BOOLEAN DEFAULT false,
			date_added TIMESTAMP NOT NULL,
			date_modified TIMESTAMP NOT NULL,
			date_deleted TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS stamp_instances (
			id VARCHAR(36) PRIMARY KEY,
			stamp_id VARCHAR(36) NOT NULL,
			condition VARCHAR(255),
			box_id VARCHAR(36),
			quantity INTEGER DEFAULT 1,
			date_added TIMESTAMP NOT NULL,
			date_modified TIMESTAMP NOT NULL,
			date_deleted TIMESTAMP,
			FOREIGN KEY (stamp_id) REFERENCES stamps(id) ON DELETE CASCADE,
			FOREIGN KEY (box_id) REFERENCES storage_boxes(id) ON DELETE SET NULL,
			UNIQUE(stamp_id, condition, box_id)
		)`,
		`CREATE TABLE IF NOT EXISTS stamp_tags (
			stamp_id VARCHAR(36),
			tag_id VARCHAR(36),
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