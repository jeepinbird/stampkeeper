package services

import (
	"database/sql"
	"fmt"
	"log"
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
	sql := `INSERT INTO stamp_instances 
		(id, stamp_id, condition, box_id, quantity, date_added, date_modified) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := s.db.Exec(sql,
		instance.ID, instance.StampID, instance.Condition, instance.BoxID, 
		instance.Quantity, instance.DateAdded, instance.DateModified)

	if err != nil {
		return nil, err
	}

	return instance, nil
}

func (s *InstanceService) UpdateStampInstance(instance *models.StampInstance) (*models.StampInstance, error) {
	query := `UPDATE stamp_instances SET 
		condition=$1, box_id=$2, quantity=$3, date_modified=$4
		WHERE id=$5 AND date_deleted IS NULL`
	
	result, err := s.db.Exec(query,
		instance.Condition, instance.BoxID, instance.Quantity, 
		instance.DateModified, instance.ID)

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
	result, err := s.db.Exec("DELETE FROM stamp_instances WHERE id = $1", id)

	if err != nil {
		return nil
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil
	}
	
	if rowsAffected == 0 {
		log.Printf("no instance found with ID: %s", id)
		return nil
	}

	log.Printf("Successfully deleted instance ID: %s", id)
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
		WHERE si.id = $1 AND si.date_deleted IS NULL`

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