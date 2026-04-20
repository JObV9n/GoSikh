package handlers

import (
	"errors"

	"learning-platform/internal/auth"
	errpkg "learning-platform/internal/errors"
	"learning-platform/internal/service"
	"learning-platform/internal/validation"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
	jwtSecret   string
}

func NewAuthHandler(authService *service.AuthService, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		jwtSecret:   jwtSecret,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req validation.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errpkg.BadRequest(c, "invalid request body")
		return
	}

	// Validate request
	if ve := req.Validate(); ve.HasErrors() {
		validation.HandleValidationError(c, ve)
		return
	}

	user, err := h.authService.Register(req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmailInUse):
			errpkg.Conflict(c, "email already in use")
		case errors.Is(err, service.ErrInvalidEmail), errors.Is(err, service.ErrInvalidPassword):
			errpkg.BadRequest(c, err.Error())
		default:
			errpkg.InternalError(c, "failed to register user")
		}
		return
	}

	token, err := auth.GenerateToken(h.jwtSecret, user.ID, user.Email)
	if err != nil {
		errpkg.InternalError(c, "failed to create token")
		return
	}

	errpkg.Created(c, gin.H{
		"token": token,
		"user": gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"created_at": user.CreatedAt,
		},
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req validation.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errpkg.BadRequest(c, "invalid request body")
		return
	}

	// Validate request
	if ve := req.Validate(); ve.HasErrors() {
		validation.HandleValidationError(c, ve)
		return
	}

	user, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			errpkg.Unauthorized(c, "invalid email or password")
			return
		}
		errpkg.InternalError(c, "failed to login")
		return
	}

	token, err := auth.GenerateToken(h.jwtSecret, user.ID, user.Email)
	if err != nil {
		errpkg.InternalError(c, "failed to create token")
		return
	}

	errpkg.OK(c, gin.H{
		"token": token,
		"user": gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"created_at": user.CreatedAt,
		},
	})
}
