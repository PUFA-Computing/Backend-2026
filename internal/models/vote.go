package models

import (
	"github.com/google/uuid"
	"time"
)

// Vote represents a vote cast by a user
type Vote struct {
	ID          int       `json:"id" example:"1"`
	VoterID     uuid.UUID `json:"voter_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	CandidateID int       `json:"candidate_id" example:"1"`
	CreatedAt   time.Time `json:"created_at" example:"2025-11-12T03:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2025-11-12T03:00:00Z"`
}

// CastVoteRequest represents the request body for casting a vote
type CastVoteRequest struct {
	CandidateID int `json:"candidate_id" binding:"required" example:"1"`
}

// VoteResponse represents a vote with additional information
type VoteResponse struct {
	ID            int       `json:"id" example:"1"`
	VoterID       uuid.UUID `json:"voter_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	CandidateID   int       `json:"candidate_id" example:"1"`
	CandidateName string    `json:"candidate_name" example:"John Doe"`
	VoterName     string    `json:"voter_name" example:"Jane Smith"`
	CreatedAt     time.Time `json:"created_at" example:"2025-11-12T03:00:00Z"`
	UpdatedAt     time.Time `json:"updated_at" example:"2025-11-12T03:00:00Z"`
}

// VoteStatusResponse represents whether a user has voted
type VoteStatusResponse struct {
	HasVoted    bool       `json:"has_voted" example:"true"`
	VoteID      *int       `json:"vote_id,omitempty" example:"1"`
	CandidateID *int       `json:"candidate_id,omitempty" example:"1"`
	VotedAt     *time.Time `json:"voted_at,omitempty" example:"2025-11-12T03:00:00Z"`
}
