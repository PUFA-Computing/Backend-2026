package app

import (
	"Backend/internal/database"
	"Backend/internal/models"
	"context"
	"database/sql"
	"github.com/google/uuid"
	"time"
)

// VoteProject inserts a new vote for a project
func VoteProject(vote *models.ProjectVote) error {
	query := `
		INSERT INTO project_votes (project_id, user_id)
		VALUES ($1, $2)
		RETURNING id, created_at`

	err := database.DB.QueryRow(
		context.Background(),
		query,
		vote.ProjectID,
		vote.UserID,
	).Scan(&vote.ID, &vote.CreatedAt)

	if err != nil {
		return err
	}

	// Update vote count in projects table
	updateQuery := `
		UPDATE projects
		SET vote_count = vote_count + 1
		WHERE id = $1`

	_, err = database.DB.Exec(context.Background(), updateQuery, vote.ProjectID)
	return err
}

// UnvoteProject removes a vote from a project
func UnvoteProject(projectID int, userID uuid.UUID) error {
	query := `DELETE FROM project_votes WHERE project_id = $1 AND user_id = $2`
	_, err := database.DB.Exec(context.Background(), query, projectID, userID)

	if err != nil {
		return err
	}

	// Update vote count in projects table
	updateQuery := `
		UPDATE projects
		SET vote_count = GREATEST(vote_count - 1, 0)
		WHERE id = $1`

	_, err = database.DB.Exec(context.Background(), updateQuery, projectID)
	return err
}

// CheckUserHasVotedProject checks if a user has already voted for a project
func CheckUserHasVotedProject(projectID int, userID uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM project_votes WHERE project_id = $1 AND user_id = $2)`
	err := database.DB.QueryRow(context.Background(), query, projectID, userID).Scan(&exists)
	return exists, err
}

// GetProjectVotesByUser retrieves all votes cast by a specific user
func GetProjectVotesByUser(userID uuid.UUID) ([]*models.ProjectVoteResponse, error) {
	query := `
		SELECT pv.id, pv.project_id, pv.user_id, pv.created_at,
		       p.title as project_title,
		       CONCAT(u.first_name, ' ', u.last_name) as user_name
		FROM project_votes pv
		LEFT JOIN projects p ON pv.project_id = p.id
		LEFT JOIN users u ON pv.user_id = u.id
		WHERE pv.user_id = $1
		ORDER BY pv.created_at DESC`

	rows, err := database.DB.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var votes []*models.ProjectVoteResponse
	for rows.Next() {
		var vote models.ProjectVoteResponse
		err := rows.Scan(
			&vote.ID,
			&vote.ProjectID,
			&vote.UserID,
			&vote.CreatedAt,
			&vote.ProjectTitle,
			&vote.UserName,
		)
		if err != nil {
			return nil, err
		}
		votes = append(votes, &vote)
	}

	return votes, nil
}

// GetProjectVotesByProject retrieves all votes for a specific project
func GetProjectVotesByProject(projectID int) ([]*models.ProjectVoteResponse, error) {
	query := `
		SELECT pv.id, pv.project_id, pv.user_id, pv.created_at,
		       p.title as project_title,
		       CONCAT(u.first_name, ' ', u.last_name) as user_name
		FROM project_votes pv
		LEFT JOIN projects p ON pv.project_id = p.id
		LEFT JOIN users u ON pv.user_id = u.id
		WHERE pv.project_id = $1
		ORDER BY pv.created_at DESC`

	rows, err := database.DB.Query(context.Background(), query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var votes []*models.ProjectVoteResponse
	for rows.Next() {
		var vote models.ProjectVoteResponse
		err := rows.Scan(
			&vote.ID,
			&vote.ProjectID,
			&vote.UserID,
			&vote.CreatedAt,
			&vote.ProjectTitle,
			&vote.UserName,
		)
		if err != nil {
			return nil, err
		}
		votes = append(votes, &vote)
	}

	return votes, nil
}

// GetProjectVoteCount returns the number of votes for a specific project
func GetProjectVoteCount(projectID int) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM project_votes WHERE project_id = $1`
	err := database.DB.QueryRow(context.Background(), query, projectID).Scan(&count)
	return count, err
}

// GetProjectVoteStatus retrieves the vote status for a user on a specific project
func GetProjectVoteStatus(projectID int, userID uuid.UUID) (*models.ProjectVoteStatusResponse, error) {
	var status models.ProjectVoteStatusResponse
	var voteID *int
	var votedAt *time.Time

	query := `
		SELECT id, created_at
		FROM project_votes
		WHERE project_id = $1 AND user_id = $2`

	err := database.DB.QueryRow(context.Background(), query, projectID, userID).Scan(
		&voteID,
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
	status.ProjectID = &projectID
	status.VotedAt = votedAt

	return &status, nil
}
