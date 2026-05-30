package app

import (
	"Backend/internal/database"
	"Backend/internal/models"
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"log"
)

func CreateUser(user *models.User) error {

	// email_verified is set to TRUE directly since Hunter.io already validates the
	// @student.president.ac.id domain during registration – no SMTP step needed.
	query := `
		INSERT INTO users (id, username, password, first_name, middle_name, last_name, email, student_id, major, year, role_id, email_verification_token, institution_name, gender, email_verified)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, TRUE)`
	_, err := database.DB.Exec(
		context.Background(),
		query,
		user.ID, user.Username, user.Password, user.FirstName, user.MiddleName, user.LastName, user.Email,
		user.StudentID, user.Major, user.Year, user.RoleID, user.EmailVerificationToken, user.InstitutionName, user.Gender,
	)
	if err != nil {
		log.Printf("Error during query execution or scanning: %v", err)
		return err
	}
	return nil
}

func AuthenticateUser(usernameOrEmail string) (*models.User, error) {
	// Use an absolute minimal query with only the essential fields needed for authentication
	query := `
		SELECT id, username, password
		FROM users
		WHERE username = $1 OR email = $1`

	log.Printf("Executing minimal login query for user: %s", usernameOrEmail)
	
	// Create a minimal user object with just the fields needed for authentication
	var user models.User
	var userIDString string
	
	err := database.DB.QueryRow(
		context.Background(),
		query,
		usernameOrEmail,
	).Scan(
		&userIDString,
		&user.Username,
		&user.Password,
	)

	if errors.Is(err, sql.ErrNoRows) {
		log.Println("No user found with username or email:", usernameOrEmail)
		return nil, err
	} else if err != nil {
		log.Println("Error during query execution or scanning:", err)
		return nil, err
	}

	// Parse the UUID
	userID, err := uuid.Parse(userIDString)
	if err != nil {
		return nil, err
	}
	
	user.ID = userID
	
	// Set default values for fields that might be needed but aren't critical for login
	user.FirstName = ""
	user.LastName = ""
	// NOTE: Do NOT set user.Email = usernameOrEmail here, because if the user
	// logged in with their username (not email), this would set a username string
	// into the Email field, causing verification emails to be sent to the wrong address.
	// The correct email will be fetched from the DB in the query below.
	user.Email = ""
	user.StudentID = ""
	user.Major = ""
	user.Year = ""
	user.Gender = ""
	user.ProfilePicture = ""
	user.EmailVerificationToken = ""
	user.PasswordResetToken = ""
	user.EmailVerified = false
	user.TwoFAEnabled = false
	
	// Now fetch additional user details if needed
	detailsQuery := `
		SELECT email, email_verified, twofa_enabled
		FROM users
		WHERE id = $1`
	
	err = database.DB.QueryRow(
		context.Background(),
		detailsQuery,
		userID,
	).Scan(
		&user.Email,
		&user.EmailVerified,
		&user.TwoFAEnabled,
	)
	
	if err != nil {
		log.Printf("Warning: Could not fetch additional user details: %v", err)
		// Continue anyway, as we have the essential fields for authentication
	}
	
	return &user, nil
}

func IsEmailVerified(email string) (bool, error) {
	var verified bool
	query := `
		SELECT email_verified
		FROM users
		WHERE email = $1 OR username = $1`
	err := database.DB.QueryRow(
		context.Background(),
		query,
		email,
	).Scan(&verified)
	if err != nil {
		log.Printf("Error during query execution or scanning: %v", err)
		return false, err
	}
	return verified, nil
}

func IsTokenVerificationEmailExists(token string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM users
			WHERE email_verification_token = $1
		)`
	err := database.DB.QueryRow(
		context.Background(),
		query,
		token,
	).Scan(&exists)
	if err != nil {
		log.Printf("Error during query execution or scanning: %v", err)
		return false, err
	}
	return exists, nil
}

func UpdateEmailVerificationToken(email, token string) error {
	query := `
		UPDATE users
		SET email_verification_token = $1
		WHERE email = $2`
	_, err := database.DB.Exec(
		context.Background(),
		query,
		token,
		email,
	)
	if err != nil {
		log.Printf("Error during query execution or scanning: %v", err)
		return err
	}
	return nil

}

// VerifyEmail updates the email_verified field in the users table and return error if verification token is invalid
func VerifyEmail(token string) error {
	query := `
		UPDATE users
		SET email_verified = TRUE
		WHERE email_verification_token = $1`
	_, err := database.DB.Exec(
		context.Background(),
		query,
		token,
	)
	if err != nil {
		log.Printf("Error during query execution or scanning: %v", err)
		return err
	}
	return nil
}

// ── Google OAuth helpers ────────────────────────────────────────────────────

// GetUserByGoogleSub returns the user linked to the given Google subject ID
// (nil, nil) on miss.
func GetUserByGoogleSub(sub string) (*models.User, error) {
	var user models.User
	var userIDStr string
	var roleID int
	var emailVerified, profileCompleted sql.NullBool
	var studentID, major, year sql.NullString

	query := `
		SELECT id, username, email, COALESCE(first_name,''), COALESCE(last_name,''),
		       role_id, email_verified, profile_completed,
		       student_id, major, year
		FROM users
		WHERE google_sub = $1`

	err := database.DB.QueryRow(context.Background(), query, sub).Scan(
		&userIDStr, &user.Username, &user.Email, &user.FirstName, &user.LastName,
		&roleID, &emailVerified, &profileCompleted,
		&studentID, &major, &year,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if user.ID, err = uuid.Parse(userIDStr); err != nil {
		return nil, err
	}
	user.RoleID = roleID
	if emailVerified.Valid {
		user.EmailVerified = emailVerified.Bool
	}
	if profileCompleted.Valid {
		user.ProfileCompleted = profileCompleted.Bool
	}
	if studentID.Valid {
		user.StudentID = studentID.String
	}
	if major.Valid {
		user.Major = major.String
	}
	if year.Valid {
		user.Year = year.String
	}
	return &user, nil
}

// LinkGoogleSub attaches a Google sub to an existing user row and flips
// auth_provider to "both" (a password user adopting Google) or "google"
// (when no password is stored). Idempotent.
func LinkGoogleSub(userID uuid.UUID, sub string) error {
	query := `
		UPDATE users
		SET google_sub = $1,
		    auth_provider = CASE
		        WHEN password IS NULL OR password = '' THEN 'google'
		        ELSE 'both'
		    END,
		    email_verified = TRUE,
		    updated_at = NOW()
		WHERE id = $2`
	_, err := database.DB.Exec(context.Background(), query, sub, userID)
	return err
}

// CreateGoogleUser inserts a row for a brand-new Google sign-up. Password is
// left NULL. StudentID/Major/Year may be empty for the second-stage form.
func CreateGoogleUser(user *models.User) error {
	query := `
		INSERT INTO users (
		    id, username, first_name, last_name, email,
		    student_id, major, year, role_id,
		    institution_name, gender,
		    email_verified, profile_completed,
		    google_sub, auth_provider
		) VALUES (
		    $1, $2, $3, $4, $5,
		    $6, $7, $8, $9,
		    $10, $11,
		    TRUE, $12,
		    $13, 'google'
		)`
	_, err := database.DB.Exec(
		context.Background(),
		query,
		user.ID, user.Username, user.FirstName, user.LastName, user.Email,
		user.StudentID, user.Major, user.Year, user.RoleID,
		user.InstitutionName, user.Gender,
		user.ProfileCompleted,
		user.GoogleSub,
	)
	if err != nil {
		log.Printf("CreateGoogleUser insert error: %v", err)
	}
	return err
}

// CompleteGoogleProfile fills in Student ID + batch on a half-registered
// Google account and promotes the role accordingly (Computizen for CS prefixes,
// Guest otherwise). Performed in a single statement to stay atomic.
func CompleteGoogleProfile(userID uuid.UUID, studentID, major, year string, roleID int) error {
	query := `
		UPDATE users
		SET student_id = $1,
		    major = $2,
		    year = $3,
		    role_id = $4,
		    profile_completed = TRUE,
		    updated_at = NOW()
		WHERE id = $5`
	_, err := database.DB.Exec(context.Background(), query, studentID, major, year, roleID, userID)
	return err
}

func GetPasswordResetToken(userID uuid.UUID) (string, error) {
	var token string
	query := `
		SELECT password_reset_token
		FROM users
		WHERE id = $1`
	err := database.DB.QueryRow(
		context.Background(),
		query,
		userID,
	).Scan(&token)
	if err != nil {
		log.Printf("Error during query execution or scanning: %v", err)
		return "", err
	}
	return token, nil
}

func UpdatePassword(userID uuid.UUID, newPassword string) error {
	query := `
		UPDATE users
		SET password = $1
		WHERE id = $2`
	_, err := database.DB.Exec(
		context.Background(),
		query,
		newPassword,
		userID,
	)
	if err != nil {
		log.Printf("Error during query execution or scanning: %v", err)
		return err
	}
	return nil
}
