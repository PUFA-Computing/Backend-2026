package models

import (
	"time"
)

// ── DB Models ──────────────────────────────────────────

type CompregenWhitelistResponse struct {
    Whitelist []*CompregenEligibleCandidate `json:"whitelist"`
}

type CompregenAddWhitelistRequest struct {
    StudentID   string `json:"student_id" binding:"required"`
    FullName    string `json:"full_name" binding:"required"`
    CampusEmail string `json:"campus_email" binding:"required"`
    Major       string `json:"major" binding:"required"`
}

type CompregenEligibleCandidate struct {
	ID           string    `json:"id"`
	StudentID    string    `json:"student_id"`
	FullName     string    `json:"full_name"`
	CampusEmail  string    `json:"campus_email"`
	Major        string    `json:"major"`
	CreatedAt    time.Time `json:"created_at"`
}

type CompregenInviteLink struct {
	ID        string     `json:"id"`
	Token     string     `json:"token"`
	Status    string     `json:"status"`
	CreatedBy string     `json:"created_by"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at"`
}

type CompregenRegistration struct {
	ID              string    `json:"id"`
	InviteLinkID    string    `json:"invite_link_id"`
	CabinetName     string    `json:"cabinet_name"`
	ConsentAccepted bool      `json:"consent_accepted"`
	Status          string    `json:"status"`
	SubmittedAt     time.Time `json:"submitted_at"`
}

type CompregenRegistrationMember struct {
	ID             string  `json:"id"`
	RegistrationID string  `json:"registration_id"`
	Role           string  `json:"role"`
	FullName       string  `json:"full_name"`
	StudentID      string  `json:"student_id"`
	Major          string  `json:"major"`
	PhoneNumber    string  `json:"phone_number"`
	PhotoUploadID  *string `json:"photo_upload_id"`
}

type CompregenPhotoUpload struct {
	ID                   string    `json:"id"`
	RegistrationMemberID *string   `json:"registration_member_id"`
	StorageKey           string    `json:"storage_key"`
	MimeType             string    `json:"mime_type"`
	SizeBytes            int       `json:"size_bytes"`
	UploadedAt           time.Time `json:"uploaded_at"`
}

// ── Request / Response shapes ──────────────────────────

type CompregenVerifyRequest struct {
	Token       string `json:"token" binding:"required"`
	StudentID   string `json:"student_id" binding:"required"`
	CampusEmail string `json:"campus_email" binding:"required"`
}

type CompregenVerifyResponse struct {
	Verified         bool    `json:"verified"`
	Reason           *string `json:"reason,omitempty"`
	AttemptsRemaining *int   `json:"attempts_remaining,omitempty"`
}

type CompregenMemberInput struct {
    FullName      string `json:"full_name" binding:"required"`
    StudentID     string `json:"student_id" binding:"required"`
    Major         string `json:"major" binding:"required"`
    PhoneNumber   string `json:"phone_number" binding:"required"`
    Nationality   string `json:"nationality" binding:"required"`
    PhotoUploadID string `json:"photo_upload_id" binding:"required"`
}

type CompregenRegisterRequest struct {
	Token           string                          `json:"token" binding:"required"`
	CabinetName     string                          `json:"cabinet_name"`
	ConsentAccepted bool                            `json:"consent_accepted" binding:"required"`
	Members         map[string]CompregenMemberInput `json:"members" binding:"required"`
}

type CompregenMemberResponse struct {
	FullName      string  `json:"full_name"`
	StudentID     string  `json:"student_id"`
	Major         string  `json:"major"`
	PhoneNumber   string  `json:"phone_number"`
	Nationality   string  `json:"nationality"`
	PhotoUploadID *string `json:"photo_upload_id"`
}

type CompregenRegistrationResponse struct {
	ID          string                             `json:"id"`
	CabinetName string                             `json:"cabinet_name"`
	SubmittedAt time.Time                          `json:"submitted_at"`
	Members     map[string]CompregenMemberResponse `json:"members"`
}