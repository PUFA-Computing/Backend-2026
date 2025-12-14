package models

import (
	"github.com/google/uuid"
	"time"
)

// Project represents a project in the system
type Project struct {
	ID          int        `json:"id" example:"1"`
	UserID      uuid.UUID  `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Title       string     `json:"title" example:"My Awesome Project"`
	Description string     `json:"description" example:"A detailed description of the project"`
	Category    *string    `json:"category" example:"Website"`
	ProjectURL  *string    `json:"project_url" example:"https://github.com/user/project"`
	ImageURL    string     `json:"image_url" example:"https://example.com/image.jpg"`
	IsPublished bool       `json:"is_published" example:"false"`
	VoteCount   int        `json:"vote_count" example:"0"`
	CreatedAt   time.Time  `json:"created_at" example:"2025-12-14T03:00:00Z"`
	UpdatedAt   time.Time  `json:"updated_at" example:"2025-12-14T03:00:00Z"`
}

// ProjectVote represents a vote on a project
type ProjectVote struct {
	ID        int       `json:"id" example:"1"`
	ProjectID int       `json:"project_id" example:"1"`
	UserID    uuid.UUID `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	CreatedAt time.Time `json:"created_at" example:"2025-12-14T03:00:00Z"`
}

// CreateProjectRequest represents the request body for creating a project
type CreateProjectRequest struct {
	Title       string  `json:"title" binding:"required" example:"My Awesome Project"`
	Description string  `json:"description" binding:"required" example:"A detailed description"`
	Category    *string `json:"category" example:"Website"`
	ProjectURL  *string `json:"project_url" example:"https://github.com/user/project"`
}

// UpdateProjectRequest represents the request body for updating a project
type UpdateProjectRequest struct {
	Title       *string `json:"title" example:"Updated Project Title"`
	Description *string `json:"description" example:"Updated description"`
	Category    *string `json:"category" example:"AI"`
	ProjectURL  *string `json:"project_url" example:"https://github.com/user/updated-project"`
}

// ProjectResponse represents a project with additional information
type ProjectResponse struct {
	ID          int        `json:"id" example:"1"`
	UserID      uuid.UUID  `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	UserName    string     `json:"user_name" example:"John Doe"`
	Title       string     `json:"title" example:"My Awesome Project"`
	Description string     `json:"description" example:"A detailed description of the project"`
	Category    *string    `json:"category" example:"Website"`
	ProjectURL  *string    `json:"project_url" example:"https://github.com/user/project"`
	ImageURL    string     `json:"image_url" example:"https://example.com/image.jpg"`
	IsPublished bool       `json:"is_published" example:"true"`
	VoteCount   int        `json:"vote_count" example:"5"`
	CreatedAt   time.Time  `json:"created_at" example:"2025-12-14T03:00:00Z"`
	UpdatedAt   time.Time  `json:"updated_at" example:"2025-12-14T03:00:00Z"`
}

// VoteProjectRequest represents the request body for voting on a project
type VoteProjectRequest struct {
	ProjectID int `json:"project_id" binding:"required" example:"1"`
}

// ProjectVoteResponse represents a vote with additional information
type ProjectVoteResponse struct {
	ID           int       `json:"id" example:"1"`
	ProjectID    int       `json:"project_id" example:"1"`
	ProjectTitle string    `json:"project_title" example:"My Awesome Project"`
	UserID       uuid.UUID `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	UserName     string    `json:"user_name" example:"Jane Smith"`
	CreatedAt    time.Time `json:"created_at" example:"2025-12-14T03:00:00Z"`
}

// ProjectVoteStatusResponse represents whether a user has voted for a project
type ProjectVoteStatusResponse struct {
	HasVoted  bool       `json:"has_voted" example:"true"`
	VoteID    *int       `json:"vote_id,omitempty" example:"1"`
	ProjectID *int       `json:"project_id,omitempty" example:"1"`
	VotedAt   *time.Time `json:"voted_at,omitempty" example:"2025-12-14T03:00:00Z"`
}
