package handlers

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"github.com/jeepinbird/stampkeeper/internal/middleware"
	"github.com/jeepinbird/stampkeeper/internal/services"
)

type PreferencesHandler struct {
	db                *sql.DB
	templates         *template.Template
	sessionMiddleware *middleware.SessionMiddleware
	stampService      *services.StampService
}

func NewPreferencesHandler(db *sql.DB, templates *template.Template, sessionMiddleware *middleware.SessionMiddleware) *PreferencesHandler {
	return &PreferencesHandler{
		db:                db,
		templates:         templates,
		sessionMiddleware: sessionMiddleware,
		stampService:      services.NewStampService(db),
	}
}

// GetPreferences returns user preferences as JSON
func (h *PreferencesHandler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	prefs := h.sessionMiddleware.GetPreferences(r)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prefs)
}

// SavePreferences saves user preferences and returns success message
func (h *PreferencesHandler) SavePreferences(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse preferences from request
	prefs := h.sessionMiddleware.UpdatePreferencesFromRequest(r)
	
	// Debug logging to see what preferences are being saved
	log.Printf("handlers.preferences.SavePreferences: %+v", prefs)
	
	// Save to cookie
	err := h.sessionMiddleware.SavePreferences(w, prefs)
	if err != nil {
		http.Error(w, "Failed to save preferences", http.StatusInternalServerError)
		return
	}

	// Return success response (for HTMX)
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<div class="alert alert-success" role="alert">
		<i class="bi bi-check-circle"></i> Preferences saved successfully!
	</div>`))
}

// GetDefaultView returns the user's preferred default view content
func (h *PreferencesHandler) GetDefaultView(w http.ResponseWriter, r *http.Request) {
	prefs := h.sessionMiddleware.GetPreferences(r)
	
	// Create a new request with user preferences injected as query parameters
	// so that the StampService can use them for sorting
	newURL := *r.URL
	query := newURL.Query()
	query.Set("sort", prefs.DefaultSort)
	query.Set("order", prefs.SortDirection)
	newURL.RawQuery = query.Encode()
	
	// Create new request with preference-enhanced URL
	newReq := r.Clone(r.Context())
	newReq.URL = &newURL
	
	// Get page from query, default to 1
	page := 1
	limit := prefs.ItemsPerPage
	
	// Get total items and stamps for the current page using enhanced request with user preferences
	totalItems, stamps, err := h.stampService.GetStampsWithCount(newReq, page, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Calculate pagination data
	totalPages := int(float64(totalItems)/float64(limit)) + 1
	if totalItems%int64(limit) == 0 && totalItems > 0 {
		totalPages--
	}
	
	// Create pagination struct
	pagination := struct {
		CurrentPage int
		TotalPages  int
		TotalItems  int64
		HasNext     bool
		HasPrev     bool
		NextPage    int
		PrevPage    int
	}{
		CurrentPage: page,
		TotalPages:  totalPages,
		TotalItems:  totalItems,
		HasNext:     page < totalPages,
		HasPrev:     page > 1,
		NextPage:    page + 1,
		PrevPage:    page - 1,
	}
	
	// Build BaseURL that points to the scroll endpoint for subsequent requests
	scrollQuery := newReq.URL.Query()
	scrollQuery.Del("page")
	baseURLWithParams := "/views/stamps/" + prefs.DefaultView + "/scroll?" + scrollQuery.Encode()
	
	// Prepare the data for the template
	data := struct {
		Stamps      interface{}
		Pagination  interface{}
		BaseURL     string
		CurrentView string
	}{
		Stamps:      stamps,
		Pagination:  pagination,
		BaseURL:     baseURLWithParams,
		CurrentView: prefs.DefaultView,
	}
	
	// Return the appropriate view template
	templateName := prefs.DefaultView + "-view.html"
	w.Header().Set("Content-Type", "text/html")
	err = h.templates.ExecuteTemplate(w, templateName, data)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
}