package services

import (
	dbApp "Backend/internal/database/app"
	"Backend/internal/models"
	"crypto/rand"
	"encoding/hex"
	"errors"
)

type CompregenService struct{}

func NewCompregenService() *CompregenService {
	return &CompregenService{}
}

func (s *CompregenService) GetLinkStatus(token string) string {
	link, err := dbApp.GetInviteLinkByToken(token)
	if err != nil {
		return "not_found"
	}
	return link.Status
}

func (s *CompregenService) Verify(token, studentID, email string) (*models.CompregenVerifyResponse, *models.CompregenInviteLink, error) {
	link, err := dbApp.GetInviteLinkByToken(token)
	if err != nil || link.Status != "active" {
		return nil, nil, errors.New("invalid_token")
	}

	locked, remaining, err := dbApp.CheckRateLimit(link.ID)
	if err != nil {
		return nil, nil, err
	}
	if locked {
		return nil, nil, errors.New("rate_limited")
	}

	eligible, err := dbApp.CheckEligibleCandidateByStudentID(studentID)
	if err != nil {
		return nil, nil, err
	}

	_ = dbApp.RecordVerifyAttempt(link.ID, studentID, email, eligible)

	if !eligible {
		reason := "not_in_whitelist"
		return &models.CompregenVerifyResponse{
			Verified:          false,
			Reason:            &reason,
			AttemptsRemaining: &remaining,
		}, link, nil
	}

	return &models.CompregenVerifyResponse{Verified: true}, link, nil
}

func (s *CompregenService) Register(token, cabinetName string, members map[string]models.CompregenMemberInput) (string, error) {
	link, err := dbApp.GetInviteLinkByToken(token)
	if err != nil || link.Status != "active" {
		return "", errors.New("link_already_used")
	}

	exists, err := dbApp.RegistrationExistsForLink(link.ID)
	if err != nil {
		return "", err
	}
	if exists {
		return "", errors.New("link_already_used")
	}

	// Validate members are whitelisted (VCP2 exempt)
	for role, m := range members {
		if role == "vcp2" {
			continue // VCP2 tidak perlu di whitelist
		}
		ok, err := dbApp.CheckEligibleCandidateByStudentID(m.StudentID)
		if err != nil {
			return "", err
		}
		if !ok {
			return "", errors.New("not_whitelisted:" + role)
		}
	}

	registrationID, err := dbApp.CreateRegistration(link.ID, cabinetName, members)
	if err != nil {
		return "", err
	}

	_ = dbApp.MarkInviteLinkUsed(link.ID)
	return registrationID, nil
}

func (s *CompregenService) GenerateInviteLink() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}