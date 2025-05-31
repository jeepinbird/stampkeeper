package database

import (
	"database/sql"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Seed populates the database with sample data if it's empty
func Seed(db *sql.DB) error {
	// Check if we already have data
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM stamps").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // Already seeded
	}

	// Create sample storage boxes
	boxID := uuid.New().String()
	_, err = db.Exec(`INSERT INTO storage_boxes (id, name, date_created) VALUES (?, ?, ?)`,
		boxID, "Box 1", time.Now().Format(time.RFC3339))
	if err != nil {
		return err
	}

	box2ID := uuid.New().String()
	_, err = db.Exec(`INSERT INTO storage_boxes (id, name, date_created) VALUES (?, ?, ?)`,
		box2ID, "Box 2", time.Now().Format(time.RFC3339))
	if err != nil {
		return err
	}

	// Create sample stamps
	stamps := []struct {
		name, scottNum, issueDate, series, condition, boxID string
		quantity                                            int
		isOwned                                             bool
	}{
		{"Lincoln 1c Green", "219", "1890-02-22", "1890-93 Regular Issue", "Used", boxID, 1, true},
		{"Washington 2c Carmine", "220", "1890-02-22", "1890-93 Regular Issue", "Mint", boxID, 1, true},
		{"Jackson 3c Purple", "221", "1890-02-22", "1890-93 Regular Issue", "Used", boxID, 1, false},
		{"German Empire 10pf", "55", "1900-01-01", "Germania", "Mint", box2ID, 2, true},
	}

	for _, s := range stamps {
		stampID := uuid.New().String()
		_, err = db.Exec(`INSERT INTO stamps 
			(id, name, scott_number, issue_date, series, condition, quantity, box_id, is_owned, date_added, date_modified) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			stampID, s.name, s.scottNum, s.issueDate, s.series, s.condition, s.quantity,
			s.boxID, s.isOwned, time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339))
		if err != nil {
			return err
		}
	}

	// Create sample tags
	tagNames := []string{"USA", "Classic", "Presidential"}
	for _, tagName := range tagNames {
		tagID := uuid.New().String()
		_, err = db.Exec(`INSERT INTO tags (id, name) VALUES (?, ?)`, tagID, tagName)
		if err != nil {
			// Ignore unique constraint violations (in case of concurrent runs)
			if !strings.Contains(err.Error(), "UNIQUE constraint failed") {
				return err
			}
		}
	}

	return nil
}