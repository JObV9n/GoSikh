package validation

// RegisterRequest is the request payload for user registration
type RegisterRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Validate validates the register request
func (r *RegisterRequest) Validate() *ValidationError {
	ve := &ValidationError{}

	if err := Email(r.Email); err != nil {
		ve.Add("email", err.Error())
	}

	if err := Password(r.Password); err != nil {
		ve.Add("password", err.Error())
	}

	if ve.HasErrors() {
		return ve
	}
	return nil
}

// LoginRequest is the request payload for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Validate validates the login request
func (r *LoginRequest) Validate() *ValidationError {
	ve := &ValidationError{}

	if err := Email(r.Email); err != nil {
		ve.Add("email", err.Error())
	}

	if r.Password == "" {
		ve.Add("password", "password is required")
	}

	if ve.HasErrors() {
		return ve
	}
	return nil
}

// CreateCourseRequest is the request payload for creating a course
type CreateCourseRequest struct {
	Slug        string `json:"slug" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

// Validate validates the create course request
func (r *CreateCourseRequest) Validate() *ValidationError {
	ve := &ValidationError{}

	if err := Slug(r.Slug); err != nil {
		ve.Add("slug", err.Error())
	}

	if err := String(r.Title, "title", 1, 255); err != nil {
		ve.Add("title", err.Error())
	}

	if r.Description != "" {
		if err := String(r.Description, "description", 0, 2000); err != nil {
			ve.Add("description", err.Error())
		}
	}

	if ve.HasErrors() {
		return ve
	}
	return nil
}

// CreateLessonRequest is the request payload for creating a lesson
type CreateLessonRequest struct {
	Slug     string `json:"slug" binding:"required"`
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content" binding:"required"`
	Position int    `json:"position"`
}

// Validate validates the create lesson request
func (r *CreateLessonRequest) Validate() *ValidationError {
	ve := &ValidationError{}

	if err := Slug(r.Slug); err != nil {
		ve.Add("slug", err.Error())
	}

	if err := String(r.Title, "title", 1, 255); err != nil {
		ve.Add("title", err.Error())
	}

	if err := String(r.Content, "content", 1, 50000); err != nil {
		ve.Add("content", err.Error())
	}

	if r.Position > 0 {
		if err := NonNegativeInt(int64(r.Position), "position"); err != nil {
			ve.Add("position", err.Error())
		}
	}

	if ve.HasErrors() {
		return ve
	}
	return nil
}

// ExecuteRequest is the request payload for code execution
type ExecuteRequest struct {
	Language string `json:"language" binding:"required"`
	Code     string `json:"code" binding:"required"`
}

// Validate validates the execute request
func (r *ExecuteRequest) Validate() *ValidationError {
	ve := &ValidationError{}

	if err := CodeLanguage(r.Language); err != nil {
		ve.Add("language", err.Error())
	}

	if err := Code(r.Code); err != nil {
		ve.Add("code", err.Error())
	}

	if ve.HasErrors() {
		return ve
	}
	return nil
}

// SubmitLessonRequest is the request payload for lesson submission and grading.
type SubmitLessonRequest struct {
	Language       string `json:"language" binding:"required"`
	Code           string `json:"code" binding:"required"`
	ExpectedOutput string `json:"expected_output"`
}

// Validate validates the lesson submission request.
func (r *SubmitLessonRequest) Validate() *ValidationError {
	ve := &ValidationError{}

	if err := CodeLanguage(r.Language); err != nil {
		ve.Add("language", err.Error())
	}

	if err := Code(r.Code); err != nil {
		ve.Add("code", err.Error())
	}

	if ve.HasErrors() {
		return ve
	}

	return nil
}

// MarkLessonProgressRequest is the request payload for marking lesson progress
type MarkLessonProgressRequest struct {
	Completed bool `json:"completed"`
}

// Validate validates the mark lesson progress request
func (r *MarkLessonProgressRequest) Validate() *ValidationError {
	// No validation needed for boolean flag
	return nil
}
