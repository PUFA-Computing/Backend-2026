package candidate

import (
	"Backend/internal/database/app"
	"Backend/internal/handlers/auth"
	"Backend/internal/models"
	"Backend/internal/services"
	"Backend/pkg/utils"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Handler struct {
	CandidateService  *services.CandidateService
	PermissionService *services.PermissionService
	AWSService        *services.S3Service
	R2Service         *services.S3Service
}

func NewCandidateHandler(candidateService *services.CandidateService, permissionService *services.PermissionService, AWSService *services.S3Service, R2Service *services.S3Service) *Handler {
	return &Handler{
		CandidateService:  candidateService,
		PermissionService: permissionService,
		AWSService:        AWSService,
		R2Service:         R2Service,
	}
}

// CreateCandidate creates a new candidate
// @Summary Create a new candidate (Admin only - role_id: 1)
// @Description Create a new candidate with profile picture upload. Requires Admin role (role_id: 1) and candidate:create permission.
// @Tags Candidates
// @Accept multipart/form-data
// @Produce json
// @Param Authorization header string true "Bearer token (Admin only)"
// @Param data formData string true "Candidate data as JSON string" example({"name":"John Doe","vision":"My vision","mission":"My mission","class":"2024","major":"informatics"})
// @Param file formData file false "Profile picture"
// @Success 201 {object} map[string]interface{} "Candidate created successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Admin role required"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /candidates/create [post]
func (h *Handler) CreateCandidate(c *gin.Context) {
	userID, err := (&auth.Handlers{}).ExtractUserIDAndCheckPermission(c, "candidate:create")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	data := c.Request.FormValue("data")
	var req models.CreateCandidateRequest
	if err := json.Unmarshal([]byte(data), &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	// Create candidate model
	candidate := &models.Candidate{
		Name:    req.Name,
		Vision:  req.Vision,
		Mission: req.Mission,
		Class:   req.Class,
		Major:   req.Major,
	}

	// Parse user_id if provided
	if req.UserID != nil && *req.UserID != "" {
		parsedUserID, err := uuid.Parse(*req.UserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{"Invalid user_id format"}})
			return
		}
		candidate.UserID = &parsedUserID
	}

	// Handle file upload
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil && err != http.ErrMissingFile {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	// Choose storage service
	upload := utils.ChooseStorageService()

	// Process image if uploaded
	if err == nil && fileHeader != nil {
		defer file.Close()

		optimizedImage, err := utils.OptimizeImage(file, 800, 800)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
			return
		}

		optimizedImageBytes, err := io.ReadAll(optimizedImage)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
			return
		}

		// Generate unique filename
		filename := utils.GenerateFriendlyURL(req.Name) + "-" + strconv.FormatInt(utils.GenerateRandomInt64(), 10)

		if upload == utils.R2Service {
			err = h.R2Service.UploadFileToR2(context.Background(), "candidates", filename, optimizedImageBytes, "image/jpeg")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
				return
			}
			profilePictureURL, _ := h.R2Service.GetFileR2("candidates", filename)
			candidate.ProfilePicture = &profilePictureURL
		} else {
			err = h.AWSService.UploadFileToAWS(context.Background(), "candidates", filename, optimizedImageBytes)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
				return
			}
			profilePictureURL, _ := h.AWSService.GetFileAWS("candidates", filename)
			candidate.ProfilePicture = &profilePictureURL
		}
	} else {
		// Use default profile picture
		if upload == utils.R2Service {
			defaultURL, _ := h.R2Service.GetFileR2("default", "candidate-profile")
			candidate.ProfilePicture = &defaultURL
		} else {
			defaultURL, _ := h.AWSService.GetFileAWS("default", "candidate-profile")
			candidate.ProfilePicture = &defaultURL
		}
	}

	if err := h.CandidateService.CreateCandidate(candidate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	log.Printf("Candidate created by user %s: %s", userID.String(), candidate.Name)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Candidate Created Successfully",
		"data":    candidate,
	})
}

// GetCandidateByID retrieves a candidate by ID
// @Summary Get candidate by ID
// @Description Get detailed information about a specific candidate including vote count
// @Tags Candidates
// @Accept json
// @Produce json
// @Param candidateID path int true "Candidate ID"
// @Success 200 {object} object{success=bool,message=string,data=models.CandidateResponse} "Candidate details"
// @Failure 400 {object} map[string]interface{} "Invalid candidate ID"
// @Failure 404 {object} map[string]interface{} "Candidate not found"
// @Router /candidates/{candidateID} [get]
func (h *Handler) GetCandidateByID(c *gin.Context) {
	candidateIDStr := c.Param("candidateID")
	candidateID, err := strconv.Atoi(candidateIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{"Invalid Candidate ID"}})
		return
	}

	candidate, err := h.CandidateService.GetCandidateWithVoteCount(candidateID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": []string{"Candidate not found"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Candidate Retrieved Successfully",
		"data":    candidate,
	})
}

// GetAllCandidates retrieves all candidates with optional filters
// @Summary List all candidates
// @Description Get a paginated list of all candidates with optional filters by major and class
// @Tags Candidates
// @Accept json
// @Produce json
// @Param major query string false "Filter by major (information_system or informatics)"
// @Param class query string false "Filter by class"
// @Param page query int false "Page number for pagination" default(1)
// @Success 200 {object} map[string]interface{} "List of candidates" example({"success":true,"data":[{"id":1,"name":"John Doe","vision":"My vision","mission":"My mission","class":"2024","major":"informatics","profile_picture":"https://...","vote_count":5}],"totalResults":1,"totalPages":1})
// @Router /candidates [get]
func (h *Handler) GetAllCandidates(c *gin.Context) {
	queryParams := make(map[string]string)
	queryParams["major"] = c.Query("major")
	queryParams["class"] = c.Query("class")
	queryParams["page"] = c.Query("page")

	candidates, totalPages, err := h.CandidateService.ListCandidates(queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"data":         candidates,
		"totalResults": len(candidates),
		"totalPages":   totalPages,
	})
}

// UpdateCandidate updates an existing candidate
// @Summary Update a candidate (Admin only - role_id: 1)
// @Description Update candidate information with optional profile picture upload. Requires Admin role (role_id: 1) and candidate:edit permission.
// @Tags Candidates
// @Accept multipart/form-data
// @Produce json
// @Param Authorization header string true "Bearer token (Admin only)"
// @Param candidateID path int true "Candidate ID"
// @Param data formData string true "Candidate data as JSON string" example({"name":"John Doe","vision":"Updated vision","major":"informatics"})
// @Param file formData file false "New profile picture"
// @Success 200 {object} map[string]interface{} "Candidate updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Admin role required"
// @Failure 404 {object} map[string]interface{} "Candidate not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /candidates/{candidateID}/edit [put]
func (h *Handler) UpdateCandidate(c *gin.Context) {
	_, err := (&auth.Handlers{}).ExtractUserIDAndCheckPermission(c, "candidate:edit")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	candidateIDStr := c.Param("candidateID")
	candidateID, err := strconv.Atoi(candidateIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{"Invalid Candidate ID"}})
		return
	}

	existingCandidate, err := h.CandidateService.GetCandidateByID(candidateID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": []string{"Candidate not found"}})
		return
	}

	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	data := c.Request.FormValue("data")
	if data == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{"No data provided"}})
		return
	}

	var req models.UpdateCandidateRequest
	if err := json.Unmarshal([]byte(data), &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	// Update candidate fields
	updatedCandidate := &models.Candidate{}
	if req.Name != nil {
		updatedCandidate.Name = *req.Name
	}
	if req.Vision != nil {
		updatedCandidate.Vision = req.Vision
	}
	if req.Mission != nil {
		updatedCandidate.Mission = req.Mission
	}
	if req.Class != nil {
		updatedCandidate.Class = req.Class
	}
	if req.Major != nil {
		updatedCandidate.Major = *req.Major
	}
	if req.UserID != nil && *req.UserID != "" {
		parsedUserID, err := uuid.Parse(*req.UserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{"Invalid user_id format"}})
			return
		}
		updatedCandidate.UserID = &parsedUserID
	}

	// Handle file upload
	file, fileHeader, err := c.Request.FormFile("file")
	hasNewImage := err == nil && fileHeader != nil

	if hasNewImage {
		defer file.Close()

		optimizedImage, err := utils.OptimizeImage(file, 800, 800)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
			return
		}

		optimizedImageBytes, err := io.ReadAll(optimizedImage)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
			return
		}

		filename := utils.GenerateFriendlyURL(existingCandidate.Name) + "-" + strconv.FormatInt(utils.GenerateRandomInt64(), 10)

		// Upload to R2 storage
		err = h.R2Service.UploadFileToR2(context.Background(), "candidates", filename, optimizedImageBytes, "image/jpeg")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
			return
		}

		profilePictureURL, err := h.R2Service.GetFileR2("candidates", filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{"Failed to get profile picture URL"}})
			return
		}

		updatedCandidate.ProfilePicture = &profilePictureURL
	} else {
		updatedCandidate.ProfilePicture = existingCandidate.ProfilePicture
	}

	if err := h.CandidateService.UpdateCandidate(candidateID, updatedCandidate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Candidate Updated Successfully",
		"data":    updatedCandidate,
	})
}

// DeleteCandidate deletes a candidate
// @Summary Delete a candidate (Admin only - role_id: 1)
// @Description Delete a candidate by ID. Requires Admin role (role_id: 1) and candidate:delete permission.
// @Tags Candidates
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token (Admin only)"
// @Param candidateID path int true "Candidate ID"
// @Success 200 {object} map[string]interface{} "Candidate deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid candidate ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Admin role required"
// @Failure 404 {object} map[string]interface{} "Candidate not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /candidates/{candidateID}/delete [delete]
func (h *Handler) DeleteCandidate(c *gin.Context) {
	_, err := (&auth.Handlers{}).ExtractUserIDAndCheckPermission(c, "candidate:delete")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	candidateIDStr := c.Param("candidateID")
	candidateID, err := strconv.Atoi(candidateIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{"Invalid Candidate ID"}})
		return
	}

	candidate, err := h.CandidateService.GetCandidateByID(candidateID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": []string{"Candidate not found"}})
		return
	}

	// Delete profile picture from storage if it exists and is not the default
	if candidate.ProfilePicture != nil && *candidate.ProfilePicture != "" {
		// Extract filename from URL (simplified approach)
		// In production, you'd want more robust URL parsing
		exists, _ := h.R2Service.FileExists(context.Background(), "candidates", candidate.Name)
		if exists {
			_ = h.R2Service.DeleteFile(context.Background(), "candidates", candidate.Name)
		}
	}

	if err := h.CandidateService.DeleteCandidate(candidateID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Candidate Deleted Successfully",
	})
}

// GetCandidatesForMyMajor retrieves all candidates that the current user can vote for based on their major
// @Summary Get candidates for my major
// @Description Get all candidates that the current user can vote for based on their major. Requires authentication.
// @Tags Candidates
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} map[string]interface{} "List of candidates for the user's major"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /candidates/my-major [get]
func (h *Handler) GetCandidatesForMyMajor(c *gin.Context) {
	token, err := utils.ExtractTokenFromHeader(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	userID, err := utils.GetUserIDFromToken(token, os.Getenv("JWT_SECRET_KEY"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	// Get user's major
	major, err := app.GetUserMajor(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{"Unable to retrieve user information"}})
		return
	}

	// Get candidates by major
	candidates, err := h.CandidateService.GetCandidatesByMajor(major)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    candidates,
		"count":   len(candidates),
		"major":   major,
	})
}
