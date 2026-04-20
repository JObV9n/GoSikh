package main

import (
	"log"

	"learning-platform/internal/config"
	"learning-platform/internal/database"
	"learning-platform/internal/http/handlers"
	"learning-platform/internal/http/router"
	"learning-platform/internal/repository"
	"learning-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	db, err := database.New(database.Config{Path: cfg.DBPath})
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		log.Fatalf("database migration failed: %v", err)
	}

	if err := seedInitialData(db.DB); err != nil {
		log.Fatalf("seed failed: %v", err)
	}

	userRepo := repository.NewUserRepository(db.DB)
	courseRepo := repository.NewCourseRepository(db.DB)
	lessonRepo := repository.NewLessonRepository(db.DB)
	progressRepo := repository.NewProgressRepository(db.DB)

	authService := service.NewAuthService(userRepo)
	courseService := service.NewCourseService(courseRepo, lessonRepo, progressRepo)
	progressService := service.NewProgressService(progressRepo, lessonRepo, courseRepo)
	executeService := service.NewExecuteService()
	submissionService := service.NewSubmissionService(executeService, courseService)

	healthHandler := handlers.NewHealthHandler(db)
	authHandler := handlers.NewAuthHandler(authService, cfg.JWTSecret)
	courseHandler := handlers.NewCourseHandler(courseService)
	lessonHandler := handlers.NewLessonHandler(courseService)
	progressHandler := handlers.NewProgressHandler(progressService)
	executeHandler := handlers.NewExecuteHandler(executeService, submissionService)

	engine := router.New(cfg, healthHandler, authHandler, courseHandler, lessonHandler, progressHandler, executeHandler)

	if gin.Mode() == gin.DebugMode {
		log.Printf("server running on :%s", cfg.Port)
	}

	if err := engine.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server run failed: %v", err)
	}
}
