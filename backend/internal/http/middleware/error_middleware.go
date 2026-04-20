package middleware

import (
	"fmt"
	"time"

	errpkg "learning-platform/internal/errors"

	"github.com/gin-gonic/gin"
)

// generateRequestID generates a unique request identifier
// Uses Gin's request ID from query params, headers, or generates a new one
func generateRequestID(c *gin.Context) string {
	// Check if request ID already exists in query
	if id := c.Query("request_id"); id != "" {
		return id
	}

	// Check if request ID exists in header
	if id := c.GetHeader("X-Request-ID"); id != "" {
		return id
	}

	// Generate new ID based on timestamp + short hash
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), c.Request.ContentLength)
}

// NotFoundHandler handles 404 errors
func NotFoundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		errpkg.NotFound(c, fmt.Sprintf("The requested endpoint %s %s does not exist", c.Request.Method, c.Request.URL.Path))
	}
}

// MethodNotAllowedHandler handles 405 errors
func MethodNotAllowedHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		errpkg.RespondError(c, 405, "Method Not Allowed", fmt.Sprintf("The HTTP method %s is not allowed for %s", c.Request.Method, c.Request.URL.Path))
	}
}
