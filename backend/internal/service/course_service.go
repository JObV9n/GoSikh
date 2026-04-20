package service

import (
	"errors"
	"fmt"

	"learning-platform/internal/models"
	"learning-platform/internal/repository"
)

var (
	ErrCourseNotFound = errors.New("course not found")
	ErrLessonNotFound = errors.New("lesson not found")
)

type CourseService struct {
	courses  *repository.CourseRepository
	lessons  *repository.LessonRepository
	progress *repository.ProgressRepository
}

func NewCourseService(
	courses *repository.CourseRepository,
	lessons *repository.LessonRepository,
	progress *repository.ProgressRepository,
) *CourseService {
	return &CourseService{
		courses:  courses,
		lessons:  lessons,
		progress: progress,
	}
}

func (s *CourseService) ListCourses() ([]models.Course, error) {
	items, err := s.courses.List()
	if err != nil {
		return nil, fmt.Errorf("list courses: %w", err)
	}
	return items, nil
}

func (s *CourseService) GetCourseByID(id int64) (*models.Course, error) {
	course, err := s.courses.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("get course by id: %w", err)
	}
	return course, nil
}

func (s *CourseService) GetCourseBySlug(slug string) (*models.Course, error) {
	course, err := s.courses.GetBySlug(slug)
	if err != nil {
		return nil, fmt.Errorf("get course by slug: %w", err)
	}
	return course, nil
}

func (s *CourseService) CreateCourse(course *models.Course) error {
	return s.courses.Create(course)
}

func (s *CourseService) ListLessonsByCourseSlug(slug string) (*models.Course, []models.Lesson, error) {
	course, err := s.courses.GetBySlug(slug)
	if err != nil {
		return nil, nil, fmt.Errorf("get course: %w", err)
	}
	if course == nil {
		return nil, nil, nil
	}

	lessons, err := s.lessons.ListByCourseID(course.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("list lessons: %w", err)
	}

	return course, lessons, nil
}

func (s *CourseService) MarkLessonProgress(userID int64, courseSlug, lessonSlug string, completed bool) error {
	course, err := s.courses.GetBySlug(courseSlug)
	if err != nil {
		return fmt.Errorf("get course: %w", err)
	}
	if course == nil {
		return ErrCourseNotFound
	}

	lesson, err := s.lessons.GetByCourseIDAndSlug(course.ID, lessonSlug)
	if err != nil {
		return fmt.Errorf("get lesson: %w", err)
	}
	if lesson == nil {
		return ErrLessonNotFound
	}

	if err := s.progress.UpsertCompletion(userID, lesson.ID, completed); err != nil {
		return fmt.Errorf("mark lesson progress: %w", err)
	}

	return nil
}

func (s *CourseService) GetCourseProgress(userID int64, courseSlug string) (*models.Course, []models.LessonProgress, error) {
	course, err := s.courses.GetBySlug(courseSlug)
	if err != nil {
		return nil, nil, fmt.Errorf("get course: %w", err)
	}
	if course == nil {
		return nil, nil, nil
	}

	items, err := s.progress.ListByUserAndCourse(userID, course.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("get course progress: %w", err)
	}

	return course, items, nil
}

func (s *CourseService) GetLessonByID(id int64) (*models.Lesson, error) {
	lesson, err := s.lessons.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("get lesson by id: %w", err)
	}
	return lesson, nil
}

func (s *CourseService) CreateLesson(lesson *models.Lesson) error {
	return s.lessons.Create(lesson)
}
