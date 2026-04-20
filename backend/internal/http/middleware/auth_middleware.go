package middleware

import (
	"strings"

	"learning-platform/internal/auth"
	errpkg "learning-platform/internal/errors"

	"github.com/gin-gonic/gin"
)

const userIDContextKey = "userID"
const userEmailContextKey = "userEmail"
const claimsContextKey = "claims"

func RequireAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authorization := c.GetHeader("Authorization")
		if authorization == "" {
			errpkg.Unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authorization, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			errpkg.Unauthorized(c, "invalid authorization header")
			c.Abort()
			return
		}

		claims, err := auth.ParseToken(secret, parts[1])
		if err != nil {
			errpkg.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}

		if claims.UserID <= 0 {
			errpkg.Unauthorized(c, "invalid token claims")
			c.Abort()
			return
		}

		// Cache the full claims object to avoid re-parsing elsewhere
		c.Set(claimsContextKey, claims)
		c.Set(userIDContextKey, claims.UserID)
		c.Set(userEmailContextKey, claims.Email)
		c.Next()
	}
}

// AuthMiddleware is a semantic alias for RequireAuth.
func AuthMiddleware(secret string) gin.HandlerFunc {
	return RequireAuth(secret)
}

func UserID(c *gin.Context) (int64, bool) {
	value, ok := c.Get(userIDContextKey)
	if !ok {
		return 0, false
	}

	userID, ok := value.(int64)
	return userID, ok
}

func UserEmail(c *gin.Context) (string, bool) {
	value, ok := c.Get(userEmailContextKey)
	if !ok {
		return "", false
	}

	email, ok := value.(string)
	return email, ok
}

// Claims returns the cached parsed JWT claims from the context.
// This avoids re-parsing the token and reduces allocations.
func Claims(c *gin.Context) (*auth.Claims, bool) {
	value, ok := c.Get(claimsContextKey)
	if !ok {
		return nil, false
	}

	claims, ok := value.(*auth.Claims)
	return claims, ok
}
