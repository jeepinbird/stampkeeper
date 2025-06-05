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
)

type ViewHandler struct {
	db           *sql.DB
	templates    *template.Template
	stampService *services.StampService
	boxService   *services.BoxService
	tagService   *services.TagService
}

func NewViewHandler(db *sql.DB, templates *template.Template) *ViewHandler {
	return &ViewHandler{
		db:           db,
		templates:    templates,
		stampService: services.NewStampService(db),
		boxService:   services.NewBoxService(db),
		tagService:   services.NewTagService(db),
	}
}

func (h *ViewHandler) GetStampsView(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	view := vars["view"]

	// Debug: Log all query parameters
	// log.Printf("=== GetStampsView Debug ===")
	// log.Printf("View: %s", view)
	// log.Printf("All Query Parameters: %v", r.URL.Query())
	// log.Printf("Sort parameter: '%s'", r.URL.Query().Get("sort"))
	// log.Printf("Order parameter: '%s'", r.URL.Query().Get("order"))
	// log.Printf("Search parameter: '%s'", r.URL.Query().Get("search"))
	// log.Printf("Owned parameter: '%s'", r.URL.Query().Get("owned"))
	// log.Printf("Box ID parameter: '%s'", r.URL.Query().Get("box_id"))
	// log.Printf("Box Filter parameter: '%s'", r.URL.Query().Get("box_filter"))
	// log.Printf("========================")

	// Get page from query, default to 1
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit := 50 // Items per page

	// Get total items for pagination
	totalItems, err := h.stampService.GetStampCount(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get stamps for the current page
	stamps, err := h.stampService.GetStamps(r, page, limit)
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

	// Prepare the full data payload for the template
	data := models.PaginatedStampsView{
		Stamps:      stamps,
		Pagination:  pagination,
		BaseURL:     r.URL.Path, // e.g., /views/stamps/gallery
		CurrentView: view,
	}

	templateName := view + "-view.html"
	err = h.templates.ExecuteTemplate(w, templateName, data)
	if err != nil {
		fmt.Printf("Template execution error: %v", err)
		return
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

	// Get total stamp count
	totalCount, err := h.stampService.GetTotalStampCount()
	if err != nil {
		log.Printf("Warning: could not get total stamp count: %v", err)
		totalCount = 0
	}

	// Get unassigned stamp count (stamps with no instances in any box)
	unassignedCount, err := h.stampService.GetUnassignedStampCount()
	if err != nil {
		log.Printf("Warning: could not get unassigned stamp count: %v", err)
		unassignedCount = 0
	}

	// Create view data with counts
	data := struct {
		Boxes           []models.StorageBox
		TotalCount      int
		UnassignedCount int
	}{
		Boxes:           boxes,
		TotalCount:      totalCount,
		UnassignedCount: unassignedCount,
	}

	err = h.templates.ExecuteTemplate(w, "box-list.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *ViewHandler) GetPopularTagsView(w http.ResponseWriter, r *http.Request) {
	// Get top 10 most used tags
	tags, err := h.tagService.GetPopularTags(10)
	if err != nil {
		log.Printf("Warning: could not fetch popular tags: %v", err)
		// Return empty template instead of failing
		tags = []models.Tag{}
	}

	data := struct {
		Tags []models.Tag
	}{
		Tags: tags,
	}

	err = h.templates.ExecuteTemplate(w, "popular-tags.html", data)
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

	// Create the view data - reusing the StampDetailView structure since it has AllBoxes
	data := models.StampDetailView{
		Stamp:    models.Stamp{}, // Empty stamp, not used in settings
		AllBoxes: allBoxes,
	}

	err = h.templates.ExecuteTemplate(w, "settings.html", data)
	if err != nil {
		fmt.Printf("Template execution error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
}