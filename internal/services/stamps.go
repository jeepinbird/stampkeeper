package services

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jeepinbird/stampkeeper/internal/models"
)

type StampService struct {
	db *sql.DB
}

func NewStampService(db *sql.DB) *StampService {
	return &StampService{db: db}
}

func (s *StampService) GetStamps(r *http.Request) ([]models.Stamp, error) {
	query := `
		SELECT s.id, s.name, s.scott_number, s.issue_date, s.series, s.condition, 
		       s.quantity, s.box_id, sb.name as box_name, s.notes, s.image_url, 
		       s.is_owned, s.date_added, s.date_modified
		FROM stamps s
		LEFT JOIN storage_boxes sb ON s.box_id = sb.id
		WHERE 1=1`

	args := []interface{}{}

	// Add filters based on query parameters
	if search := r.URL.Query().Get("search"); search != "" {
		query += ` AND (s.name ILIKE ? OR s.scott_number ILIKE ? OR s.series ILIKE ?)`
		searchParam := "%" + search + "%"
		args = append(args, searchParam, searchParam, searchParam)
	}

	if owned := r.URL.Query().Get("owned"); owned != "" {
		if owned == "true" {
			query += ` AND s.is_owned = true`
		} else if owned == "false" {
			query += ` AND s.is_owned = false`
		}
	}

	if boxID := r.URL.Query().Get("box_id"); boxID != "" {
		query += ` AND s.box_id = ?`
		args = append(args, boxID)
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
	case "scott_number":
		query += ` ORDER BY s.scott_number ` + order
	case "name":
		query += ` ORDER BY s.name ` + order
	case "issue_date":
		query += ` ORDER BY s.issue_date ` + order
	case "date_added":
		query += ` ORDER BY s.date_added DESC`
	default:
		query += ` ORDER BY s.date_added DESC`
	}

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
			&stamp.Condition, &stamp.Quantity, &stamp.BoxID, &stamp.BoxName, &stamp.Notes, &stamp.ImageURL,
			&stamp.IsOwned, &dateAdded, &dateModified)
		if err != nil {
			return nil, err
		}

		stamp.DateAdded, _ = time.Parse(time.RFC3339, dateAdded)
		stamp.DateModified, _ = time.Parse(time.RFC3339, dateModified)
		stamp.Tags, _ = s.getStampTags(stamp.ID)
		stamps = append(stamps, stamp)
	}
	return stamps, nil
}

func (s *StampService) GetStampByID(id string) (*models.Stamp, error) {
	var stamp models.Stamp
	var dateAdded, dateModified string
	query := `
		SELECT s.id, s.name, s.scott_number, s.issue_date, s.series, s.condition, 
		       s.quantity, s.box_id, sb.name as box_name, s.notes, s.image_url, 
		       s.is_owned, s.date_added, s.date_modified
		FROM stamps s
		LEFT JOIN storage_boxes sb ON s.box_id = sb.id
		WHERE s.id = ?`

	err := s.db.QueryRow(query, id).Scan(&stamp.ID, &stamp.Name, &stamp.ScottNumber, &stamp.IssueDate,
		&stamp.Series, &stamp.Condition, &stamp.Quantity, &stamp.BoxID, &stamp.BoxName, &stamp.Notes,
		&stamp.ImageURL, &stamp.IsOwned, &dateAdded, &dateModified)

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

	return &stamp, nil
}

func (s *StampService) CreateStamp(stamp *models.Stamp) (*models.Stamp, error) {
	_, err := s.db.Exec(`INSERT INTO stamps 
		(id, name, scott_number, issue_date, series, condition, quantity, box_id, notes, image_url, is_owned, date_added, date_modified) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		stamp.ID, stamp.Name, stamp.ScottNumber, stamp.IssueDate, stamp.Series, stamp.Condition, stamp.Quantity,
		stamp.BoxID, stamp.Notes, stamp.ImageURL, stamp.IsOwned, stamp.DateAdded.Format(time.RFC3339), stamp.DateModified.Format(time.RFC3339))

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
	_, err := s.db.Exec(`UPDATE stamps SET 
		name=?, scott_number=?, issue_date=?, series=?, condition=?, quantity=?, 
		box_id=?, notes=?, image_url=?, is_owned=?, date_modified=?
		WHERE id=?`,
		stamp.Name, stamp.ScottNumber, stamp.IssueDate, stamp.Series, stamp.Condition, stamp.Quantity,
		stamp.BoxID, stamp.Notes, stamp.ImageURL, stamp.IsOwned, stamp.DateModified.Format(time.RFC3339), stamp.ID)

	if err != nil {
		return nil, err
	}

	// Update tags
	s.updateStampTags(stamp.ID, stamp.Tags)

	return stamp, nil
}

func (s *StampService) DeleteStamp(id string) error {
	// First, delete associations in stamp_tags
	_, err := s.db.Exec("DELETE FROM stamp_tags WHERE stamp_id = ?", id)
	if err != nil {
		return err
	}

	// Then, delete the stamp
	_, err = s.db.Exec("DELETE FROM stamps WHERE id = ?", id)
	return err
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

func (s *StampService) updateStampTags(stampID string, tags []string) error {
	// Begin a transaction
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