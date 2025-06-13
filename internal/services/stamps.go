package services

import (
	"database/sql"
	"net/http"
	"strings"
	"time"
	"log"
	"fmt"

	"github.com/google/uuid"
	"github.com/jeepinbird/stampkeeper/internal/database"
	"github.com/jeepinbird/stampkeeper/internal/models"
)

type StampService struct {
	db *sql.DB
}

// StampFilters holds all filter parameters for stamp queries
type StampFilters struct {
	Search string
	Owned  string
	BoxID  string
	Sort   string
	Order  string
	Limit  int
	Offset int
}

// NewStampFiltersFromRequest creates StampFilters from HTTP request parameters
func NewStampFiltersFromRequest(r *http.Request, page, limit int) StampFilters {
	order := r.URL.Query().Get("order")
	if order == "" {
		order = "ASC"
	}
	order = strings.ToUpper(order)
	if order != "ASC" && order != "DESC" {
		order = "ASC"
	}

	// Handle both old 'owned' parameter and new 'owned_filter' parameter
	owned := r.URL.Query().Get("owned")
	if owned == "" {
		ownedFilter := r.URL.Query().Get("owned_filter")
		if ownedFilter == "all" {
			owned = ""
		} else {
			owned = ownedFilter
		}
	}

	return StampFilters{
		Search: r.URL.Query().Get("search"),
		Owned:  owned,
		BoxID:  r.URL.Query().Get("box_id"),
		Sort:   r.URL.Query().Get("sort"),
		Order:  order,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}
}

func NewStampService(db *sql.DB) *StampService {
	return &StampService{db: db}
}

// Gets the total count of unique stamps (not instances) matching filters
func (s *StampService) GetStampCount(r *http.Request) (int64, error) {
	filters := NewStampFiltersFromRequest(r, 1, 1) // Page/limit not needed for count
	
	qb := database.NewQueryBuilder(`
		SELECT COUNT(DISTINCT s.id) 
		FROM stamps s 
		LEFT JOIN stamp_instances si ON s.id = si.stamp_id AND si.date_deleted IS NULL
		WHERE s.date_deleted IS NULL`)

	qb.AddSearchFilter(filters.Search, "s")
	
	if filters.Owned != "" {
		if filters.Owned == "true" {
			qb.AddCondition(` AND EXISTS (SELECT 1 FROM stamp_instances si2 WHERE si2.stamp_id = s.id AND si2.date_deleted IS NULL)`)
		} else if filters.Owned == "false" {
			qb.AddCondition(` AND NOT EXISTS (SELECT 1 FROM stamp_instances si2 WHERE si2.stamp_id = s.id AND si2.date_deleted IS NULL)`)
		}
	}

	if filters.BoxID != "" {
		qb.AddCondition(` AND EXISTS (SELECT 1 FROM stamp_instances si3 WHERE si3.stamp_id = s.id AND si3.box_id = ? AND si3.date_deleted IS NULL)`, filters.BoxID)
	}

	query, args := qb.GetQuery()
	var count int64
	err := s.db.QueryRow(query, args...).Scan(&count)
	return count, err
}

func (s *StampService) GetStamps(r *http.Request, page, limit int) ([]models.Stamp, error) {
	filters := NewStampFiltersFromRequest(r, page, limit)
	
	if filters.BoxID != "" {
		return s.getStampsInBox(filters)
	}
	return s.getGeneralStamps(filters)
}

func (s *StampService) getGeneralStamps(filters StampFilters) ([]models.Stamp, error) {
	qb := database.NewQueryBuilder(`
		SELECT s.id, s.name, s.scott_number, s.issue_date, s.series, 
		       s.notes, s.image_url, s.date_added, s.date_modified,
		       CASE WHEN COUNT(si.id) > 0 THEN true ELSE false END as is_owned
		FROM stamps s
		LEFT JOIN stamp_instances si ON s.id = si.stamp_id AND si.date_deleted IS NULL
		WHERE s.date_deleted IS NULL`)

	qb.AddSearchFilter(filters.Search, "s")
	qb.AddCondition(` GROUP BY s.id, s.name, s.scott_number, s.issue_date, s.series, s.notes, s.image_url, s.date_added, s.date_modified`)
	qb.AddOwnedFilter(filters.Owned, "si")
	qb.AddSortAndLimit(filters.Sort, filters.Order, filters.Limit, filters.Offset, "s")

	query, args := qb.GetQuery()
	return s.executeStampQuery(query, args)
}

func (s *StampService) getStampsInBox(filters StampFilters) ([]models.Stamp, error) {
	qb := database.NewQueryBuilder(`
		SELECT DISTINCT s.id, s.name, s.scott_number, s.issue_date, s.series, 
		       s.notes, s.image_url, s.date_added, s.date_modified, true as is_owned
		FROM stamps s
		JOIN stamp_instances si ON s.id = si.stamp_id 
		WHERE s.date_deleted IS NULL AND si.date_deleted IS NULL`)

	qb.AddBoxFilter(filters.BoxID, "si")
	qb.AddSearchFilter(filters.Search, "s")
	qb.AddSortAndLimit(filters.Sort, filters.Order, filters.Limit, filters.Offset, "s")

	query, args := qb.GetQuery()
	return s.executeStampQuery(query, args)
}

func (s *StampService) executeStampQuery(query string, args []interface{}) ([]models.Stamp, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stamps []models.Stamp
	for rows.Next() {
		var stamp models.Stamp
		var dateAdded, dateModified time.Time
		err := rows.Scan(&stamp.ID, &stamp.Name, &stamp.ScottNumber, &stamp.IssueDate, &stamp.Series,
			&stamp.Notes, &stamp.ImageURL, &dateAdded, &dateModified, &stamp.IsOwned)
		if err != nil {
			return nil, err
		}

		stamp.DateAdded = dateAdded
		stamp.DateModified = dateModified
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
	sql := `SELECT s.id, s.name, s.scott_number, s.issue_date, s.series, 
		       s.notes, s.image_url, s.date_added, s.date_modified
		FROM stamps s
		WHERE s.id = $1 AND s.date_deleted IS NULL`

	var stamp models.Stamp
	var dateAdded, dateModified time.Time
	err := s.db.QueryRow(sql, id).Scan(&stamp.ID, &stamp.Name, &stamp.ScottNumber, &stamp.IssueDate,
		&stamp.Series, &stamp.Notes, &stamp.ImageURL, &dateAdded, &dateModified)

	if err != nil {
		return nil, err
	}

	stamp.DateAdded = dateAdded
	stamp.DateModified = dateModified

	// Get tags
	stamp.Tags, _ = s.getStampTags(stamp.ID)
	
	// Get all instances
	stamp.Instances, _ = s.getStampInstances(stamp.ID)
	
	// Set IsOwned based on whether we have any instances
	stamp.IsOwned = len(stamp.Instances) > 0

	return &stamp, nil
}

func (s *StampService) CreateStamp(stamp *models.Stamp) (*models.Stamp, error) {
	sql := `INSERT INTO stamps 
		(id, name, scott_number, issue_date, series, notes, image_url, is_owned, date_added, date_modified) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	
	_, err := s.db.Exec(sql,
		stamp.ID, stamp.Name, stamp.ScottNumber, stamp.IssueDate, stamp.Series, 
		stamp.Notes, stamp.ImageURL, stamp.IsOwned, 
		stamp.DateAdded, stamp.DateModified)

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
		name=$1, scott_number=$2, issue_date=$3, series=$4, notes=$5, image_url=$6, 
		is_owned=$7, date_modified=$8
		WHERE id=$9 AND date_deleted IS NULL`
	
	result, err := s.db.Exec(query,
		stamp.Name, stamp.ScottNumber, stamp.IssueDate, stamp.Series, stamp.Notes, stamp.ImageURL,
		stamp.IsOwned, stamp.DateModified, stamp.ID)

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

	now := time.Now()
	
	// Soft delete all instances
	_, err = tx.Exec("UPDATE stamp_instances SET date_deleted = $1 WHERE stamp_id = $2 AND date_deleted IS NULL", now, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Remove tag associations
	_, err = tx.Exec("DELETE FROM stamp_tags WHERE stamp_id = $1", id)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Soft delete the stamp
	_, err = tx.Exec("UPDATE stamps SET date_deleted = $1 WHERE id = $2 AND date_deleted IS NULL", now, id)
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
		WHERE st.stamp_id = $1`, stampID)
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
		WHERE si.stamp_id = $1 AND si.date_deleted IS NULL
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
	_, err = tx.Exec("DELETE FROM stamp_tags WHERE stamp_id = $1", stampID)
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
		err := tx.QueryRow("SELECT id FROM tags WHERE name = $1", tagName).Scan(&tagID)
		if err == sql.ErrNoRows {
			// Create new tag
			tagID = uuid.New().String()
			_, err = tx.Exec("INSERT INTO tags (id, name) VALUES ($1, $2)", tagID, tagName)
			if err != nil {
				tx.Rollback()
				return err
			}
		} else if err != nil {
			tx.Rollback()
			return err
		}

		// Link stamp to tag
		_, err = tx.Exec("INSERT INTO stamp_tags (stamp_id, tag_id) VALUES ($1, $2)", stampID, tagID)
		if err != nil {
			if !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
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
		WHERE si.stamp_id = $1 AND si.date_deleted IS NULL AND si.box_id IS NOT NULL
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