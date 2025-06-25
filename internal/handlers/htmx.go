package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
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
	boxService   *services.BoxService
}

func NewHTMXHandler(db *sql.DB, templates *template.Template) *HTMXHandler {
	return &HTMXHandler{
		db:           db,
		templates:    templates,
		stampService: services.NewStampService(db),
		tagService:   services.NewTagService(db),
		boxService:   services.NewBoxService(db),
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

	log.Printf("handlers.htmx.RemoveStampTag: tagName = %v", tagName)

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
	log.Printf("handlers.htmx.RemoveStampTag: stamp = %+v", stamp)
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

// CreateBox creates a new storage box and returns the updated boxes table
func (h *HTMXHandler) CreateBox(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	boxName := strings.TrimSpace(r.FormValue("name"))
	if boxName == "" {
		http.Error(w, "Box name is required", http.StatusBadRequest)
		return
	}

	box := &models.StorageBox{
		ID:          uuid.New().String(),
		Name:        boxName,
		DateCreated: time.Now(),
	}

	log.Printf("handlers.htmx.CreateBox: %+v", box)

	_, err := h.boxService.CreateBox(box)
	if err != nil {
		http.Error(w, "Failed to create box", http.StatusInternalServerError)
		return
	}

	// Get all boxes for the updated table
	allBoxes, err := h.boxService.GetBoxes()
	if err != nil {
		http.Error(w, "Failed to fetch boxes", http.StatusInternalServerError)
		return
	}

	// Return the updated boxes table
	data := models.SettingsView{AllBoxes: allBoxes}
	
	w.Header().Set("Content-Type", "text/html")
	err = h.templates.ExecuteTemplate(w, "boxes-table", data)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
}

// UpdateBoxName updates a box name and returns the updated row
func (h *HTMXHandler) UpdateBoxName(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	boxID := vars["id"]
	
	boxName := strings.TrimSpace(r.FormValue("name"))
	if boxName == "" {
		http.Error(w, "Box name is required", http.StatusBadRequest)
		return
	}

	// Get the current box
	box, err := h.boxService.GetBoxByID(boxID)
	if err != nil {
		http.Error(w, "Box not found", http.StatusNotFound)
		return
	}

	// Update the name
	box.Name = boxName
	
	log.Printf("handlers.htmx.UpdateBoxName: %+v", box)

	_, err = h.boxService.UpdateBox(box)
	if err != nil {
		http.Error(w, "Failed to update box", http.StatusInternalServerError)
		return
	}

	// Return the updated box row
	w.Header().Set("Content-Type", "text/html")
	err = h.templates.ExecuteTemplate(w, "box-row", map[string]interface{}{"Box": box})
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
}

// DeleteBox deletes a box and returns the updated boxes table
func (h *HTMXHandler) DeleteBox(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	boxID := vars["id"]

	log.Printf("handlers.htmx.DeleteBox: %v", boxID)

	err := h.boxService.DeleteBox(boxID)
	if err != nil {
		http.Error(w, "Failed to delete box", http.StatusInternalServerError)
		return
	}

	// Get all boxes for the updated table
	allBoxes, err := h.boxService.GetBoxes()
	if err != nil {
		http.Error(w, "Failed to fetch boxes", http.StatusInternalServerError)
		return
	}

	// Return the updated boxes table
	data := models.SettingsView{AllBoxes: allBoxes}
	
	w.Header().Set("Content-Type", "text/html")
	err = h.templates.ExecuteTemplate(w, "boxes-table", data)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
}