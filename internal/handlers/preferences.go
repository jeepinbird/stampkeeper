package handlers

import (
	"database/sql"
	"encoding/json"
	"html/template"
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

// GetDefaultView returns the user's preferred default view with proper navigation
func (h *PreferencesHandler) GetDefaultView(w http.ResponseWriter, r *http.Request) {
	prefs := h.sessionMiddleware.GetPreferences(r)
	
	// Redirect to the user's preferred view with current filters
	viewPath := "/views/stamps/" + prefs.DefaultView
	
	// Preserve any existing query parameters
	if r.URL.RawQuery != "" {
		viewPath += "?" + r.URL.RawQuery
	}
	
	// Use HTMX redirect header if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", viewPath)
		w.WriteHeader(http.StatusOK)
		return
	}
	
	// Regular redirect for non-HTMX requests
	http.Redirect(w, r, viewPath, http.StatusSeeOther)
}