package app

import (
	"Backend/internal/database"
	"Backend/internal/models"
	"context"
	"fmt"
	"github.com/google/uuid"
	"strconv"
)

// CreateCandidate inserts a new candidate into the database
func CreateCandidate(candidate *models.Candidate) error {
	query := `
		INSERT INTO candidates (name, vision, mission, class, user_id, major, profile_picture)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`

	err := database.DB.QueryRow(
		context.Background(),
		query,
		candidate.Name,
		candidate.Vision,
		candidate.Mission,
		candidate.Class,
		candidate.UserID,
		candidate.Major,
		candidate.ProfilePicture,
	).Scan(&candidate.ID, &candidate.CreatedAt, &candidate.UpdatedAt)

	return err
}

// UpdateCandidate updates an existing candidate in the database
func UpdateCandidate(candidateID int, candidate *models.Candidate) error {
	query := `
		UPDATE candidates
		SET name = $1, vision = $2, mission = $3, class = $4, user_id = $5, 
		    major = $6, profile_picture = $7, updated_at = NOW()
		WHERE id = $8`

	_, err := database.DB.Exec(
		context.Background(),
		query,
		candidate.Name,
		candidate.Vision,
		candidate.Mission,
		candidate.Class,
		candidate.UserID,
		candidate.Major,
		candidate.ProfilePicture,
		candidateID,
	)

	return err
}

// DeleteCandidate removes a candidate from the database
func DeleteCandidate(candidateID int) error {
	query := `DELETE FROM candidates WHERE id = $1`
	_, err := database.DB.Exec(context.Background(), query, candidateID)
	return err
}

// GetCandidateByID retrieves a candidate by their ID
func GetCandidateByID(candidateID int) (*models.Candidate, error) {
	var candidate models.Candidate
	query := `
		SELECT id, name, vision, mission, class, user_id, major, profile_picture, created_at, updated_at
		FROM candidates
		WHERE id = $1`

	err := database.DB.QueryRow(context.Background(), query, candidateID).Scan(
		&candidate.ID,
		&candidate.Name,
		&candidate.Vision,
		&candidate.Mission,
		&candidate.Class,
		&candidate.UserID,
		&candidate.Major,
		&candidate.ProfilePicture,
		&candidate.CreatedAt,
		&candidate.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &candidate, nil
}

// GetCandidateWithVoteCount retrieves a candidate with their vote count
func GetCandidateWithVoteCount(candidateID int) (*models.CandidateResponse, error) {
	var candidate models.CandidateResponse
	query := `
		SELECT c.id, c.name, c.vision, c.mission, c.class, c.user_id, c.major, 
		       c.profile_picture, c.created_at, c.updated_at,
		       COALESCE(COUNT(v.id), 0) as vote_count
		FROM candidates c
		LEFT JOIN votes v ON c.id = v.candidate_id
		WHERE c.id = $1
		GROUP BY c.id`

	err := database.DB.QueryRow(context.Background(), query, candidateID).Scan(
		&candidate.ID,
		&candidate.Name,
		&candidate.Vision,
		&candidate.Mission,
		&candidate.Class,
		&candidate.UserID,
		&candidate.Major,
		&candidate.ProfilePicture,
		&candidate.CreatedAt,
		&candidate.UpdatedAt,
		&candidate.VoteCount,
	)

	if err != nil {
		return nil, err
	}

	return &candidate, nil
}

// ListCandidates returns a list of candidates based on query parameters
func ListCandidates(queryParams map[string]string) ([]*models.CandidateResponse, int, error) {
	limit := 20
	query := `
		SELECT c.id, c.name, c.vision, c.mission, c.class, c.user_id, c.major, 
		       c.profile_picture, c.created_at, c.updated_at,
		       COALESCE(COUNT(v.id), 0) as vote_count
		FROM candidates c
		LEFT JOIN votes v ON c.id = v.candidate_id
		WHERE 1 = 1`

	// Add filters
	if major := queryParams["major"]; major != "" {
		query += " AND c.major = '" + major + "'"
	}

	if class := queryParams["class"]; class != "" {
		query += " AND c.class = '" + class + "'"
	}

	query += " GROUP BY c.id"

	// Count total records
	countQuery := `SELECT COUNT(*) FROM candidates WHERE 1 = 1`
	if major := queryParams["major"]; major != "" {
		countQuery += " AND major = '" + major + "'"
	}
	if class := queryParams["class"]; class != "" {
		countQuery += " AND class = '" + class + "'"
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
		query += fmt.Sprintf(" ORDER BY c.created_at DESC LIMIT %d OFFSET %d", limit, offset)
	} else {
		query += fmt.Sprintf(" ORDER BY c.created_at DESC LIMIT %d", limit)
	}

	rows, err := database.DB.Query(context.Background(), query)
	if err != nil {
		return nil, totalPages, err
	}
	defer rows.Close()

	var candidates []*models.CandidateResponse
	for rows.Next() {
		var candidate models.CandidateResponse
		err := rows.Scan(
			&candidate.ID,
			&candidate.Name,
			&candidate.Vision,
			&candidate.Mission,
			&candidate.Class,
			&candidate.UserID,
			&candidate.Major,
			&candidate.ProfilePicture,
			&candidate.CreatedAt,
			&candidate.UpdatedAt,
			&candidate.VoteCount,
		)
		if err != nil {
			return nil, totalPages, err
		}
		candidates = append(candidates, &candidate)
	}

	return candidates, totalPages, nil
}

// GetCandidatesByMajor retrieves all candidates for a specific major
func GetCandidatesByMajor(major string) ([]*models.CandidateResponse, error) {
	query := `
		SELECT c.id, c.name, c.vision, c.mission, c.class, c.user_id, c.major, 
		       c.profile_picture, c.created_at, c.updated_at,
		       COALESCE(COUNT(v.id), 0) as vote_count
		FROM candidates c
		LEFT JOIN votes v ON c.id = v.candidate_id
		WHERE c.major = $1
		GROUP BY c.id
		ORDER BY c.created_at DESC`

	rows, err := database.DB.Query(context.Background(), query, major)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var candidates []*models.CandidateResponse
	for rows.Next() {
		var candidate models.CandidateResponse
		err := rows.Scan(
			&candidate.ID,
			&candidate.Name,
			&candidate.Vision,
			&candidate.Mission,
			&candidate.Class,
			&candidate.UserID,
			&candidate.Major,
			&candidate.ProfilePicture,
			&candidate.CreatedAt,
			&candidate.UpdatedAt,
			&candidate.VoteCount,
		)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, &candidate)
	}

	return candidates, nil
}

// CheckCandidateExists checks if a candidate exists
func CheckCandidateExists(candidateID int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM candidates WHERE id = $1)`
	err := database.DB.QueryRow(context.Background(), query, candidateID).Scan(&exists)
	return exists, err
}

// GetCandidateMajor retrieves the major of a candidate
func GetCandidateMajor(candidateID int) (string, error) {
	var major string
	query := `SELECT major FROM candidates WHERE id = $1`
	err := database.DB.QueryRow(context.Background(), query, candidateID).Scan(&major)
	return major, err
}

// GetUserMajor retrieves the major of a user
func GetUserMajor(userID uuid.UUID) (string, error) {
	var major string
	query := `SELECT major FROM users WHERE id = $1`
	err := database.DB.QueryRow(context.Background(), query, userID).Scan(&major)
	return major, err
}

// GetUserYear retrieves the year of a user
func GetUserYear(userID uuid.UUID) (string, error) {
	var year string
	query := `SELECT year FROM users WHERE id = $1`
	err := database.DB.QueryRow(context.Background(), query, userID).Scan(&year)
	return year, err
}
