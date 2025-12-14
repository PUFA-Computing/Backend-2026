package project

import (
	"Backend/internal/services"
	"Backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"strconv"
)

type VoteHandler struct {
	ProjectVoteService *services.ProjectVoteService
}

func NewProjectVoteHandler(projectVoteService *services.ProjectVoteService) *VoteHandler {
	return &VoteHandler{
		ProjectVoteService: projectVoteService,
	}
}

// VoteProject allows a user to vote for a project
// @Summary Vote for a project
// @Description Vote for a published project. One vote per user per project.
// @Tags Project Votes
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param projectID path int true "Project ID"
// @Success 201 {object} map[string]interface{} "Vote cast successfully"
// @Failure 400 {object} map[string]interface{} "Bad request, already voted, or project not published"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Project not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /projects/{projectID}/vote [post]
func (h *VoteHandler) VoteProject(c *gin.Context) {
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

	vote, err := h.ProjectVoteService.VoteProject(userID, projectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Vote Cast Successfully",
		"data":    vote,
	})
}

// UnvoteProject allows a user to remove their vote from a project
// @Summary Remove vote from a project
// @Description Remove your vote from a project
// @Tags Project Votes
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param projectID path int true "Project ID"
// @Success 200 {object} map[string]interface{} "Vote removed successfully"
// @Failure 400 {object} map[string]interface{} "Bad request or haven't voted"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Project not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /projects/{projectID}/unvote [delete]
func (h *VoteHandler) UnvoteProject(c *gin.Context) {
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

	if err := h.ProjectVoteService.UnvoteProject(userID, projectID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Vote Removed Successfully",
	})
}

// GetMyProjectVotes retrieves all votes cast by the current user
// @Summary Get my project votes
// @Description Get all projects voted by the current authenticated user
// @Tags Project Votes
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} map[string]interface{} "List of user's votes"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /projects/votes/my-votes [get]
func (h *VoteHandler) GetMyProjectVotes(c *gin.Context) {
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

	votes, err := h.ProjectVoteService.GetProjectVotesByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    votes,
		"count":   len(votes),
	})
}

// GetProjectVoteCount returns the vote count for a specific project
// @Summary Get vote count for project (Public)
// @Description Get the total number of votes for a specific project. No authentication required.
// @Tags Project Votes
// @Accept json
// @Produce json
// @Param projectID path int true "Project ID"
// @Success 200 {object} map[string]interface{} "Vote count"
// @Failure 400 {object} map[string]interface{} "Invalid project ID"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /projects/{projectID}/votes/count [get]
func (h *VoteHandler) GetProjectVoteCount(c *gin.Context) {
	projectIDStr := c.Param("projectID")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{"Invalid Project ID"}})
		return
	}

	count, err := h.ProjectVoteService.GetProjectVoteCount(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"project_id": projectID,
		"vote_count": count,
	})
}

// CheckHasVoted checks if the current user has voted for a specific project
// @Summary Check if user has voted for project
// @Description Check if the current authenticated user has voted for a specific project
// @Tags Project Votes
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param projectID path int true "Project ID"
// @Success 200 {object} map[string]interface{} "Vote status"
// @Failure 400 {object} map[string]interface{} "Invalid project ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /projects/{projectID}/votes/check [get]
func (h *VoteHandler) CheckHasVoted(c *gin.Context) {
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

	status, err := h.ProjectVoteService.CheckUserHasVotedProject(projectID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}
