package services

import (
	"Backend/internal/database/app"
	"Backend/internal/models"
	"errors"
	"github.com/google/uuid"
)

type ProjectVoteService struct {
}

func NewProjectVoteService() *ProjectVoteService {
	return &ProjectVoteService{}
}

// VoteProject allows a user to vote for a project
func (pvs *ProjectVoteService) VoteProject(userID uuid.UUID, projectID int) (*models.ProjectVote, error) {
	// Check if project exists
	exists, err := app.CheckProjectExists(projectID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("project not found")
	}

	// Check if project is published
	project, err := app.GetProjectByID(projectID)
	if err != nil {
		return nil, errors.New("project not found")
	}
	if !project.IsPublished {
		return nil, errors.New("you can only vote for published projects")
	}

	// Check if user has already voted for this project
	hasVoted, err := app.CheckUserHasVotedProject(projectID, userID)
	if err != nil {
		return nil, err
	}
	if hasVoted {
		return nil, errors.New("you have already voted for this project")
	}

	// Create the vote
	vote := &models.ProjectVote{
		ProjectID: projectID,
		UserID:    userID,
	}

	if err := app.VoteProject(vote); err != nil {
		return nil, err
	}

	return vote, nil
}

// UnvoteProject allows a user to remove their vote from a project
func (pvs *ProjectVoteService) UnvoteProject(userID uuid.UUID, projectID int) error {
	// Check if project exists
	exists, err := app.CheckProjectExists(projectID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("project not found")
	}

	// Check if user has voted for this project
	hasVoted, err := app.CheckUserHasVotedProject(projectID, userID)
	if err != nil {
		return err
	}
	if !hasVoted {
		return errors.New("you have not voted for this project")
	}

	if err := app.UnvoteProject(projectID, userID); err != nil {
		return err
	}

	return nil
}

// GetProjectVotesByUser retrieves all votes cast by a user
func (pvs *ProjectVoteService) GetProjectVotesByUser(userID uuid.UUID) ([]*models.ProjectVoteResponse, error) {
	votes, err := app.GetProjectVotesByUser(userID)
	if err != nil {
		return nil, err
	}
	return votes, nil
}

// GetProjectVotesByProject retrieves all votes for a specific project
func (pvs *ProjectVoteService) GetProjectVotesByProject(projectID int) ([]*models.ProjectVoteResponse, error) {
	// Check if project exists
	exists, err := app.CheckProjectExists(projectID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("project not found")
	}

	votes, err := app.GetProjectVotesByProject(projectID)
	if err != nil {
		return nil, err
	}
	return votes, nil
}

// GetProjectVoteCount returns the number of votes for a project
func (pvs *ProjectVoteService) GetProjectVoteCount(projectID int) (int, error) {
	count, err := app.GetProjectVoteCount(projectID)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// CheckUserHasVotedProject checks if a user has voted for a project
func (pvs *ProjectVoteService) CheckUserHasVotedProject(projectID int, userID uuid.UUID) (*models.ProjectVoteStatusResponse, error) {
	status, err := app.GetProjectVoteStatus(projectID, userID)
	if err != nil {
		return nil, err
	}
	return status, nil
}
