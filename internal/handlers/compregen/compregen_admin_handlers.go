package compregen

import (
	dbApp "Backend/internal/database/app"
	"net/http"

	"github.com/gin-gonic/gin"
)

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