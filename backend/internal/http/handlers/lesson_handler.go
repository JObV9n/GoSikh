package handlers

import (
	errpkg "learning-platform/internal/errors"
	"learning-platform/internal/models"
	"learning-platform/internal/service"
	"learning-platform/internal/validation"

	"github.com/gin-gonic/gin"
)

type LessonHandler struct {
	service *service.CourseService
}

func NewLessonHandler(service *service.CourseService) *LessonHandler {
	return &LessonHandler{service: service}
}

func (h *LessonHandler) CreateLesson(c *gin.Context) {
	var req validation.CreateLessonRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		errpkg.BadRequest(c, "invalid request body")
		return
	}

	// Validate request
	if ve := req.Validate(); ve.HasErrors() {
		validation.HandleValidationError(c, ve)
		return
	}

	courseSlug := c.Param("courseSlug")
	if courseSlug == "" {
		errpkg.BadRequest(c, "course slug is required")
		return
	}

	// Get the course by slug to get its ID
	course, err := h.service.GetCourseBySlug(courseSlug)
	if err != nil {
		errpkg.InternalError(c, "failed to get course")
		return
	}
	if course == nil {
		errpkg.NotFound(c, "course not found")
		return
	}

	lesson := &models.Lesson{
		CourseID: course.ID,
		Slug:     req.Slug,
		Title:    req.Title,
		Content:  req.Content,
		Position: req.Position,
	}

	if err := h.service.CreateLesson(lesson); err != nil {
		errpkg.InternalError(c, "failed to create lesson")
		return
	}

	errpkg.Created(c, gin.H{"lesson": lesson})
}

func (h *LessonHandler) GetLessonByID(c *gin.Context) {
	id, ok := validation.ParamInt64(c, "id")
	if !ok {
		return
	}

	lesson, err := h.service.GetLessonByID(id)
	if err != nil {
		errpkg.InternalError(c, "failed to get lesson")
		return
	}
	if lesson == nil {
		errpkg.NotFound(c, "lesson not found")
		return
	}

	errpkg.OK(c, gin.H{"lesson": lesson})
}
