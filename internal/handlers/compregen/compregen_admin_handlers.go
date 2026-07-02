package compregen

import (
	dbApp "Backend/internal/database/app"
	"Backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET /compregen/admin/links/active
func (h *Handler) GetActiveLink(c *gin.Context) {
    link, err := dbApp.GetActiveInviteLink()
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "no active link found"})
        return
    }
    c.JSON(http.StatusOK, gin.H{
        "token":  link.Token,
        "status": link.Status,
        "url":    "https://compsci.president.ac.id/compregen/cp-vcp/" + link.Token,
    })
}

// GET /compregen/admin/attempts
func (h *Handler) GetVerifyAttempts(c *gin.Context) {
    attempts, err := dbApp.GetAllVerifyAttempts()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    if attempts == nil {
        attempts = []map[string]interface{}{}
    }
    c.JSON(http.StatusOK, gin.H{"attempts": attempts})
}

// GET /compregen/admin/registrations
func (h *Handler) GetRegistrations(c *gin.Context) {
	registrations, err := dbApp.GetAllRegistrations()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"registrations": registrations})
}

// POST /compregen/admin/links
func (h *Handler) GenerateLink(c *gin.Context) {
	token, err := h.Service.GenerateInviteLink()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	link, err := dbApp.CreateInviteLink(token, "admin")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save link"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token": link.Token,
		"url":   "https://compsci.president.ac.id/compregen/cp-vcp/" + link.Token,
	})
}

// GET /compregen/admin/whitelist
func (h *Handler) GetWhitelist(c *gin.Context) {
    candidates, err := dbApp.GetAllEligibleCandidates()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"whitelist": candidates})
}

// POST /compregen/admin/whitelist
func (h *Handler) AddWhitelistMember(c *gin.Context) {
    var req models.CompregenAddWhitelistRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    candidate, err := dbApp.AddEligibleCandidate(req.StudentID, req.FullName, req.CampusEmail, req.Major)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, gin.H{"success": true, "member": candidate})
}