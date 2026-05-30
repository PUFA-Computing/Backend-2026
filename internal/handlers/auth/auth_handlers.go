package auth

import (
	"Backend/internal/models"
	"Backend/internal/services"
	"Backend/pkg/utils"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log"
	"net/http"
	"os"
	"strings"
)

type Handlers struct {
	AuthService       *services.AuthService
	PermissionService *services.PermissionService
	UserService       *services.UserService
}

func NewAuthHandlers(authService *services.AuthService, permissionService *services.PermissionService, userService *services.UserService) *Handlers {
	return &Handlers{
		AuthService:       authService,
		PermissionService: permissionService,
		UserService:       userService,
	}
}

// RegisterUser registers a new user (default role: Computizen - role_id 2)
// @Summary Register a new user
// @Description Register a new user account. New users get Computizen role (role_id: 2) by default. Only users with year 2025 can vote.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.User true "User registration details"
// @Success 201 {object} map[string]interface{} "User registered successfully"
// @Failure 400 {object} map[string]interface{} "Bad request or validation error"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/register [post]
func (h *Handlers) RegisterUser(c *gin.Context) {
	log.Println("=== Starting RegisterUser function ===")
	var newUser models.User
	suffix := "@student.president.ac.id"

	if err := c.BindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	// Remove whitespace from firstname and lastname
	newUser.FirstName = utils.RemoveWhitespace(newUser.FirstName)
	newUser.LastName = utils.RemoveWhitespace(newUser.LastName)
	newUser.Username = utils.RemoveWhitespace(newUser.Username)

	// Check if username or email already exists
	// // if username exists add something to username because its generate from firstname and lastname
	if exists, err := h.AuthService.IsUsernameExists(newUser.Username); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	} else if exists {
		// Generate a random string of characters
		randomBytes := make([]byte, 4) // Adjust length as needed
		if _, err := rand.Read(randomBytes); err != nil {
			// Handle error if random generation fails
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{"Failed to generate random string"}})
			return
		}
		randomString := base64.URLEncoding.EncodeToString(randomBytes)
		randomString = randomString[0:4] // Keep only the first 4 characters

		// Append the random string to the username
		newUser.Username = fmt.Sprintf("%s%s", newUser.Username, randomString)
		log.Println("New Username: ", newUser.Username)
	}

	log.Printf("Validating email: %s against suffix: %s", newUser.Email, suffix)
	if err := validateEmail(newUser.Email, suffix); err != nil {
		log.Printf("Email validation failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}
	log.Println("Email validation passed")

	if exists, err := h.AuthService.IsEmailExists(newUser.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	} else if exists {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{"Email already exists"}})
		return
	}

	if err := validateStudentID(newUser.StudentID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	// Check student ID exists
	if exists, err := h.AuthService.IsStudentIDExists(newUser.StudentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	} else if exists {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{"Student ID already exists"}})
		return
	}

	// Email Verification Token is generated but not sent via SMTP.
	// All @student.president.ac.id accounts are auto-verified on creation.
	token := utils.GenerateRandomString(32)
	newUser.EmailVerificationToken = token

	if err := h.AuthService.RegisterUser(&newUser); err != nil {
		// Distinguish validation errors (400) from server errors (500)
		if strings.Contains(err.Error(), "email") ||
			strings.Contains(err.Error(), "invalid") ||
			strings.Contains(err.Error(), "disposable") ||
			strings.Contains(err.Error(), "verify") ||
			strings.Contains(err.Error(), "validation") {
			log.Printf("Email validation error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		} else {
			log.Printf("Server error during registration: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		}
		return
	}

	log.Println("Registration completed successfully – account auto-verified (no SMTP)")
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Account created successfully! You can now sign in with your email and password.",
	})
}

// Validate student ID
func validateStudentID(studentID string) error {
	if len(studentID) != 12 {
		return errors.New("student ID must be 12 characters long")
	} else if studentID[:3] != "001" && studentID[:3] != "012" && studentID[:3] != "013" && studentID[:3] != "025" {
		return errors.New("you are not a student of faculty of computer science")
	} else if studentID[3:7] < "2010" {
		return errors.New("you are not eligible to register an account")
	}
	return nil
}

func validateEmail(email, suffix string) error {
	if len(email) < len(suffix) || email[len(email)-len(suffix):] != suffix {
		return errors.New("email must be a President University student email")
	}
	return nil
}

// Login authenticates a user and returns JWT token
// @Summary Login user
// @Description Login with username/email and password. Returns JWT token for authentication. Admin (role_id: 1) can create candidates. Computizen (role_id: 2) with year 2025 can vote.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body object{username=string,password=string,passcode=string} true "Login credentials"
// @Success 200 {object} map[string]interface{} "Login successful with JWT token"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Invalid credentials or email not verified"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/login [post]
func (h *Handlers) Login(c *gin.Context) {
	var loginRequest struct {
		Username string  `json:"username"`
		Password string  `json:"password"`
		Passcode *string `json:"passcode"`
	}

	if err := c.BindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	// Lowercase the username
	loginRequest.Username = strings.ToLower(loginRequest.Username)

	// Add debug logging
	log.Printf("Attempting login for user: %s", loginRequest.Username)
	user, err := h.AuthService.LoginUser(loginRequest.Username, loginRequest.Password)
	if err != nil {
		// Log the actual error for debugging
		log.Printf("Login error: %v", err)
		
		// Check if it's an unauthorized error or another type of error
		var unauthorizedErr *utils.UnauthorizedError
		if errors.As(err, &unauthorizedErr) {
			// This is an invalid credentials error
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid Credentials"})
		} else {
			// This is a server error
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Login Failed", "error": err.Error()})
		}
		return
	}

	// If there is no passcode, but 2FA is enabled, return otp required
	if loginRequest.Passcode == nil {

		if user.TwoFAEnabled {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Two Factor Authentication Required"})
			return
		}
	}

	if loginRequest.Passcode != nil {
		if !user.TwoFAEnabled {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "2FA is not enabled for this account"})
			return
		}

		_, err := h.UserService.VerifyTwoFA(user.ID, *loginRequest.Passcode)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid 2FA Code"})
			return
		}

	}

	// Account is always auto-verified on creation (no email gate).
	// This block is intentionally removed – all registered users can log in directly.

	log.Printf("Email verification check skipped – accounts are auto-verified at registration")

	token, err := utils.GenerateJWTToken(user.ID, os.Getenv("JWT_SECRET_KEY"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	// Try to store token in Redis, but continue even if it fails
	_ = utils.StoreTokenInRedis(user.ID, token)
	// No need to check for errors since we've made Redis optional

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Login Successful",
		"data":    gin.H{"access_token": token, "token_type": "Bearer", "user_id": user.ID.String()},
	})
}

// Logout logs out the current user
// @Summary Logout user
// @Description Logout the current user and invalidate the JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} map[string]interface{} "Logout successful"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Security BearerAuth
// @Router /auth/logout [post]
func (h *Handlers) Logout(c *gin.Context) {
	tokenString, err := utils.ExtractTokenFromHeader(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": []string{"Unauthorized"}})
		return
	}

	_, err = utils.ValidateToken(tokenString, os.Getenv("JWT_SECRET_KEY"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": []string{"Unauthorized"}})
		return
	}

	err = utils.RevokeToken(tokenString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Logout Successful"})
}

func (h *Handlers) RefreshToken(c *gin.Context) {
	tokenString, err := utils.ExtractTokenFromHeader(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": []string{"Unauthorized"}})
		return
	}

	claims, err := utils.ValidateToken(tokenString, os.Getenv("JWT_SECRET_KEY"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": []string{"Unauthorized"}})
		return
	}

	userID := claims.UserID
	token, err := utils.GenerateJWTToken(userID, os.Getenv("JWT_SECRET_KEY"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	if err := utils.StoreTokenInRedis(userID, token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Access Token Refreshed Successfully",
		"data": gin.H{
			"access_token": token,
			"token_type":   "Bearer",
			"user_id":      userID.String(),
		},
	})
}

func (h *Handlers) ExtractUserIDAndCheckPermission(c *gin.Context, permissionType string) (uuid.UUID, error) {
	token, err := utils.ExtractTokenFromHeader(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return uuid.UUID{}, err
	}

	userID, err := utils.GetUserIDFromToken(token, os.Getenv("JWT_SECRET_KEY"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return uuid.UUID{}, err
	}

	hasPermission, err := (&services.PermissionService{}).CheckPermission(context.Background(), userID, permissionType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return uuid.UUID{}, err
	}

	if !hasPermission {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": []string{"Unauthorized"}})
		return uuid.UUID{}, err
	}

	return userID, nil
}

// ── Google OAuth ─────────────────────────────────────────────────────────────

// GoogleSignIn accepts a Google ID token from the frontend (issued by
// NextAuth's GoogleProvider), verifies it server-side, then either logs the
// user in or creates / links the account. The response always carries a
// backend JWT so the SPA can call protected endpoints the same way as
// password login.
//
// Body: { "id_token": "<google-id-token>" }
//
// Response on success:
//
//	{
//	  "success": true,
//	  "data": {
//	    "access_token": "...",
//	    "token_type":   "Bearer",
//	    "user_id":      "<uuid>",
//	    "needs_completion": true|false,
//	    "is_new_user":      true|false,
//	    "was_linked":       true|false,
//	    "role_id":          6
//	  }
//	}
//
// needs_completion=true tells the frontend to route the user to
// /auth/complete-profile to enter Student ID + batch.
//
// @Summary Sign in / register with Google
// @Tags Auth
// @Accept json
// @Produce json
// @Router /auth/google [post]
func (h *Handlers) GoogleSignIn(c *gin.Context) {
	var body struct {
		IDToken string `json:"id_token"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "id_token is required"})
		return
	}

	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	claims, err := utils.VerifyGoogleIDToken(body.IDToken, clientID)
	if err != nil {
		log.Printf("Google id_token rejected: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid Google token"})
		return
	}

	outcome, err := h.AuthService.GoogleSignInOrLink(claims)
	if err != nil {
		log.Printf("GoogleSignInOrLink failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	if outcome.WasLinked {
		// Linked account logic can go here
	}

	token, err := utils.GenerateJWTToken(outcome.User.ID, os.Getenv("JWT_SECRET_KEY"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	_ = utils.StoreTokenInRedis(outcome.User.ID, token)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Google sign-in successful",
		"data": gin.H{
			"access_token":     token,
			"token_type":       "Bearer",
			"user_id":          outcome.User.ID.String(),
			"role_id":          outcome.User.RoleID,
			"needs_completion": outcome.NeedsCompletion,
			"is_new_user":      outcome.IsNewUser,
			"was_linked":       outcome.WasLinked,
			"email":            outcome.User.Email,
		},
	})
}

// CompleteGoogleProfile finishes a fresh Google sign-up by accepting Student
// ID + batch year. Promotes to Computizen (role 2) for CS prefixes; otherwise
// the row stays as Guest (role 6).
//
// Body: { "student_id": "001202301234", "year": "2023" }
//
// @Summary Complete Google profile with Student ID + batch
// @Tags Auth
// @Accept json
// @Produce json
// @Router /auth/google/complete [post]
func (h *Handlers) CompleteGoogleProfile(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Unauthorized"})
		return
	}

	var body struct {
		StudentID string `json:"student_id"`
		Year      string `json:"year"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	roleID, major, err := h.AuthService.CompleteStudentProfile(userID, strings.TrimSpace(body.StudentID), strings.TrimSpace(body.Year))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Profile completed",
		"data": gin.H{
			"role_id": roleID,
			"major":   major,
		},
	})
}

// LinkGoogleAccount is the "Link Google" button on the profile page. It
// requires a valid backend JWT (password user) AND a fresh Google id_token.
//
// Body: { "id_token": "<google-id-token>" }
//
// @Summary Link a Google account to the currently authenticated user
// @Tags Auth
// @Accept json
// @Produce json
// @Router /auth/google/link [post]
func (h *Handlers) LinkGoogleAccount(c *gin.Context) {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Unauthorized"})
		return
	}

	var body struct {
		IDToken string `json:"id_token"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "id_token is required"})
		return
	}

	claims, err := utils.VerifyGoogleIDToken(body.IDToken, os.Getenv("GOOGLE_CLIENT_ID"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid Google token"})
		return
	}

	if err := h.AuthService.LinkGoogleToExistingAccount(userID, claims); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Google account linked"})
}

func (h *Handlers) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{"Token is required"}})
		return
	}

	exists, err := h.AuthService.IsTokenVerificationEmailExists(token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{"Invalid Token"}})
		return
	}

	if err := h.AuthService.VerifyEmail(token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Email Verified Successfully"})
}



// GetAccountStatus checks if an account exists and returns its auth status
// @Summary Get account status by email
// @Tags Auth
// @Produce json
// @Param email query string true "User email"
// @Router /auth/account-status [get]
func (h *Handlers) GetAccountStatus(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Email is required"})
		return
	}

	user, err := h.AuthService.GetUserByEmail(email)
	if err != nil || user == nil {
		// Just return false if not found, don't leak errors for enumeration
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"exists": false,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"exists":        true,
			"auth_provider": user.AuthProvider,
			"has_password":  user.Password != "",
		},
	})
}

