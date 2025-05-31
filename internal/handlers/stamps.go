package handlers

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jeepinbird/stampkeeper/internal/models"
	"github.com/jeepinbird/stampkeeper/internal/services"
)

type StampHandler struct {
	db        *sql.DB
	templates *template.Template
	service   *services.StampService
}

func NewStampHandler(db *sql.DB, templates *template.Template) *StampHandler {
	return &StampHandler{
		db:        db,
		templates: templates,
		service:   services.NewStampService(db),
	}
}

func (h *StampHandler) GetStamps(w http.ResponseWriter, r *http.Request) {
	stamps, err := h.service.GetStamps(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stamps)
}

func (h *StampHandler) GetStamp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	stamp, err := h.service.GetStampByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Stamp not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stamp)
}

func (h *StampHandler) CreateStamp(w http.ResponseWriter, r *http.Request) {
	var stamp models.Stamp
	if err := json.NewDecoder(r.Body).Decode(&stamp); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set default values
	stamp.ID = uuid.New().String()
	stamp.DateAdded = time.Now()
	stamp.DateModified = time.Now()
	if stamp.Quantity == 0 {
		stamp.Quantity = 1
	}

	createdStamp, err := h.service.CreateStamp(&stamp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdStamp)
}

func (h *StampHandler) UpdateStamp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var stamp models.Stamp
	if err := json.NewDecoder(r.Body).Decode(&stamp); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	stamp.ID = id
	stamp.DateModified = time.Now()

	updatedStamp, err := h.service.UpdateStamp(&stamp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedStamp)
}

func (h *StampHandler) DeleteStamp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.service.DeleteStamp(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}