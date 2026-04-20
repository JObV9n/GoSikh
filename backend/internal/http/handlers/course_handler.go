package handlers

import (
	"encoding/json"
	"errors"

	errpkg "learning-platform/internal/errors"
	"learning-platform/internal/http/middleware"
	"learning-platform/internal/http/response"
	"learning-platform/internal/models"
	"learning-platform/internal/service"
	"learning-platform/internal/validation"

	"github.com/gin-gonic/gin"
)

type CourseHandler struct {
	service *service.CourseService
}

func NewCourseHandler(service *service.CourseService) *CourseHandler {
	return &CourseHandler{service: service}
}

func (h *CourseHandler) ListCourses(c *gin.Context) {
	courses, err := h.service.ListCourses()
	if err != nil {
		errpkg.InternalError(c, "failed to list courses")
		return
	}

	errpkg.OK(c, gin.H{"courses": courses})
}

func (h *CourseHandler) GetCourseByID(c *gin.Context) {
	id, ok := validation.ParamInt64(c, "id")
	if !ok {
		return
	}

	course, err := h.service.GetCourseByID(id)
	if err != nil {
		errpkg.InternalError(c, "failed to get course")
		return
	}
	if course == nil {
		errpkg.NotFound(c, "course not found")
		return
	}

	errpkg.OK(c, gin.H{"course": course})
}

func (h *CourseHandler) CreateCourse(c *gin.Context) {
	var req validation.CreateCourseRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		errpkg.BadRequest(c, "invalid request body")
		return
	}

	// Validate request
	if ve := req.Validate(); ve.HasErrors() {
		validation.HandleValidationError(c, ve)
		return
	}

	course := &models.Course{
		Slug:        req.Slug,
		Title:       req.Title,
		Description: req.Description,
	}

	if err := h.service.CreateCourse(course); err != nil {
		errpkg.InternalError(c, "failed to create course")
		return
	}

	errpkg.Created(c, gin.H{"course": course})
}

func (h *CourseHandler) ListLessons(c *gin.Context) {
	courseSlug := c.Param("courseSlug")
	if courseSlug == "" {
		errpkg.BadRequest(c, "course slug is required")
		return
	}

	course, lessons, err := h.service.ListLessonsByCourseSlug(courseSlug)
	if err != nil {
		errpkg.InternalError(c, "failed to list lessons")
		return
	}
	if course == nil {
		errpkg.NotFound(c, "course not found")
		return
	}

	errpkg.OK(c, gin.H{
		"course":  course,
		"lessons": lessons,
	})
}

func (h *CourseHandler) MarkLessonProgress(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		errpkg.Unauthorized(c, "unauthorized")
		return
	}

	// Use direct JSON decoder instead of ShouldBindJSON to reduce allocations
	var req validation.MarkLessonProgressRequest
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&req); err != nil {
		errpkg.BadRequest(c, "invalid request body")
		return
	}

	courseSlug := c.Param("courseSlug")
	lessonSlug := c.Param("lessonSlug")

	if courseSlug == "" || lessonSlug == "" {
		errpkg.BadRequest(c, "course and lesson slugs are required")
		return
	}

	if err := h.service.MarkLessonProgress(userID, courseSlug, lessonSlug, req.Completed); err != nil {
		switch {
		case errors.Is(err, service.ErrCourseNotFound):
			errpkg.NotFound(c, "course not found")
		case errors.Is(err, service.ErrLessonNotFound):
			errpkg.NotFound(c, "lesson not found")
		default:
			errpkg.InternalError(c, "failed to update progress")
		}
		return
	}

	// Use optimized response to minimize allocations
	// For the StatusOK response, use pre-encoded bytes to avoid any JSON encoding allocations
	c.Data(200, "application/json", response.GetStatusOKEnvelopeBytes())
}

func (h *CourseHandler) GetCourseProgress(c *gin.Context) {
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

	course, items, err := h.service.GetCourseProgress(userID, courseSlug)
	if err != nil {
		errpkg.InternalError(c, "failed to fetch progress")
		return
	}
	if course == nil {
		errpkg.NotFound(c, "course not found")
		return
	}

	resp := struct {
		Course   *models.Course          `json:"course"`
		Progress []models.LessonProgress `json:"progress"`
	}{
		Course:   course,
		Progress: items,
	}
	errpkg.OK(c, resp)
}
