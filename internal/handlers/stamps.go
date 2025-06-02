package handlers

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"time"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

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

	// Add some logging to debug
	log.Printf("UpdateStamp called for ID: %s", id)

	// First, get the existing stamp
	existingStamp, err := h.service.GetStampByID(id)
	if err != nil {
		log.Printf("Error getting existing stamp: %v", err)
		if err == sql.ErrNoRows {
			http.Error(w, "Stamp not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, fmt.Sprintf("Error reading request body: %v", err), http.StatusBadRequest)
		return
	}
	// log.Printf("Request body: %s", string(body))

	// Parse the incoming JSON into a map to handle partial updates
	var updates map[string]interface{}
	if err := json.Unmarshal(body, &updates); err != nil {
		log.Printf("Error unmarshaling JSON: %v", err)
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}
	// log.Printf("Parsed updates: %+v", updates)

	// Apply updates to the existing stamp
	if name, ok := updates["name"].(string); ok {
		existingStamp.Name = name
		log.Printf("Updated name to: %s", name)
	}

	if scottNumber, ok := updates["scott_number"]; ok {
		if scottNumber == nil || scottNumber == "" {
			existingStamp.ScottNumber = nil
		} else if scottNumberStr, ok := scottNumber.(string); ok {
			existingStamp.ScottNumber = &scottNumberStr
			log.Printf("Updated scott_number to: %s", scottNumberStr)
		}
	}

	if issueDate, ok := updates["issue_date"]; ok {
		if issueDate == nil || issueDate == "" {
			existingStamp.IssueDate = nil
		} else if issueDateStr, ok := issueDate.(string); ok {
			existingStamp.IssueDate = &issueDateStr
			log.Printf("Updated issue_date to: %s", issueDateStr)
		}
	}

	if series, ok := updates["series"]; ok {
		if series == nil || series == "" {
			existingStamp.Series = nil
		} else if seriesStr, ok := series.(string); ok {
			existingStamp.Series = &seriesStr
			log.Printf("Updated series to: %s", seriesStr)
		}
	}

	if condition, ok := updates["condition"]; ok {
		if condition == nil || condition == "" {
			existingStamp.Condition = nil
		} else if conditionStr, ok := condition.(string); ok {
			existingStamp.Condition = &conditionStr
			log.Printf("Updated condition to: %s", conditionStr)
		}
	}

	if quantity, ok := updates["quantity"]; ok {
		if quantityFloat, ok := quantity.(float64); ok {
			existingStamp.Quantity = int(quantityFloat)
			log.Printf("Updated quantity to: %d", int(quantityFloat))
		}
	}

	if boxID, ok := updates["box_id"]; ok {
		if boxID == nil || boxID == "" {
			existingStamp.BoxID = nil
		} else if boxIDStr, ok := boxID.(string); ok {
			existingStamp.BoxID = &boxIDStr
			log.Printf("Updated box_id to: %s", boxIDStr)
		}
	}

	if notes, ok := updates["notes"]; ok {
		if notes == nil || notes == "" {
			existingStamp.Notes = nil
		} else if notesStr, ok := notes.(string); ok {
			existingStamp.Notes = &notesStr
			log.Printf("Updated notes to: %s", notesStr)
		}
	}

	if imageURL, ok := updates["image_url"]; ok {
		if imageURL == nil || imageURL == "" {
			existingStamp.ImageURL = nil
		} else if imageURLStr, ok := imageURL.(string); ok {
			existingStamp.ImageURL = &imageURLStr
			log.Printf("Updated image_url to: %s", imageURLStr)
		}
	}

	if isOwned, ok := updates["is_owned"]; ok {
		if isOwnedBool, ok := isOwned.(bool); ok {
			existingStamp.IsOwned = isOwnedBool
			log.Printf("Updated is_owned to: %t", isOwnedBool)
		}
	}

	// Handle tags array
	if tagsInterface, ok := updates["tags"]; ok {
		log.Printf("Processing tags update: %+v", tagsInterface)
		if tagsArray, ok := tagsInterface.([]interface{}); ok {
			var tags []string
			for _, tag := range tagsArray {
				if tagStr, ok := tag.(string); ok {
					tags = append(tags, tagStr)
				}
			}
			existingStamp.Tags = tags
			log.Printf("Updated tags to: %+v", tags)
		}
	}

	// Update the modified timestamp
	existingStamp.DateModified = time.Now()

	// Save the updated stamp
	// log.Printf("Saving updated stamp: %+v", existingStamp)
	updatedStamp, err := h.service.UpdateStamp(existingStamp)
	if err != nil {
		log.Printf("Error updating stamp in service: %v", err)
		http.Error(w, fmt.Sprintf("Failed to update stamp: %v", err), http.StatusInternalServerError)
		return
	}

	log.Print("Stamp updated successfully")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedStamp)
}

func (h *StampHandler) UploadStampImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stampID := vars["id"]
	
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
	defer file.Close()
	
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
	file.Seek(0, 0)
	
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
	existingStamp, err := h.service.GetStampByID(stampID)
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
			
			// Check if the current image file exists
			if _, err := os.Stat(currentFilepath); err == nil {
				// Create backup by renaming with .bak extension
				backupFilepath := currentFilepath + ".bak"
				err = os.Rename(currentFilepath, backupFilepath)
				if err != nil {
					log.Printf("Warning: Could not backup existing image: %v", err)
					// Continue anyway - don't fail the upload for backup issues
				} else {
					log.Printf("Backed up existing image to: %s", backupFilepath)
				}
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
	log.Printf("Uploading file to: %v", filepath)
	dst, err := os.Create(filepath)
	if err != nil {
		http.Error(w, "Error creating file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	
	// Copy the uploaded file to destination
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	log.Print("File uploaded successfully")
	
	// Update the stamp record with the new image URL
	imageURL := fmt.Sprintf("/static/images/stamps/%s", filename)
	existingStamp.ImageURL = &imageURL
	existingStamp.DateModified = time.Now()
	
	_, err = h.service.UpdateStamp(existingStamp)
	if err != nil {
		http.Error(w, "Error updating stamp", http.StatusInternalServerError)
		return
	}
	log.Printf("ImageURL for stamp_id %v updated to point to the new file", stampID)
	
	// Return the new image URL as JSON
	response := map[string]string{"image_url": imageURL}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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