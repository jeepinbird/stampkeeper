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

	// Total owned stamps
	s.db.QueryRow("SELECT COUNT(*) FROM stamps WHERE is_owned = true").Scan(&stats.TotalOwned)

	// Unique stamps (distinct scott numbers, but handle nulls)
	s.db.QueryRow("SELECT COUNT(DISTINCT scott_number) FROM stamps WHERE scott_number IS NOT NULL").Scan(&stats.UniqueStamps)

	// Stamps needed
	s.db.QueryRow("SELECT COUNT(*) FROM stamps WHERE is_owned = false").Scan(&stats.StampsNeeded)

	// Storage boxes
	s.db.QueryRow("SELECT COUNT(*) FROM storage_boxes").Scan(&stats.StorageBoxes)

	return &stats, nil
}