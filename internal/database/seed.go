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
	_, err = db.Exec(`INSERT INTO storage_boxes (id, name, date_created) VALUES ($1, $2, $3)`,
		boxID, "Box 1", time.Now())
	if err != nil {
		return err
	}

	box2ID := uuid.New().String()
	_, err = db.Exec(`INSERT INTO storage_boxes (id, name, date_created) VALUES ($1, $2, $3)`,
		box2ID, "Box 2", time.Now())
	if err != nil {
		return err
	}

	// Create sample stamps (designs)
	stamps := []struct {
		id, name, scottNum, issueDate, series string
	}{
		{uuid.New().String(), "Lincoln 1c Green", "219", "1890-02-22", "1890-93 Regular Issue"},
		{uuid.New().String(), "Washington 2c Carmine", "220", "1890-02-22", "1890-93 Regular Issue"},
		{uuid.New().String(), "Jackson 3c Purple", "221", "1890-02-22", "1890-93 Regular Issue"},
		{uuid.New().String(), "German Empire 10pf", "55", "1900-01-01", "Germania"},
	}

	for _, s := range stamps {
		_, err = db.Exec(`INSERT INTO stamps 
			(id, name, scott_number, issue_date, series, is_owned, date_added, date_modified) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			s.id, s.name, s.scottNum, s.issueDate, s.series, false,
			time.Now(), time.Now())
		if err != nil {
			return err
		}
	}

	// Create sample stamp instances (grouped by condition/box with quantities)
	instances := []struct {
		stampIndex int    // Index into the stamps slice above
		condition  string
		boxID      string
		quantity   int
	}{
		{0, "Used", boxID, 1},      // 1 Used Lincoln in Box 1
		{1, "Mint", boxID, 1},      // 1 Mint Washington in Box 1  
		{1, "Used", box2ID, 2},     // 2 Used Washington in Box 2
		{3, "Mint", box2ID, 1},     // 1 Mint German in Box 2
		{3, "Used", box2ID, 2},     // 2 Used German in Box 2
		// Note: Jackson stamp (index 2) has no instances - it's a "needed" stamp
	}

	for _, inst := range instances {
		instanceID := uuid.New().String()
		stampID := stamps[inst.stampIndex].id
		
		_, err = db.Exec(`INSERT INTO stamp_instances 
			(id, stamp_id, condition, box_id, quantity, date_added, date_modified) 
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			instanceID, stampID, inst.condition, inst.boxID, inst.quantity,
			time.Now(), time.Now())
		if err != nil {
			return err
		}
	}

	// Create sample tags
	tagNames := []string{"USA", "Classic", "Presidential", "Germany"}
	tagIDs := make(map[string]string)
	
	for _, tagName := range tagNames {
		tagID := uuid.New().String()
		tagIDs[tagName] = tagID
		_, err = db.Exec(`INSERT INTO tags (id, name) VALUES ($1, $2)`, tagID, tagName)
		if err != nil {
			// Ignore unique constraint violations (in case of concurrent runs)
			if !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				return err
			}
		}
	}

	// Add some tag associations
	stampTags := []struct {
		stampIndex int
		tags       []string
	}{
		{0, []string{"USA", "Classic", "Presidential"}}, // Lincoln
		{1, []string{"USA", "Classic", "Presidential"}}, // Washington  
		{2, []string{"USA", "Classic", "Presidential"}}, // Jackson (needed stamp)
		{3, []string{"Germany", "Classic"}},             // German Empire
	}

	for _, st := range stampTags {
		stampID := stamps[st.stampIndex].id
		for _, tagName := range st.tags {
			if tagID, exists := tagIDs[tagName]; exists {
				_, err = db.Exec(`INSERT INTO stamp_tags (stamp_id, tag_id) VALUES ($1, $2)`, stampID, tagID)
				if err != nil && !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
					return err
				}
			}
		}
	}

	return nil
}