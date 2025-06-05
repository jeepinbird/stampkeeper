package services

import (
	"database/sql"
	"net/http"
	"strings"
	"time"
	"log"
	"fmt"

	"github.com/google/uuid"
	"github.com/jeepinbird/stampkeeper/internal/models"
)

type StampService struct {
	db *sql.DB
}

func NewStampService(db *sql.DB) *StampService {
	return &StampService{db: db}
}

// Gets the total count of unique stamps (not instances) matching filters
func (s *StampService) GetStampCount(r *http.Request) (int64, error) {
	query := `
		SELECT COUNT(DISTINCT s.id) 
		FROM stamps s 
		LEFT JOIN stamp_instances si ON s.id = si.stamp_id AND si.date_deleted IS NULL
		WHERE s.date_deleted IS NULL`
	args := []interface{}{}

	// Build WHERE clause based on filters
	if search := r.URL.Query().Get("search"); search != "" {
		query += ` AND (LOWER(s.name) LIKE LOWER(?) OR LOWER(s.scott_number) LIKE LOWER(?) OR LOWER(s.series) LIKE LOWER(?))`
		searchParam := "%" + search + "%"
		args = append(args, searchParam, searchParam, searchParam)
	}

	if owned := r.URL.Query().Get("owned"); owned != "" {
		if owned == "true" {
			query += ` AND EXISTS (SELECT 1 FROM stamp_instances si2 WHERE si2.stamp_id = s.id AND si2.date_deleted IS NULL)`
		} else if owned == "false" {
			query += ` AND NOT EXISTS (SELECT 1 FROM stamp_instances si2 WHERE si2.stamp_id = s.id AND si2.date_deleted IS NULL)`
		}
	}

	if boxID := r.URL.Query().Get("box_id"); boxID != "" {
		query += ` AND EXISTS (SELECT 1 FROM stamp_instances si3 WHERE si3.stamp_id = s.id AND si3.box_id = ? AND si3.date_deleted IS NULL)`
		args = append(args, boxID)
	}

	var count int64
	err := s.db.QueryRow(query, args...).Scan(&count)
	return count, err
}

func (s *StampService) GetStamps(r *http.Request, page, limit int) ([]models.Stamp, error) {
	query := `
		SELECT s.id, s.name, s.scott_number, s.issue_date, s.series, 
		       s.notes, s.image_url, s.date_added, s.date_modified,
		       CASE WHEN COUNT(si.id) > 0 THEN true ELSE false END as is_owned
		FROM stamps s
		LEFT JOIN stamp_instances si ON s.id = si.stamp_id AND si.date_deleted IS NULL
		WHERE s.date_deleted IS NULL`

	args := []interface{}{}

	// Add filters based on query parameters
	if search := r.URL.Query().Get("search"); search != "" {
		query += ` AND (LOWER(s.name) LIKE LOWER(?) OR LOWER(s.scott_number) LIKE LOWER(?) OR LOWER(s.series) LIKE LOWER(?))`
		searchParam := "%" + search + "%"
		args = append(args, searchParam, searchParam, searchParam)
	}

	// Group by stamp before applying owned filter
	query += ` GROUP BY s.id, s.name, s.scott_number, s.issue_date, s.series, s.notes, s.image_url, s.date_added, s.date_modified`

	if owned := r.URL.Query().Get("owned"); owned != "" {
		if owned == "true" {
			query += ` HAVING COUNT(si.id) > 0`
		} else if owned == "false" {
			query += ` HAVING COUNT(si.id) = 0`
		}
	}

	if boxID := r.URL.Query().Get("box_id"); boxID != "" {
		// This is trickier - we need stamps that have instances in this specific box
		query = `
			SELECT DISTINCT s.id, s.name, s.scott_number, s.issue_date, s.series, 
			       s.notes, s.image_url, s.date_added, s.date_modified,
			       true as is_owned
			FROM stamps s
			JOIN stamp_instances si ON s.id = si.stamp_id 
			WHERE s.date_deleted IS NULL AND si.date_deleted IS NULL AND si.box_id = ?`
		
		// Reset args and add boxID at the beginning
		newArgs := []interface{}{boxID}
		
		if search := r.URL.Query().Get("search"); search != "" {
			query += ` AND (LOWER(s.name) LIKE LOWER(?) OR LOWER(s.scott_number) LIKE LOWER(?) OR LOWER(s.series) LIKE LOWER(?))`
			searchParam := "%" + search + "%"
			newArgs = append(newArgs, searchParam, searchParam, searchParam)
		}
		
		args = newArgs
	}

	// Add sorting
	sortBy := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order")
	if order == "" {
		order = "ASC"
	}
	order = strings.ToUpper(order)
	if order != "ASC" && order != "DESC" {
		order = "ASC"
	}

	switch sortBy {
	case "name":
		query += ` ORDER BY s.name ` + order
	case "issue_date":
		query += ` ORDER BY s.issue_date ` + order
	case "date_added":
		query += ` ORDER BY s.date_added ` + order
	default:
		query += ` ORDER BY s.scott_number ` + order
	}

	// Calculate offset from page and limit
	offset := (page - 1) * limit
	query += ` LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stamps []models.Stamp
	for rows.Next() {
		var stamp models.Stamp
		var dateAdded, dateModified string
		err := rows.Scan(&stamp.ID, &stamp.Name, &stamp.ScottNumber, &stamp.IssueDate, &stamp.Series,
			&stamp.Notes, &stamp.ImageURL, &dateAdded, &dateModified, &stamp.IsOwned)
		if err != nil {
			return nil, err
		}

		stamp.DateAdded, _ = time.Parse(time.RFC3339, dateAdded)
		stamp.DateModified, _ = time.Parse(time.RFC3339, dateModified)
		stamp.Tags, _ = s.getStampTags(stamp.ID)
		
		// Load instances for this stamp
		stamp.Instances, _ = s.getStampInstances(stamp.ID)
		
		// Populate BoxNames for list view display
		stamp.BoxNames, _ = s.getStampBoxNames(stamp.ID)
		
		stamps = append(stamps, stamp)
	}
	return stamps, nil
}

func (s *StampService) GetStampByID(id string) (*models.Stamp, error) {
	var stamp models.Stamp
	var dateAdded, dateModified string
	query := `
		SELECT s.id, s.name, s.scott_number, s.issue_date, s.series, 
		       s.notes, s.image_url, s.date_added, s.date_modified
		FROM stamps s
		WHERE s.id = ? AND s.date_deleted IS NULL`

	err := s.db.QueryRow(query, id).Scan(&stamp.ID, &stamp.Name, &stamp.ScottNumber, &stamp.IssueDate,
		&stamp.Series, &stamp.Notes, &stamp.ImageURL, &dateAdded, &dateModified)

	if err != nil {
		return nil, err
	}

	// Parse timestamps
	if stamp.DateAdded, err = time.Parse(time.RFC3339, dateAdded); err != nil {
		stamp.DateAdded = time.Now()
	}
	if stamp.DateModified, err = time.Parse(time.RFC3339, dateModified); err != nil {
		stamp.DateModified = time.Now()
	}

	// Get tags
	stamp.Tags, _ = s.getStampTags(stamp.ID)
	
	// Get all instances
	stamp.Instances, _ = s.getStampInstances(stamp.ID)
	
	// Set IsOwned based on whether we have any instances
	stamp.IsOwned = len(stamp.Instances) > 0

	return &stamp, nil
}

func (s *StampService) CreateStamp(stamp *models.Stamp) (*models.Stamp, error) {
	_, err := s.db.Exec(`INSERT INTO stamps 
		(id, name, scott_number, issue_date, series, notes, image_url, is_owned, date_added, date_modified) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		stamp.ID, stamp.Name, stamp.ScottNumber, stamp.IssueDate, stamp.Series, 
		stamp.Notes, stamp.ImageURL, stamp.IsOwned, 
		stamp.DateAdded.Format(time.RFC3339), stamp.DateModified.Format(time.RFC3339))

	if err != nil {
		return nil, err
	}

	// Handle tags
	if len(stamp.Tags) > 0 {
		s.updateStampTags(stamp.ID, stamp.Tags)
	}

	return stamp, nil
}

func (s *StampService) UpdateStamp(stamp *models.Stamp) (*models.Stamp, error) {
	log.Printf("Updating stamp with ID: %s", stamp.ID)
	
	query := `UPDATE stamps SET 
		name=?, scott_number=?, issue_date=?, series=?, notes=?, image_url=?, 
		is_owned=?, date_modified=?
		WHERE id=? AND date_deleted IS NULL`
	
	result, err := s.db.Exec(query,
		stamp.Name, stamp.ScottNumber, stamp.IssueDate, stamp.Series, stamp.Notes, stamp.ImageURL,
		stamp.IsOwned, stamp.DateModified.Format(time.RFC3339), stamp.ID)

	if err != nil {
		log.Printf("Error executing UPDATE query: %v", err)
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return nil, err
	}
	
	if rowsAffected == 0 {
		log.Printf("Warning: No rows were updated for stamp ID: %s", stamp.ID)
		return nil, fmt.Errorf("no stamp found with ID: %s", stamp.ID)
	}

	// Update tags
	err = s.updateStampTags(stamp.ID, stamp.Tags)
	if err != nil {
		log.Printf("Error updating tags: %v", err)
		return nil, fmt.Errorf("failed to update tags: %v", err)
	}

	return stamp, nil
}

func (s *StampService) DeleteStamp(id string) error {
	// Soft delete the stamp and all its instances
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	now := time.Now().Format(time.RFC3339)
	
	// Soft delete all instances
	_, err = tx.Exec("UPDATE stamp_instances SET date_deleted = ? WHERE stamp_id = ? AND date_deleted IS NULL", now, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Remove tag associations
	_, err = tx.Exec("DELETE FROM stamp_tags WHERE stamp_id = ?", id)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Soft delete the stamp
	_, err = tx.Exec("UPDATE stamps SET date_deleted = ? WHERE id = ? AND date_deleted IS NULL", now, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// Helper functions

func (s *StampService) getStampTags(stampID string) ([]string, error) {
	rows, err := s.db.Query(`
		SELECT t.name 
		FROM tags t 
		JOIN stamp_tags st ON t.id = st.tag_id 
		WHERE st.stamp_id = ?`, stampID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func (s *StampService) getStampInstances(stampID string) ([]models.StampInstance, error) {
	rows, err := s.db.Query(`
		SELECT si.id, si.stamp_id, si.condition, si.box_id, sb.name as box_name,
		       si.quantity, si.date_added, si.date_modified
		FROM stamp_instances si
		LEFT JOIN storage_boxes sb ON si.box_id = sb.id
		WHERE si.stamp_id = ? AND si.date_deleted IS NULL
		ORDER BY si.condition, sb.name`, stampID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []models.StampInstance
	for rows.Next() {
		var instance models.StampInstance
		var dateAdded, dateModified string
		
		err := rows.Scan(&instance.ID, &instance.StampID, &instance.Condition, 
			&instance.BoxID, &instance.BoxName, &instance.Quantity, &dateAdded, &dateModified)
		if err != nil {
			return nil, err
		}

		instance.DateAdded, _ = time.Parse(time.RFC3339, dateAdded)
		instance.DateModified, _ = time.Parse(time.RFC3339, dateModified)
		
		instances = append(instances, instance)
	}
	return instances, nil
}

func (s *StampService) updateStampTags(stampID string, tags []string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	// Remove existing tags for this stamp
	_, err = tx.Exec("DELETE FROM stamp_tags WHERE stamp_id = ?", stampID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Add new tags
	for _, tagName := range tags {
		tagName = strings.TrimSpace(tagName)
		if tagName == "" {
			continue
		}

		// Get or create tag
		var tagID string
		err := tx.QueryRow("SELECT id FROM tags WHERE name = ?", tagName).Scan(&tagID)
		if err == sql.ErrNoRows {
			// Create new tag
			tagID = uuid.New().String()
			_, err = tx.Exec("INSERT INTO tags (id, name) VALUES (?, ?)", tagID, tagName)
			if err != nil {
				tx.Rollback()
				return err
			}
		} else if err != nil {
			tx.Rollback()
			return err
		}

		// Link stamp to tag
		_, err = tx.Exec("INSERT INTO stamp_tags (stamp_id, tag_id) VALUES (?, ?)", stampID, tagID)
		if err != nil {
			if !strings.Contains(err.Error(), "UNIQUE constraint failed") {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit()
}

func (s *StampService) getStampBoxNames(stampID string) ([]string, error) {
	rows, err := s.db.Query(`
		SELECT DISTINCT sb.name 
		FROM stamp_instances si
		JOIN storage_boxes sb ON si.box_id = sb.id
		WHERE si.stamp_id = ? AND si.date_deleted IS NULL AND si.box_id IS NOT NULL
		ORDER BY sb.name`, stampID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var boxNames []string
	for rows.Next() {
		var boxName string
		if err := rows.Scan(&boxName); err != nil {
			return nil, err
		}
		boxNames = append(boxNames, boxName)
	}
	return boxNames, nil
}