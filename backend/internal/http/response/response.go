package response

import (
	"encoding/json"
	"net/http"

	"learning-platform/internal/models"

	"github.com/gin-gonic/gin"
)

// StatusResponse is a minimal response for simple acknowledgment (e.g., "ok").
type StatusResponse struct {
	Status string `json:"status"`
}

// CourseListResponse is an optimized response for course list endpoints.
type CourseListResponse struct {
	Courses []*models.Course `json:"courses"`
}

// CourseWithLessonsResponse combines course and lessons in a single typed response.
type CourseWithLessonsResponse struct {
	Course  *models.Course    `json:"course"`
	Lessons []*models.Lesson  `json:"lessons"`
}

// PreallocatedResponses avoids allocating new response objects for static responses.
var (
	// StatusOK is a reusable "ok" response for simple endpoints.
	StatusOK = &StatusResponse{Status: "ok"}

	// These are pre-encoded versions of common responses to avoid per-request allocations
	// These byte slices are immutable and safe to use without allocation
	statusOKEnvelopeBytes = []byte(`{"success":true,"data":{"status":"ok"},"error":null}`)
)

// GetStatusOKEnvelopeBytes returns the pre-encoded successful "ok" response envelope.
// This is byte-perfect compatible with the APIResponse envelope format {success:true, data:{status:"ok"}, error:null}
// Using this avoids all allocations for the response body on the hot path.
func GetStatusOKEnvelopeBytes() []byte {
	return statusOKEnvelopeBytes
}

// DirectJSON writes a response directly as JSON without envelope wrapping.
// Used for hot paths where we want to minimize allocations.
func DirectJSON(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, data)
}

// DirectEnvelopedJSON writes a response with the APIResponse envelope using the optimized pre-encoded bytes
// for StatusOK responses. This significantly reduces allocations on hot paths.
func DirectEnvelopedJSON(c *gin.Context, statusCode int, data *StatusResponse) {
	if data.Status == "ok" {
		c.Data(statusCode, "application/json", statusOKEnvelopeBytes)
		return
	}

	// Fallback for non-standard status
	type APIResponse struct {
		Success bool        `json:"success"`
		Data    interface{} `json:"data"`
		Error   interface{} `json:"error"`
	}
	c.JSON(statusCode, APIResponse{
		Success: true,
		Data:    data,
		Error:   nil,
	})
}

// EnvelopedJSON writes a response with the APIResponse envelope.
// This is the standard response format with success/data/error fields.
func EnvelopedJSON(c *gin.Context, statusCode int, data interface{}) {
	type APIResponse struct {
		Success bool        `json:"success"`
		Data    interface{} `json:"data"`
		Error   interface{} `json:"error"`
	}

	c.JSON(statusCode, APIResponse{
		Success: true,
		Data:    data,
		Error:   nil,
	})
}

// DirectJSONWithStatus writes JSON and returns quickly for minimal-overhead paths.
// Useful for status endpoints or simple acknowledgments.
func DirectJSONWithStatus(c *gin.Context, statusCode int, response interface{}) {
	c.Header("Content-Type", "application/json")
	c.Status(statusCode)
	json.NewEncoder(c.Writer).Encode(response)
}

// RawBytes writes raw JSON bytes directly to the response writer.
// Used for pre-serialized or zero-copy paths.
func RawBytes(c *gin.Context, statusCode int, contentType string, body []byte) {
	c.Data(statusCode, contentType, body)
}

// NoContent writes a 204 No Content response.
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}
