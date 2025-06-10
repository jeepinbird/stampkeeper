package services

import (
	"database/sql"

	"github.com/jeepinbird/stampkeeper/internal/models"
)

type TagService struct {
	db *sql.DB
}

func NewTagService(db *sql.DB) *TagService {
	return &TagService{db: db}
}

func (s *TagService) GetTags() ([]models.Tag, error) {
	query := `
		SELECT t.id, t.name, COUNT(st.stamp_id) as stamp_count
		FROM tags t
		LEFT JOIN stamp_tags st ON t.id = st.tag_id
		GROUP BY t.id, t.name
		ORDER BY t.name`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		err := rows.Scan(&tag.ID, &tag.Name, &tag.StampCount)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

func (s *TagService) CreateTag(tag *models.Tag) (*models.Tag, error) {
	_, err := s.db.Exec(`INSERT INTO tags (id, name) VALUES ($1, $2)`, tag.ID, tag.Name)
	if err != nil {
		return nil, err
	}

	return tag, nil
}

func (s *TagService) UpdateTag(tag *models.Tag) (*models.Tag, error) {
	_, err := s.db.Exec(`UPDATE tags SET name = $1 WHERE id = $2`, tag.Name, tag.ID)
	if err != nil {
		return nil, err
	}

	return tag, nil
}

func (s *TagService) DeleteTag(id string) error {
	_, err := s.db.Exec("DELETE FROM tags WHERE id = $1", id)
	return err
}