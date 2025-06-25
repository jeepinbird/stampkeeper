package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"time"
)

// UserPreferences represents user-specific application preferences
type UserPreferences struct {
	DefaultView     string `json:"defaultView"`     // "gallery" or "list"
	DefaultSort     string `json:"defaultSort"`     // "name", "date", etc.
	SortDirection   string `json:"sortDirection"`   // "ASC" or "DESC"
	ItemsPerPage    int    `json:"itemsPerPage"`    // Number of items per page
	LastUpdated     time.Time `json:"lastUpdated"`
}

// DefaultPreferences returns the default user preferences
func DefaultPreferences() UserPreferences {
	return UserPreferences{
		DefaultView:   "gallery",
		DefaultSort:   "name",
		SortDirection: "ASC",
		ItemsPerPage:  50,
		LastUpdated:   time.Now(),
	}
}

// SessionMiddleware provides session management for user preferences
type SessionMiddleware struct {
	cookieName string
	maxAge     int // in seconds
}

// NewSessionMiddleware creates a new session middleware instance
func NewSessionMiddleware() *SessionMiddleware {
	return &SessionMiddleware{
		cookieName: "stampkeeper_preferences",
		maxAge:     30 * 24 * 60 * 60, // 30 days
	}
}

// SessionHandler wraps HTTP handlers to provide session functionality
func (sm *SessionMiddleware) SessionHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add preferences to request context if available
		prefs := sm.GetPreferences(r)
		ctx := WithPreferences(r.Context(), prefs)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// GetPreferences retrieves user preferences from the cookie
func (sm *SessionMiddleware) GetPreferences(r *http.Request) UserPreferences {
	cookie, err := r.Cookie(sm.cookieName)
	if err != nil {
		// Return defaults if no cookie found
		return DefaultPreferences()
	}

	// URL-decode the cookie value
	decodedValue, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		log.Printf("middleware.sessions.GetPreferences: Failed to decode cookie value: %v", err)
		return DefaultPreferences()
	}

	var prefs UserPreferences
	err = json.Unmarshal([]byte(decodedValue), &prefs)
	if err != nil {
		log.Printf("middleware.sessions.GetPreferences: Failed to unmarshal JSON: %v, value: %s", err, decodedValue)
		// Return defaults if cookie is corrupted
		return DefaultPreferences()
	}

	// Validate preferences and use defaults for invalid values
	if prefs.DefaultView != "gallery" && prefs.DefaultView != "list" {
		prefs.DefaultView = "gallery"
	}
	if prefs.SortDirection != "ASC" && prefs.SortDirection != "DESC" {
		prefs.SortDirection = "ASC"
	}
	if prefs.ItemsPerPage <= 0 || prefs.ItemsPerPage > 200 {
		prefs.ItemsPerPage = 50
	}

	return prefs
}

// SavePreferences saves user preferences to a cookie
func (sm *SessionMiddleware) SavePreferences(w http.ResponseWriter, prefs UserPreferences) error {
	prefs.LastUpdated = time.Now()

	data, err := json.Marshal(prefs)
	if err != nil {
		return err
	}

	// URL-encode the JSON data to handle special characters in cookie values
	encodedData := url.QueryEscape(string(data))
	
	// Debug logging to see what's being saved
	log.Printf("middleware.sessions.SavePreferences: JSON data: %s", string(data))
	log.Printf("middleware.sessions.SavePreferences: Encoded cookie value: %s", encodedData)

	cookie := &http.Cookie{
		Name:     sm.cookieName,
		Value:    encodedData,
		Path:     "/",
		MaxAge:   sm.maxAge,
		HttpOnly: false, // Allow JavaScript access for client-side reading
		SameSite: http.SameSiteStrictMode,
		Secure:   false, // Set to true in production with HTTPS
	}

	http.SetCookie(w, cookie)
	return nil
}

// UpdatePreferencesFromRequest updates preferences based on request parameters
func (sm *SessionMiddleware) UpdatePreferencesFromRequest(r *http.Request) UserPreferences {
	current := sm.GetPreferences(r)

	// Update from form values if present
	if view := r.FormValue("defaultView"); view != "" {
		if view == "gallery" || view == "list" {
			current.DefaultView = view
		}
	}

	if sort := r.FormValue("defaultSort"); sort != "" {
		current.DefaultSort = sort
	}

	if direction := r.FormValue("sortDirection"); direction != "" {
		if direction == "ASC" || direction == "DESC" {
			current.SortDirection = direction
		}
	}

	if itemsStr := r.FormValue("itemsPerPage"); itemsStr != "" {
		if items := parseIntSafe(itemsStr, 50); items > 0 && items <= 200 {
			current.ItemsPerPage = items
		}
	}

	return current
}

// Helper function to safely parse integers
func parseIntSafe(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	
	// Simple integer parsing without importing strconv
	result := 0
	for _, r := range s {
		if r >= '0' && r <= '9' {
			result = result*10 + int(r-'0')
		} else {
			return defaultVal
		}
	}
	
	if result == 0 {
		return defaultVal
	}
	return result
}