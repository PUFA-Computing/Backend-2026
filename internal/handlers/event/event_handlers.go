package event

import (
	"Backend/internal/handlers/auth"
	"Backend/internal/handlers/user"
	"Backend/internal/models"
	"Backend/internal/services"
	"Backend/pkg/utils"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	EventService      *services.EventService
	PermissionService *services.PermissionService
	AWSService        *services.S3Service
	R2Service         *services.S3Service
}

func NewEventHandlers(eventService *services.EventService, permissionService *services.PermissionService, AWSService *services.S3Service, R2Service *services.S3Service) *Handlers {
	return &Handlers{
		EventService:      eventService,
		PermissionService: permissionService,
		AWSService:        AWSService,
		R2Service:         R2Service,
	}
}

func (h *Handlers) CreateEvent(c *gin.Context) {
	userID, err := (&auth.Handlers{}).ExtractUserIDAndCheckPermission(c, "events:create")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	// Log request body
	log.Println(c.Request.Body)

	parse := c.Request.ParseMultipartForm(10 << 20)
	if parse != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{parse.Error()}})
		return
	}

	log.Println(parse)

	data := c.Request.FormValue("data")
	var newEvent models.Event
	if err := json.Unmarshal([]byte(data), &newEvent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	log.Println(data)

	newEvent.UserID = userID

	if newEvent.Title != "" {
		newEvent.Slug = utils.GenerateFriendlyURL(newEvent.Title)
	}

	if newEvent.StartDate.After(newEvent.EndDate) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{"Start Date cannot be after End Date"}})
		return
	}

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "No file uploaded"})
		return
	}

	log.Println(file)

	optimizedImage, err := utils.OptimizeImage(file, 2800, 1080)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	optimizedImageBytes, err := io.ReadAll(optimizedImage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	// Upload image to R2 storage
	err = h.R2Service.UploadFileToR2(context.Background(), "event", newEvent.Slug, optimizedImageBytes, "image/jpeg")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	newEvent.Thumbnail, _ = h.R2Service.GetFileR2("event", newEvent.Slug)

	if err := h.EventService.CreateEvent(&newEvent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Event Created Successfully",
		"data":    newEvent,
		"relationships": gin.H{
			"author": gin.H{
				"id": userID,
			},
		},
	})
}

func (h *Handlers) EditEvent(c *gin.Context) {
	log.Println("=== EditEvent handler started ===")
	userID, err := (&auth.Handlers{}).ExtractUserIDAndCheckPermission(c, "events:edit")
	if err != nil {
		log.Println("Auth error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}
	log.Println("User authenticated:", userID)

	eventIDStr := c.Param("eventID")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{"Invalid Event ID"}})
		return
	}

	// Get existing event first
	existingEvent, err := h.EventService.GetEventByID(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	// Parse form data - don't return error if no multipart form
	err = c.Request.ParseMultipartForm(10 << 20)
	if err != nil {
		log.Println("Warning: Could not parse multipart form, continuing with regular form data:", err)
		// Continue anyway, as we might just have JSON data without a file
	}

	data := c.Request.FormValue("data")
	if data == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{"No event data provided"}})
		return
	}

	var updatedEvent models.Event
	if err := json.Unmarshal([]byte(data), &updatedEvent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	log.Println("Received updated event data:", updatedEvent)

	if updatedEvent.StartDate.After(updatedEvent.EndDate) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{"Start Date cannot be after End Date"}})
		return
	}

	// Set slug based on title change
	if updatedEvent.Title != "" && updatedEvent.Title != existingEvent.Title {
		updatedEvent.Slug = utils.GenerateFriendlyURL(updatedEvent.Title)
	} else {
		updatedEvent.Slug = existingEvent.Slug
	}

	// Try to get file from form, but don't require it
	file, fileHeader, err := c.Request.FormFile("file")

	// Check if a valid file was uploaded
	if err == nil && fileHeader != nil && fileHeader.Size > 0 && fileHeader.Filename != "empty.txt" {
		// File was uploaded, process it
		log.Println("File uploaded, processing image:", fileHeader.Filename, fileHeader.Size, "bytes")
		optimizedImage, err := utils.OptimizeImage(file, 2800, 1080)
		if err != nil {
			log.Println("Error optimizing image:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
			return
		}

		optimizedImageBytes, err := io.ReadAll(optimizedImage)
		if err != nil {
			log.Println("Error reading optimized image:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
			return
		}

		// Make sure we have a valid slug for the file name
		if updatedEvent.Slug == "" {
			updatedEvent.Slug = existingEvent.Slug
		}

		// Upload new image
		log.Println("Uploading image to R2 with slug:", updatedEvent.Slug)
		err = h.R2Service.UploadFileToR2(context.Background(), "event", updatedEvent.Slug, optimizedImageBytes, "image/jpeg")
		if err != nil {
			log.Println("Error uploading to R2:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
			return
		}

		// Update thumbnail URL
		thumbnailURL, err := h.R2Service.GetFileR2("event", updatedEvent.Slug)
		if err != nil {
			log.Println("Error getting R2 URL:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
			return
		}

		log.Println("New thumbnail URL:", thumbnailURL)
		updatedEvent.Thumbnail = thumbnailURL
	} else {
		// No file uploaded or empty file, keep existing thumbnail
		log.Println("No valid file uploaded, keeping existing thumbnail. Error:", err)
		updatedEvent.Thumbnail = existingEvent.Thumbnail
	}

	utils.ReflectiveUpdate(existingEvent, updatedEvent)

	if err := h.EventService.EditEvent(eventID, existingEvent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Event Updated Successfully",
		"data":    existingEvent,
		"relationships": gin.H{
			"author": gin.H{
				"id": userID,
			},
		},
	})
}

func (h *Handlers) DeleteEvent(c *gin.Context) {
	_, err := (&auth.Handlers{}).ExtractUserIDAndCheckPermission(c, "events:delete")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	eventIDStr := c.Param("eventID")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{"Invalid Event ID"}})
		return
	}

	event, err := h.EventService.GetEventByID(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	// Check image exists on AWS or R2
	exists, _ := h.AWSService.FileExists(context.Background(), "event", event.Slug)
	if exists {
		if err := h.AWSService.DeleteFile(context.Background(), "event", event.Slug); err != nil {
			log.Println("Error deleting file from AWS")
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
			return
		}
	} else {
		exists, _ := h.R2Service.FileExists(context.Background(), "event", event.Slug)
		if exists {
			if err := h.R2Service.DeleteFile(context.Background(), "event", event.Slug); err != nil {
				log.Println("Error deleting file from R2")
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
				return
			}
		}
	}

	if err := h.EventService.DeleteEvent(eventID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Event Deleted Successfully",
	})
}

// GetEventByID retrieves an event by its ID
func (h *Handlers) GetEventByID(c *gin.Context) {
	eventIDStr := c.Param("eventID")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{"Invalid Event ID"}})
		return
	}

	event, err := h.EventService.GetEventByID(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Event Retrieved Successfully",
		"data":    event,
	})
}

// GetEventBySlug retrieves an event by its slug
func (h *Handlers) GetEventBySlug(c *gin.Context) {
	slug := c.Param("eventID")

	event, err := h.EventService.GetEventBySlug(slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Event Retrieved Successfully",
		"data":    event,
	})

}

// ListEvents retrieves a list of events based on the query parameters
func (h *Handlers) ListEvents(c *gin.Context) {
	log.Println("List Events Begin")

	queryParams := map[string]string{
		"organization_id": c.Query("organization_id"),
		"status":          c.Query("status"),
		"page":            c.Query("page"),
	}

	events, totalPages, err := h.EventService.ListEvents(queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"data":         events,
		"totalResults": len(events),
		"totalPages":   totalPages,
	})
}

func (h *Handlers) RegisterForEvent(c *gin.Context) {
	log.Println("Register for Event Begin")
	userID, err := (&auth.Handlers{}).ExtractUserIDAndCheckPermission(c, "events:register")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	// Check Role id and if it is 6 cannot register for event
	roleID, err := (&user.Handlers{}).GetRoleIDByUserID(c, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	if roleID == 8 {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "message": []string{"You are not eligible to this event"}})
		return
	}

	eventIDStr := c.Param("eventID")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{"Invalid Event ID"}})
		return
	}

	log.Println("Register for Event Middle")

	// Handle multipart form for file upload
	form, err := c.MultipartForm()
	if err != nil {
		// If not multipart form, try to bind JSON for backward compatibility
		var eventRegistration models.EventRegistration
		if err := c.BindJSON(&eventRegistration); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{"Invalid request format"}})
			return
		}

		log.Println(eventRegistration.AdditionalNotes)

		if err := h.EventService.RegisterForEvent(userID, eventID, eventRegistration.AdditionalNotes, ""); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Registered Successfully",
			"relationships": gin.H{
				"user": gin.H{
					"data": gin.H{
						"id": userID,
					},
				},
				"event": gin.H{
					"data": gin.H{
						"id": eventID,
					},
				},
			},
		})
		return
	}

	// Get additional notes from form
	additionalNotes := ""
	if len(form.Value["additional_notes"]) > 0 {
		additionalNotes = form.Value["additional_notes"][0]
	}
	log.Println(additionalNotes)

	// Handle file upload
	var filePath string
	files := form.File["file"]
	if len(files) > 0 {
		file, err := files[0].Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{"Error opening uploaded file"}})
			return
		}
		defer file.Close()

		// Read file bytes
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{"Error reading uploaded file"}})
			return
		}

		// Get file content type
		fileType := files[0].Header.Get("Content-Type")
		log.Printf("Uploading file with content type: %s", fileType)

		// Validate file type
		allowedTypes := []string{"application/pdf", "image/jpeg", "image/jpg", "image/png", "application/zip", "application/x-zip-compressed"}
		validType := false
		for _, t := range allowedTypes {
			if t == fileType {
				validType = true
				break
			}
		}

		if !validType {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{"Invalid file type. Only PDF, ZIP, and images (JPEG, PNG) are allowed"}})
			return
		}

		// Generate unique filename with timestamp to prevent collisions
		timestamp := time.Now().UnixNano() / int64(time.Millisecond)
		filename := fmt.Sprintf("event_reg_%d_%s_%d", eventID, userID.String(), timestamp)

		// Log file details before upload
		log.Printf("Uploading file for event registration - Type: %s, Size: %d bytes", fileType, len(fileBytes))

		// Upload to R2 with file type
		err = h.R2Service.UploadFileToR2(context.Background(), "event_registrations", filename, fileBytes, fileType)
		if err != nil {
			log.Printf("Error uploading file to R2: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{"Error uploading file"}})
			return
		}

		// Log successful upload
		log.Printf("File successfully uploaded to R2 for event registration")

		// Get file path
		var err2 error
		filePath, err2 = h.R2Service.GetFileR2("event_registrations", filename)
		if err2 != nil {
			log.Printf("Error getting file path: %v", err2)
			// Continue anyway, we'll just store a blank file path
			filePath = ""
		}
		log.Printf("File path for registration: %s", filePath)
	}

	log.Println("Register for Event Middle 2")

	log.Printf("Attempting to register user %s for event %d with filePath: %s", userID.String(), eventID, filePath)
	if err := h.EventService.RegisterForEvent(userID, eventID, additionalNotes, filePath); err != nil {
		log.Printf("Error registering for event: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	log.Println("Register for Event End")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Registered Successfully",
		"relationships": gin.H{
			"user": gin.H{
				"data": gin.H{
					"id": userID,
				},
			},
			"event": gin.H{
				"data": gin.H{
					"id": eventID,
				},
			},
		},
	})
}

func (h *Handlers) ListRegisteredUsers(c *gin.Context) {
	userID, err := (&auth.Handlers{}).ExtractUserIDAndCheckPermission(c, "events:listRegisteredUsers")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	eventIDStr := c.Param("eventID")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{"Invalid Event ID"}})
		return
	}

	users, err := h.EventService.ListRegisteredUsers(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Registered Users Retrieved Successfully",
		"data":    users,
		"relationships": gin.H{
			"user": gin.H{
				"data": gin.H{
					"id": userID,
				},
			},
			"event": gin.H{
				"data": gin.H{
					"id": eventID,
				},
			},
		},
	})
}

func (h *Handlers) ListEventsRegisteredByUser(c *gin.Context) {
	userID, err := (&auth.Handlers{}).ExtractUserIDAndCheckPermission(c, "users:edit")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	events, err := h.EventService.ListEventsRegisteredByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Registered Events Retrieved Successfully",
		"data":    events,
	})
}

func (h *Handlers) TotalRegisteredUsers(c *gin.Context) {
	eventIDStr := c.Param("eventID")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{"Invalid Event ID"}})
		return
	}

	total, err := h.EventService.TotalRegisteredUsers(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Total Registered Users Retrieved Successfully",
		"data":    total,
	})
}
