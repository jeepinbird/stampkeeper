package services

import (
	"database/sql"
	"log"

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
	query := "SELECT COALESCE(SUM(quantity), 0) FROM stamp_instances WHERE date_deleted IS NULL"
	if err := s.db.QueryRow(query).Scan(&stats.TotalOwned); err != nil {
		return nil, err
	}
	log.Printf("services.stats.GetStats: TotalOwned query result = %d", stats.TotalOwned)

	// Unique stamps owned (distinct stamp designs)
	if err := s.db.QueryRow("SELECT COALESCE(COUNT(DISTINCT stamp_id),0) FROM stamp_instances WHERE date_deleted IS NULL").Scan(&stats.UniqueStamps); err != nil {
		return nil, err
	}

	// Storage boxes
	if err := s.db.QueryRow("SELECT COUNT(*) FROM storage_boxes").Scan(&stats.StorageBoxes); err != nil {
		return nil, err
	}

	return &stats, nil
}
