package repository

import (
	"database/sql"
	"fmt"

	"learning-platform/internal/models"
)

type ProgressRepository struct {
	db *sql.DB
}

func NewProgressRepository(db *sql.DB) *ProgressRepository {
	return &ProgressRepository{db: db}
}

func (r *ProgressRepository) UpsertCompletion(userID, lessonID int64, completed bool) error {
	completedInt := 0
	if completed {
		completedInt = 1
	}

	_, err := r.db.Exec(`
INSERT INTO progress (user_id, lesson_id, completed, completed_at, updated_at)
VALUES (?, ?, ?, CASE WHEN ? = 1 THEN CURRENT_TIMESTAMP ELSE NULL END, CURRENT_TIMESTAMP)
ON CONFLICT(user_id, lesson_id)
DO UPDATE SET
	completed = excluded.completed,
	completed_at = CASE WHEN excluded.completed = 1 THEN CURRENT_TIMESTAMP ELSE NULL END,
	updated_at = CURRENT_TIMESTAMP
`, userID, lessonID, completedInt, completedInt)
	if err != nil {
		return fmt.Errorf("upsert progress: %w", err)
	}

	return nil
}

func (r *ProgressRepository) ListByUserAndCourse(userID, courseID int64) ([]models.LessonProgress, error) {
	rows, err := r.db.Query(`
SELECT p.lesson_id, p.completed, p.completed_at
FROM progress p
JOIN lessons l ON l.id = p.lesson_id
WHERE p.user_id = ? AND l.course_id = ?
ORDER BY l.position ASC, l.id ASC
`, userID, courseID)
	if err != nil {
		return nil, fmt.Errorf("list progress: %w", err)
	}
	defer rows.Close()

	items := make([]models.LessonProgress, 0)
	for rows.Next() {
		var item models.LessonProgress
		var completedInt int
		if err := rows.Scan(&item.LessonID, &completedInt, &item.CompletedAt); err != nil {
			return nil, fmt.Errorf("scan progress: %w", err)
		}
		item.Completed = completedInt == 1
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate progress: %w", err)
	}

	return items, nil
}

// GetCompletedCountForCourse returns the number of completed lessons for a user in a course
// Optimized for fast completion percentage calculation
func (r *ProgressRepository) GetCompletedCountForCourse(userID, courseID int64) (int, error) {
	var count int
	err := r.db.QueryRow(`
SELECT COUNT(*)
FROM progress p
JOIN lessons l ON l.id = p.lesson_id
WHERE p.user_id = ? AND l.course_id = ? AND p.completed = 1
`, userID, courseID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("get completed count: %w", err)
	}
	return count, nil
}

// GetLessonProgress retrieves progress for a specific user and lesson
func (r *ProgressRepository) GetLessonProgress(userID, lessonID int64) (*models.LessonProgress, error) {
	var item models.LessonProgress
	var completedInt int

	err := r.db.QueryRow(`
SELECT lesson_id, completed, completed_at
FROM progress
WHERE user_id = ? AND lesson_id = ?
`, userID, lessonID).Scan(&item.LessonID, &completedInt, &item.CompletedAt)

	if err == sql.ErrNoRows {
		// Progress record doesn't exist yet, return nil
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get lesson progress: %w", err)
	}

	item.Completed = completedInt == 1
	return &item, nil
}

// IsLessonCompleted returns whether a lesson is marked complete by a user
func (r *ProgressRepository) IsLessonCompleted(userID, lessonID int64) (bool, error) {
	var completed int
	err := r.db.QueryRow(`
SELECT completed
FROM progress
WHERE user_id = ? AND lesson_id = ?
`, userID, lessonID).Scan(&completed)

	if err == sql.ErrNoRows {
		// Progress record doesn't exist, lesson not completed
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("is lesson completed: %w", err)
	}

	return completed == 1, nil
}

// DeleteProgress removes a progress record
func (r *ProgressRepository) DeleteProgress(userID, lessonID int64) error {
	_, err := r.db.Exec(`
DELETE FROM progress
WHERE user_id = ? AND lesson_id = ?
`, userID, lessonID)
	if err != nil {
		return fmt.Errorf("delete progress: %w", err)
	}
	return nil
}

// ListCourseProgressSummaries returns progress summaries for all courses using one query.
func (r *ProgressRepository) ListCourseProgressSummaries(userID int64) ([]models.CourseProgressSummary, error) {
	rows, err := r.db.Query(`
SELECT
	c.id,
	c.slug,
	c.title,
	c.description,
	c.created_at,
	COUNT(l.id) AS total_lessons,
	COALESCE(SUM(CASE WHEN p.completed = 1 THEN 1 ELSE 0 END), 0) AS completed_count
FROM courses c
LEFT JOIN lessons l ON l.course_id = c.id
LEFT JOIN progress p ON p.lesson_id = l.id AND p.user_id = ?
GROUP BY c.id, c.slug, c.title, c.description, c.created_at
ORDER BY c.id ASC
`, userID)
	if err != nil {
		return nil, fmt.Errorf("list course progress summaries: %w", err)
	}
	defer rows.Close()

	summaries := make([]models.CourseProgressSummary, 0)
	for rows.Next() {
		var course models.Course
		var summary models.CourseProgressSummary

		if err := rows.Scan(
			&course.ID,
			&course.Slug,
			&course.Title,
			&course.Description,
			&course.CreatedAt,
			&summary.TotalLessons,
			&summary.CompletedCount,
		); err != nil {
			return nil, fmt.Errorf("scan course progress summary: %w", err)
		}

		summary.Course = &course
		if summary.TotalLessons > 0 {
			summary.CompletionPercentage = (float64(summary.CompletedCount) / float64(summary.TotalLessons)) * 100.0
		}

		summaries = append(summaries, summary)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate course progress summaries: %w", err)
	}

	return summaries, nil
}
