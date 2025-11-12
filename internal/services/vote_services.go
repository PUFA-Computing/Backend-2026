package services

import (
	"Backend/internal/database/app"
	"Backend/internal/models"
	"errors"
	"github.com/google/uuid"
)

type VoteService struct {
}

func NewVoteService() *VoteService {
	return &VoteService{}
}

// CastVote allows a user to vote for a candidate
func (vs *VoteService) CastVote(voterID uuid.UUID, candidateID int) (*models.Vote, error) {
	// Check if user's year is 2025
	voterYear, err := app.GetUserYear(voterID)
	if err != nil {
		return nil, errors.New("unable to verify voter information")
	}
	if voterYear != "2025" {
		return nil, errors.New("only users with year 2025 are eligible to vote")
	}

	// Check if user has already voted
	hasVoted, err := app.CheckUserHasVoted(voterID)
	if err != nil {
		return nil, err
	}
	if hasVoted {
		return nil, errors.New("you have already voted")
	}

	// Check if candidate exists
	exists, err := app.CheckCandidateExists(candidateID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("candidate not found")
	}

	// Validate that voter and candidate have the same major
	voterMajor, err := app.GetUserMajor(voterID)
	if err != nil {
		return nil, errors.New("unable to verify voter information")
	}

	candidateMajor, err := app.GetCandidateMajor(candidateID)
	if err != nil {
		return nil, errors.New("unable to verify candidate information")
	}

	if voterMajor != candidateMajor {
		return nil, errors.New("you can only vote for candidates from your major")
	}

	// Create the vote
	vote := &models.Vote{
		VoterID:     voterID,
		CandidateID: candidateID,
	}

	if err := app.CastVote(vote); err != nil {
		return nil, err
	}

	return vote, nil
}

// GetVoteByVoterID retrieves a vote by voter ID
func (vs *VoteService) GetVoteByVoterID(voterID uuid.UUID) (*models.Vote, error) {
	vote, err := app.GetVoteByVoterID(voterID)
	if err != nil {
		return nil, errors.New("vote not found")
	}
	return vote, nil
}

// GetVoteStatus checks if a user has voted and returns their vote status
func (vs *VoteService) GetVoteStatus(voterID uuid.UUID) (*models.VoteStatusResponse, error) {
	status, err := app.GetVoteStatus(voterID)
	if err != nil {
		return nil, err
	}
	return status, nil
}

// ListVotes retrieves all votes (admin only)
func (vs *VoteService) ListVotes(queryParams map[string]string) ([]*models.VoteResponse, int, error) {
	votes, totalPages, err := app.ListVotes(queryParams)
	if err != nil {
		return nil, 0, err
	}
	return votes, totalPages, nil
}

// GetVoteCountByCandidate returns the number of votes for a candidate
func (vs *VoteService) GetVoteCountByCandidate(candidateID int) (int, error) {
	count, err := app.GetVoteCountByCandidate(candidateID)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// DeleteVote removes a vote (admin only)
func (vs *VoteService) DeleteVote(voteID int) error {
	if err := app.DeleteVote(voteID); err != nil {
		return err
	}
	return nil
}

// GetVotesByCandidateID retrieves all votes for a specific candidate
func (vs *VoteService) GetVotesByCandidateID(candidateID int) ([]*models.VoteResponse, error) {
	// Check if candidate exists
	exists, err := app.CheckCandidateExists(candidateID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("candidate not found")
	}

	votes, err := app.GetVotesByCandidateID(candidateID)
	if err != nil {
		return nil, err
	}
	return votes, nil
}

// CheckUserCanVote validates if a user can vote (hasn't voted yet, has valid major, and year is 2025)
func (vs *VoteService) CheckUserCanVote(voterID uuid.UUID) (bool, string, error) {
	// Check if user's year is 2025
	voterYear, err := app.GetUserYear(voterID)
	if err != nil {
		return false, "unable to verify user information", err
	}
	if voterYear != "2025" {
		return false, "only users with year 2025 are eligible to vote", nil
	}

	// Check if user has already voted
	hasVoted, err := app.CheckUserHasVoted(voterID)
	if err != nil {
		return false, "unable to verify vote status", err
	}
	if hasVoted {
		return false, "you have already voted", nil
	}

	// Check if user has a valid major
	voterMajor, err := app.GetUserMajor(voterID)
	if err != nil {
		return false, "unable to verify user information", err
	}

	if voterMajor != "information system" && voterMajor != "informatics" {
		return false, "invalid major", nil
	}

	return true, "", nil
}
