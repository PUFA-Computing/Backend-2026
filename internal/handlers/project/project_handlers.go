package project

import (
	"Backend/internal/handlers/auth"
	"Backend/internal/models"
	"Backend/internal/services"
	"Backend/pkg/utils"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Handler struct {
	ProjectService     *services.ProjectService
	PermissionService  *services.PermissionService
	AWSService         *services.S3Service
	R2Service          *services.S3Service
}

func NewProjectHandler(projectService *services.ProjectService, permissionService *services.PermissionService, AWSService *services.S3Service, R2Service *services.S3Service) *Handler {
	return &Handler{
		ProjectService:    projectService,
		PermissionService: permissionService,
		AWSService:        AWSService,
		R2Service:         R2Service,
	}
}

// CreateProject creates a new project
// @Summary Create a new project
// @Description Create a new project with image upload. Requires authentication. Project will be unpublished by default.
// @Tags Projects
// @Accept multipart/form-data
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param data formData string true "Project data as JSON string" example({"title":"My Project","description":"Project description","category":"Website","project_url":"https://github.com/user/project"})
// @Param file formData file true "Project image"
// @Success 201 {object} map[string]interface{} "Project created successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /projects/create [post]
func (h *Handler) CreateProject(c *gin.Context) {
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

	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	data := c.Request.FormValue("data")
	var req models.CreateProjectRequest
	if err := json.Unmarshal([]byte(data), &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	// Create project model
	project := &models.Project{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		ProjectURL:  req.ProjectURL,
	}

	// Handle file upload (required)
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{"Project image is required"}})
		return
	}
	defer file.Close()

	// Choose storage service
	upload := utils.ChooseStorageService()

	// Process image
	optimizedImage, err := utils.OptimizeImage(file, 1200, 800)
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
	filename := utils.GenerateFriendlyURL(req.Title) + "-" + strconv.FormatInt(utils.GenerateRandomInt64(), 10)

	if upload == utils.R2Service {
		err = h.R2Service.UploadFileToR2(context.Background(), "projects", filename, optimizedImageBytes, "image/jpeg")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
			return
		}
		imageURL, _ := h.R2Service.GetFileR2("projects", filename)
		project.ImageURL = imageURL
	} else {
		err = h.AWSService.UploadFileToAWS(context.Background(), "projects", filename, optimizedImageBytes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
			return
		}
		imageURL, _ := h.AWSService.GetFileAWS("projects", filename)
		project.ImageURL = imageURL
	}

	if err := h.ProjectService.CreateProject(project); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	log.Printf("Project created by user %s: %s", userID.String(), project.Title)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Project Created Successfully",
		"data":    project,
	})
}

// GetProjectByID retrieves a project by ID
// @Summary Get project by ID
// @Description Get detailed information about a specific project including vote count
// @Tags Projects
// @Accept json
// @Produce json
// @Param projectID path int true "Project ID"
// @Success 200 {object} object{success=bool,message=string,data=models.ProjectResponse} "Project details"
// @Failure 400 {object} map[string]interface{} "Invalid project ID"
// @Failure 404 {object} map[string]interface{} "Project not found"
// @Router /projects/{projectID} [get]
func (h *Handler) GetProjectByID(c *gin.Context) {
	projectIDStr := c.Param("projectID")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{"Invalid Project ID"}})
		return
	}

	project, err := h.ProjectService.GetProjectWithVoteCount(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": []string{"Project not found"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Project Retrieved Successfully",
		"data":    project,
	})
}

// GetAllProjects retrieves all projects with optional filters
// @Summary List all projects
// @Description Get a paginated list of all published projects with optional filters by category
// @Tags Projects
// @Accept json
// @Produce json
// @Param category query string false "Filter by category (Website, AI, System, etc)"
// @Param page query int false "Page number for pagination" default(1)
// @Success 200 {object} map[string]interface{} "List of projects"
// @Router /projects [get]
func (h *Handler) GetAllProjects(c *gin.Context) {
	queryParams := make(map[string]string)
	queryParams["category"] = c.Query("category")
	queryParams["page"] = c.Query("page")
	// Only show published projects by default
	queryParams["is_published"] = "true"

	projects, totalPages, err := h.ProjectService.ListProjects(queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"data":         projects,
		"totalResults": len(projects),
		"totalPages":   totalPages,
	})
}

// UpdateProject updates an existing project
// @Summary Update a project
// @Description Update project information with optional image upload. Only project owner can update.
// @Tags Projects
// @Accept multipart/form-data
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param projectID path int true "Project ID"
// @Param data formData string true "Project data as JSON string" example({"title":"Updated Title","description":"Updated description"})
// @Param file formData file false "New project image"
// @Success 200 {object} map[string]interface{} "Project updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Not project owner"
// @Failure 404 {object} map[string]interface{} "Project not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /projects/{projectID}/edit [put]
func (h *Handler) UpdateProject(c *gin.Context) {
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

	projectIDStr := c.Param("projectID")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{"Invalid Project ID"}})
		return
	}

	existingProject, err := h.ProjectService.GetProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": []string{"Project not found"}})
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

	var req models.UpdateProjectRequest
	if err := json.Unmarshal([]byte(data), &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	// Update project fields
	updatedProject := &models.Project{
		Title:       existingProject.Title,
		Description: existingProject.Description,
		Category:    existingProject.Category,
		ProjectURL:  existingProject.ProjectURL,
		ImageURL:    existingProject.ImageURL,
	}

	if req.Title != nil {
		updatedProject.Title = *req.Title
	}
	if req.Description != nil {
		updatedProject.Description = *req.Description
	}
	if req.Category != nil {
		updatedProject.Category = req.Category
	}
	if req.ProjectURL != nil {
		updatedProject.ProjectURL = req.ProjectURL
	}

	// Handle file upload
	file, fileHeader, err := c.Request.FormFile("file")
	hasNewImage := err == nil && fileHeader != nil

	if hasNewImage {
		defer file.Close()

		optimizedImage, err := utils.OptimizeImage(file, 1200, 800)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
			return
		}

		optimizedImageBytes, err := io.ReadAll(optimizedImage)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
			return
		}

		filename := utils.GenerateFriendlyURL(updatedProject.Title) + "-" + strconv.FormatInt(utils.GenerateRandomInt64(), 10)

		// Upload to R2 storage
		err = h.R2Service.UploadFileToR2(context.Background(), "projects", filename, optimizedImageBytes, "image/jpeg")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
			return
		}

		imageURL, err := h.R2Service.GetFileR2("projects", filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{"Failed to get image URL"}})
			return
		}

		updatedProject.ImageURL = imageURL
	}

	if err := h.ProjectService.UpdateProject(projectID, userID, updatedProject); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Project Updated Successfully",
		"data":    updatedProject,
	})
}

// DeleteProject deletes a project
// @Summary Delete a project
// @Description Delete a project by ID. Only project owner or admin can delete.
// @Tags Projects
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param projectID path int true "Project ID"
// @Success 200 {object} map[string]interface{} "Project deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid project ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Not project owner or admin"
// @Failure 404 {object} map[string]interface{} "Project not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /projects/{projectID}/delete [delete]
func (h *Handler) DeleteProject(c *gin.Context) {
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

	projectIDStr := c.Param("projectID")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{"Invalid Project ID"}})
		return
	}

	// Check if user has admin permission
	hasPermission, _ := h.PermissionService.CheckPermission(c.Request.Context(), userID, "project:delete")

	if err := h.ProjectService.DeleteProject(projectID, userID, hasPermission); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Project Deleted Successfully",
	})
}

// PublishProject publishes a project (admin only)
// @Summary Publish a project (Admin only)
// @Description Publish a project to make it visible to all users. Requires project:publish permission.
// @Tags Projects
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token (Admin only)"
// @Param projectID path int true "Project ID"
// @Success 200 {object} map[string]interface{} "Project published successfully"
// @Failure 400 {object} map[string]interface{} "Invalid project ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Admin role required"
// @Failure 404 {object} map[string]interface{} "Project not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /projects/{projectID}/publish [put]
func (h *Handler) PublishProject(c *gin.Context) {
	_, err := (&auth.Handlers{}).ExtractUserIDAndCheckPermission(c, "project:publish")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	projectIDStr := c.Param("projectID")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{"Invalid Project ID"}})
		return
	}

	if err := h.ProjectService.PublishProject(projectID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Project Published Successfully",
	})
}

// GetMyProjects retrieves all projects created by the current user
// @Summary Get my projects
// @Description Get all projects created by the current authenticated user
// @Tags Projects
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} map[string]interface{} "List of user's projects"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /projects/my-projects [get]
func (h *Handler) GetMyProjects(c *gin.Context) {
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

	projects, err := h.ProjectService.GetProjectsByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    projects,
		"count":   len(projects),
	})
}
