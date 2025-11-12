package app

import (
	"Backend/internal/database"
	"Backend/internal/models"
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"strconv"
	"time"
)

// CastVote inserts a new vote into the database
func CastVote(vote *models.Vote) error {
	query := `
		INSERT INTO votes (voter_id, candidate_id)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at`

	err := database.DB.QueryRow(
		context.Background(),
		query,
		vote.VoterID,
		vote.CandidateID,
	).Scan(&vote.ID, &vote.CreatedAt, &vote.UpdatedAt)

	return err
}

// GetVoteByVoterID retrieves a vote by voter ID
func GetVoteByVoterID(voterID uuid.UUID) (*models.Vote, error) {
	var vote models.Vote
	query := `
		SELECT id, voter_id, candidate_id, created_at, updated_at
		FROM votes
		WHERE voter_id = $1`

	err := database.DB.QueryRow(context.Background(), query, voterID).Scan(
		&vote.ID,
		&vote.VoterID,
		&vote.CandidateID,
		&vote.CreatedAt,
		&vote.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &vote, nil
}

// GetVoteByID retrieves a vote by its ID
func GetVoteByID(voteID int) (*models.Vote, error) {
	var vote models.Vote
	query := `
		SELECT id, voter_id, candidate_id, created_at, updated_at
		FROM votes
		WHERE id = $1`

	err := database.DB.QueryRow(context.Background(), query, voteID).Scan(
		&vote.ID,
		&vote.VoterID,
		&vote.CandidateID,
		&vote.CreatedAt,
		&vote.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &vote, nil
}

// CheckUserHasVoted checks if a user has already voted
func CheckUserHasVoted(voterID uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM votes WHERE voter_id = $1)`
	err := database.DB.QueryRow(context.Background(), query, voterID).Scan(&exists)
	return exists, err
}

// GetVoteStatus retrieves the vote status for a user
func GetVoteStatus(voterID uuid.UUID) (*models.VoteStatusResponse, error) {
	var status models.VoteStatusResponse
	var voteID *int
	var candidateID *int
	var votedAt *time.Time

	query := `
		SELECT id, candidate_id, created_at
		FROM votes
		WHERE voter_id = $1`

	err := database.DB.QueryRow(context.Background(), query, voterID).Scan(
		&voteID,
		&candidateID,
		&votedAt,
	)

	if err != nil {
		// Check if it's a "no rows" error (user hasn't voted)
		if err == sql.ErrNoRows {
			status.HasVoted = false
			return &status, nil
		}
		// For other errors, return the error
		return nil, err
	}

	// User has voted
	status.HasVoted = true
	status.VoteID = voteID
	status.CandidateID = candidateID
	status.VotedAt = votedAt

	return &status, nil
}

// ListVotes returns a list of all votes with additional information
func ListVotes(queryParams map[string]string) ([]*models.VoteResponse, int, error) {
	limit := 50
	query := `
		SELECT v.id, v.voter_id, v.candidate_id, v.created_at, v.updated_at,
		       c.name as candidate_name,
		       CONCAT(u.first_name, ' ', u.last_name) as voter_name
		FROM votes v
		LEFT JOIN candidates c ON v.candidate_id = c.id
		LEFT JOIN users u ON v.voter_id = u.id
		WHERE 1 = 1`

	// Add filters
	if candidateIDStr := queryParams["candidate_id"]; candidateIDStr != "" {
		query += " AND v.candidate_id = " + candidateIDStr
	}

	// Count total records
	countQuery := `SELECT COUNT(*) FROM votes WHERE 1 = 1`
	if candidateIDStr := queryParams["candidate_id"]; candidateIDStr != "" {
		countQuery += " AND candidate_id = " + candidateIDStr
	}

	var totalRecords int
	err := database.DB.QueryRow(context.Background(), countQuery).Scan(&totalRecords)
	if err != nil {
		return nil, 0, err
	}

	totalPages := (totalRecords + limit - 1) / limit

	// Add pagination
	if pageStr := queryParams["page"]; pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			return nil, totalPages, err
		}
		offset := (page - 1) * limit
		query += fmt.Sprintf(" ORDER BY v.created_at DESC LIMIT %d OFFSET %d", limit, offset)
	} else {
		query += fmt.Sprintf(" ORDER BY v.created_at DESC LIMIT %d", limit)
	}

	rows, err := database.DB.Query(context.Background(), query)
	if err != nil {
		return nil, totalPages, err
	}
	defer rows.Close()

	var votes []*models.VoteResponse
	for rows.Next() {
		var vote models.VoteResponse
		err := rows.Scan(
			&vote.ID,
			&vote.VoterID,
			&vote.CandidateID,
			&vote.CreatedAt,
			&vote.UpdatedAt,
			&vote.CandidateName,
			&vote.VoterName,
		)
		if err != nil {
			return nil, totalPages, err
		}
		votes = append(votes, &vote)
	}

	return votes, totalPages, nil
}

// GetVoteCountByCandidate returns the number of votes for a specific candidate
func GetVoteCountByCandidate(candidateID int) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM votes WHERE candidate_id = $1`
	err := database.DB.QueryRow(context.Background(), query, candidateID).Scan(&count)
	return count, err
}

// DeleteVote removes a vote from the database (admin only)
func DeleteVote(voteID int) error {
	query := `DELETE FROM votes WHERE id = $1`
	_, err := database.DB.Exec(context.Background(), query, voteID)
	return err
}

// GetVotesByCandidateID retrieves all votes for a specific candidate
func GetVotesByCandidateID(candidateID int) ([]*models.VoteResponse, error) {
	query := `
		SELECT v.id, v.voter_id, v.candidate_id, v.created_at, v.updated_at,
		       c.name as candidate_name,
		       CONCAT(u.first_name, ' ', u.last_name) as voter_name
		FROM votes v
		LEFT JOIN candidates c ON v.candidate_id = c.id
		LEFT JOIN users u ON v.voter_id = u.id
		WHERE v.candidate_id = $1
		ORDER BY v.created_at DESC`

	rows, err := database.DB.Query(context.Background(), query, candidateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var votes []*models.VoteResponse
	for rows.Next() {
		var vote models.VoteResponse
		err := rows.Scan(
			&vote.ID,
			&vote.VoterID,
			&vote.CandidateID,
			&vote.CreatedAt,
			&vote.UpdatedAt,
			&vote.CandidateName,
			&vote.VoterName,
		)
		if err != nil {
			return nil, err
		}
		votes = append(votes, &vote)
	}

	return votes, nil
}
