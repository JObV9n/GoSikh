package router

import (
	"learning-platform/internal/config"
	"learning-platform/internal/http/handlers"
	"learning-platform/internal/http/middleware"
	"strings"

	"github.com/gin-gonic/gin"
)

func New(
	cfg config.Config,
	health *handlers.HealthHandler,
	auth *handlers.AuthHandler,
	courses *handlers.CourseHandler,
	lessons *handlers.LessonHandler,
	progress *handlers.ProgressHandler,
	execute *handlers.ExecuteHandler,
) *gin.Engine {
	r := gin.New()

	// Global middleware stack order: recovery first, request logging second.
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.RequestLogger())

	// CORS middleware
	r.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if isAllowedOrigin(origin, cfg.FrontendURL) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Vary", "Origin")
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Handle 404 and 405 errors
	r.NoRoute(middleware.NotFoundHandler())
	r.NoMethod(middleware.MethodNotAllowedHandler())

	r.GET("/health", health.Check)

	api := r.Group("/api")
	{
		api.POST("/auth/register", auth.Register)
		api.POST("/auth/login", auth.Login)

		api.GET("/courses", courses.ListCourses)
		api.GET("/courses/id/:id", courses.GetCourseByID)
		api.POST("/courses", courses.CreateCourse)
		api.GET("/courses/:courseSlug/lessons", courses.ListLessons)
		api.GET("/lessons/:id", lessons.GetLessonByID)
		api.POST("/execute", execute.Execute)

		protected := api.Group("")
		protected.Use(middleware.RequireAuth(cfg.JWTSecret))
		{
			protected.GET("/courses/:courseSlug/progress", courses.GetCourseProgress)
			protected.PUT("/courses/:courseSlug/lessons/:lessonSlug/progress", courses.MarkLessonProgress)
			protected.POST("/courses/:courseSlug/lessons/:lessonSlug/submit", execute.SubmitLesson)
			protected.POST("/courses/:courseSlug/lessons", lessons.CreateLesson)

			// Progress tracking endpoints
			protected.POST("/progress/lessons/:lessonID/complete", progress.MarkLessonCompleted)
			protected.POST("/progress/lessons/:lessonID/incomplete", progress.MarkLessonIncomplete)
			protected.GET("/progress/courses/id/:courseID", progress.GetCourseProgress)
			protected.GET("/progress/courses/:courseSlug/progress", progress.GetCourseProgressBySlug)
			protected.GET("/progress/all", progress.GetAllProgress)

			// Code execution endpoint
			protected.POST("/run-code", execute.Execute)
		}
	}

	return r
}

func isAllowedOrigin(origin, configuredFrontendURL string) bool {
	if origin == "" {
		return false
	}

	if origin == configuredFrontendURL {
		return true
	}

	// Allow local frontend dev servers regardless of port.
	return strings.HasPrefix(origin, "http://localhost:") ||
		strings.HasPrefix(origin, "http://127.0.0.1:") ||
		strings.HasPrefix(origin, "https://localhost:") ||
		strings.HasPrefix(origin, "https://127.0.0.1:")
}
