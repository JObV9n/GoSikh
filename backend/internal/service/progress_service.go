package service

import (
	"fmt"

	"learning-platform/internal/models"
	"learning-platform/internal/repository"
)

type ProgressService struct {
	progress *repository.ProgressRepository
	lessons  *repository.LessonRepository
	courses  *repository.CourseRepository
}

func NewProgressService(
	progress *repository.ProgressRepository,
	lessons *repository.LessonRepository,
	courses *repository.CourseRepository,
) *ProgressService {
	return &ProgressService{
		progress: progress,
		lessons:  lessons,
		courses:  courses,
	}
}

// MarkLessonCompleted marks a lesson as completed for a user
func (s *ProgressService) MarkLessonCompleted(userID, lessonID int64) error {
	if err := s.progress.UpsertCompletion(userID, lessonID, true); err != nil {
		return fmt.Errorf("mark lesson completed: %w", err)
	}
	return nil
}

// MarkLessonIncomplete marks a lesson as incomplete for a user
func (s *ProgressService) MarkLessonIncomplete(userID, lessonID int64) error {
	if err := s.progress.UpsertCompletion(userID, lessonID, false); err != nil {
		return fmt.Errorf("mark lesson incomplete: %w", err)
	}
	return nil
}

// GetUserCourseProgress retrieves all progress for a user in a specific course
// Returns the course, all lesson progress items, and calculated completion percentage
func (s *ProgressService) GetUserCourseProgress(userID int64, courseID int64) (*models.Course, []models.LessonProgress, float64, error) {
	// Fetch course
	course, err := s.courses.GetByID(courseID)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("get course: %w", err)
	}
	if course == nil {
		return nil, nil, 0, fmt.Errorf("course not found")
	}

	// Fetch lesson count only to avoid loading full lesson content for percentage math.
	totalLessons, err := s.lessons.CountByCourseID(courseID)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("count lessons: %w", err)
	}

	if totalLessons == 0 {
		// No lessons in course, return 0% completion
		return course, []models.LessonProgress{}, 0.0, nil
	}

	// Fetch user progress for this course
	progress, err := s.progress.ListByUserAndCourse(userID, courseID)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("get progress: %w", err)
	}

	// Calculate completion percentage
	completionPercentage := calculateCompletionPercentage(progress, totalLessons)

	return course, progress, completionPercentage, nil
}

// GetUserCourseProgressBySlug retrieves progress using course slug
func (s *ProgressService) GetUserCourseProgressBySlug(userID int64, courseSlug string) (*models.Course, []models.LessonProgress, float64, error) {
	course, err := s.courses.GetBySlug(courseSlug)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("get course: %w", err)
	}
	if course == nil {
		return nil, nil, 0, fmt.Errorf("course not found")
	}

	return s.GetUserCourseProgress(userID, course.ID)
}

// GetUserAllProgress retrieves all courses with their progress and completion percentages
func (s *ProgressService) GetUserAllProgress(userID int64) ([]models.CourseProgressSummary, error) {
	summaries, err := s.progress.ListCourseProgressSummaries(userID)
	if err != nil {
		return nil, fmt.Errorf("list all progress summaries: %w", err)
	}

	return summaries, nil
}

// calculateCompletionPercentage computes the percentage of lessons completed
func calculateCompletionPercentage(progress []models.LessonProgress, totalLessons int) float64 {
	if totalLessons == 0 {
		return 0.0
	}

	completed := 0
	for _, p := range progress {
		if p.Completed {
			completed++
		}
	}

	return (float64(completed) / float64(totalLessons)) * 100.0
}

