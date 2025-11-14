package handlers

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"github.com/jeepinbird/stampkeeper/internal/middleware"
)

type PreferencesHandler struct {
	db                *sql.DB
	templates         *template.Template
	sessionMiddleware *middleware.SessionMiddleware
}

func NewPreferencesHandler(db *sql.DB, templates *template.Template, sessionMiddleware *middleware.SessionMiddleware) *PreferencesHandler {
	return &PreferencesHandler{
		db:                db,
		templates:         templates,
		sessionMiddleware: sessionMiddleware,
	}
}

// GetPreferences returns user preferences as JSON
func (h *PreferencesHandler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	prefs := h.sessionMiddleware.GetPreferences(r)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(prefs); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
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
	if _, err := w.Write([]byte(`<div class="alert alert-success" role="alert">
		<i class="bi bi-check-circle"></i> Preferences saved successfully!
	</div>`)); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

// GetDefaultView redirects to the user's preferred default view
// HTMX will follow the redirect and swap the final HTML response
func (h *PreferencesHandler) GetDefaultView(w http.ResponseWriter, r *http.Request) {
	prefs := h.sessionMiddleware.GetPreferences(r)

	// Build redirect URL with user preferences as query parameters
	query := r.URL.Query()
	query.Set("sort", prefs.DefaultSort)
	query.Set("order", prefs.SortDirection)

	redirectURL := "/views/stamps/" + prefs.DefaultView + "?" + query.Encode()
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}
