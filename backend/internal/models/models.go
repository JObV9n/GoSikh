package models

import "time"

type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type Course struct {
	ID          int64     `json:"id"`
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type Lesson struct {
	ID        int64     `json:"id"`
	CourseID  int64     `json:"course_id"`
	Slug      string    `json:"slug"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
}

type LessonProgress struct {
	LessonID    int64      `json:"lesson_id"`
	Completed   bool       `json:"completed"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type CourseProgressSummary struct {
	Course               *Course `json:"course"`
	CompletedCount       int     `json:"completed_count"`
	TotalLessons         int     `json:"total_lessons"`
	CompletionPercentage float64 `json:"completion_percentage"`
}
