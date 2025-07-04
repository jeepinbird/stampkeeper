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
	stampHandler := handlers.NewStampHandler(db, templates)
	instanceHandler := handlers.NewInstanceHandler(db, templates)
	boxHandler := handlers.NewBoxHandler(db, templates)
	tagHandler := handlers.NewTagHandler(db, templates)
	statsHandler := handlers.NewStatsHandler(db, templates)
	viewHandler := handlers.NewViewHandler(db, templates, sessionMiddleware)
	preferencesHandler := handlers.NewPreferencesHandler(db, templates, sessionMiddleware)
	htmxHandler := handlers.NewHTMXHandler(db, templates)
	
	// Create main router
	r := mux.NewRouter()
	
	// JSON API routes
	api := r.PathPrefix("/api").Subrouter()

	// Stamp design endpoints
	api.HandleFunc("/stamps", stampHandler.GetStamps).Methods("GET")
	api.HandleFunc("/stamps", stampHandler.CreateStamp).Methods("POST")
	api.HandleFunc("/stamps/{id}", stampHandler.GetStamp).Methods("GET")
	api.HandleFunc("/stamps/{id}", stampHandler.UpdateStamp).Methods("PUT")
	api.HandleFunc("/stamps/{id}", stampHandler.DeleteStamp).Methods("DELETE")
	api.HandleFunc("/stamps/{id}/upload-image", stampHandler.UploadStampImage).Methods("POST")

	// Stamp instance endpoints (moved to instanceHandler)
	api.HandleFunc("/instances/{stamp_id}", instanceHandler.CreateStampInstance).Methods("POST")
	api.HandleFunc("/instances/{instance_id}", instanceHandler.GetStampInstance).Methods("GET")
	api.HandleFunc("/instances/{instance_id}", instanceHandler.UpdateStampInstance).Methods("PUT")
	api.HandleFunc("/instances/{instance_id}", instanceHandler.DeleteStampInstance).Methods("DELETE")

	// Storage boxes endpoints
	api.HandleFunc("/boxes", boxHandler.GetBoxes).Methods("GET")
	api.HandleFunc("/boxes", boxHandler.CreateBox).Methods("POST")
	api.HandleFunc("/boxes/{id}", boxHandler.GetBox).Methods("GET")
	api.HandleFunc("/boxes/{id}", boxHandler.UpdateBox).Methods("PUT")
	api.HandleFunc("/boxes/{id}", boxHandler.DeleteBox).Methods("DELETE")

	// Tags endpoints
	api.HandleFunc("/tags", tagHandler.GetTags).Methods("GET")
	api.HandleFunc("/tags", tagHandler.CreateTag).Methods("POST")
	api.HandleFunc("/tags/{id}", tagHandler.UpdateTag).Methods("PUT")
	api.HandleFunc("/tags/{id}", tagHandler.DeleteTag).Methods("DELETE")

	// Stats endpoint
	api.HandleFunc("/stats", statsHandler.GetStats).Methods("GET")

	// User preferences endpoints
	api.HandleFunc("/preferences", preferencesHandler.GetPreferences).Methods("GET")
	api.HandleFunc("/preferences", preferencesHandler.SavePreferences).Methods("POST")

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
	r.HandleFunc("/htmx/stamps/{id}/field/{field}", htmxHandler.UpdateStampField).Methods("POST")
	r.HandleFunc("/htmx/stamps/{id}/tags", htmxHandler.AddStampTag).Methods("POST")
	r.HandleFunc("/htmx/stamps/{id}/tags/{tag}", htmxHandler.RemoveStampTag).Methods("DELETE")
	r.HandleFunc("/htmx/boxes", htmxHandler.CreateBox).Methods("POST")
	r.HandleFunc("/htmx/boxes/{id}", htmxHandler.UpdateBoxName).Methods("PUT")
	r.HandleFunc("/htmx/boxes/{id}", htmxHandler.DeleteBox).Methods("DELETE")

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