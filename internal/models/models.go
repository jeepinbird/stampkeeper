package models

import "time"

// StampInstance represents a group of physical copies with the same condition in the same box.
// For example: "3 Used copies in Box 1" would be one instance with Quantity=3.
type StampInstance struct {
	ID           string     `json:"id"`
	StampID      string     `json:"stamp_id"`
	Condition    *string    `json:"condition,omitempty"`
	BoxID        *string    `json:"box_id,omitempty"`
	BoxName      *string    `json:"box_name,omitempty"` // For joined queries
	Quantity     int        `json:"quantity"`
	DateAdded    time.Time  `json:"date_added"`
	DateModified time.Time  `json:"date_modified"`
	DateDeleted  *time.Time `json:"date_deleted,omitempty"` // For soft deletes
}

// Stamp represents the abstract design of a stamp.
// It holds information common to all instances of that stamp design.
type Stamp struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	ScottNumber  *string         `json:"scott_number,omitempty"`
	IssueDate    *string         `json:"issue_date,omitempty"`
	Series       *string         `json:"series,omitempty"`
	Notes        *string         `json:"notes,omitempty"` // Notes about the stamp design itself
	ImageURL     *string         `json:"image_url,omitempty"`
	IsOwned      bool            `json:"is_owned"` // Calculated: true if any instances exist
	DateAdded    time.Time       `json:"date_added"`
	DateModified time.Time       `json:"date_modified"`
	DateDeleted  *time.Time      `json:"date_deleted,omitempty"` // For soft deletes
	Tags         []string        `json:"tags,omitempty"`
	Instances    []StampInstance `json:"instances,omitempty"` // Groups of physical copies
	BoxNames     []string        `json:"box_names,omitempty"` // Comma-separated list of box names for display
}

type StorageBox struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	DateCreated time.Time `json:"date_created"`
	StampCount  int       `json:"stamp_count,omitempty"` // Total quantity of all instances in this box
}

type Tag struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	StampCount int    `json:"stamp_count,omitempty"` // Number of different stamp designs with this tag
}

// Stats calculated from instances and stamps
type Stats struct {
	TotalOwned   int `json:"total_owned"`   // Sum of all instance quantities
	UniqueStamps int `json:"unique_stamps"` // Count of distinct stamp designs
	StampsNeeded int `json:"stamps_needed"` // Stamp designs with no instances
	StorageBoxes int `json:"storage_boxes"` // Count of storage boxes
}

// --- View-specific Models ---

// PaginatedStampsView holds data for the gallery/list view.
type PaginatedStampsView struct {
	Stamps      []Stamp
	Pagination  Pagination
	BaseURL     string
	CurrentView string
	FilteredBox *StorageBox // Box being filtered on, if any
}

// Pagination holds calculated pagination data.
type Pagination struct {
	CurrentPage int
	TotalPages  int
	TotalItems  int64
	HasNext     bool
	HasPrev     bool
	NextPage    int
	PrevPage    int
}

// StampDetailView holds all data needed for the stamp detail page.
type StampDetailView struct {
	Stamp    Stamp
	AllBoxes []StorageBox // For dropdowns when editing instances
}

// SettingsView holds all data needed for the settings page.
type SettingsView struct {
	AllBoxes    []StorageBox
	Preferences UserPreferences
}

// UserPreferences represents user-specific application preferences.
type UserPreferences struct {
	DefaultView   string `json:"defaultView"`
	DefaultSort   string `json:"defaultSort"`
	SortDirection string `json:"sortDirection"`
	ItemsPerPage  int    `json:"itemsPerPage"`
}