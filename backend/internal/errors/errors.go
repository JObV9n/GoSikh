package errors

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// APIResponse is the standard envelope for all API responses.
type APIResponse struct {
	Success bool      `json:"success"`
	Data    any       `json:"data"`
	Error   *APIError `json:"error"`
}

// APIError contains structured error metadata.
type APIError struct {
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Status    int               `json:"status"`
	Path      string            `json:"path"`
	RequestID string            `json:"request_id,omitempty"`
	Timestamp string            `json:"timestamp"`
	Details   map[string]string `json:"details,omitempty"`
}

func requestIDFromContext(c *gin.Context) string {
	if id := c.GetString("requestID"); id != "" {
		return id
	}
	if id := c.GetHeader("X-Request-ID"); id != "" {
		return id
	}
	return ""
}

func statusCodeLabel(statusCode int) string {
	text := http.StatusText(statusCode)
	text = strings.ToUpper(strings.ReplaceAll(text, " ", "_"))
	if text == "" {
		return "UNKNOWN_ERROR"
	}
	return text
}

func buildAPIError(c *gin.Context, statusCode int, message string, details map[string]string) *APIError {
	if message == "" {
		message = http.StatusText(statusCode)
	}

	return &APIError{
		Code:      statusCodeLabel(statusCode),
		Message:   message,
		Status:    statusCode,
		Path:      c.Request.URL.Path,
		RequestID: requestIDFromContext(c),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Details:   details,
	}
}

// Response builder for standard errors
func RespondError(c *gin.Context, statusCode int, errorMsg, message string) {
	requestID := requestIDFromContext(c)
	log.Printf(
		"[ERROR] request_id=%s method=%s path=%s status=%d error=%s message=%s",
		requestID,
		c.Request.Method,
		c.Request.URL.Path,
		statusCode,
		errorMsg,
		message,
	)

	c.JSON(statusCode, APIResponse{
		Success: false,
		Data:    nil,
		Error:   buildAPIError(c, statusCode, message, nil),
	})
}

// Response with additional details
func RespondErrorWithDetails(c *gin.Context, statusCode int, errorMsg, message string, details map[string]string) {
	requestID := requestIDFromContext(c)
	log.Printf(
		"[ERROR] request_id=%s method=%s path=%s status=%d error=%s message=%s details=%v",
		requestID,
		c.Request.Method,
		c.Request.URL.Path,
		statusCode,
		errorMsg,
		message,
		details,
	)

	c.JSON(statusCode, APIResponse{
		Success: false,
		Data:    nil,
		Error:   buildAPIError(c, statusCode, message, details),
	})
}

// Common error responses

func BadRequest(c *gin.Context, message string) {
	RespondError(c, http.StatusBadRequest, "Bad Request", message)
}

func BadRequestWithDetails(c *gin.Context, message string, details map[string]string) {
	RespondErrorWithDetails(c, http.StatusBadRequest, "Bad Request", message, details)
}

func Unauthorized(c *gin.Context, message string) {
	RespondError(c, http.StatusUnauthorized, "Unauthorized", message)
}

func Forbidden(c *gin.Context, message string) {
	RespondError(c, http.StatusForbidden, "Forbidden", message)
}

func NotFound(c *gin.Context, message string) {
	RespondError(c, http.StatusNotFound, "Not Found", message)
}

func Conflict(c *gin.Context, message string) {
	RespondError(c, http.StatusConflict, "Conflict", message)
}

func InternalError(c *gin.Context, message string) {
	RespondError(c, http.StatusInternalServerError, "Internal Server Error", message)
}

func ServiceUnavailable(c *gin.Context, message string) {
	RespondError(c, http.StatusServiceUnavailable, "Service Unavailable", message)
}

// ValidationError for form/input validation errors
func ValidationError(c *gin.Context, details map[string]string) {
	RespondErrorWithDetails(
		c,
		http.StatusBadRequest,
		"Validation Error",
		"One or more fields failed validation",
		details,
	)
}

func RespondSuccess(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Data:    data,
		Error:   nil,
	})
}

func RespondSuccessWithMessage(c *gin.Context, statusCode int, message string, data interface{}) {
	body := data
	if message != "" {
		if dataMap, ok := data.(gin.H); ok {
			dataMap["message"] = message
			body = dataMap
		} else {
			body = gin.H{"result": data, "message": message}
		}
	}

	c.JSON(statusCode, APIResponse{
		Success: true,
		Data:    body,
		Error:   nil,
	})
}

// Quick helpers
func OK(c *gin.Context, data interface{}) {
	RespondSuccess(c, http.StatusOK, data)
}

func Created(c *gin.Context, data interface{}) {
	RespondSuccess(c, http.StatusCreated, data)
}

func CreatedWithMessage(c *gin.Context, message string, data interface{}) {
	RespondSuccessWithMessage(c, http.StatusCreated, message, data)
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// OKWithPreEncoded writes a pre-encoded response for ultra-low-allocation responses.
// This is used for hot paths where the response data is always the same (like {"status":"ok"}).
//
// This bypasses the normal envelope wrapping and uses raw bytes to avoid allocations.
func OKWithRawEnvelope(c *gin.Context, envelope []byte) {
	c.Data(http.StatusOK, "application/json", envelope)
}

