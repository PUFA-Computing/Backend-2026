package models

import (
	"github.com/google/uuid"
	"time"
)

// Project represents a project in the system
type Project struct {
	ID               int        `json:"id" example:"1"`
	UserID           uuid.UUID  `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Title            string     `json:"title" example:"My Awesome Project"`
	Description      string     `json:"description" example:"A detailed description of the project"`
	Category         *string    `json:"category" example:"Website"`
	ProjectURL       *string    `json:"project_url" example:"https://github.com/user/project"`
	ImageURL         string     `json:"image_url" example:"https://example.com/image.jpg"`
	ProjectMembers   []string   `json:"project_members" example:"[\"John Doe\",\"Jane Smith\"]"`
	LinkedInProfiles []string   `json:"linkedin_profiles" example:"[\"https://linkedin.com/in/johndoe\",\"https://linkedin.com/in/janesmith\"]"`
	Major            *string    `json:"major" example:"information_system"`
	Batch            *int       `json:"batch" example:"2025"`
	IsPublished      bool       `json:"is_published" example:"false"`
	VoteCount        int        `json:"vote_count" example:"0"`
	ApprovedBy       *uuid.UUID `json:"approved_by,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
	ApprovedAt       *time.Time `json:"approved_at,omitempty" example:"2025-12-15T03:00:00Z"`
	RejectionReason  *string    `json:"rejection_reason,omitempty" example:"Does not meet quality standards"`
	CreatedAt        time.Time  `json:"created_at" example:"2025-12-14T03:00:00Z"`
	UpdatedAt        time.Time  `json:"updated_at" example:"2025-12-14T03:00:00Z"`
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
	Title            string   `json:"title" binding:"required" example:"My Awesome Project"`
	Description      string   `json:"description" binding:"required" example:"A detailed description"`
	Category         *string  `json:"category" example:"Website"`
	ProjectURL       *string  `json:"project_url" example:"https://github.com/user/project"`
	ProjectMembers   []string `json:"project_members" binding:"required,min=1,max=10,dive,required" example:"[\"John Doe\",\"Jane Smith\"]"`
	LinkedInProfiles []string `json:"linkedin_profiles" binding:"required,dive,required,url" example:"[\"https://linkedin.com/in/johndoe\",\"https://linkedin.com/in/janesmith\"]"`
	Major            string   `json:"major" binding:"required,oneof=information_system informatics" example:"information_system"`
	Batch            int      `json:"batch" binding:"required,min=2021,max=2025" example:"2025"`
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
	ID              int        `json:"id" example:"1"`
	UserID          uuid.UUID  `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	UserName        string     `json:"user_name" example:"John Doe"`
	Title           string     `json:"title" example:"My Awesome Project"`
	Description     string     `json:"description" example:"A detailed description of the project"`
	Category        *string    `json:"category" example:"Website"`
	ProjectURL      *string    `json:"project_url" example:"https://github.com/user/project"`
	ImageURL        string     `json:"image_url" example:"https://example.com/image.jpg"`
	ProjectMembers   []string   `json:"project_members" example:"[\"John Doe\",\"Jane Smith\"]"`
	LinkedInProfiles []string   `json:"linkedin_profiles" example:"[\"https://linkedin.com/in/johndoe\",\"https://linkedin.com/in/janesmith\"]"`
	Major            *string    `json:"major" example:"information_system"`
	Batch            *int       `json:"batch" example:"2025"`
	IsPublished     bool       `json:"is_published" example:"true"`
	VoteCount       int        `json:"vote_count" example:"5"`
	ApprovedBy      *uuid.UUID `json:"approved_by,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
	ApprovedByName  *string    `json:"approved_by_name,omitempty" example:"Admin User"`
	ApprovedAt      *time.Time `json:"approved_at,omitempty" example:"2025-12-15T03:00:00Z"`
	RejectionReason *string    `json:"rejection_reason,omitempty" example:"Does not meet quality standards"`
	CreatedAt       time.Time  `json:"created_at" example:"2025-12-14T03:00:00Z"`
	UpdatedAt       time.Time  `json:"updated_at" example:"2025-12-14T03:00:00Z"`
}

// VoteProjectRequest represents the request body for voting on a project (no longer needed)
type VoteProjectRequest struct {
	// Empty - voting no longer requires team info
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

// ApproveProjectRequest represents the request body for approving a project
type ApproveProjectRequest struct {
	Note *string `json:"note" example:"Project meets all quality standards"`
}

// RejectProjectRequest represents the request body for rejecting a project
type RejectProjectRequest struct {
	Reason string `json:"reason" binding:"required" example:"Does not meet quality standards"`
}
