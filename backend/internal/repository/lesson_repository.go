package repository

import (
	"database/sql"
	"fmt"

	"learning-platform/internal/models"
)

type LessonRepository struct {
	db *sql.DB
}

func NewLessonRepository(db *sql.DB) *LessonRepository {
	return &LessonRepository{db: db}
}

func (r *LessonRepository) ListByCourseID(courseID int64) ([]models.Lesson, error) {
	rows, err := r.db.Query(
		"SELECT id, course_id, slug, title, content, position, created_at FROM lessons WHERE course_id = ? ORDER BY position ASC, id ASC",
		courseID,
	)
	if err != nil {
		return nil, fmt.Errorf("list lessons by course id: %w", err)
	}
	defer rows.Close()

	lessons := make([]models.Lesson, 0)
	for rows.Next() {
		var lesson models.Lesson
		if err := rows.Scan(&lesson.ID, &lesson.CourseID, &lesson.Slug, &lesson.Title, &lesson.Content, &lesson.Position, &lesson.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan lesson: %w", err)
		}
		lessons = append(lessons, lesson)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate lessons: %w", err)
	}

	return lessons, nil
}

func (r *LessonRepository) GetByCourseIDAndSlug(courseID int64, slug string) (*models.Lesson, error) {
	lesson := &models.Lesson{}
	err := r.db.QueryRow(
		"SELECT id, course_id, slug, title, content, position, created_at FROM lessons WHERE course_id = ? AND slug = ?",
		courseID,
		slug,
	).Scan(&lesson.ID, &lesson.CourseID, &lesson.Slug, &lesson.Title, &lesson.Content, &lesson.Position, &lesson.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get lesson: %w", err)
	}

	return lesson, nil
}

func (r *LessonRepository) GetByID(id int64) (*models.Lesson, error) {
	lesson := &models.Lesson{}
	err := r.db.QueryRow(
		"SELECT id, course_id, slug, title, content, position, created_at FROM lessons WHERE id = ?",
		id,
	).Scan(&lesson.ID, &lesson.CourseID, &lesson.Slug, &lesson.Title, &lesson.Content, &lesson.Position, &lesson.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get lesson by id: %w", err)
	}

	return lesson, nil
}

func (r *LessonRepository) Create(lesson *models.Lesson) error {
	result, err := r.db.Exec(
		"INSERT INTO lessons (course_id, slug, title, content, position) VALUES (?, ?, ?, ?, ?)",
		lesson.CourseID,
		lesson.Slug,
		lesson.Title,
		lesson.Content,
		lesson.Position,
	)
	if err != nil {
		return fmt.Errorf("create lesson: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get inserted lesson id: %w", err)
	}
	lesson.ID = id

	return nil
}

func (r *LessonRepository) CountByCourseID(courseID int64) (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM lessons WHERE course_id = ?", courseID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count lessons by course id: %w", err)
	}

	return count, nil
}
