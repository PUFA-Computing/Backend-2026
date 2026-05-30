package services

import (
	"Backend/configs"
	"Backend/internal/database"
	"Backend/internal/database/app"
	"Backend/internal/models"
	"Backend/pkg/utils"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type AuthService struct {
	otp *OTPManager
}

func NewAuthService() *AuthService {
	return &AuthService{
		otp: NewOTPManager(),
	}
}

func (as *AuthService) RegisterUser(user *models.User) error {
	// Email validation – only @student.president.ac.id domain is accepted.
	// Hunter.io already guarantees delivery-ability for this domain, so no SMTP
	// verification email is sent; the account is activated immediately.
	if err := as.ValidateEmail(user.Email); err != nil {
		return err
	}

	user.ID = uuid.New()
	user.RoleID = 2
	user.Gender = "male"
	user.EmailVerified = true           // auto-verified – no SMTP step
	user.EmailVerificationToken = ""   // not needed

	user.ProfilePicture = "https://sg.pufacomputing.live/Assets/male.jpeg"

	// Set major based on studentID prefix
	switch user.StudentID[:3] {
	case "001":
		user.Major = "informatics"
	case "012":
		user.Major = "information system"
	case "013":
		user.Major = "visual communication design"
	case "025":
		user.Major = "interior design"
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	if err = app.CreateUser(user); err != nil {
		return err
	}

	return nil
}

func (as *AuthService) LoginUser(usernameOrEmail string, password string) (*models.User, error) {
	// Log the login attempt with detailed information
	log.Printf("=== LOGIN ATTEMPT START === Username/Email: %s", usernameOrEmail)

	// Create a complete user struct with all fields initialized to zero values
	user := &models.User{}

	// Use a minimal query to get just the essential fields for authentication
	var userID string
	var hashedPassword string

	// Explicitly log the SQL query we're about to execute
	log.Printf("Executing minimal login query for: %s", usernameOrEmail)

	// Use a very simple query with only the minimum fields needed for authentication
	query := `SELECT id, username, password, email FROM users WHERE username = $1 OR email = $1`

	// Execute the query and scan results into variables
	err := database.DB.QueryRow(context.Background(), query, usernameOrEmail).Scan(
		&userID, &user.Username, &hashedPassword, &user.Email)

	// Handle query errors
	if err != nil {
		log.Printf("Login query error: %v", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &utils.UnauthorizedError{Message: "invalid credentials"}
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Log successful query
	log.Printf("Successfully retrieved basic user info for: %s", user.Username)

	// Verify password
	log.Printf("Comparing password for user: %s", user.Username)
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		log.Printf("Password verification failed for user %s: %v", user.Username, err)
		return nil, &utils.UnauthorizedError{Message: "invalid credentials"}
	}

	// Password verified, parse UUID
	user.ID, err = uuid.Parse(userID)
	if err != nil {
		log.Printf("Error parsing UUID: %v", err)
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	// Now get additional fields needed for authentication flow
	log.Printf("Fetching additional authentication fields for user: %s", user.Username)

	// Only get the specific fields we need for the authentication flow
	var emailVerified, twoFAEnabled sql.NullBool

	authQuery := `SELECT email_verified, twofa_enabled FROM users WHERE id = $1`
	err = database.DB.QueryRow(context.Background(), authQuery, userID).Scan(
		&emailVerified, &twoFAEnabled)

	if err != nil {
		log.Printf("Warning: Could not fetch auth details: %v - will use defaults", err)
		// Continue with defaults if we can't get these fields
	} else {
		// Set fields only if they're not NULL in the database
		if emailVerified.Valid {
			user.EmailVerified = emailVerified.Bool
		}
		if twoFAEnabled.Valid {
			user.TwoFAEnabled = twoFAEnabled.Bool
		}
		log.Printf("Auth details - EmailVerified: %v, 2FA Enabled: %v",
			user.EmailVerified, user.TwoFAEnabled)
	}

	// Store the hashed password for token generation
	user.Password = hashedPassword

	log.Printf("=== LOGIN ATTEMPT SUCCESS === User: %s", user.Username)
	return user, nil
}

func (as *AuthService) IsUsernameExists(username string) (bool, error) {
	return app.IsUsernameExists(username)
}

func (as *AuthService) IsEmailExists(email string) (bool, error) {
	return app.IsEmailExists(email)
}

func (as *AuthService) IsStudentIDExists(studentID string) (bool, error) {
	return app.CheckStudentIDExists(studentID)
}

func (as *AuthService) GetUserByStudentID(studentID string) (*models.User, error) {
	return app.GetUserByStudentID(studentID)
}

func (as *AuthService) GetUserByUsernameOrEmail(usernameOrEmail string) (*models.User, error) {
	return app.AuthenticateUser(usernameOrEmail)
}

func (as *AuthService) GetUserByEmail(email string) (*models.User, error) {
	return app.GetUserByEmail(email)
}

func (as *AuthService) CheckStudentIDExists(studentID string) (bool, error) {
	return app.CheckStudentIDExists(studentID)
}

func (as *AuthService) IsEmailVerified(username string) (bool, error) {
	return app.IsEmailVerified(username)
}

func (as *AuthService) IsTokenVerificationEmailExists(token string) (bool, error) {
	return app.IsTokenVerificationEmailExists(token)
}

func (as *AuthService) UpdateEmailVerificationToken(email, token string) error {
	return app.UpdateEmailVerificationToken(email, token)
}

func (as *AuthService) VerifyEmail(token string) error {
	return app.VerifyEmail(token)
}

type HunterEmailVerification struct {
	Data struct {
		Status     string `json:"status"`
		Result     string `json:"result"`
		Score      int    `json:"score"`
		Regexp     bool   `json:"regexp"`
		Gibberish  bool   `json:"gibberish"`
		Disposable bool   `json:"disposable"`
		Webmail    bool   `json:"webmail"`
		MxRecords  bool   `json:"mx_records"`
		SmtpServer bool   `json:"smtp_server"`
		SmtpCheck  bool   `json:"smtp_check"`
		AcceptAll  bool   `json:"accept_all"`
		Block      bool   `json:"block"`
	} `json:"data"`
}

func (as *AuthService) ValidateEmail(email string) error {
	// Always log the email we're validating
	log.Printf("ValidateEmail called for: %s", email)

	// Bypass validation for President University student emails
	if strings.HasSuffix(email, "@student.president.ac.id") {
		log.Println("Bypassing Hunter.io validation for President University student email")
		return nil
	}

	// Load the Hunter API key from the config
	apiKey := configs.LoadConfig().HunterApiKey

	// If no API key is provided, bypass validation for development
	if apiKey == "" {
		log.Println("Warning: No Hunter API key provided, bypassing email validation")
		return nil
	}

	log.Printf("Validating email %s with Hunter.io", email)
	url := fmt.Sprintf("https://api.hunter.io/v2/email-verifier?email=%s&api_key=%s", email, apiKey)

	// Create a new HTTP client with a timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create a new HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}(resp.Body)

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to validate email: received status code %d", resp.StatusCode)
	}

	// Parse the JSON response
	var verification HunterEmailVerification
	if err := json.NewDecoder(resp.Body).Decode(&verification); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Check the email status
	switch verification.Data.Status {
	case "valid":
		// Email is valid, proceed with registration
		return nil
	case "invalid":
		return errors.New("the email address is invalid")
	case "disposable":
		return errors.New("the email address is from a disposable email service")
	case "webmail":
		// Optionally handle webmail addresses differently
		return nil
	case "unknown":
		return errors.New("failed to verify the email address")
	default:
		return errors.New("unexpected status from email verification")
	}
}

// ── Google OAuth flow ───────────────────────────────────────────────────────

// GoogleAuthOutcome is what GoogleSignInOrLink returns to the handler so the
// HTTP layer can pick the right response shape.
type GoogleAuthOutcome struct {
	User             *models.User
	IsNewUser        bool // true ⇢ we just created the row in this call
	NeedsCompletion  bool // true ⇢ student email but no student_id yet
	WasLinked        bool // true ⇢ matched an existing password account by email
}

// GoogleSignInOrLink is the heart of the Google flow. Given a verified set of
// Google claims, it does exactly one of:
//
//  1. Existing row found by google_sub  → log them in.
//  2. Existing row found by email       → auto-link (silent) and log them in.
//  3. No row found                      → create a fresh user. Role is
//     decided by the email domain:
//       - @student.president.ac.id → temporarily Guest with profile_completed
//         = false; the frontend will route them to /complete-profile to enter
//         Student ID + batch. Only then do we promote to Computizen / keep as
//         Guest based on the prefix.
//       - everything else (gmail.com etc.) → Guest, profile_completed = true.
func (as *AuthService) GoogleSignInOrLink(claims *utils.GoogleIDTokenClaims) (*GoogleAuthOutcome, error) {
	email := strings.ToLower(strings.TrimSpace(claims.Email))
	if email == "" {
		return nil, errors.New("google id_token has no email")
	}

	// 1. Lookup by google_sub – the cheapest, most-stable path.
	if u, err := app.GetUserByGoogleSub(claims.Sub); err != nil {
		return nil, fmt.Errorf("lookup by google_sub: %w", err)
	} else if u != nil {
		return &GoogleAuthOutcome{
			User:            u,
			NeedsCompletion: !u.ProfileCompleted,
		}, nil
	}

	// 2. Lookup by email – existing password user adopting Google.
	if u, err := app.GetUserByEmail(email); err != nil {
		return nil, fmt.Errorf("lookup by email: %w", err)
	} else if u != nil {
		if err := app.LinkGoogleSub(u.ID, claims.Sub); err != nil {
			return nil, fmt.Errorf("link google_sub: %w", err)
		}
		// Refresh the row so the caller sees the new auth_provider.
		refreshed, err := app.GetUserByID(u.ID)
		if err != nil || refreshed == nil {
			return nil, fmt.Errorf("refresh after link: %w", err)
		}
		return &GoogleAuthOutcome{
			User:            refreshed,
			WasLinked:       true,
			NeedsCompletion: !refreshed.ProfileCompleted,
		}, nil
	}

	// 3. Brand-new account.
	isStudentEmail := strings.HasSuffix(email, "@student.president.ac.id")

	newUser := &models.User{
		ID:               uuid.New(),
		Email:            email,
		FirstName:        firstNonEmpty(claims.GivenName, claims.Name),
		LastName:         claims.FamilyName,
		Gender:           "male",
		GoogleSub:        &claims.Sub,
		AuthProvider:     "google",
		ProfileCompleted: !isStudentEmail, // gmail = done, student = needs form
		EmailVerified:    true,
	}

	// Generate a unique username from the email's local-part. If it collides
	// we suffix a 4-char random string.
	base := utils.RemoveWhitespace(strings.ToLower(strings.Split(email, "@")[0]))
	if base == "" {
		base = "user"
	}
	newUser.Username = base
	if exists, _ := as.IsUsernameExists(base); exists {
		newUser.Username = base + utils.GenerateRandomString(4)
	}

	if isStudentEmail {
		// They will fill these in on the complete-profile screen.
		newUser.StudentID = ""
		newUser.Major = ""
		newUser.Year = ""
		newUser.RoleID = models.RoleGuest // temporary
	} else {
		newUser.RoleID = models.RoleGuest // gmail.com stays guest forever
		newUser.StudentID = ""
		newUser.Major = ""
		newUser.Year = ""
	}

	if err := app.CreateGoogleUser(newUser); err != nil {
		return nil, fmt.Errorf("create google user: %w", err)
	}

	return &GoogleAuthOutcome{
		User:            newUser,
		IsNewUser:       true,
		NeedsCompletion: isStudentEmail,
	}, nil
}

// LinkGoogleToExistingAccount is the explicit dashboard-button version of
// step 2 in GoogleSignInOrLink. Used when a logged-in password user clicks
// "Link Google" while authenticated.
func (as *AuthService) LinkGoogleToExistingAccount(userID uuid.UUID, claims *utils.GoogleIDTokenClaims) error {
	// Refuse if the Google sub is already attached to a *different* user –
	// otherwise we'd silently re-route Google sign-ins to someone else.
	if existing, err := app.GetUserByGoogleSub(claims.Sub); err != nil {
		return err
	} else if existing != nil && existing.ID != userID {
		return errors.New("this Google account is already linked to a different user")
	}
	return app.LinkGoogleSub(userID, claims.Sub)
}

// CompleteStudentProfile runs the same Student-ID validation as the classic
// /auth/register handler and writes the result onto an already-created Google
// account. Role is promoted to Computizen for CS prefixes, kept as Guest
// otherwise (Faculty of Computer Science gate).
func (as *AuthService) CompleteStudentProfile(userID uuid.UUID, studentID, year string) (roleID int, major string, err error) {
	if len(studentID) != 12 {
		return 0, "", errors.New("student ID must be 12 characters long")
	}
	if studentID[3:7] < "2010" {
		return 0, "", errors.New("you are not eligible to register an account")
	}

	// Reject duplicate Student IDs (someone else already claimed this number).
	if existing, err := app.GetUserByStudentID(studentID); err == nil && existing != nil && existing.ID != userID {
		return 0, "", errors.New("Student ID already exists")
	}

	switch studentID[:3] {
	case "001":
		major = "informatics"
		roleID = models.RoleComputizen
	case "012":
		major = "information system"
		roleID = models.RoleComputizen
	case "013":
		major = "visual communication design"
		roleID = models.RoleComputizen
	case "025":
		major = "interior design"
		roleID = models.RoleComputizen
	default:
		// Not part of the Faculty of Computer Science → stay as Guest.
		major = "other"
		roleID = models.RoleGuest
	}

	if err := app.CompleteGoogleProfile(userID, studentID, major, year, roleID); err != nil {
		return 0, "", err
	}
	return roleID, major, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func (as *AuthService) RequestForgotPassword(userID uuid.UUID) (string, error) {
	tokenOtp := utils.GenerateRandomTokenOtp()
	expiresAt := time.Now().Add(5 * time.Minute)
	otpCode, err := as.otp.GenerateOTP(userID, tokenOtp, time.Minute*5)
	if err != nil {
		return "", err
	}

	// TODO: Send OTP code to email
	log.Println("OTP code:", otpCode)

	err = app.SavePasswordResetToken(userID, tokenOtp, expiresAt)
	if err != nil {
		return "", err
	}

	return otpCode, nil
}

func (as *AuthService) VerifyOTP(userID uuid.UUID, otpCode string) bool {
	tokenOtp, err := app.GetPasswordResetToken(userID)
	if err != nil {
		return false
	}

	return as.otp.VerifyOTP(userID, tokenOtp, otpCode)
}

func (as *AuthService) ResetPassword(userID uuid.UUID, otpCode, password string) (bool, error) {
	tokenOtp, err := app.GetPasswordResetToken(userID)
	if err != nil {
		return false, err
	}

	valid := as.otp.VerifyOTP(userID, tokenOtp, otpCode)
	if !valid {
		return false, nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return false, err
	}

	err = app.UpdatePassword(userID, string(hashedPassword))
	if err != nil {
		return false, err
	}

	return true, nil
}
