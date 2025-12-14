package app

import (
	"Backend/internal/database"
	"Backend/internal/models"
	"context"
	"fmt"
	"github.com/google/uuid"
	"strconv"
)

// CreateProject inserts a new project into the database
func CreateProject(project *models.Project) error {
	query := `
		INSERT INTO projects (user_id, title, description, category, project_url, image_url, is_published, vote_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	err := database.DB.QueryRow(
		context.Background(),
		query,
		project.UserID,
		project.Title,
		project.Description,
		project.Category,
		project.ProjectURL,
		project.ImageURL,
		project.IsPublished,
		project.VoteCount,
	).Scan(&project.ID, &project.CreatedAt, &project.UpdatedAt)

	return err
}

// UpdateProject updates an existing project in the database
func UpdateProject(projectID int, project *models.Project) error {
	query := `
		UPDATE projects
		SET title = $1, description = $2, category = $3, project_url = $4, 
		    image_url = $5, updated_at = NOW()
		WHERE id = $6`

	_, err := database.DB.Exec(
		context.Background(),
		query,
		project.Title,
		project.Description,
		project.Category,
		project.ProjectURL,
		project.ImageURL,
		projectID,
	)

	return err
}

// DeleteProject removes a project from the database
func DeleteProject(projectID int) error {
	query := `DELETE FROM projects WHERE id = $1`
	_, err := database.DB.Exec(context.Background(), query, projectID)
	return err
}

// GetProjectByID retrieves a project by its ID
func GetProjectByID(projectID int) (*models.Project, error) {
	var project models.Project
	query := `
		SELECT id, user_id, title, description, category, project_url, image_url, 
		       is_published, vote_count, created_at, updated_at
		FROM projects
		WHERE id = $1`

	err := database.DB.QueryRow(context.Background(), query, projectID).Scan(
		&project.ID,
		&project.UserID,
		&project.Title,
		&project.Description,
		&project.Category,
		&project.ProjectURL,
		&project.ImageURL,
		&project.IsPublished,
		&project.VoteCount,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &project, nil
}

// GetProjectWithVoteCount retrieves a project with its vote count and user information
func GetProjectWithVoteCount(projectID int) (*models.ProjectResponse, error) {
	var project models.ProjectResponse
	query := `
		SELECT p.id, p.user_id, p.title, p.description, p.category, p.project_url, 
		       p.image_url, p.is_published, p.vote_count, p.created_at, p.updated_at,
		       CONCAT(u.first_name, ' ', u.last_name) as user_name
		FROM projects p
		LEFT JOIN users u ON p.user_id = u.id
		WHERE p.id = $1`

	err := database.DB.QueryRow(context.Background(), query, projectID).Scan(
		&project.ID,
		&project.UserID,
		&project.Title,
		&project.Description,
		&project.Category,
		&project.ProjectURL,
		&project.ImageURL,
		&project.IsPublished,
		&project.VoteCount,
		&project.CreatedAt,
		&project.UpdatedAt,
		&project.UserName,
	)

	if err != nil {
		return nil, err
	}

	return &project, nil
}

// ListProjects returns a list of projects based on query parameters
func ListProjects(queryParams map[string]string) ([]*models.ProjectResponse, int, error) {
	limit := 20
	query := `
		SELECT p.id, p.user_id, p.title, p.description, p.category, p.project_url, 
		       p.image_url, p.is_published, p.vote_count, p.created_at, p.updated_at,
		       CONCAT(u.first_name, ' ', u.last_name) as user_name
		FROM projects p
		LEFT JOIN users u ON p.user_id = u.id
		WHERE 1 = 1`

	// Add filters
	if category := queryParams["category"]; category != "" {
		query += " AND p.category = '" + category + "'"
	}

	if isPublished := queryParams["is_published"]; isPublished != "" {
		query += " AND p.is_published = " + isPublished
	} else {
		// By default, only show published projects for public listing
		query += " AND p.is_published = true"
	}

	if userID := queryParams["user_id"]; userID != "" {
		query += " AND p.user_id = '" + userID + "'"
	}

	// Count total records
	countQuery := `SELECT COUNT(*) FROM projects WHERE 1 = 1`
	if category := queryParams["category"]; category != "" {
		countQuery += " AND category = '" + category + "'"
	}
	if isPublished := queryParams["is_published"]; isPublished != "" {
		countQuery += " AND is_published = " + isPublished
	} else {
		countQuery += " AND is_published = true"
	}
	if userID := queryParams["user_id"]; userID != "" {
		countQuery += " AND user_id = '" + userID + "'"
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
		query += fmt.Sprintf(" ORDER BY p.created_at DESC LIMIT %d OFFSET %d", limit, offset)
	} else {
		query += fmt.Sprintf(" ORDER BY p.created_at DESC LIMIT %d", limit)
	}

	rows, err := database.DB.Query(context.Background(), query)
	if err != nil {
		return nil, totalPages, err
	}
	defer rows.Close()

	var projects []*models.ProjectResponse
	for rows.Next() {
		var project models.ProjectResponse
		err := rows.Scan(
			&project.ID,
			&project.UserID,
			&project.Title,
			&project.Description,
			&project.Category,
			&project.ProjectURL,
			&project.ImageURL,
			&project.IsPublished,
			&project.VoteCount,
			&project.CreatedAt,
			&project.UpdatedAt,
			&project.UserName,
		)
		if err != nil {
			return nil, totalPages, err
		}
		projects = append(projects, &project)
	}

	return projects, totalPages, nil
}

// PublishProject sets a project's is_published status to true
func PublishProject(projectID int) error {
	query := `
		UPDATE projects
		SET is_published = true, updated_at = NOW()
		WHERE id = $1`

	_, err := database.DB.Exec(context.Background(), query, projectID)
	return err
}

// CheckProjectExists checks if a project exists
func CheckProjectExists(projectID int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1)`
	err := database.DB.QueryRow(context.Background(), query, projectID).Scan(&exists)
	return exists, err
}

// GetProjectsByUser retrieves all projects created by a specific user
func GetProjectsByUser(userID uuid.UUID) ([]*models.ProjectResponse, error) {
	query := `
		SELECT p.id, p.user_id, p.title, p.description, p.category, p.project_url, 
		       p.image_url, p.is_published, p.vote_count, p.created_at, p.updated_at,
		       CONCAT(u.first_name, ' ', u.last_name) as user_name
		FROM projects p
		LEFT JOIN users u ON p.user_id = u.id
		WHERE p.user_id = $1
		ORDER BY p.created_at DESC`

	rows, err := database.DB.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*models.ProjectResponse
	for rows.Next() {
		var project models.ProjectResponse
		err := rows.Scan(
			&project.ID,
			&project.UserID,
			&project.Title,
			&project.Description,
			&project.Category,
			&project.ProjectURL,
			&project.ImageURL,
			&project.IsPublished,
			&project.VoteCount,
			&project.CreatedAt,
			&project.UpdatedAt,
			&project.UserName,
		)
		if err != nil {
			return nil, err
		}
		projects = append(projects, &project)
	}

	return projects, nil
}

// GetProjectOwnerID retrieves the user_id of a project owner
func GetProjectOwnerID(projectID int) (uuid.UUID, error) {
	var ownerID uuid.UUID
	query := `SELECT user_id FROM projects WHERE id = $1`
	err := database.DB.QueryRow(context.Background(), query, projectID).Scan(&ownerID)
	return ownerID, err
}
