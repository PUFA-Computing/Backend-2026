package compregen

import (
	dbApp "Backend/internal/database/app"
	"Backend/internal/models"
	"Backend/internal/services"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Service   *services.CompregenService
	R2Service *services.S3Service
}

func NewCompregenHandler(s *services.CompregenService, r2 *services.S3Service) *Handler {
	return &Handler{Service: s, R2Service: r2}
}

// GET /compregen/links/:token
func (h *Handler) GetLinkStatus(c *gin.Context) {
	token := c.Param("token")
	status := h.Service.GetLinkStatus(token)
	c.JSON(http.StatusOK, gin.H{"status": status})
}

// POST /compregen/verify
func (h *Handler) Verify(c *gin.Context) {
	var req models.CompregenVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	resp, link, err := h.Service.Verify(req.Token, req.StudentID, req.CampusEmail)
	if err != nil {
		if err.Error() == "rate_limited" {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate_limited", "retry_after": 1800})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if resp.Verified {
		// Issue session cookie scoped to this token
		c.SetCookie("compregen_session", link.ID+":"+req.Token, 3600, "/", "", true, true)
	}

	c.JSON(http.StatusOK, resp)
}

// POST /compregen/register
func (h *Handler) Register(c *gin.Context) {
	var req models.CompregenRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	if !req.ConsentAccepted {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "consent_required"})
		return
	}

	roles := []string{"cp", "vcp1", "vcp2"}
	for _, role := range roles {
		if _, ok := req.Members[role]; !ok {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error":  "validation_failed",
				"fields": gin.H{"members." + role: "missing"},
			})
			return
		}
	}

	registrationID, err := h.Service.Register(req.Token, req.CabinetName, req.Members)
	if err != nil {
		if err.Error() == "link_already_used" {
			c.JSON(http.StatusConflict, gin.H{"error": "link_already_used"})
			return
		}
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "validation_failed", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"registration_id": registrationID})
}

// POST /compregen/upload/photo
func (h *Handler) UploadPhoto(c *gin.Context) {
	token := c.PostForm("token")
	role := c.PostForm("role")

	if token == "" || role == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token and role required"})
		return
	}

	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}
	defer file.Close()

	// Validate mime type
	mimeType := fileHeader.Header.Get("Content-Type")
	allowed := map[string]bool{"image/jpeg": true, "image/png": true, "image/webp": true}
	if !allowed[mimeType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file type"})
		return
	}

	// Max 5MB
	if fileHeader.Size > 5<<20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large"})
		return
	}

	storageKey := "compregen/" + role + "/" + token + "-" + fileHeader.Filename

	fileBytes := make([]byte, fileHeader.Size)
	if _, err := file.Read(fileBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}

	if err := h.R2Service.UploadFileToR2(c.Request.Context(), "compregen", storageKey, fileBytes, mimeType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed"})
		return
	}

	upload, err := dbApp.SavePhotoUpload(storageKey, mimeType, int(fileHeader.Size))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save upload record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"photo_upload_id": upload.ID})
}



// Admin middleware
func AdminKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-Admin-Api-Key")
		expected := os.Getenv("COMPREGEN_ADMIN_API_KEY")
		if key == "" || key != expected {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}