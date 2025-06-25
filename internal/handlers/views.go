package handlers

import (
	"database/sql"
	"html/template"
	"math"
	"net/http"
	"fmt"
	"log"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jeepinbird/stampkeeper/internal/models"
	"github.com/jeepinbird/stampkeeper/internal/services"
	"github.com/jeepinbird/stampkeeper/internal/middleware"
)

type ViewHandler struct {
	db                *sql.DB
	templates         *template.Template
	stampService      *services.StampService
	boxService        *services.BoxService
	sessionMiddleware *middleware.SessionMiddleware
}

func NewViewHandler(db *sql.DB, templates *template.Template, sessionMiddleware *middleware.SessionMiddleware) *ViewHandler {
	return &ViewHandler{
		db:                db,
		templates:         templates,
		stampService:      services.NewStampService(db),
		boxService:        services.NewBoxService(db),
		sessionMiddleware: sessionMiddleware,
	}
}

func (h *ViewHandler) GetStampsView(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	view := vars["view"]

	// Get page from query, default to 1
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	prefs := h.sessionMiddleware.GetPreferences(r)
	// Items per page
	limit := prefs.ItemsPerPage
	if limit <= 0 {
		limit = 50
	}

	// Get total items and stamps for the current page
	totalItems, stamps, err := h.stampService.GetStampsWithCount(r, page, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate pagination data
	totalPages := int(math.Ceil(float64(totalItems) / float64(limit)))
	pagination := models.Pagination{
		CurrentPage: page,
		TotalPages:  totalPages,
		TotalItems:  totalItems,
		HasNext:     page < totalPages,
		HasPrev:     page > 1,
		NextPage:    page + 1,
		PrevPage:    page - 1,
	}

	// Build a BaseURL that points to the new /scroll endpoint for subsequent requests
	query := r.URL.Query()
	query.Del("page")
	// The BaseURL must point to the /scroll endpoint
	baseURLWithParams := fmt.Sprintf("/views/stamps/%s/scroll?%s", view, query.Encode())

	// Get box details if filtering by box
	var filteredBox *models.StorageBox
	boxID := r.URL.Query().Get("box_id")
	if boxID != "" {
		box, err := h.boxService.GetBoxByID(boxID)
		if err == nil {
			filteredBox = box
		}
	}

	// Prepare the full data payload for the template
	data := models.PaginatedStampsView{
		Stamps:      stamps,
		Pagination:  pagination,
		BaseURL:     baseURLWithParams, // e.g., /views/stamps/gallery
		CurrentView: view,
		FilteredBox: filteredBox,
	}

	templateName := view + "-view.html"
	err = h.templates.ExecuteTemplate(w, templateName, data)
	if err != nil {
		fmt.Printf("Template execution error: %v", err)
		return
	}
}

// Add this new handler function to your ViewHandler
func (h *ViewHandler) GetStampsScroll(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	view := vars["view"] // "gallery" or "list"

	// --- This logic is the same as in GetStampsView ---
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	prefs := h.sessionMiddleware.GetPreferences(r)
	limit := prefs.ItemsPerPage
	if limit <= 0 {
		limit = 50
	}

	totalItems, stamps, err := h.stampService.GetStampsWithCount(r, page, limit)
	if err != nil {
		w.Write([]byte(""))
		return
	}

	totalPages := int(math.Ceil(float64(totalItems) / float64(limit)))
	pagination := models.Pagination{
		CurrentPage: page,
		TotalPages:  totalPages,
		HasNext:     page < totalPages,
		NextPage:    page + 1,
		// Other fields are not strictly necessary for the partial
	}

	// Build the BaseURL for the *next* scroll request
	query := r.URL.Query()
	query.Del("page")
	// IMPORTANT: The BaseURL must point to the /scroll endpoint for subsequent loads
	baseURLWithParams := fmt.Sprintf("/views/stamps/%s/scroll?%s", view, query.Encode())

	data := models.PaginatedStampsView{
		Stamps:     stamps,
		Pagination: pagination,
		BaseURL:    baseURLWithParams,
	}
	// --- End of repeated logic ---

	// Determine which partial to render
	var templateName string
	if view == "gallery" {
		templateName = "_gallery-page.html"
	} else {
		templateName = "_list-rows.html"
	}

	err = h.templates.ExecuteTemplate(w, templateName, data)
	if err != nil {
		log.Printf("Template execution error for scroll: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func (h *ViewHandler) GetStampDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	stamp, err := h.stampService.GetStampByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Stamp not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Get all boxes for the dropdown
	allBoxes, err := h.boxService.GetBoxes()
	if err != nil {
		// Log the error but don't fail the whole request
		log.Printf("Warning: could not fetch boxes for dropdown: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create the view data
	data := models.StampDetailView{
		Stamp:    *stamp,
		AllBoxes: allBoxes,
	}

	err = h.templates.ExecuteTemplate(w, "stamp-detail.html", data)
	if err != nil {
		fmt.Printf("Template execution error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
}

func (h *ViewHandler) GetBoxesView(w http.ResponseWriter, r *http.Request) {
	boxes, err := h.boxService.GetBoxes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get user preferences to pass to template
	prefs := h.sessionMiddleware.GetPreferences(r)

	// Get the currently active box ID from query parameters
	activeBoxID := r.URL.Query().Get("box_id")

	// Create data structure that includes both boxes and preferences
	data := struct {
		Boxes       interface{}
		Preferences middleware.UserPreferences
		ActiveBoxID string
	}{
		Boxes:       boxes,
		Preferences: prefs,
		ActiveBoxID: activeBoxID,
	}

	err = h.templates.ExecuteTemplate(w, "box-list.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *ViewHandler) GetNewInstanceRow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stampID := vars["id"]

	allBoxes, err := h.boxService.GetBoxes()
	if err != nil {
		log.Printf("Warning: could not fetch boxes for new instance row: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := models.StampDetailView{
		Stamp:    models.Stamp{ID: stampID},
		AllBoxes: allBoxes,
	}

	err = h.templates.ExecuteTemplate(w, "new-instance-row.html", data)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
}

func (h *ViewHandler) GetNewStampForm(w http.ResponseWriter, r *http.Request) {
	// Get all boxes for potential future use
	allBoxes, err := h.boxService.GetBoxes()
	if err != nil {
		// Log the error but don't fail the whole request
		log.Printf("Warning: could not fetch boxes for new stamp form: %v", err)
		allBoxes = []models.StorageBox{} // Empty slice as fallback
	}

	// Create minimal data for the template
	data := models.StampDetailView{
		Stamp:    models.Stamp{ID: "new", Name: "New Stamp"},
		AllBoxes: allBoxes,
	}

	err = h.templates.ExecuteTemplate(w, "new-stamp-form.html", data)
	if err != nil {
		fmt.Printf("Template execution error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
}

func (h *ViewHandler) GetSettingsView(w http.ResponseWriter, r *http.Request) {
	// Get all boxes for the storage box management section
	allBoxes, err := h.boxService.GetBoxes()
	if err != nil {
		log.Printf("Warning: could not fetch boxes for settings page: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get fresh user preferences directly from cookie to ensure we have the latest values
	prefs := h.sessionMiddleware.GetPreferences(r)
	
	// Debug logging to see what preferences are actually retrieved
	log.Printf("handlers.views.GetSettingsView: %+v", prefs)

	// Create the view data
	data := models.SettingsView{
		AllBoxes: allBoxes,
		Preferences: models.UserPreferences{
			DefaultView:   prefs.DefaultView,
			DefaultSort:   prefs.DefaultSort,
			SortDirection: prefs.SortDirection,
			ItemsPerPage:  prefs.ItemsPerPage,
		},
	}

	err = h.templates.ExecuteTemplate(w, "settings.html", data)
	if err != nil {
		fmt.Printf("Template execution error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
}

func (h *ViewHandler) GetIndexView(w http.ResponseWriter, r *http.Request) {
	// Get fresh user preferences directly from cookie
	prefs := h.sessionMiddleware.GetPreferences(r)
	
	// Debug logging to see what preferences are retrieved for index
	log.Printf("handlers.views.GetIndexView: %+v", prefs)

	// Create the view data with preferences
	data := struct {
		Preferences middleware.UserPreferences
	}{
		Preferences: prefs,
	}

	err := h.templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		fmt.Printf("Template execution error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
}