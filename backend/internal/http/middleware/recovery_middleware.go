package middleware

import (
	"log"
	"runtime/debug"

	errpkg "learning-platform/internal/errors"

	"github.com/gin-gonic/gin"
)

// RecoveryMiddleware catches panics and returns deterministic JSON errors.
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := generateRequestID(c)
		defer func() {
			if recovered := recover(); recovered != nil {
				if c.GetString("requestID") == "" {
					c.Set("requestID", requestID)
					c.Writer.Header().Set("X-Request-ID", requestID)
				}

				log.Printf(
					"[PANIC] request_id=%s method=%s path=%s error=%v stack=%s",
					requestID,
					c.Request.Method,
					c.Request.URL.Path,
					recovered,
					debug.Stack(),
				)

				errpkg.InternalError(c, "An unexpected error occurred")
				c.Abort()
			}
		}()

		c.Next()
	}
}

// ErrorHandler is kept as a compatibility alias for older router wiring.
func ErrorHandler() gin.HandlerFunc {
	return RecoveryMiddleware()
}

// RecoverHandler is kept as a compatibility alias for older call sites.
func RecoverHandler() gin.HandlerFunc {
	return RecoveryMiddleware()
}
