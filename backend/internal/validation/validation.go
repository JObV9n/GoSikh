package validation

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// Email regex pattern
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	// Slug regex pattern (alphanumeric and hyphens)
	slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
)

// ValidationError represents validation errors with field-level details
type ValidationError struct {
	Fields map[string]string // field name -> error message
}

func (ve *ValidationError) Error() string {
	var msgs []string
	for field, msg := range ve.Fields {
		msgs = append(msgs, fmt.Sprintf("%s: %s", field, msg))
	}
	return strings.Join(msgs, "; ")
}

// Add adds a validation error for a field
func (ve *ValidationError) Add(field, message string) {
	if ve.Fields == nil {
		ve.Fields = make(map[string]string)
	}
	if _, exists := ve.Fields[field]; !exists {
		ve.Fields[field] = message
	}
}

// HasErrors returns true if there are any validation errors
func (ve *ValidationError) HasErrors() bool {
	return ve != nil && len(ve.Fields) > 0
}

// Email validates email format
func Email(email string) error {
	if email = strings.TrimSpace(email); email == "" {
		return fmt.Errorf("email is required")
	}
	if len(email) > 254 {
		return fmt.Errorf("email must be at most 254 characters")
	}
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("email format is invalid")
	}
	return nil
}

// Password validates password strength
func Password(password string) error {
	if password == "" {
		return fmt.Errorf("password is required")
	}
	if len(password) < 6 {
		return fmt.Errorf("password must be at least 6 characters")
	}
	if len(password) > 128 {
		return fmt.Errorf("password must be at most 128 characters")
	}
	return nil
}

// String validates string field
func String(value string, fieldName string, minLen, maxLen int) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	if len(value) < minLen {
		return fmt.Errorf("%s must be at least %d characters", fieldName, minLen)
	}
	if maxLen > 0 && len(value) > maxLen {
		return fmt.Errorf("%s must be at most %d characters", fieldName, maxLen)
	}
	return nil
}

// Slug validates slug format (lowercase alphanumeric with hyphens)
func Slug(slug string) error {
	if slug = strings.TrimSpace(slug); slug == "" {
		return fmt.Errorf("slug is required")
	}
	if len(slug) < 1 {
		return fmt.Errorf("slug must have at least 1 character")
	}
	if len(slug) > 100 {
		return fmt.Errorf("slug must be at most 100 characters")
	}
	if !slugRegex.MatchString(slug) {
		return fmt.Errorf("slug must contain only lowercase letters, numbers, and hyphens")
	}
	return nil
}

// IntInRange validates integer is within range
func IntInRange(value int64, fieldName string, min, max int64) error {
	if value < min {
		return fmt.Errorf("%s must be at least %d", fieldName, min)
	}
	if max >= 0 && value > max {
		return fmt.Errorf("%s must be at most %d", fieldName, max)
	}
	return nil
}

// PositiveInt validates that integer is positive
func PositiveInt(value int64, fieldName string) error {
	if value <= 0 {
		return fmt.Errorf("%s must be a positive number", fieldName)
	}
	return nil
}

// NonNegativeInt validates that integer is non-negative
func NonNegativeInt(value int64, fieldName string) error {
	if value < 0 {
		return fmt.Errorf("%s must be a non-negative number", fieldName)
	}
	return nil
}

// OneOf validates that value is one of allowed options
func OneOf(value string, fieldName string, allowed ...string) error {
	if value == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	for _, opt := range allowed {
		if value == opt {
			return nil
		}
	}
	return fmt.Errorf("%s must be one of: %s", fieldName, strings.Join(allowed, ", "))
}

// CodeLanguage validates supported programming languages
func CodeLanguage(language string) error {
	allowedLanguages := []string{"javascript", "python", "go", "c", "php", "sql"}
	return OneOf(language, "language", allowedLanguages...)
}

// Code validates code content
func Code(code string) error {
	if strings.TrimSpace(code) == "" {
		return fmt.Errorf("code is required")
	}
	if len(code) > 100000 {
		return fmt.Errorf("code must be at most 100000 characters")
	}
	return nil
}
