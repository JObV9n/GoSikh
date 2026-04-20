package handlers

import (
	"errors"
	"log"
	"time"

	errpkg "learning-platform/internal/errors"
	"learning-platform/internal/http/middleware"
	"learning-platform/internal/service"
	"learning-platform/internal/validation"

	"github.com/gin-gonic/gin"
)

type ExecuteHandler struct {
	service    *service.ExecuteService
	submission *service.SubmissionService
}

func NewExecuteHandler(service *service.ExecuteService, submission *service.SubmissionService) *ExecuteHandler {
	return &ExecuteHandler{service: service, submission: submission}
}

// Execute handles code execution requests
// POST /api/run-code
func (h *ExecuteHandler) Execute(c *gin.Context) {
	requestID := c.GetString("requestID")
	start := time.Now()

	var req validation.ExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[EXECUTE] request_id=%s event=reject reason=invalid_request_body path=%s", requestID, c.Request.URL.Path)
		errpkg.BadRequest(c, "invalid request body")
		return
	}

	log.Printf(
		"[EXECUTE] request_id=%s event=start language=%s code_bytes=%d path=%s",
		requestID,
		req.Language,
		len(req.Code),
		c.Request.URL.Path,
	)

	// Validate request
	if ve := req.Validate(); ve.HasErrors() {
		log.Printf("[EXECUTE] request_id=%s event=reject reason=validation_failed path=%s", requestID, c.Request.URL.Path)
		validation.HandleValidationError(c, ve)
		return
	}

	result, err := h.service.Execute(c.Request.Context(), service.ExecuteRequest{
		Language: req.Language,
		Code:     req.Code,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUnsupportedLanguage):
			log.Printf("[EXECUTE] request_id=%s event=failed reason=unsupported_language latency_ms=%d", requestID, time.Since(start).Milliseconds())
			errpkg.BadRequest(c, "unsupported language")
		case errors.Is(err, service.ErrEmptyCode):
			log.Printf("[EXECUTE] request_id=%s event=failed reason=empty_code latency_ms=%d", requestID, time.Since(start).Milliseconds())
			errpkg.BadRequest(c, "code is required")
		default:
			log.Printf("[EXECUTE] request_id=%s event=failed reason=execution_error latency_ms=%d error=%v", requestID, time.Since(start).Milliseconds(), err)
			errpkg.InternalError(c, "execution failed")
		}
		return
	}

	log.Printf(
		"[EXECUTE] request_id=%s event=success language=%s runner=%s sandboxed=%t exit_code=%d timed_out=%t latency_ms=%d stdout_bytes=%d stderr_bytes=%d",
		requestID,
		result.Language,
		result.Runner,
		result.Sandboxed,
		result.ExitCode,
		result.TimedOut,
		time.Since(start).Milliseconds(),
		len(result.Stdout),
		len(result.Stderr),
	)

	errpkg.OK(c, result)
}

// SubmitLesson handles lesson submission, grading, and progress update.
// POST /api/courses/:courseSlug/lessons/:lessonSlug/submit
func (h *ExecuteHandler) SubmitLesson(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		errpkg.Unauthorized(c, "unauthorized")
		return
	}

	courseSlug := c.Param("courseSlug")
	lessonSlug := c.Param("lessonSlug")
	if courseSlug == "" || lessonSlug == "" {
		errpkg.BadRequest(c, "course and lesson slugs are required")
		return
	}

	var req validation.SubmitLessonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errpkg.BadRequest(c, "invalid request body")
		return
	}

	if ve := req.Validate(); ve.HasErrors() {
		validation.HandleValidationError(c, ve)
		return
	}

	result, err := h.submission.SubmitLesson(c.Request.Context(), service.LessonSubmissionRequest{
		UserID:         userID,
		CourseSlug:     courseSlug,
		LessonSlug:     lessonSlug,
		Language:       req.Language,
		Code:           req.Code,
		ExpectedOutput: req.ExpectedOutput,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUnsupportedLanguage):
			errpkg.BadRequest(c, "unsupported language")
		case errors.Is(err, service.ErrEmptyCode):
			errpkg.BadRequest(c, "code is required")
		case errors.Is(err, service.ErrCourseNotFound):
			errpkg.NotFound(c, "course not found")
		case errors.Is(err, service.ErrLessonNotFound):
			errpkg.NotFound(c, "lesson not found")
		default:
			errpkg.InternalError(c, "submission failed")
		}
		return
	}

	errpkg.OK(c, result)
}
