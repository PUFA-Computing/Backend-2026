package news

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
	"time"
)

type Handler struct {
	NewsService       *services.NewsService
	PermissionService *services.PermissionService
	AWSService        *services.S3Service
	R2Service         *services.S3Service
}

func NewNewsHandler(newsService *services.NewsService, permissionService *services.PermissionService, AWSService *services.S3Service, R2Service *services.S3Service) *Handler {
	return &Handler{
		NewsService:       newsService,
		PermissionService: permissionService,
		AWSService:        AWSService,
		R2Service:         R2Service,
	}
}

func (h *Handler) CreateNews(c *gin.Context) {
	userID, err := (&auth.Handlers{}).ExtractUserIDAndCheckPermission(c, "news:create")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	data := c.Request.FormValue("data")
	var newNews models.News
	if err := json.Unmarshal([]byte(data), &newNews); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	newNews.UserID = userID

	if newNews.Title != "" {
		newNews.Slug = utils.GenerateFriendlyURL(newNews.Title)
	}

	if newNews.PublishDate.IsZero() {
		newNews.PublishDate = time.Now()
	}

	// Check if a file was uploaded
	file, fileHeader, err := c.Request.FormFile("file")
	
	// If error is not because of missing file, return error
	if err != nil && err != http.ErrMissingFile {
		
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{err.Error()}})
		
		return
	
	}

	
	// Choose storage service to upload image to (AWS or R2)
	upload := utils.ChooseStorageService()

	
	
	// Check if file was uploaded
	if err == nil && fileHeader != nil {
		
		// Process the uploaded image
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

		
		
		if upload == utils.R2Service {
			
			err = h.R2Service.UploadFileToR2(context.Background(), "news", newNews.Slug, optimizedImageBytes, "image/jpeg")
			
			if err != nil {
				
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
				
				return
			
			}

			
			newNews.Thumbnail, _ = h.R2Service.GetFileR2("news", newNews.Slug)
		
		} else {
			
			err = h.AWSService.UploadFileToAWS(context.Background(), "news", newNews.Slug, optimizedImageBytes)
			
			if err != nil {
				
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
				
				return
			
			}

			
			newNews.Thumbnail, _ = h.AWSService.GetFileAWS("news", newNews.Slug)
		
		}
	
	} else {
		
		// No image uploaded, use default image
		if upload == utils.R2Service {
			
			// Use the default news thumbnail from R2
			newNews.Thumbnail, _ = h.R2Service.GetFileR2("default", "news-thumbnail")
		
		} else {
			
			// Use the default news thumbnail from AWS
			newNews.Thumbnail, _ = h.AWSService.GetFileAWS("default", "news-thumbnail")
		
		}
	
	}

	if err := h.NewsService.CreateNews(&newNews); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "News Created Successfully",
		"data":    newNews,
	})
}

func (h *Handler) GetNewsByID(c *gin.Context) {
	newsIDStr := c.Param("newsID")
	newsID, err := strconv.Atoi(newsIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{"Invalid News ID"}})
		return
	}

	news, err := h.NewsService.GetNewsByID(newsID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": []string{"News not found"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "News Retrieved Successfully",
		"data":    news,
	})
}

func (h *Handler) GetNewsBySlug(c *gin.Context) {
	newsSlug := c.Param("newsID")

	news, err := h.NewsService.GetNewsBySlug(newsSlug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": []string{"News not found"}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "News Retrieved Successfully",
		"data":    news,
	})

}

func (h *Handler) EditNews(c *gin.Context) {
	_, err := (&auth.Handlers{}).ExtractUserIDAndCheckPermission(c, "news:edit")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	newsIDStr := c.Param("newsID")
	newsID, err := strconv.Atoi(newsIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{"Invalid News ID"}})
		return
	}

	existingNews, err := h.NewsService.GetNewsByID(newsID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": []string{"News not found"}})
		return
	}

	// Parse form data
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		log.Printf("Error parsing multipart form: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	// Get the updated news data
	data := c.Request.FormValue("data")
	if data == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{"No data provided"}})
		return
	}

	log.Printf("Received data: %s", data)

	var updatedNews models.News
	if err := json.Unmarshal([]byte(data), &updatedNews); err != nil {
		log.Printf("Error unmarshaling JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	// Handle slug generation
	if updatedNews.Title != "" && updatedNews.Title != existingNews.Title {
		updatedNews.Slug = utils.GenerateFriendlyURL(updatedNews.Title)
		log.Printf("Generated new slug: %s", updatedNews.Slug)
	} else {
		updatedNews.Slug = existingNews.Slug
		log.Printf("Using existing slug: %s", updatedNews.Slug)
	}

	// Check if a new file is being uploaded
	file, fileHeader, err := c.Request.FormFile("file")
	hasNewImage := err == nil && fileHeader != nil

	log.Printf("Has new image: %v", hasNewImage)

	// Only process image if a new one is provided
	if hasNewImage {
		log.Printf("Processing new image")
		defer file.Close()

		optimizedImage, err := utils.OptimizeImage(file, 2800, 1080)
		if err != nil {
			log.Printf("Error optimizing image: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
			return
		}

		optimizedImageBytes, err := io.ReadAll(optimizedImage)
		if err != nil {
			log.Printf("Error reading optimized image: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
			return
		}

		// Upload image to R2 storage with the correct slug
		log.Printf("Uploading image to R2 with slug: %s", updatedNews.Slug)
		err = h.R2Service.UploadFileToR2(context.Background(), "news", updatedNews.Slug, optimizedImageBytes, "image/jpeg")
		if err != nil {
			log.Printf("Error uploading to R2: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
			return
		}

		// Get the correct URL from R2 service
		thumbnailURL, err := h.R2Service.GetFileR2("news", updatedNews.Slug)
		if err != nil {
			log.Printf("Error getting thumbnail URL: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{"Failed to get thumbnail URL: " + err.Error()}})
			return
		}

		log.Printf("Setting thumbnail URL: %s", thumbnailURL)
		updatedNews.Thumbnail = thumbnailURL
	} else {
		// No new image, keep the existing thumbnail
		log.Printf("No new image provided, keeping existing thumbnail: %s", existingNews.Thumbnail)
		updatedNews.Thumbnail = existingNews.Thumbnail
	}

	// Update the news in the database
	utils.ReflectiveUpdate(existingNews, &updatedNews)

	if err := h.NewsService.EditNews(newsID, &updatedNews); err != nil {
		log.Printf("Error updating news in database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	log.Printf("News updated successfully with ID: %d", newsID)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "News Updated Successfully",
		"data":    existingNews,
	})
}

func (h *Handler) DeleteNews(c *gin.Context) {
	_, err := (&auth.Handlers{}).ExtractUserIDAndCheckPermission(c, "news:delete")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	newsIDStr := c.Param("newsID")
	newsID, err := strconv.Atoi(newsIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{"Invalid News ID"}})
		return
	}

	news, err := h.NewsService.GetNewsByID(newsID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": []string{"News not found"}})
		return
	}

	// Check image exists on AWS or R2
	exists, _ := h.AWSService.FileExists(context.Background(), "news", news.Slug)
	if exists {
		if err := h.AWSService.DeleteFile(context.Background(), "news", news.Slug); err != nil {
			log.Println("Error deleting file from AWS")
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
			return
		}
	} else {
		exists, _ := h.R2Service.FileExists(context.Background(), "news", news.Slug)
		if exists {
			if err := h.R2Service.DeleteFile(context.Background(), "news", news.Slug); err != nil {
				log.Println("Error deleting file from R2")
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
				return
			}
		}
	}

	if err := h.NewsService.DeleteNews(newsID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "News Deleted Successfully",
	})
}

func (h *Handler) ListNews(c *gin.Context) {
	queryParams := make(map[string]string)
	queryParams["organization_id"] = c.Query("organization_id")
	queryParams["page"] = c.Query("page")

	news, totalPages, err := h.NewsService.ListNews(queryParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"data":         news,
		"totalResults": len(news),
		"totalPages":   totalPages,
	})
}

func (h *Handler) LikeNews(c *gin.Context) {
	token, err := utils.ExtractTokenFromHeader(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}
	userID, err := utils.GetUserIDFromToken(token, os.Getenv("JWT_SECRET_KEY"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	newsIDStr := c.Param("newsID")
	newsID, err := strconv.Atoi(newsIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{"Invalid News ID"}})
		return
	}

	if err := h.NewsService.LikeNews(userID, newsID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "News Liked Successfully",
	})
}

//func (h *Handler) UnlikeNews(c *gin.Context) {
//	token, err := utils.ExtractTokenFromHeader(c)
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
//		return
//	}
//	userID, err := utils.GetUserIDFromToken(token, os.Getenv("JWT_SECRET_KEY"))
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
//		return
//	}
//
//	newsIDStr := c.Param("newsID")
//	newsID, err := strconv.Atoi(newsIDStr)
//	if err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": []string{"Invalid News ID"}})
//		return
//	}
//
//	if err := h.NewsService.UnlikeNews(userID, newsID); err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": []string{err.Error()}})
//		return
//	}
//
//	c.JSON(http.StatusOK, gin.H{
//		"success": true,
//		"message": "News Unliked Successfully",
//	})
//}
