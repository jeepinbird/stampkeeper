package handlers

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jeepinbird/stampkeeper/internal/models"
	"github.com/jeepinbird/stampkeeper/internal/services"
)

type BoxHandler struct {
	db        *sql.DB
	templates *template.Template
	service   *services.BoxService
}

func NewBoxHandler(db *sql.DB, templates *template.Template) *BoxHandler {
	return &BoxHandler{
		db:        db,
		templates: templates,
		service:   services.NewBoxService(db),
	}
}

func (h *BoxHandler) GetBoxes(w http.ResponseWriter, r *http.Request) {
	boxes, err := h.service.GetBoxes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(boxes)
}

func (h *BoxHandler) GetBox(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	box, err := h.service.GetBoxByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Box not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(box)
}

func (h *BoxHandler) CreateBox(w http.ResponseWriter, r *http.Request) {
	var box models.StorageBox
	if err := json.NewDecoder(r.Body).Decode(&box); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	box.ID = uuid.New().String()
	box.DateCreated = time.Now()

	log.Printf("handlers.boxes.CreateBox: %+v", box)

	createdBox, err := h.service.CreateBox(&box)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdBox)
}

func (h *BoxHandler) UpdateBox(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var box models.StorageBox
	if err := json.NewDecoder(r.Body).Decode(&box); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	box.ID = id

	log.Printf("handlers.boxes.UpdateBox: %+v", box)

	updatedBox, err := h.service.UpdateBox(&box)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedBox)
}

func (h *BoxHandler) DeleteBox(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	log.Printf("handlers.boxes.DeleteBox: %v", id)

	if err := h.service.DeleteBox(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}