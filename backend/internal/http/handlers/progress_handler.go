package handlers

import (
	errpkg "learning-platform/internal/errors"
	"learning-platform/internal/http/middleware"
	"learning-platform/internal/service"
	"learning-platform/internal/validation"

	"github.com/gin-gonic/gin"
)

type ProgressHandler struct {
	service *service.ProgressService
}

func NewProgressHandler(service *service.ProgressService) *ProgressHandler {
	return &ProgressHandler{service: service}
}

// MarkLessonCompleted marks a lesson as completed for the authenticated user
// POST /api/progress/lessons/:lessonID/complete
func (h *ProgressHandler) MarkLessonCompleted(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		errpkg.Unauthorized(c, "unauthorized")
		return
	}

	lessonID, ok := validation.ParamInt64(c, "lessonID")
	if !ok {
		return
	}

	if err := h.service.MarkLessonCompleted(userID, lessonID); err != nil {
		errpkg.InternalError(c, "failed to mark lesson completed")
		return
	}

	errpkg.OK(c, gin.H{"status": "ok", "message": "lesson marked as completed"})
}

// MarkLessonIncomplete marks a lesson as incomplete for the authenticated user
// POST /api/progress/lessons/:lessonID/incomplete
func (h *ProgressHandler) MarkLessonIncomplete(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		errpkg.Unauthorized(c, "unauthorized")
		return
	}

	lessonID, ok := validation.ParamInt64(c, "lessonID")
	if !ok {
		return
	}

	if err := h.service.MarkLessonIncomplete(userID, lessonID); err != nil {
		errpkg.InternalError(c, "failed to mark lesson incomplete")
		return
	}

	errpkg.OK(c, gin.H{"status": "ok", "message": "lesson marked as incomplete"})
}

// GetCourseProgress retrieves progress for a user in a specific course with completion percentage
// GET /api/progress/courses/id/:courseID
func (h *ProgressHandler) GetCourseProgress(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		errpkg.Unauthorized(c, "unauthorized")
		return
	}

	courseID, ok := validation.ParamInt64(c, "courseID")
	if !ok {
		return
	}

	course, progress, percentage, err := h.service.GetUserCourseProgress(userID, courseID)
	if err != nil {
		if err.Error() == "course not found" {
			errpkg.NotFound(c, "course not found")
			return
		}
		errpkg.InternalError(c, "failed to fetch progress")
		return
	}

	errpkg.OK(c, gin.H{
		"course":                course,
		"progress":              progress,
		"completion_percentage": percentage,
	})
}

// GetCourseProgressBySlug retrieves progress for a user in a specific course using slug
// GET /api/progress/courses/:courseSlug/progress
func (h *ProgressHandler) GetCourseProgressBySlug(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		errpkg.Unauthorized(c, "unauthorized")
		return
	}

	courseSlug := c.Param("courseSlug")
	if courseSlug == "" {
		errpkg.BadRequest(c, "course slug is required")
		return
	}

	course, progress, percentage, err := h.service.GetUserCourseProgressBySlug(userID, courseSlug)
	if err != nil {
		if err.Error() == "course not found" {
			errpkg.NotFound(c, "course not found")
			return
		}
		errpkg.InternalError(c, "failed to fetch progress")
		return
	}

	errpkg.OK(c, gin.H{
		"course":                course,
		"progress":              progress,
		"completion_percentage": percentage,
	})
}

// GetAllProgress retrieves all courses with progress and completion percentages for the authenticated user
// GET /api/progress/all
func (h *ProgressHandler) GetAllProgress(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		errpkg.Unauthorized(c, "unauthorized")
		return
	}

	summaries, err := h.service.GetUserAllProgress(userID)
	if err != nil {
		errpkg.InternalError(c, "failed to fetch progress")
		return
	}

	errpkg.OK(c, gin.H{
		"courses": summaries,
	})
}
