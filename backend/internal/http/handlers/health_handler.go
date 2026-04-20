package handlers

import (
	"context"
	"time"

	"learning-platform/internal/database"
	errpkg "learning-platform/internal/errors"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	db *database.DB
}

func NewHealthHandler(db *database.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) Check(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if err := h.db.HealthCheck(ctx); err != nil {
		errpkg.ServiceUnavailable(c, "database is down")
		return
	}

	errpkg.OK(c, gin.H{
		"status": "healthy",
		"db":     "up",
	})
}
