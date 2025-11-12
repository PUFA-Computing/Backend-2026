package vote

import (
	"Backend/internal/handlers/auth"
	"Backend/internal/models"
	"Backend/internal/services"
	"Backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"strconv"
)

type Handler struct {
	VoteService       *services.VoteService
	PermissionService *services.PermissionService
}

func NewVoteHandler(voteService *services.VoteService, permissionService *services.PermissionService) *Handler {
	return &Handler{
		VoteService:       voteService,
		PermissionService: permissionService,
	}
}

// CastVote allows a user to vote for a candidate
// @Summary Cast a vote (Year 2025 users only)
// @Description Cast a vote for a candidate. Any user with year 2025 can vote once for a candidate from their major. One vote per user, same major required.
// @Tags Votes
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token (Year 2025 user)"
// @Param request body models.CastVoteRequest true "Vote request"
// @Success 201 {object} map[string]interface{} "Vote cast successfully"
// @Failure 400 {object} map[string]interface{} "Bad request, already voted, not year 2025, or different major"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /votes/cast [post]
func (h *Handler) CastVote(c *gin.Context) {
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

	var req models.CastVoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	vote, err := h.VoteService.CastVote(userID, req.CandidateID)
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

// GetVoteStatus checks if a user has voted
// @Summary Check vote status
// @Description Check if the current user has already voted
// @Tags Votes
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} map[string]interface{} "Vote status"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /votes/status [get]
func (h *Handler) GetVoteStatus(c *gin.Context) {
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

	status, err := h.VoteService.GetVoteStatus(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}

// GetMyVote retrieves the current user's vote
// @Summary Get my vote
// @Description Get the current user's vote details
// @Tags Votes
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} map[string]interface{} "User's vote details"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Vote not found"
// @Security BearerAuth
// @Router /votes/my-vote [get]
func (h *Handler) GetMyVote(c *gin.Context) {
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

	vote, err := h.VoteService.GetVoteByVoterID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": []string{"You have not voted yet"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    vote,
	})
}

// ListVotes retrieves all votes (admin only)
// @Summary List all votes
// @Description Get a paginated list of all votes (requires vote:view permission)
// @Tags Votes
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param candidate_id query int false "Filter by candidate ID"
// @Param page query int false "Page number for pagination" default(1)
// @Success 200 {object} map[string]interface{} "List of votes"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /votes [get]
func (h *Handler) ListVotes(c *gin.Context) {
	_, err := (&auth.Handlers{}).ExtractUserIDAndCheckPermission(c, "vote:view")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	queryParams := make(map[string]string)
	queryParams["candidate_id"] = c.Query("candidate_id")
	queryParams["page"] = c.Query("page")

	votes, totalPages, err := h.VoteService.ListVotes(queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"data":         votes,
		"totalResults": len(votes),
		"totalPages":   totalPages,
	})
}

// GetVoteCountByCandidate returns the vote count for a specific candidate
// @Summary Get vote count for candidate (Public)
// @Description Get the total number of votes for a specific candidate. No authentication required.
// @Tags Votes
// @Accept json
// @Produce json
// @Param candidateID path int true "Candidate ID"
// @Success 200 {object} map[string]interface{} "Vote count"
// @Failure 400 {object} map[string]interface{} "Invalid candidate ID"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /votes/candidate/{candidateID}/count [get]
func (h *Handler) GetVoteCountByCandidate(c *gin.Context) {
	candidateIDStr := c.Param("candidateID")
	candidateID, err := strconv.Atoi(candidateIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{"Invalid Candidate ID"}})
		return
	}

	count, err := h.VoteService.GetVoteCountByCandidate(candidateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"candidate_id": candidateID,
		"vote_count":   count,
	})
}

// GetVotesByCandidateID retrieves all votes for a specific candidate (admin only)
// @Summary Get votes by candidate
// @Description Get all votes for a specific candidate (requires vote:view permission)
// @Tags Votes
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param candidateID path int true "Candidate ID"
// @Success 200 {object} map[string]interface{} "List of votes for the candidate"
// @Failure 400 {object} map[string]interface{} "Invalid candidate ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /votes/candidate/{candidateID} [get]
func (h *Handler) GetVotesByCandidateID(c *gin.Context) {
	_, err := (&auth.Handlers{}).ExtractUserIDAndCheckPermission(c, "vote:view")
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

	votes, err := h.VoteService.GetVotesByCandidateID(candidateID)
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

// DeleteVote removes a vote (admin only)
// @Summary Delete a vote
// @Description Delete a vote by ID (requires vote:delete permission)
// @Tags Votes
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param voteID path int true "Vote ID"
// @Success 200 {object} map[string]interface{} "Vote deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid vote ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /votes/{voteID}/delete [delete]
func (h *Handler) DeleteVote(c *gin.Context) {
	_, err := (&auth.Handlers{}).ExtractUserIDAndCheckPermission(c, "vote:delete")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	voteIDStr := c.Param("voteID")
	voteID, err := strconv.Atoi(voteIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{"Invalid Vote ID"}})
		return
	}

	if err := h.VoteService.DeleteVote(voteID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Vote Deleted Successfully",
	})
}

// CheckCanVote checks if the current user can vote
// @Summary Check if user can vote
// @Description Check if the current user is eligible to vote (year must be 2025, hasn't voted yet, valid major)
// @Tags Votes
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} map[string]interface{} "Eligibility status"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /votes/can-vote [get]
func (h *Handler) CheckCanVote(c *gin.Context) {
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

	canVote, message, err := h.VoteService.CheckUserCanVote(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"can_vote": canVote,
		"message":  message,
	})
}
