package models

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID                     uuid.UUID  `pg:"type:uuid" json:"id"`
	Username               string     `json:"username"`
	Password               string     `json:"password"`
	FirstName              string     `json:"first_name"`
	MiddleName             *string    `json:"middle_name"`
	LastName               string     `json:"last_name"`
	Email                  string     `json:"email"`
	StudentID              string     `json:"student_id"`
	Major                  string     `json:"major"`
	ProfilePicture         string     `json:"profile_picture"`
	DateOfBirth            *time.Time `json:"date_of_birth"`
	RoleID                 int        `json:"role_id"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
	Year                   string     `json:"year"`
	EmailVerified          bool       `json:"email_verified"`
	EmailVerificationToken string     `json:"email_verification_token"`
	PasswordResetToken     string     `json:"password_reset_token"`
	PasswordResetExpires   *time.Time `json:"password_reset_expires"`
	StudentIDVerified      bool       `json:"student_id_verified"`
	StudentIDVerification  *string    `json:"student_id_verification"`
	InstitutionName        *string    `json:"institution_name"`
	Gender                 string     `json:"gender"`
	AdditionalNotes        *string    `json:"additional_notes"`
	FilePath               *string    `json:"file_path"`
	TwoFAEnabled           bool       `json:"twofa_enabled"`
	TwoFAImage             *string    `json:"twofa_image"`
	TwoFASecret            *string    `json:"twofa_secret"`
	// Google OAuth + account linking ─────────────────────────────────────────
	// GoogleSub is Google's stable subject ID (`sub` claim). Used to look up
	// the row on subsequent Google sign-ins so renaming a Google email still
	// finds the right account.
	GoogleSub *string `json:"google_sub,omitempty"`
	// AuthProvider tells us which login methods are wired up: "password",
	// "google", or "both". Driven by whether google_sub / password is set.
	AuthProvider string `json:"auth_provider"`
	// ProfileCompleted is false only for fresh Google sign-ups that still
	// need to enter Student ID + batch on the complete-profile screen.
	ProfileCompleted bool `json:"profile_completed"`
}

// Pre-defined role IDs that match migrations/000001_roles.up.sql.
const (
	RoleAdmin      = 1
	RoleComputizen = 2
	RolePufaCS     = 3
	RolePumaIT     = 4
	RolePumaIS     = 5
	RoleGuest      = 6
)
