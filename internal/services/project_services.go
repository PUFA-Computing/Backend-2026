package services

import (
	"Backend/internal/database/app"
	"Backend/internal/models"
	"errors"
	"github.com/google/uuid"
)

type ProjectService struct {
}

func NewProjectService() *ProjectService {
	return &ProjectService{}
}

// CreateProject creates a new project
func (ps *ProjectService) CreateProject(project *models.Project) error {
	// Validate required fields
	if project.Title == "" {
		return errors.New("title is required")
	}
	if project.Description == "" {
		return errors.New("description is required")
	}
	if project.ImageURL == "" {
		return errors.New("image URL is required")
	}

	// Set default values
	project.IsPublished = false
	project.VoteCount = 0

	if err := app.CreateProject(project); err != nil {
		return err
	}

	return nil
}

// UpdateProject updates an existing project
func (ps *ProjectService) UpdateProject(projectID int, userID uuid.UUID, project *models.Project) error {
	// Check if project exists
	existingProject, err := app.GetProjectByID(projectID)
	if err != nil {
		return errors.New("project not found")
	}

	// Check if user is the owner
	if existingProject.UserID != userID {
		return errors.New("you are not authorized to update this project")
	}

	if err := app.UpdateProject(projectID, project); err != nil {
		return err
	}

	return nil
}

// DeleteProject deletes a project
func (ps *ProjectService) DeleteProject(projectID int, userID uuid.UUID, isAdmin bool) error {
	// Check if project exists
	existingProject, err := app.GetProjectByID(projectID)
	if err != nil {
		return errors.New("project not found")
	}

	// Check if user is the owner or admin
	if existingProject.UserID != userID && !isAdmin {
		return errors.New("you are not authorized to delete this project")
	}

	if err := app.DeleteProject(projectID); err != nil {
		return err
	}

	return nil
}

// GetProjectByID retrieves a project by ID
func (ps *ProjectService) GetProjectByID(projectID int) (*models.Project, error) {
	project, err := app.GetProjectByID(projectID)
	if err != nil {
		return nil, errors.New("project not found")
	}
	return project, nil
}

// GetProjectWithVoteCount retrieves a project with vote count
func (ps *ProjectService) GetProjectWithVoteCount(projectID int) (*models.ProjectResponse, error) {
	project, err := app.GetProjectWithVoteCount(projectID)
	if err != nil {
		return nil, errors.New("project not found")
	}
	return project, nil
}

// ListProjects retrieves all projects with filters
func (ps *ProjectService) ListProjects(queryParams map[string]string) ([]*models.ProjectResponse, int, error) {
	projects, totalPages, err := app.ListProjects(queryParams)
	if err != nil {
		return nil, 0, err
	}
	return projects, totalPages, nil
}

// PublishProject publishes a project (admin only)
func (ps *ProjectService) PublishProject(projectID int) error {
	// Check if project exists
	exists, err := app.CheckProjectExists(projectID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("project not found")
	}

	if err := app.PublishProject(projectID); err != nil {
		return err
	}

	return nil
}

// GetProjectsByUser retrieves all projects by a user
func (ps *ProjectService) GetProjectsByUser(userID uuid.UUID) ([]*models.ProjectResponse, error) {
	projects, err := app.GetProjectsByUser(userID)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

// CheckProjectOwnership checks if a user owns a project
func (ps *ProjectService) CheckProjectOwnership(projectID int, userID uuid.UUID) (bool, error) {
	ownerID, err := app.GetProjectOwnerID(projectID)
	if err != nil {
		return false, err
	}
	return ownerID == userID, nil
}
