package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/jeepinbird/stampkeeper/internal/models"
	"github.com/jeepinbird/stampkeeper/internal/services"
)

// HTMXHandler handles HTMX-specific endpoints that return HTML fragments
type HTMXHandler struct {
	db           *sql.DB
	templates    *template.Template
	stampService *services.StampService
	tagService   *services.TagService
}

func NewHTMXHandler(db *sql.DB, templates *template.Template) *HTMXHandler {
	return &HTMXHandler{
		db:           db,
		templates:    templates,
		stampService: services.NewStampService(db),
		tagService:   services.NewTagService(db),
	}
}

// UpdateStampField handles individual field updates for stamps
func (h *HTMXHandler) UpdateStampField(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	stampID := vars["id"]
	field := vars["field"]

	// Parse the field value from the form
	var value interface{}
	var err error

	switch field {
	case "name", "scott_number", "series", "notes":
		value = strings.TrimSpace(r.FormValue("value"))
		if value == "" {
			value = nil
		}
	case "issue_date":
		dateStr := strings.TrimSpace(r.FormValue("value"))
		if dateStr == "" {
			value = nil
		} else {
			value = dateStr
		}
	default:
		http.Error(w, "Invalid field", http.StatusBadRequest)
		return
	}

	// Get the current stamp
	stamp, err := h.stampService.GetStampByID(stampID)
	if err != nil {
		http.Error(w, "Stamp not found", http.StatusNotFound)
		return
	}

	// Update the specific field
	switch field {
	case "name":
		if value != nil {
			stamp.Name = value.(string)
		}
	case "scott_number":
		if value != nil {
			valueStr := value.(string)
			stamp.ScottNumber = &valueStr
		} else {
			stamp.ScottNumber = nil
		}
	case "series":
		if value != nil {
			valueStr := value.(string)
			stamp.Series = &valueStr
		} else {
			stamp.Series = nil
		}
	case "issue_date":
		if value != nil {
			valueStr := value.(string)
			stamp.IssueDate = &valueStr
		} else {
			stamp.IssueDate = nil
		}
	case "notes":
		if value != nil {
			valueStr := value.(string)
			stamp.Notes = &valueStr
		} else {
			stamp.Notes = nil
		}
	}

	// Update timestamp
	stamp.DateModified = time.Now()

	// Save the updated stamp
	_, err = h.stampService.UpdateStamp(stamp)
	if err != nil {
		http.Error(w, "Failed to update stamp", http.StatusInternalServerError)
		return
	}

	// Return success indicator (green flash)
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<div class="field-update-success"></div>`))
}

// AddStampTag adds a new tag to a stamp and returns the updated tags section
func (h *HTMXHandler) AddStampTag(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	stampID := vars["id"]
	tagName := strings.TrimSpace(r.FormValue("tag_name"))

	if tagName == "" {
		http.Error(w, "Tag name is required", http.StatusBadRequest)
		return
	}

	// Get the current stamp
	stamp, err := h.stampService.GetStampByID(stampID)
	if err != nil {
		http.Error(w, "Stamp not found", http.StatusNotFound)
		return
	}

	// Check if tag already exists on this stamp
	for _, existingTag := range stamp.Tags {
		if strings.EqualFold(existingTag, tagName) {
			http.Error(w, "Tag already exists", http.StatusConflict)
			return
		}
	}

	// Add the new tag
	stamp.Tags = append(stamp.Tags, tagName)
	stamp.DateModified = time.Now()

	// Update the stamp
	_, err = h.stampService.UpdateStamp(stamp)
	if err != nil {
		http.Error(w, "Failed to add tag", http.StatusInternalServerError)
		return
	}

	// Return the updated tags section
	data := models.StampDetailView{Stamp: *stamp}
	
	w.Header().Set("Content-Type", "text/html")
	err = h.templates.ExecuteTemplate(w, "stamp-tags-section", data)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
}

// RemoveStampTag removes a tag from a stamp and returns the updated tags section
func (h *HTMXHandler) RemoveStampTag(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	stampID := vars["id"]
	tagName := vars["tag"]

	// Get the current stamp
	stamp, err := h.stampService.GetStampByID(stampID)
	if err != nil {
		http.Error(w, "Stamp not found", http.StatusNotFound)
		return
	}

	// Remove the tag
	var newTags []string
	for _, existingTag := range stamp.Tags {
		if !strings.EqualFold(existingTag, tagName) {
			newTags = append(newTags, existingTag)
		}
	}

	stamp.Tags = newTags
	stamp.DateModified = time.Now()

	// Update the stamp
	_, err = h.stampService.UpdateStamp(stamp)
	if err != nil {
		http.Error(w, "Failed to remove tag", http.StatusInternalServerError)
		return
	}

	// Return the updated tags section
	data := models.StampDetailView{Stamp: *stamp}
	
	w.Header().Set("Content-Type", "text/html")
	err = h.templates.ExecuteTemplate(w, "stamp-tags-section", data)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
}

// GetFieldUpdateIndicator returns a visual indicator for successful field updates
func (h *HTMXHandler) GetFieldUpdateIndicator(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<div class="field-success-indicator" style="background-color: #d4edda; padding: 2px; border-radius: 3px; animation: fadeOut 2s forwards;">✓</div>`))
}