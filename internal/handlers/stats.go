package handlers

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/jeepinbird/stampkeeper/internal/services"
)

type StatsHandler struct {
	db        *sql.DB
	templates *template.Template
	service   *services.StatsService
}

func NewStatsHandler(db *sql.DB, templates *template.Template) *StatsHandler {
	return &StatsHandler{
		db:        db,
		templates: templates,
		service:   services.NewStatsService(db),
	}
}

func (h *StatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}