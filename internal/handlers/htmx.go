package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jeepinbird/stampkeeper/internal/models"
	"github.com/jeepinbird/stampkeeper/internal/services"
)

// HTMXHandler handles HTMX-specific endpoints that return HTML fragments
type HTMXHandler struct {
	db              *sql.DB
	templates       *template.Template
	stampService    *services.StampService
	tagService      *services.TagService
	boxService      *services.BoxService
	instanceService *services.InstanceService
}

func NewHTMXHandler(db *sql.DB, templates *template.Template) *HTMXHandler {
	return &HTMXHandler{
		db:              db,
		templates:       templates,
		stampService:    services.NewStampService(db),
		tagService:      services.NewTagService(db),
		boxService:      services.NewBoxService(db),
		instanceService: services.NewInstanceService(db),
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
	var value any
	var err error

	switch field {
	case "name", "scott_number", "series", "notes", "image_url":
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
	case "image_url":
		if value != nil {
			valueStr := value.(string)
			stamp.ImageURL = &valueStr
		} else {
			stamp.ImageURL = nil
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
	if _, err := w.Write([]byte(`<div class="field-update-success"></div>`)); err != nil {
		log.Printf("handlers.htmx.UpdateStampField: Error writing response: %v", err)
	}
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

	log.Printf("handlers.htmx.AddStampTag: Called for stampID=%s, tagName=%s", stampID, tagName)

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
	if _, err := w.Write([]byte(`<div class="field-success-indicator" style="background-color: #d4edda; padding: 2px; border-radius: 3px; animation: fadeOut 2s forwards;">✓</div>`)); err != nil {
		log.Printf("handlers.htmx.GetFieldUpdateIndicator: Error writing response: %v", err)
	}
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

	// Return the updated box row (display mode)
	w.Header().Set("Content-Type", "text/html")
	err = h.templates.ExecuteTemplate(w, "box-row", box)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
}

// GetBoxEditForm returns the edit form for a box
func (h *HTMXHandler) GetBoxEditForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	boxID := vars["id"]

	box, err := h.boxService.GetBoxByID(boxID)
	if err != nil {
		http.Error(w, "Box not found", http.StatusNotFound)
		return
	}

	// Return the edit mode row
	w.Header().Set("Content-Type", "text/html")
	err = h.templates.ExecuteTemplate(w, "box-row-edit", box)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
}

// CancelBoxEdit returns the display mode for a box
func (h *HTMXHandler) CancelBoxEdit(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	boxID := vars["id"]

	box, err := h.boxService.GetBoxByID(boxID)
	if err != nil {
		http.Error(w, "Box not found", http.StatusNotFound)
		return
	}

	// Return the display mode row
	w.Header().Set("Content-Type", "text/html")
	err = h.templates.ExecuteTemplate(w, "box-row", box)
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

// DeleteStamp deletes a stamp (soft delete)
func (h *HTMXHandler) DeleteStamp(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	stampID := vars["id"]

	log.Printf("handlers.htmx.DeleteStamp: %v", stampID)

	err := h.stampService.DeleteStamp(stampID)
	if err != nil {
		http.Error(w, "Failed to delete stamp", http.StatusInternalServerError)
		return
	}

	// Return success with no content (client will handle redirect)
	w.WriteHeader(http.StatusNoContent)
}

// CreateStamp creates a new stamp and redirects to detail page
func (h *HTMXHandler) CreateStamp(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Get basic fields
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		http.Error(w, "Stamp name is required", http.StatusBadRequest)
		return
	}

	// Helper function to parse optional string fields
	parseOptionalString := func(value string) *string {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return nil
		}
		return &trimmed
	}

	// Create stamp model
	stamp := &models.Stamp{
		ID:           uuid.New().String(),
		Name:         name,
		ScottNumber:  parseOptionalString(r.FormValue("scott_number")),
		Series:       parseOptionalString(r.FormValue("series")),
		IssueDate:    parseOptionalString(r.FormValue("issue_date")),
		Notes:        parseOptionalString(r.FormValue("notes")),
		IsOwned:      false,
		DateAdded:    time.Now(),
		DateModified: time.Now(),
	}

	// Parse tags (can be multiple)
	tags := r.Form["tags"]
	var cleanTags []string
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed != "" {
			cleanTags = append(cleanTags, trimmed)
		}
	}
	stamp.Tags = cleanTags

	log.Printf("handlers.htmx.CreateStamp: Creating stamp: %+v", stamp)

	// Create stamp via service
	_, err = h.stampService.CreateStamp(stamp)
	if err != nil {
		LogAndReturnError(w, err, "CreateStamp", http.StatusInternalServerError)
		return
	}

	log.Printf("handlers.htmx.CreateStamp: Stamp created successfully with ID: %s", stamp.ID)

	// Fetch the complete stamp with all details
	createdStamp, err := h.stampService.GetStampByID(stamp.ID)
	if err != nil {
		log.Printf("handlers.htmx.CreateStamp: Error fetching created stamp: %v", err)
		http.Error(w, "Failed to fetch created stamp", http.StatusInternalServerError)
		return
	}

	// Get all boxes for the template
	allBoxes, err := h.boxService.GetBoxes()
	if err != nil {
		log.Printf("handlers.htmx.CreateStamp: Warning: could not fetch boxes: %v", err)
		allBoxes = []models.StorageBox{} // Empty slice as fallback
	}

	// Create the view data
	data := models.StampDetailView{
		Stamp:    *createdStamp,
		AllBoxes: allBoxes,
	}

	// Return the stamp detail HTML
	w.Header().Set("Content-Type", "text/html")
	err = h.templates.ExecuteTemplate(w, "stamp-detail.html", data)
	if err != nil {
		log.Printf("handlers.htmx.CreateStamp: Template error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
}

// UploadStampImage handles image upload for stamps
func (h *HTMXHandler) UploadStampImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stampID := vars["id"]

	// Validate stampID is a valid UUID to prevent path traversal
	if _, err := uuid.Parse(stampID); err != nil {
		http.Error(w, "Invalid stamp ID", http.StatusBadRequest)
		return
	}

	// Parse multipart form with 5MB limit
	err := r.ParseMultipartForm(5 << 20) // 5MB
	if err != nil {
		http.Error(w, "File too large. Maximum size is 5MB.", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Error closing file: %v", err)
		}
	}()

	// Validate file size
	if handler.Size > 5<<20 {
		http.Error(w, "File too large. Maximum size is 5MB.", http.StatusBadRequest)
		return
	}

	// Validate file type by reading the first 512 bytes
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	// Reset file pointer
	if _, err := file.Seek(0, 0); err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	// Check if it's an image
	contentType := http.DetectContentType(buffer)
	if !strings.HasPrefix(contentType, "image/") {
		http.Error(w, "File must be an image", http.StatusBadRequest)
		return
	}

	// Create stamps directory if it doesn't exist
	imagesDir := "./static/images/stamps"
	err = os.MkdirAll(imagesDir, 0755)
	if err != nil {
		http.Error(w, "Error creating directory", http.StatusInternalServerError)
		return
	}

	// Get existing stamp to check for current image
	existingStamp, err := h.stampService.GetStampByID(stampID)
	if err != nil {
		http.Error(w, "Stamp not found", http.StatusNotFound)
		return
	}

	// Backup existing image if it exists
	if existingStamp.ImageURL != nil && *existingStamp.ImageURL != "" {
		// Extract filename from the current image URL
		currentImageURL := *existingStamp.ImageURL
		if strings.HasPrefix(currentImageURL, "/static/images/stamps/") {
			currentFilename := strings.TrimPrefix(currentImageURL, "/static/images/stamps/")
			currentFilepath := filepath.Join(imagesDir, currentFilename)
			backupFilepath := currentFilepath + ".bak"

			// Atomic rename operation - no race condition between check and rename
			// If the file doesn't exist, rename will fail which is fine
			err = os.Rename(currentFilepath, backupFilepath)
			if err != nil {
				log.Printf("handlers.htmx.UploadStampImage: Warning: Could not backup existing image: %v", err)
				// Continue anyway - don't fail the upload for backup issues
			} else {
				log.Printf("handlers.htmx.UploadStampImage: Backed up existing image to: %s", backupFilepath)
			}
		}
	}

	// Generate unique filename
	ext := filepath.Ext(handler.Filename)
	if ext == "" {
		// Determine extension from content type
		switch contentType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/gif":
			ext = ".gif"
		case "image/webp":
			ext = ".webp"
		default:
			ext = ".jpg"
		}
	}

	filename := fmt.Sprintf("%s%s", stampID, ext)
	filepath := filepath.Join(imagesDir, filename)

	// Create the destination file
	log.Printf("handlers.htmx.UploadStampImage: Uploading file to: %v", filepath)
	dst, err := os.Create(filepath)
	if err != nil {
		http.Error(w, "Error creating file", http.StatusInternalServerError)
		return
	}

	// Cleanup flag - only keep file if everything succeeds
	cleanup := true
	defer func() {
		if err := dst.Close(); err != nil {
			log.Printf("handlers.htmx.UploadStampImage: Error closing destination file: %v", err)
		}
		if cleanup {
			if err := os.Remove(filepath); err != nil {
				log.Printf("handlers.htmx.UploadStampImage: Error removing partial upload: %v", err)
			}
		}
	}()

	// Copy the uploaded file to destination
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	// Success - don't delete the file
	cleanup = false
	log.Print("handlers.htmx.UploadStampImage: File uploaded successfully")

	// Update the stamp record with the new image URL
	imageURL := fmt.Sprintf("/static/images/stamps/%s", filename)
	existingStamp.ImageURL = &imageURL
	existingStamp.DateModified = time.Now()

	_, err = h.stampService.UpdateStamp(existingStamp)
	if err != nil {
		http.Error(w, "Error updating stamp", http.StatusInternalServerError)
		return
	}
	log.Printf("handlers.htmx.UploadStampImage: ImageURL for stamp_id %v updated to point to the new file", stampID)

	// Return the new image URL as JSON
	response := map[string]string{"image_url": imageURL}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("handlers.htmx.UploadStampImage: Error encoding JSON response: %v", err)
	}
}

// CreateStampInstance creates a new instance and returns the updated instances section
func (h *HTMXHandler) CreateStampInstance(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	stampID := vars["stampId"]

	log.Printf("handlers.htmx.CreateStampInstance: Called for stampID=%q", stampID)

	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Get form values
	condition := strings.TrimSpace(r.FormValue("condition"))
	boxName := strings.TrimSpace(r.FormValue("box_name"))
	quantityStr := strings.TrimSpace(r.FormValue("quantity"))

	log.Printf("handlers.htmx.CreateStampInstance: Form values - condition=%q, boxName=%q, quantity=%v", condition, boxName, quantityStr)

	// Validate quantity
	var quantity int
	if quantityStr == "" || quantityStr == "0" {
		http.Error(w, "Quantity must be at least 1", http.StatusBadRequest)
		return
	}
	_, err = fmt.Sscanf(quantityStr, "%d", &quantity)
	if err != nil || quantity < 1 {
		http.Error(w, "Invalid quantity", http.StatusBadRequest)
		return
	}

	// Handle box - either find existing or create new
	boxID, err := h.getOrCreateBox(boxName)
	if err != nil {
		LogAndReturnError(w, err, "CreateStampInstance-BoxLookup", http.StatusInternalServerError)
		return
	}

	// Parse condition as optional
	var conditionPtr *string
	if condition != "" {
		conditionPtr = &condition
	}

	// Create instance
	instance := &models.StampInstance{
		ID:           uuid.New().String(),
		StampID:      stampID,
		Condition:    conditionPtr,
		BoxID:        boxID,
		Quantity:     quantity,
		DateAdded:    time.Now(),
		DateModified: time.Now(),
	}

	_, err = h.instanceService.CreateStampInstance(instance)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") || strings.Contains(err.Error(), "duplicate key") {
			http.Error(w, "An instance with this condition and box already exists", http.StatusConflict)
			return
		}
		log.Printf("handlers.htmx.CreateStampInstance: Error creating instance: %v", err)
		http.Error(w, "Failed to create instance", http.StatusInternalServerError)
		return
	}

	conditionStr := "none"
	if instance.Condition != nil {
		conditionStr = *instance.Condition
	}
	boxIDStr := "none"
	if instance.BoxID != nil {
		boxIDStr = *instance.BoxID
	}
	log.Printf("handlers.htmx.CreateStampInstance: Successfully created instance %q for stamp %q (condition=%q, boxID=%q, quantity=%d)",
		instance.ID, stampID, conditionStr, boxIDStr, instance.Quantity)

	// Reload the stamp to get updated instances
	stamp, err := h.stampService.GetStampByID(stampID)
	if err != nil {
		http.Error(w, "Failed to reload stamp", http.StatusInternalServerError)
		return
	}

	// Get all boxes for the template
	allBoxes, err := h.boxService.GetBoxes()
	if err != nil {
		http.Error(w, "Failed to fetch boxes", http.StatusInternalServerError)
		return
	}

	// Return the updated your-copies-section HTML
	data := models.StampDetailView{
		Stamp:    *stamp,
		AllBoxes: allBoxes,
	}

	w.Header().Set("Content-Type", "text/html")
	err = h.templates.ExecuteTemplate(w, "your-copies-section", data)
	if err != nil {
		LogAndReturnError(w, err, "CreateStampInstance-Template", http.StatusInternalServerError)
		return
	}
}

// UpdateInstanceField updates a single field of an instance
func (h *HTMXHandler) UpdateInstanceField(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	instanceID := vars["instanceId"]
	field := vars["field"]

	log.Printf("handlers.htmx.UpdateInstanceField called: instanceID=%q, field=%q", instanceID, field)

	// Get the current instance
	instance, err := h.instanceService.GetStampInstance(instanceID)
	if err != nil {
		http.Error(w, "Instance not found", http.StatusNotFound)
		return
	}

	// Parse and update the field
	switch field {
	case "condition":
		condition := strings.TrimSpace(r.FormValue("condition"))
		log.Printf("handlers.htmx.UpdateInstanceField: condition value=%q", condition)
		if condition == "" {
			instance.Condition = nil
		} else {
			instance.Condition = &condition
		}

	case "quantity":
		quantityStr := strings.TrimSpace(r.FormValue("quantity"))
		log.Printf("handlers.htmx.UpdateInstanceField: quantity value=%q", quantityStr)
		var quantity int
		_, err := fmt.Sscanf(quantityStr, "%d", &quantity)
		if err != nil || quantity < 0 {
			http.Error(w, "Invalid quantity", http.StatusBadRequest)
			return
		}
		instance.Quantity = quantity

	case "box_id":
		boxName := strings.TrimSpace(r.FormValue("box_name"))
		log.Printf("handlers.htmx.UpdateInstanceField: box_name value=%q", boxName)
		boxID, err := h.getOrCreateBox(boxName)
		if err != nil {
			LogAndReturnError(w, err, "UpdateInstanceField-BoxLookup", http.StatusInternalServerError)
			return
		}
		instance.BoxID = boxID

	default:
		http.Error(w, "Invalid field", http.StatusBadRequest)
		return
	}

	// Update timestamp
	instance.DateModified = time.Now()

	// Check if quantity is 0 - if so, delete the instance
	if instance.Quantity == 0 {
		err = h.instanceService.DeleteStampInstance(instanceID)
		if err != nil {
			http.Error(w, "Failed to delete instance", http.StatusInternalServerError)
			return
		}
		// Return 204 to signal deletion - client will remove the row
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Save the updated instance
	_, err = h.instanceService.UpdateStampInstance(instance)
	if err != nil {
		http.Error(w, "Failed to update instance", http.StatusInternalServerError)
		return
	}

	// Get all boxes for the template
	allBoxes, err := h.boxService.GetBoxes()
	if err != nil {
		http.Error(w, "Failed to fetch boxes", http.StatusInternalServerError)
		return
	}

	// Reload the instance to get the box name
	instance, err = h.instanceService.GetStampInstance(instanceID)
	if err != nil {
		http.Error(w, "Failed to reload instance", http.StatusInternalServerError)
		return
	}

	// Return the updated row HTML
	data := struct {
		Instance *models.StampInstance
		AllBoxes []models.StorageBox
	}{
		Instance: instance,
		AllBoxes: allBoxes,
	}

	w.Header().Set("Content-Type", "text/html")
	err = h.templates.ExecuteTemplate(w, "instance-row", data)
	if err != nil {
		LogAndReturnError(w, err, "UpdateInstanceField-Template", http.StatusInternalServerError)
		return
	}
}

// DeleteStampInstance deletes an instance
func (h *HTMXHandler) DeleteStampInstance(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	instanceID := vars["instanceId"]

	log.Printf("handlers.htmx.DeleteStampInstance: %v", instanceID)

	err := h.instanceService.DeleteStampInstance(instanceID)
	if err != nil {
		http.Error(w, "Failed to delete instance", http.StatusInternalServerError)
		return
	}

	// Return 204 to signal successful deletion - client will remove the row
	w.WriteHeader(http.StatusNoContent)
}

// AdjustInstanceQuantity adjusts the quantity by a delta (+1 or -1)
func (h *HTMXHandler) AdjustInstanceQuantity(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	instanceID := vars["instanceId"]
	deltaStr := r.FormValue("delta")

	var delta int
	_, err := fmt.Sscanf(deltaStr, "%d", &delta)
	if err != nil {
		http.Error(w, "Invalid delta", http.StatusBadRequest)
		return
	}

	// Get current instance
	instance, err := h.instanceService.GetStampInstance(instanceID)
	if err != nil {
		http.Error(w, "Instance not found", http.StatusNotFound)
		return
	}

	// Calculate new quantity (can't go below 0)
	newQuantity := max(instance.Quantity+delta, 0)

	// If quantity is 0, delete the instance
	if newQuantity == 0 {
		err = h.instanceService.DeleteStampInstance(instanceID)
		if err != nil {
			http.Error(w, "Failed to delete instance", http.StatusInternalServerError)
			return
		}
		// Return 204 to signal deletion
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Update quantity
	instance.Quantity = newQuantity
	instance.DateModified = time.Now()

	_, err = h.instanceService.UpdateStampInstance(instance)
	if err != nil {
		http.Error(w, "Failed to update instance", http.StatusInternalServerError)
		return
	}

	// Get all boxes for the template
	allBoxes, err := h.boxService.GetBoxes()
	if err != nil {
		http.Error(w, "Failed to fetch boxes", http.StatusInternalServerError)
		return
	}

	// Reload the instance to get the box name
	instance, err = h.instanceService.GetStampInstance(instanceID)
	if err != nil {
		http.Error(w, "Failed to reload instance", http.StatusInternalServerError)
		return
	}

	// Return the updated row HTML
	data := struct {
		Instance *models.StampInstance
		AllBoxes []models.StorageBox
	}{
		Instance: instance,
		AllBoxes: allBoxes,
	}

	w.Header().Set("Content-Type", "text/html")
	err = h.templates.ExecuteTemplate(w, "instance-row", data)
	if err != nil {
		LogAndReturnError(w, err, "AdjustInstanceQuantity-Template", http.StatusInternalServerError)
		return
	}
}

// GetTagInputRow returns a new tag input field HTML fragment
func (h *HTMXHandler) GetTagInputRow(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	_, err := w.Write([]byte(`<div class="tag-input-row mb-2">
		<input type="text" name="tags" class="form-control form-control-sm" placeholder="Enter a tag (optional)">
	</div>`))
	if err != nil {
		log.Printf("handlers.htmx.GetTagInputRow: Error writing tag input row response: %v", err)
	}
}

// getOrCreateBox looks up a box by name or creates it if it doesn't exist
// Returns the box ID or nil if no box name provided
func (h *HTMXHandler) getOrCreateBox(boxName string) (*string, error) {
	if boxName == "" {
		return nil, nil
	}

	// Look up existing box (case-insensitive)
	// Use limit of 1000 for box lookups - reasonable cap for dropdown searches
	boxes, err := h.boxService.GetBoxes(1000)
	if err != nil {
		return nil, fmt.Errorf("handlers.htmx.getOrCreateBox: failed to lookup boxes: %w", err)
	}

	for _, box := range boxes {
		if strings.EqualFold(box.Name, boxName) {
			return &box.ID, nil
		}
	}

	// Create new box
	newBox := &models.StorageBox{
		ID:          uuid.New().String(),
		Name:        boxName,
		DateCreated: time.Now(),
	}

	createdBox, err := h.boxService.CreateBox(newBox)
	if err != nil {
		return nil, fmt.Errorf("failed to create box: %w", err)
	}

	return &createdBox.ID, nil
}
