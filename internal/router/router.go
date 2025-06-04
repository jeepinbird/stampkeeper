package router

import (
	"database/sql"
	"html/template"
	"net/http"
	"encoding/json"
	
	"github.com/gorilla/mux"
	"github.com/jeepinbird/stampkeeper/internal/handlers"
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
	
	// Initialize handlers with dependencies
	stampHandler := handlers.NewStampHandler(db, templates)
	instanceHandler := handlers.NewInstanceHandler(db, templates)
	boxHandler := handlers.NewBoxHandler(db, templates)
	tagHandler := handlers.NewTagHandler(db, templates)
	statsHandler := handlers.NewStatsHandler(db, templates)
	viewHandler := handlers.NewViewHandler(db, templates)
	
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
	api.HandleFunc("/stamps/{stamp_id}/instances", instanceHandler.CreateStampInstance).Methods("POST")
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

	// --- HTMX View Endpoints (return HTML fragments) ---
	r.HandleFunc("/views/stamps/{view:gallery|list}", viewHandler.GetStampsView).Methods("GET")
	r.HandleFunc("/views/stamps/detail/{id}", viewHandler.GetStampDetail).Methods("GET")
	r.HandleFunc("/views/boxes-list", viewHandler.GetBoxesView).Methods("GET")

	// --- Static File Server ---
	// Serves CSS, JS, images, etc. from the 'static' directory
	fs := http.FileServer(http.Dir("./static/"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	// --- Main Application Route ---
	// Serves the main index.html file on the root path
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	}).Methods("GET")

	// Optional: Add some helpful middleware
	//r.Use(loggingMiddleware)
	//r.Use(corsMiddleware)

	return r
}