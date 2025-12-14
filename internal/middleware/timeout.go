package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// TimeoutMiddleware adds a timeout to all requests
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a context with timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		
		// Replace request context
		c.Request = c.Request.WithContext(ctx)
		
		// Channel to signal completion
		finished := make(chan struct{})
		
		go func() {
			c.Next()
			close(finished)
		}()
		
		select {
		case <-finished:
			// Request completed successfully
			return
		case <-ctx.Done():
			// Timeout occurred
			c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{
				"error":   "Request timeout",
				"message": "The request took too long to process. Please try again.",
			})
			return
		}
	}
}
