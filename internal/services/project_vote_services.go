package services

import (
	"Backend/internal/database/app"
	"Backend/internal/models"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"strings"
)

// TeamInfoRequest represents team information for validation
type TeamInfoRequest struct {
	ProjectMembers   []string
	LinkedInProfiles []string
	Major            string
	Batch            int
}

// ValidateTeamInfo validates the team information
func ValidateTeamInfo(req *TeamInfoRequest) error {
	// Check members count
	if len(req.ProjectMembers) == 0 {
		return errors.New("at least one project member is required")
	}
	if len(req.ProjectMembers) > 10 {
		return errors.New("maximum 10 project members allowed")
	}

	// Check for empty member names
	for i, member := range req.ProjectMembers {
		if strings.TrimSpace(member) == "" {
			return fmt.Errorf("project member at index %d cannot be empty", i)
		}
	}

	// Check LinkedIn profiles match members
	if len(req.LinkedInProfiles) != len(req.ProjectMembers) {
		return errors.New("number of LinkedIn profiles must match number of project members")
	}

	// Validate LinkedIn URLs
	for i, url := range req.LinkedInProfiles {
		if !isValidLinkedInURL(url) {
			return fmt.Errorf("invalid LinkedIn URL at index %d: %s", i, url)
		}
	}

	return nil
}

// isValidLinkedInURL checks if a URL is a valid LinkedIn profile URL
func isValidLinkedInURL(url string) bool {
	url = strings.TrimSpace(url)
	if url == "" {
		return false
	}

	// Basic LinkedIn URL validation
	// Accept both http and https, with or without www
	validPrefixes := []string{
		"https://www.linkedin.com/in/",
		"https://linkedin.com/in/",
		"http://www.linkedin.com/in/",
		"http://linkedin.com/in/",
	}

	for _, prefix := range validPrefixes {
		if strings.HasPrefix(strings.ToLower(url), prefix) {
			// Check if there's something after the prefix
			if len(url) > len(prefix) {
				return true
			}
		}
	}

	return false
}

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
