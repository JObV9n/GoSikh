package repository

import (
	"database/sql"
	"fmt"

	"learning-platform/internal/models"
)

type CourseRepository struct {
	db *sql.DB
}

func NewCourseRepository(db *sql.DB) *CourseRepository {
	return &CourseRepository{db: db}
}

func (r *CourseRepository) List() ([]models.Course, error) {
	rows, err := r.db.Query("SELECT id, slug, title, description, created_at FROM courses ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("list courses: %w", err)
	}
	defer rows.Close()

	courses := make([]models.Course, 0)
	for rows.Next() {
		var course models.Course
		if err := rows.Scan(&course.ID, &course.Slug, &course.Title, &course.Description, &course.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan course: %w", err)
		}
		courses = append(courses, course)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate courses: %w", err)
	}

	return courses, nil
}

func (r *CourseRepository) GetByID(id int64) (*models.Course, error) {
	course := &models.Course{}
	err := r.db.QueryRow(
		"SELECT id, slug, title, description, created_at FROM courses WHERE id = ?",
		id,
	).Scan(&course.ID, &course.Slug, &course.Title, &course.Description, &course.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get course by id: %w", err)
	}

	return course, nil
}

func (r *CourseRepository) GetBySlug(slug string) (*models.Course, error) {
	course := &models.Course{}
	err := r.db.QueryRow(
		"SELECT id, slug, title, description, created_at FROM courses WHERE slug = ?",
		slug,
	).Scan(&course.ID, &course.Slug, &course.Title, &course.Description, &course.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get course by slug: %w", err)
	}

	return course, nil
}

func (r *CourseRepository) Create(course *models.Course) error {
	result, err := r.db.Exec(
		"INSERT INTO courses (slug, title, description) VALUES (?, ?, ?)",
		course.Slug,
		course.Title,
		course.Description,
	)
	if err != nil {
		return fmt.Errorf("create course: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get inserted course id: %w", err)
	}
	course.ID = id

	return nil
}
