package router

import (
    "database/sql"
    "html/template"
    "net/http"
    
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
    var templates *template.Template // For HTML templates
    
    templates = template.New("").Funcs(template.FuncMap{
		"substr": substr,
	})
	templates = template.Must(templates.ParseGlob("templates/*.html"))
    
    // Initialize handlers with dependencies
    stampHandler := handlers.NewStampHandler(db, templates)
    boxHandler := handlers.NewBoxHandler(db, templates)
    tagHandler := handlers.NewTagHandler(db, templates)
	statsHandler := handlers.NewStatsHandler(db, templates)
	viewHandler := handlers.NewViewHandler(db, templates)
    
    // Create main router
    r := mux.NewRouter()
    
    // JSON API routes
    api := r.PathPrefix("/api").Subrouter()

    // Stamps endpoints
    api.HandleFunc("/stamps", stampHandler.GetStamps).Methods("GET")
	api.HandleFunc("/stamps", stampHandler.CreateStamp).Methods("POST")
	api.HandleFunc("/stamps/{id}", stampHandler.GetStamp).Methods("GET")
	api.HandleFunc("/stamps/{id}", stampHandler.UpdateStamp).Methods("PUT")
	api.HandleFunc("/stamps/{id}", stampHandler.DeleteStamp).Methods("DELETE")

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

// loggingMiddleware logs each request
// func loggingMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// You could use a proper logger here like logrus or zap
// 		// For now, just basic logging
// 		next.ServeHTTP(w, r)
// 	})
// }

// corsMiddleware adds CORS headers (useful for development)
// func corsMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.Header().Set("Access-Control-Allow-Origin", "*")
// 		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
// 		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

// 		// Handle preflight requests
// 		if r.Method == "OPTIONS" {
// 			w.WriteHeader(http.StatusOK)
// 			return
// 		}

// 		next.ServeHTTP(w, r)
// 	})
// }