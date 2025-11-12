package models

import (
	"github.com/google/uuid"
	"time"
)

// Candidate represents a candidate in the election
type Candidate struct {
	ID             int        `json:"id" example:"1"`
	Name           string     `json:"name" example:"John Doe"`
	Vision         *string    `json:"vision" example:"My vision for better education"`
	Mission        *string    `json:"mission" example:"Serve all students equally"`
	Class          *string    `json:"class" example:"2024"`
	UserID         *uuid.UUID `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Major          string     `json:"major" example:"informatics"`
	ProfilePicture *string    `json:"profile_picture" example:"https://example.com/image.jpg"`
	CreatedAt      time.Time  `json:"created_at" example:"2025-11-12T03:00:00Z"`
	UpdatedAt      time.Time  `json:"updated_at" example:"2025-11-12T03:00:00Z"`
}

// CreateCandidateRequest represents the request body for creating a candidate
type CreateCandidateRequest struct {
	Name           string  `json:"name" binding:"required" example:"John Doe"`
	Vision         *string `json:"vision" example:"My vision for better education"`
	Mission        *string `json:"mission" example:"Serve all students equally"`
	Class          *string `json:"class" example:"2024"`
	UserID         *string `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Major          string  `json:"major" binding:"required" example:"informatics"`
	ProfilePicture *string `json:"profile_picture" example:"https://example.com/image.jpg"`
}

// UpdateCandidateRequest represents the request body for updating a candidate
type UpdateCandidateRequest struct {
	Name           *string `json:"name" example:"John Doe Updated"`
	Vision         *string `json:"vision" example:"Updated vision"`
	Mission        *string `json:"mission" example:"Updated mission"`
	Class          *string `json:"class" example:"2024"`
	UserID         *string `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Major          *string `json:"major" example:"informatics"`
	ProfilePicture *string `json:"profile_picture" example:"https://example.com/new-image.jpg"`
}

// CandidateResponse represents the response with vote count
type CandidateResponse struct {
	ID             int        `json:"id" example:"1"`
	Name           string     `json:"name" example:"John Doe"`
	Vision         *string    `json:"vision" example:"My vision for better education"`
	Mission        *string    `json:"mission" example:"Serve all students equally"`
	Class          *string    `json:"class" example:"2024"`
	UserID         *uuid.UUID `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Major          string     `json:"major" example:"informatics"`
	ProfilePicture *string    `json:"profile_picture" example:"https://example.com/image.jpg"`
	VoteCount      int        `json:"vote_count" example:"5"`
	CreatedAt      time.Time  `json:"created_at" example:"2025-11-12T03:00:00Z"`
	UpdatedAt      time.Time  `json:"updated_at" example:"2025-11-12T03:00:00Z"`
}
