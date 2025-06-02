package handlers

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"time"
	"fmt"

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
	// Get pagination params from query string for the API
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 50 // Default limit for API calls
	}

	// Call the service with the new arguments
	stamps, err := h.service.GetStamps(r, page, limit)
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

	// First, get the existing stamp
	existingStamp, err := h.service.GetStampByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Stamp not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Parse the incoming JSON into a map to handle partial updates
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Apply updates to the existing stamp
	if name, ok := updates["name"].(string); ok {
		existingStamp.Name = name
	}

	if scottNumber, ok := updates["scott_number"].(string); ok {
		if scottNumber == "" {
			existingStamp.ScottNumber = nil
		} else {
			existingStamp.ScottNumber = &scottNumber
		}
	}

	if issueDate, ok := updates["issue_date"].(string); ok {
		if issueDate == "" {
			existingStamp.IssueDate = nil
		} else {
			existingStamp.IssueDate = &issueDate
		}
	}

	if series, ok := updates["series"].(string); ok {
		if series == "" {
			existingStamp.Series = nil
		} else {
			existingStamp.Series = &series
		}
	}

	if condition, ok := updates["condition"].(string); ok {
		if condition == "" {
			existingStamp.Condition = nil
		} else {
			existingStamp.Condition = &condition
		}
	}

	if quantity, ok := updates["quantity"].(float64); ok {
		existingStamp.Quantity = int(quantity)
	}

	if boxID, ok := updates["box_id"]; ok {
		if boxID == nil || boxID == "" {
			existingStamp.BoxID = nil
		} else if boxIDStr, ok := boxID.(string); ok {
			existingStamp.BoxID = &boxIDStr
		}
	}

	if notes, ok := updates["notes"].(string); ok {
		if notes == "" {
			existingStamp.Notes = nil
		} else {
			existingStamp.Notes = &notes
		}
	}

	if imageURL, ok := updates["image_url"].(string); ok {
		if imageURL == "" {
			existingStamp.ImageURL = nil
		} else {
			existingStamp.ImageURL = &imageURL
		}
	}

	if isOwned, ok := updates["is_owned"].(bool); ok {
		existingStamp.IsOwned = isOwned
	}

	// Handle tags array
	if tagsInterface, ok := updates["tags"]; ok {
		if tagsArray, ok := tagsInterface.([]interface{}); ok {
			var tags []string
			for _, tag := range tagsArray {
				if tagStr, ok := tag.(string); ok {
					tags = append(tags, tagStr)
				}
			}
			existingStamp.Tags = tags
		}
	}

	// Update the modified timestamp
	existingStamp.DateModified = time.Now()

	// Save the updated stamp
	updatedStamp, err := h.service.UpdateStamp(existingStamp)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update stamp: %v", err), http.StatusInternalServerError)
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