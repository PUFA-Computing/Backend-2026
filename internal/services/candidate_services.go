package services

import (
	"Backend/internal/database/app"
	"Backend/internal/models"
	"Backend/pkg/utils"
	"errors"
	"github.com/google/uuid"
)

type CandidateService struct {
}

func NewCandidateService() *CandidateService {
	return &CandidateService{}
}

// CreateCandidate creates a new candidate
func (cs *CandidateService) CreateCandidate(candidate *models.Candidate) error {
	// Validate major
	if !isValidMajor(candidate.Major) {
		return errors.New("invalid major: must be 'information system' or 'informatics'")
	}

	if err := app.CreateCandidate(candidate); err != nil {
		return err
	}

	return nil
}

// UpdateCandidate updates an existing candidate
func (cs *CandidateService) UpdateCandidate(candidateID int, updatedCandidate *models.Candidate) error {
	existingCandidate, err := app.GetCandidateByID(candidateID)
	if err != nil {
		return errors.New("candidate not found")
	}

	// Validate major if it's being updated
	if updatedCandidate.Major != "" && !isValidMajor(updatedCandidate.Major) {
		return errors.New("invalid major: must be 'information system' or 'informatics'")
	}

	// Use reflective update to only update non-zero fields
	utils.ReflectiveUpdate(existingCandidate, updatedCandidate)

	if err := app.UpdateCandidate(candidateID, existingCandidate); err != nil {
		return err
	}

	return nil
}

// DeleteCandidate deletes a candidate
func (cs *CandidateService) DeleteCandidate(candidateID int) error {
	// Check if candidate exists
	exists, err := app.CheckCandidateExists(candidateID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("candidate not found")
	}

	if err := app.DeleteCandidate(candidateID); err != nil {
		return err
	}

	return nil
}

// GetCandidateByID retrieves a candidate by ID
func (cs *CandidateService) GetCandidateByID(candidateID int) (*models.Candidate, error) {
	candidate, err := app.GetCandidateByID(candidateID)
	if err != nil {
		return nil, errors.New("candidate not found")
	}
	return candidate, nil
}

// GetCandidateWithVoteCount retrieves a candidate with their vote count
func (cs *CandidateService) GetCandidateWithVoteCount(candidateID int) (*models.CandidateResponse, error) {
	candidate, err := app.GetCandidateWithVoteCount(candidateID)
	if err != nil {
		return nil, errors.New("candidate not found")
	}
	return candidate, nil
}

// ListCandidates retrieves all candidates with optional filters
func (cs *CandidateService) ListCandidates(queryParams map[string]string) ([]*models.CandidateResponse, int, error) {
	// Validate major filter if provided
	if major := queryParams["major"]; major != "" && !isValidMajor(major) {
		return nil, 0, errors.New("invalid major filter: must be 'information system' or 'informatics'")
	}

	candidates, totalPages, err := app.ListCandidates(queryParams)
	if err != nil {
		return nil, 0, err
	}
	return candidates, totalPages, nil
}

// GetCandidatesByMajor retrieves all candidates for a specific major
func (cs *CandidateService) GetCandidatesByMajor(major string) ([]*models.CandidateResponse, error) {
	// Validate major
	if !isValidMajor(major) {
		return nil, errors.New("invalid major: must be 'information system' or 'informatics'")
	}

	candidates, err := app.GetCandidatesByMajor(major)
	if err != nil {
		return nil, err
	}
	return candidates, nil
}

// isValidMajor checks if the major is valid
func isValidMajor(major string) bool {
	return major == "information system" || major == "informatics"
}

// ValidateCandidateOwnership checks if a user owns a candidate
func (cs *CandidateService) ValidateCandidateOwnership(candidateID int, userID uuid.UUID) (bool, error) {
	candidate, err := app.GetCandidateByID(candidateID)
	if err != nil {
		return false, err
	}

	if candidate.UserID == nil {
		return false, nil
	}

	return *candidate.UserID == userID, nil
}
