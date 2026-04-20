package validation

import (
	"learning-platform/internal/errors"

	"github.com/gin-gonic/gin"
)

// HandleValidationError converts validation errors to HTTP responses
func HandleValidationError(c *gin.Context, ve *ValidationError) bool {
	if !ve.HasErrors() {
		return false // No errors, continue processing
	}

	errors.BadRequestWithDetails(c, "validation failed", ve.Fields)
	return true // Error handled, stop processing
}

// BindAndValidate binds JSON request and validates it
// Returns true if there were errors (meaning handler should return early)
func BindAndValidate(c *gin.Context, req interface {
	Validate() *ValidationError
}) bool {
	// First, bind JSON
	if err := c.ShouldBindJSON(req); err != nil {
		errors.BadRequest(c, "invalid request body")
		return true
	}

	// Then, validate
	if ve := req.Validate(); ve.HasErrors() {
		HandleValidationError(c, ve)
		return true
	}

	return false // No errors
}

// ParamInt64 extracts and validates an int64 parameter
func ParamInt64(c *gin.Context, paramName string) (int64, bool) {
	var value int64
	if _, err := parseParamInt64(c.Param(paramName), &value); err != nil {
		errors.BadRequest(c, "invalid "+paramName)
		return 0, false
	}
	return value, true
}

// QueryInt64 extracts and validates an int64 query parameter
func QueryInt64(c *gin.Context, queryName string) (int64, bool) {
	queryValue := c.DefaultQuery(queryName, "")
	if queryValue == "" {
		return 0, true // Optional query param not provided
	}

	var value int64
	if _, err := parseParamInt64(queryValue, &value); err != nil {
		errors.BadRequest(c, "invalid "+queryName)
		return 0, false
	}
	return value, true
}

// Helper function to parse int64 from string
func parseParamInt64(value string, result *int64) (int64, error) {
	var parsed int64
	_, err := simpleParseInt(value, &parsed)
	if err != nil {
		return 0, err
	}
	*result = parsed
	return parsed, nil
}

// Simple int64 parser using basic arithmetic
func simpleParseInt(s string, result *int64) (int64, error) {
	if s == "" {
		return 0, ErrInvalidFormat
	}

	var negative bool
	var startIdx int

	if s[0] == '-' {
		negative = true
		startIdx = 1
	}

	if startIdx >= len(s) {
		return 0, ErrInvalidFormat
	}

	var value int64
	for i := startIdx; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return 0, ErrInvalidFormat
		}
		value = value*10 + int64(c-'0')
	}

	if negative {
		value = -value
	}

	*result = value
	return value, nil
}

// Custom error types
var (
	ErrInvalidFormat = errInvalidFormat{}
)

type errInvalidFormat struct{}

func (e errInvalidFormat) Error() string {
	return "invalid format"
}
