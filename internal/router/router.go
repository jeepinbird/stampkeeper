package router

import (
	"database/sql"
	"html/template"
	"net/http"
	"encoding/json"
	
	"github.com/gorilla/mux"
	"github.com/jeepinbird/stampkeeper/internal/handlers"
	"github.com/jeepinbird/stampkeeper/internal/middleware"
)

func substr(s string, start, length int) string {
	if start < 0 {
		start = 0
	}
	if start > len(s) {
		return ""
	}
	if start+length > len(s) {
		length = len(s) - start
	}
	return s[start : start+length]
}

func Setup(db *sql.DB) *mux.Router {
	var templates *template.Template

	// Create custom template functions
	funcMap := template.FuncMap{
		"substr": substr,
		"deref": func(s *string) string {
			if s == nil {
				return ""
			}
			return *s
		},
		"json": func(v interface{}) string {
			bytes, err := json.Marshal(v)
			if err != nil {
				return "[]"
			}
			return string(bytes)
		},
		"eq": func(a, b interface{}) bool {
			return a == b
		},
		"add": func(a, b int) int {
			return a + b
		},
	}
	
	templates = template.New("").Funcs(funcMap)
	templates = template.Must(templates.ParseGlob("templates/*.html"))
	
	// Initialize session middleware
	sessionMiddleware := middleware.NewSessionMiddleware()

	// Initialize handlers with dependencies
	viewHandler := handlers.NewViewHandler(db, templates, sessionMiddleware)
	preferencesHandler := handlers.NewPreferencesHandler(db, templates, sessionMiddleware)
	htmxHandler := handlers.NewHTMXHandler(db, templates)

	// Create main router
	r := mux.NewRouter()

	// --- HTMX View Endpoints (return HTML fragments) ---
	r.HandleFunc("/views/stamps/{view:gallery|list}", viewHandler.GetStampsView).Methods("GET")
	r.HandleFunc("/views/stamps/{view:gallery|list}/scroll", viewHandler.GetStampsScroll).Methods("GET")
	r.HandleFunc("/views/stamps/detail/{id}", viewHandler.GetStampDetail).Methods("GET")
	r.HandleFunc("/views/boxes-list", viewHandler.GetBoxesView).Methods("GET")
	r.HandleFunc("/views/stamps/{id}/new-instance-row", viewHandler.GetNewInstanceRow).Methods("GET")
	r.HandleFunc("/views/stamps/new", viewHandler.GetNewStampForm).Methods("GET")
	r.HandleFunc("/views/settings", viewHandler.GetSettingsView).Methods("GET")
	r.HandleFunc("/views/default", preferencesHandler.GetDefaultView).Methods("GET")

	// --- HTMX-specific endpoints (return HTML fragments) ---
	r.HandleFunc("/htmx/stamps", htmxHandler.CreateStamp).Methods("POST")
	r.HandleFunc("/htmx/stamps/{id}/field/{field}", htmxHandler.UpdateStampField).Methods("POST")
	r.HandleFunc("/htmx/stamps/{id}/tags", htmxHandler.AddStampTag).Methods("POST")
	r.HandleFunc("/htmx/stamps/{id}/tags/{tag}", htmxHandler.RemoveStampTag).Methods("DELETE")
	r.HandleFunc("/htmx/stamps/{id}", htmxHandler.DeleteStamp).Methods("DELETE")
	r.HandleFunc("/htmx/boxes", htmxHandler.CreateBox).Methods("POST")
	r.HandleFunc("/htmx/boxes/{id}/edit", htmxHandler.GetBoxEditForm).Methods("GET")
	r.HandleFunc("/htmx/boxes/{id}/cancel", htmxHandler.CancelBoxEdit).Methods("GET")
	r.HandleFunc("/htmx/boxes/{id}", htmxHandler.UpdateBoxName).Methods("PUT")
	r.HandleFunc("/htmx/boxes/{id}", htmxHandler.DeleteBox).Methods("DELETE")
	r.HandleFunc("/htmx/instances/{stampId}", htmxHandler.CreateStampInstance).Methods("POST")
	r.HandleFunc("/htmx/instances/{instanceId}/field/{field}", htmxHandler.UpdateInstanceField).Methods("POST")
	r.HandleFunc("/htmx/instances/{instanceId}/quantity/adjust", htmxHandler.AdjustInstanceQuantity).Methods("POST")
	r.HandleFunc("/htmx/instances/{instanceId}", htmxHandler.DeleteStampInstance).Methods("DELETE")

	// API endpoints (kept for specific functionality)
	r.HandleFunc("/api/stamps/{id}/upload-image", htmxHandler.UploadStampImage).Methods("POST")
	r.HandleFunc("/api/preferences", preferencesHandler.GetPreferences).Methods("GET")
	r.HandleFunc("/api/preferences", preferencesHandler.SavePreferences).Methods("POST")

	// --- Static File Server ---
	// Serves CSS, JS, images, etc. from the 'static' directory
	fs := http.FileServer(http.Dir("./static/"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	// --- Main Application Route ---
	// Serves the main index.html template with user preferences
	r.HandleFunc("/", viewHandler.GetIndexView).Methods("GET")

	// Apply session middleware to all routes
	r.Use(sessionMiddleware.SessionHandler)

	return r
}