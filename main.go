package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	_ "github.com/marcboeker/go-duckdb"
)

// Structs for our data models
type Stamp struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	ScottNumber  *string   `json:"scott_number,omitempty"`
	IssueDate    *string   `json:"issue_date,omitempty"`
	Series       *string   `json:"series,omitempty"`
	Condition    *string   `json:"condition,omitempty"`
	Quantity     int       `json:"quantity"`
	BoxID        *string   `json:"box_id,omitempty"`
	BoxName      *string   `json:"box_name,omitempty"` // For joined queries
	Notes        *string   `json:"notes,omitempty"`
	ImageURL     *string   `json:"image_url,omitempty"`
	IsOwned      bool      `json:"is_owned"`
	DateAdded    time.Time `json:"date_added"`
	DateModified time.Time `json:"date_modified"`
	Tags         []string  `json:"tags,omitempty"`
}

type StorageBox struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	DateCreated time.Time `json:"date_created"`
	StampCount  int       `json:"stamp_count,omitempty"` // For summary queries
}

type Tag struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	StampCount int    `json:"stamp_count,omitempty"` // For summary queries
}

type Stats struct {
	TotalOwned   int `json:"total_owned"`
	UniqueStamps int `json:"unique_stamps"`
	StampsNeeded int `json:"stamps_needed"`
	StorageBoxes int `json:"storage_boxes"`
}

var db *sql.DB
var templates *template.Template // For HTML templates

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

func main() {
	var err error
	// Initialize database
	db, err = sql.Open("duckdb", "stampkeeper.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Create tables
	if err := createTables(); err != nil {
		log.Fatal("Failed to create tables:", err)
	}

	// Seed with sample data (optional)
	if err := seedSampleData(); err != nil {
		log.Println("Warning: Failed to seed sample data:", err)
	}

	// Parse HTML templates
	templates = template.New("").Funcs(template.FuncMap{
		"substr": substr,
	})
	templates = template.Must(templates.ParseGlob("templates/*.html"))

	// Setup routes
	r := mux.NewRouter()

	// API routes (JSON)
	api := r.PathPrefix("/api").Subrouter()

	// Stamps endpoints
	api.HandleFunc("/stamps", getStamps).Methods("GET")
	api.HandleFunc("/stamps", createStamp).Methods("POST")
	api.HandleFunc("/stamps/{id}", getStamp).Methods("GET")
	api.HandleFunc("/stamps/{id}", updateStamp).Methods("PUT")
	api.HandleFunc("/stamps/{id}", deleteStamp).Methods("DELETE")

	// Storage boxes endpoints
	api.HandleFunc("/boxes", getBoxes).Methods("GET")
	api.HandleFunc("/boxes", createBox).Methods("POST")
	api.HandleFunc("/boxes/{id}", getBox).Methods("GET")
	api.HandleFunc("/boxes/{id}", updateBox).Methods("PUT")
	api.HandleFunc("/boxes/{id}", deleteBox).Methods("DELETE")

	// Tags endpoints
	api.HandleFunc("/tags", getTags).Methods("GET")
	api.HandleFunc("/tags", createTag).Methods("POST")
	api.HandleFunc("/tags/{id}", updateTag).Methods("PUT")
	api.HandleFunc("/tags/{id}", deleteTag).Methods("DELETE")

	// Stats endpoint
	api.HandleFunc("/stats", getStats).Methods("GET")

	// --- HTMX View Endpoints ---
	r.HandleFunc("/views/stamps/{view:gallery|list}", handleGetStampsView).Methods("GET")
	r.HandleFunc("/views/boxes-list", handleGetBoxesView).Methods("GET")

	// --- Static File Server ---
	// This serves files out of the 'static' directory
	fs := http.FileServer(http.Dir("./static/"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	// This serves the index.html file on the root path
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})

	fmt.Println("StampKeeper server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

// --- Handler to render HTML for stamps ---
func handleGetStampsView(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	view := vars["view"]

	stamps, err := fetchStampsFromDB(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	templateName := view + "-view.html"
	err = templates.ExecuteTemplate(w, templateName, stamps)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// --- Handler to render HTML for the box list ---
func handleGetBoxesView(w http.ResponseWriter, r *http.Request) {
	boxes, err := fetchBoxesFromDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = templates.ExecuteTemplate(w, "box-list.html", boxes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// --- Logic to fetch stamps (used by both API and HTML views) ---
func fetchStampsFromDB(r *http.Request) ([]Stamp, error) {
	query := `
		SELECT s.id, s.name, s.scott_number, s.issue_date, s.series, s.condition, 
		       s.quantity, s.box_id, sb.name as box_name, s.notes, s.image_url, 
		       s.is_owned, s.date_added, s.date_modified
		FROM stamps s
		LEFT JOIN storage_boxes sb ON s.box_id = sb.id
		WHERE 1=1`

	args := []interface{}{}

	// Add filters based on query parameters
	if search := r.URL.Query().Get("search"); search != "" {
		query += ` AND (s.name ILIKE ? OR s.scott_number ILIKE ? OR s.series ILIKE ?)`
		searchParam := "%" + search + "%"
		args = append(args, searchParam, searchParam, searchParam)
	}

	if owned := r.URL.Query().Get("owned"); owned != "" {
		if owned == "true" {
			query += ` AND s.is_owned = true`
		} else if owned == "false" {
			query += ` AND s.is_owned = false`
		}
	}

	if boxID := r.URL.Query().Get("box_id"); boxID != "" {
		query += ` AND s.box_id = ?`
		args = append(args, boxID)
	}

	// Add sorting
	sortBy := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order")
	if order == "" {
		order = "ASC" // Default order
	}
	order = strings.ToUpper(order)
	if order != "ASC" && order != "DESC" {
		order = "ASC" // Sanitize
	}


	switch sortBy {
	case "scott_number":
		query += ` ORDER BY s.scott_number ` + order
	case "name":
		query += ` ORDER BY s.name ` + order
	case "issue_date":
		query += ` ORDER BY s.issue_date ` + order
	case "date_added":
		query += ` ORDER BY s.date_added DESC` // always newest first
	default:
		query += ` ORDER BY s.date_added DESC`
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stamps []Stamp
	for rows.Next() {
		var s Stamp
		var dateAdded, dateModified string
		err := rows.Scan(&s.ID, &s.Name, &s.ScottNumber, &s.IssueDate, &s.Series,
			&s.Condition, &s.Quantity, &s.BoxID, &s.BoxName, &s.Notes, &s.ImageURL,
			&s.IsOwned, &dateAdded, &dateModified)
		if err != nil {
			return nil, err
		}

		if s.ImageURL == nil || *s.ImageURL == "" {
			placeholder := "https://via.placeholder.com/200x200.png?text=No+Image"
			s.ImageURL = &placeholder
		}

		s.DateAdded, _ = time.Parse(time.RFC3339, dateAdded)
		s.DateModified, _ = time.Parse(time.RFC3339, dateModified)
		s.Tags, _ = getStampTags(s.ID)
		stamps = append(stamps, s)
	}
	return stamps, nil
}

// --- Logic to fetch boxes ---
func fetchBoxesFromDB() ([]StorageBox, error) {
	query := `
		SELECT sb.id, sb.name, sb.date_created, COUNT(s.id) as stamp_count
		FROM storage_boxes sb
		LEFT JOIN stamps s ON sb.id = s.box_id
		GROUP BY sb.id, sb.name, sb.date_created
		ORDER BY sb.name`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var boxes []StorageBox
	for rows.Next() {
		var b StorageBox
		var dateCreated string
		err := rows.Scan(&b.ID, &b.Name, &dateCreated, &b.StampCount)
		if err != nil {
			return nil, err
		}
		b.DateCreated, _ = time.Parse(time.RFC3339, dateCreated)
		boxes = append(boxes, b)
	}
	return boxes, nil
}


// --- Handler JSON API ---
func getStamps(w http.ResponseWriter, r *http.Request) {
	stamps, err := fetchStampsFromDB(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stamps)
}

// --- Handler JSON API ---
func getBoxes(w http.ResponseWriter, r *http.Request) {
	boxes, err := fetchBoxesFromDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(boxes)
}


func createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS storage_boxes (
			id TEXT PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			date_created TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS stamps (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			scott_number TEXT,
			issue_date TEXT,
			series TEXT,
			condition TEXT,
			quantity INTEGER DEFAULT 1,
			box_id TEXT,
			notes TEXT,
			image_url TEXT,
			is_owned BOOLEAN DEFAULT true,
			date_added TEXT NOT NULL,
			date_modified TEXT NOT NULL,
			FOREIGN KEY (box_id) REFERENCES storage_boxes(id)
		)`,
		`CREATE TABLE IF NOT EXISTS tags (
			id TEXT PRIMARY KEY,
			name TEXT UNIQUE NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS stamp_tags (
			stamp_id TEXT,
			tag_id TEXT,
			PRIMARY KEY (stamp_id, tag_id),
			FOREIGN KEY (stamp_id) REFERENCES stamps(id),
			FOREIGN KEY (tag_id) REFERENCES tags(id)
		)`,
	}

	for _, query := range queries {
		// NOTE: Using ILIKE for case-insensitive search in DuckDB for some text fields could be beneficial
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query %s: %v", query, err)
		}
	}

	return nil
}

func seedSampleData() error {
	// Check if we already have data
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM stamps").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // Already seeded
	}

	// Create sample storage box
	boxID := uuid.New().String()
	_, err = db.Exec(`INSERT INTO storage_boxes (id, name, date_created) VALUES (?, ?, ?)`,
		boxID, "Box 1", time.Now().Format(time.RFC3339))
	if err != nil {
		return err
	}
	box2ID := uuid.New().String()
	_, err = db.Exec(`INSERT INTO storage_boxes (id, name, date_created) VALUES (?, ?, ?)`,
		box2ID, "Box 2", time.Now().Format(time.RFC3339))
	if err != nil {
		return err
	}

	// Create sample stamps
	stamps := []struct {
		name, scottNum, issueDate, series, condition, boxID string
		quantity                                      int
		isOwned                                       bool
	}{
		{"Lincoln 1c Green", "219", "1890-02-22", "1890-93 Regular Issue", "Used", boxID, 1, true},
		{"Washington 2c Carmine", "220", "1890-02-22", "1890-93 Regular Issue", "Mint", boxID, 1, true},
		{"Jackson 3c Purple", "221", "1890-02-22", "1890-93 Regular Issue", "Used", boxID, 1, false},
		{"German Empire 10pf", "55", "1900-01-01", "Germania", "Mint", box2ID, 2, true},
	}

	for _, s := range stamps {
		stampID := uuid.New().String()
		_, err = db.Exec(`INSERT INTO stamps 
			(id, name, scott_number, issue_date, series, condition, quantity, box_id, is_owned, date_added, date_modified) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			stampID, s.name, s.scottNum, s.issueDate, s.series, s.condition, s.quantity,
			s.boxID, s.isOwned, time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339))
		if err != nil {
			return err
		}
	}

	// Create sample tags
	tagNames := []string{"USA", "Classic", "Presidential"}
	for _, tagName := range tagNames {
		tagID := uuid.New().String()
		_, err = db.Exec(`INSERT INTO tags (id, name) VALUES (?, ?)`, tagID, tagName)
		if err != nil {
			// ignore unique constraint violation
			if !strings.Contains(err.Error(), "UNIQUE constraint failed") {
				return err
			}
		}
	}

	return nil
}

// Stamps handlers
func getStamp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var s Stamp
	var dateAdded, dateModified string
	query := `
		SELECT s.id, s.name, s.scott_number, s.issue_date, s.series, s.condition, 
		       s.quantity, s.box_id, sb.name as box_name, s.notes, s.image_url, 
		       s.is_owned, s.date_added, s.date_modified
		FROM stamps s
		LEFT JOIN storage_boxes sb ON s.box_id = sb.id
		WHERE s.id = ?`

	err := db.QueryRow(query, id).Scan(&s.ID, &s.Name, &s.ScottNumber, &s.IssueDate,
		&s.Series, &s.Condition, &s.Quantity, &s.BoxID, &s.BoxName, &s.Notes,
		&s.ImageURL, &s.IsOwned, &dateAdded, &dateModified)
	
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Stamp not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Parse timestamps
	if s.DateAdded, err = time.Parse(time.RFC3339, dateAdded); err != nil {
		s.DateAdded = time.Now() // fallback
	}
	if s.DateModified, err = time.Parse(time.RFC3339, dateModified); err != nil {
		s.DateModified = time.Now() // fallback
	}

	// Get tags
	s.Tags, _ = getStampTags(s.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}

func createStamp(w http.ResponseWriter, r *http.Request) {
	var s Stamp
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.ID = uuid.New().String()
	s.DateAdded = time.Now()
	s.DateModified = time.Now()
	if s.Quantity == 0 {
		s.Quantity = 1
	}

	_, err := db.Exec(`INSERT INTO stamps 
		(id, name, scott_number, issue_date, series, condition, quantity, box_id, notes, image_url, is_owned, date_added, date_modified) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.ID, s.Name, s.ScottNumber, s.IssueDate, s.Series, s.Condition, s.Quantity,
		s.BoxID, s.Notes, s.ImageURL, s.IsOwned, s.DateAdded.Format(time.RFC3339), s.DateModified.Format(time.RFC3339))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Handle tags
	if len(s.Tags) > 0 {
		updateStampTags(s.ID, s.Tags)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(s)
}

func updateStamp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var s Stamp
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.ID = id
	s.DateModified = time.Now()

	_, err := db.Exec(`UPDATE stamps SET 
		name=?, scott_number=?, issue_date=?, series=?, condition=?, quantity=?, 
		box_id=?, notes=?, image_url=?, is_owned=?, date_modified=?
		WHERE id=?`,
		s.Name, s.ScottNumber, s.IssueDate, s.Series, s.Condition, s.Quantity,
		s.BoxID, s.Notes, s.ImageURL, s.IsOwned, s.DateModified.Format(time.RFC3339), s.ID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update tags
	updateStampTags(s.ID, s.Tags)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}

func deleteStamp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	// First, delete associations in stamp_tags
	_, err := db.Exec("DELETE FROM stamp_tags WHERE stamp_id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Then, delete the stamp
	_, err = db.Exec("DELETE FROM stamps WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Storage box handlers
func getBox(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var b StorageBox
	var dateCreated string
	err := db.QueryRow(`SELECT id, name, date_created FROM storage_boxes WHERE id = ?`, id).
		Scan(&b.ID, &b.Name, &dateCreated)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Box not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Parse timestamp
	if b.DateCreated, err = time.Parse(time.RFC3339, dateCreated); err != nil {
		b.DateCreated = time.Now() // fallback
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(b)
}

func createBox(w http.ResponseWriter, r *http.Request) {
	var b StorageBox
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	b.ID = uuid.New().String()
	b.DateCreated = time.Now()

	_, err := db.Exec(`INSERT INTO storage_boxes (id, name, date_created) VALUES (?, ?, ?)`,
		b.ID, b.Name, b.DateCreated.Format(time.RFC3339))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(b)
}

func updateBox(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var b StorageBox
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec(`UPDATE storage_boxes SET name = ? WHERE id = ?`, b.Name, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(b)
}

func deleteBox(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Clear box_id from stamps first
	_, err := db.Exec("UPDATE stamps SET box_id = NULL WHERE box_id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Delete the box
	_, err = db.Exec("DELETE FROM storage_boxes WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Tag handlers
func getTags(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT t.id, t.name, COUNT(st.stamp_id) as stamp_count
		FROM tags t
		LEFT JOIN stamp_tags st ON t.id = st.tag_id
		GROUP BY t.id, t.name
		ORDER BY t.name`

	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var t Tag
		err := rows.Scan(&t.ID, &t.Name, &t.StampCount)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tags = append(tags, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tags)
}

func createTag(w http.ResponseWriter, r *http.Request) {
	var t Tag
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	t.ID = uuid.New().String()
	_, err := db.Exec(`INSERT INTO tags (id, name) VALUES (?, ?)`, t.ID, t.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

func updateTag(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var t Tag
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec(`UPDATE tags SET name = ? WHERE id = ?`, t.Name, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

func deleteTag(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	_, err := db.Exec("DELETE FROM tags WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Stats handler
func getStats(w http.ResponseWriter, r *http.Request) {
	var stats Stats

	// Total owned stamps
	db.QueryRow("SELECT COUNT(*) FROM stamps WHERE is_owned = true").Scan(&stats.TotalOwned)
	
	// Unique stamps (distinct scott numbers, but handle nulls)
	db.QueryRow("SELECT COUNT(DISTINCT scott_number) FROM stamps WHERE scott_number IS NOT NULL").Scan(&stats.UniqueStamps)
	
	// Stamps needed
	db.QueryRow("SELECT COUNT(*) FROM stamps WHERE is_owned = false").Scan(&stats.StampsNeeded)
	
	// Storage boxes
	db.QueryRow("SELECT COUNT(*) FROM storage_boxes").Scan(&stats.StorageBoxes)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// Helper functions
func getStampTags(stampID string) ([]string, error) {
	rows, err := db.Query(`
		SELECT t.name 
		FROM tags t 
		JOIN stamp_tags st ON t.id = st.tag_id 
		WHERE st.stamp_id = ?`, stampID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func updateStampTags(stampID string, tags []string) error {
	// Begin a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Remove existing tags for this stamp
	_, err = tx.Exec("DELETE FROM stamp_tags WHERE stamp_id = ?", stampID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Add new tags
	for _, tagName := range tags {
		tagName = strings.TrimSpace(tagName)
		if tagName == "" {
			continue
		}

		// Get or create tag
		var tagID string
		err := tx.QueryRow("SELECT id FROM tags WHERE name = ?", tagName).Scan(&tagID)
		if err == sql.ErrNoRows {
			// Create new tag
			tagID = uuid.New().String()
			_, err = tx.Exec("INSERT INTO tags (id, name) VALUES (?, ?)", tagID, tagName)
			if err != nil {
				tx.Rollback()
				return err
			}
		} else if err != nil {
			tx.Rollback()
			return err
		}

		// Link stamp to tag
		_, err = tx.Exec("INSERT INTO stamp_tags (stamp_id, tag_id) VALUES (?, ?)", stampID, tagID)
		if err != nil {
			// Ignore if the link already exists (e.g., from a concurrent request)
			if !strings.Contains(err.Error(), "UNIQUE constraint failed") {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit()
}