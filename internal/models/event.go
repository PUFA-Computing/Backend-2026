package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID              int       `json:"id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	UserID          uuid.UUID `json:"user_id"`
	Status          string    `json:"status"`
	Slug            string    `json:"slug"`
	Thumbnail       string    `json:"thumbnail"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updatedAt"`
	OrganizationID  int       `json:"organization_id"`
	MaxRegistration *int      `json:"max_registration"`
	Organization    string    `json:"organization"`
	Author          string    `json:"author"`
	TotalRegistered int       `json:"total_registered"`
}

type EventRegistration struct {
	ID               int       `json:"id"`
	EventID          int       `json:"event_id"`
	UserID           uuid.UUID `json:"user_id"`
	RegistrationDate time.Time `json:"registration_date"`
	AdditionalNotes  string    `json:"additional_notes"`
	FilePath         string    `json:"file_path"`
	FilePaths        []string  `json:"file_paths" gorm:"-"`
}

// ProcessFilePaths processes the comma-separated file paths and populates the FilePaths slice
func (er *EventRegistration) ProcessFilePaths() {
	if er.FilePath == "" {
		er.FilePaths = []string{}
		return
	}

	// Split the file paths by comma
	paths := strings.Split(er.FilePath, ",")

	// Trim spaces and filter out empty paths
	var validPaths []string
	for _, path := range paths {
		trimmedPath := strings.TrimSpace(path)
		if trimmedPath != "" {
			validPaths = append(validPaths, trimmedPath)
		}
	}

	er.FilePaths = validPaths
}
