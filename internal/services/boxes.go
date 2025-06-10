package services

import (
	"database/sql"
	"time"

	"github.com/jeepinbird/stampkeeper/internal/models"
)

type BoxService struct {
	db *sql.DB
}

func NewBoxService(db *sql.DB) *BoxService {
	return &BoxService{db: db}
}

func (s *BoxService) GetBoxes() ([]models.StorageBox, error) {
	query := `
		SELECT sb.id, sb.name, sb.date_created, COALESCE(SUM(si.quantity), 0) as instance_count
		FROM storage_boxes sb
		LEFT JOIN stamp_instances si ON sb.id = si.box_id AND si.date_deleted IS NULL
		GROUP BY sb.id, sb.name, sb.date_created
		ORDER BY sb.name`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var boxes []models.StorageBox
	for rows.Next() {
		var box models.StorageBox
		var dateCreated string
		err := rows.Scan(&box.ID, &box.Name, &dateCreated, &box.StampCount)
		if err != nil {
			return nil, err
		}
		box.DateCreated, _ = time.Parse(time.RFC3339, dateCreated)
		boxes = append(boxes, box)
	}
	return boxes, nil
}

func (s *BoxService) GetBoxByID(id string) (*models.StorageBox, error) {
	var box models.StorageBox
	var dateCreated string
	err := s.db.QueryRow(`SELECT id, name, date_created FROM storage_boxes WHERE id = $1`, id).
		Scan(&box.ID, &box.Name, &dateCreated)

	if err != nil {
		return nil, err
	}

	// Parse timestamp
	if box.DateCreated, err = time.Parse(time.RFC3339, dateCreated); err != nil {
		box.DateCreated = time.Now()
	}

	return &box, nil
}

func (s *BoxService) CreateBox(box *models.StorageBox) (*models.StorageBox, error) {
	_, err := s.db.Exec(`INSERT INTO storage_boxes (id, name, date_created) VALUES ($1, $2, $3)`,
		box.ID, box.Name, box.DateCreated)

	if err != nil {
		return nil, err
	}

	return box, nil
}

func (s *BoxService) UpdateBox(box *models.StorageBox) (*models.StorageBox, error) {
	_, err := s.db.Exec(`UPDATE storage_boxes SET name = $1 WHERE id = $2`, box.Name, box.ID)
	if err != nil {
		return nil, err
	}

	return box, nil
}

func (s *BoxService) DeleteBox(id string) error {
	// Set box_id to NULL for all instances in this box
	_, err := s.db.Exec("UPDATE stamp_instances SET box_id = NULL WHERE box_id = $1", id)
	if err != nil {
		return err
	}

	// Delete the box
	_, err = s.db.Exec("DELETE FROM storage_boxes WHERE id = $1", id)
	return err
}