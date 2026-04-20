package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger logs inbound requests and outbound responses.
// Optimized to reduce allocations by using string formatting directly
// instead of creating intermediate string slices.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID := generateRequestID(c)
		c.Set("requestID", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()
		// Optimized: Use Printf directly with fewer intermediate values
		log.Printf("[REQUEST] request_id=%v method=%s path=%s status=%d latency_ms=%d client_ip=%s",
			requestID,
			c.Request.Method,
			c.Request.URL.Path,
			status,
			duration.Milliseconds(),
			c.ClientIP(),
		)
	}
}
