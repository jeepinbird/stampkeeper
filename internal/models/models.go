package models

import "time"

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