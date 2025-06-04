package services

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jeepinbird/stampkeeper/internal/models"
)

type InstanceService struct {
	db *sql.DB
}

func NewInstanceService(db *sql.DB) *InstanceService {
	return &InstanceService{db: db}
}

func (s *InstanceService) CreateStampInstance(instance *models.StampInstance) (*models.StampInstance, error) {
	_, err := s.db.Exec(`INSERT INTO stamp_instances 
		(id, stamp_id, condition, box_id, quantity, date_added, date_modified) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		instance.ID, instance.StampID, instance.Condition, instance.BoxID, 
		instance.Quantity, instance.DateAdded.Format(time.RFC3339), instance.DateModified.Format(time.RFC3339))

	if err != nil {
		return nil, err
	}

	return instance, nil
}

func (s *InstanceService) UpdateStampInstance(instance *models.StampInstance) (*models.StampInstance, error) {
	query := `UPDATE stamp_instances SET 
		condition=?, box_id=?, quantity=?, date_modified=?
		WHERE id=? AND date_deleted IS NULL`
	
	result, err := s.db.Exec(query,
		instance.Condition, instance.BoxID, instance.Quantity, 
		instance.DateModified.Format(time.RFC3339), instance.ID)

	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	
	if rowsAffected == 0 {
		return nil, fmt.Errorf("no instance found with ID: %s", instance.ID)
	}

	return instance, nil
}

func (s *InstanceService) DeleteStampInstance(id string) error {
	now := time.Now().Format(time.RFC3339)
	_, err := s.db.Exec("UPDATE stamp_instances SET date_deleted = ? WHERE id = ? AND date_deleted IS NULL", now, id)
	return err
}

func (s *InstanceService) GetStampInstance(id string) (*models.StampInstance, error) {
	var instance models.StampInstance
	var dateAdded, dateModified string
	
	query := `
		SELECT si.id, si.stamp_id, si.condition, si.box_id, sb.name as box_name, 
		       si.quantity, si.date_added, si.date_modified
		FROM stamp_instances si
		LEFT JOIN storage_boxes sb ON si.box_id = sb.id
		WHERE si.id = ? AND si.date_deleted IS NULL`

	err := s.db.QueryRow(query, id).Scan(&instance.ID, &instance.StampID, &instance.Condition, 
		&instance.BoxID, &instance.BoxName, &instance.Quantity, &dateAdded, &dateModified)

	if err != nil {
		return nil, err
	}

	instance.DateAdded, _ = time.Parse(time.RFC3339, dateAdded)
	instance.DateModified, _ = time.Parse(time.RFC3339, dateModified)

	return &instance, nil
}

// GetStampInstances returns all instances for a given stamp ID
func (s *InstanceService) GetStampInstances(stampID string) ([]models.StampInstance, error) {
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