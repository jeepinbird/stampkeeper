package services

import (
	"database/sql"

	"github.com/jeepinbird/stampkeeper/internal/models"
)

type StatsService struct {
	db *sql.DB
}

func NewStatsService(db *sql.DB) *StatsService {
	return &StatsService{db: db}
}

func (s *StatsService) GetStats() (*models.Stats, error) {
	var stats models.Stats

	// Total owned instances (sum of quantities)
	s.db.QueryRow("SELECT COALESCE(SUM(quantity), 0) FROM stamp_instances WHERE date_deleted IS NULL").Scan(&stats.TotalOwned)

	// Unique stamps (distinct stamp designs)
	s.db.QueryRow("SELECT COUNT(DISTINCT scott_number) FROM stamps WHERE scott_number IS NOT NULL AND date_deleted IS NULL").Scan(&stats.UniqueStamps)

	// Stamps needed (stamp designs with no instances)
	s.db.QueryRow(`
		SELECT COUNT(*) 
		FROM stamps s 
		WHERE s.date_deleted IS NULL 
		AND NOT EXISTS (
			SELECT 1 FROM stamp_instances si 
			WHERE si.stamp_id = s.id AND si.date_deleted IS NULL
		)`).Scan(&stats.StampsNeeded)

	// Storage boxes
	s.db.QueryRow("SELECT COUNT(*) FROM storage_boxes").Scan(&stats.StorageBoxes)

	return &stats, nil
}