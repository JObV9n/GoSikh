package service

import (
	"context"
	"fmt"
	"strings"
)

type LessonSubmissionRequest struct {
	UserID         int64
	CourseSlug     string
	LessonSlug     string
	Language       string
	Code           string
	ExpectedOutput string
}

type LessonSubmissionResult struct {
	ExecuteResponse
	Passed         bool   `json:"passed"`
	ExpectedOutput string `json:"expected_output,omitempty"`
	ActualOutput   string `json:"actual_output"`
	ProgressMarked bool   `json:"progress_marked"`
}

type SubmissionService struct {
	execute *ExecuteService
	course  *CourseService
}

func NewSubmissionService(execute *ExecuteService, course *CourseService) *SubmissionService {
	return &SubmissionService{
		execute: execute,
		course:  course,
	}
}

func normalizeOutput(value string) string {
	return strings.TrimSpace(strings.ReplaceAll(value, "\r\n", "\n"))
}

func (s *SubmissionService) SubmitLesson(ctx context.Context, req LessonSubmissionRequest) (LessonSubmissionResult, error) {
	execution, err := s.execute.Execute(ctx, ExecuteRequest{
		Language: req.Language,
		Code:     req.Code,
	})
	if err != nil {
		return LessonSubmissionResult{}, err
	}

	actual := normalizeOutput(execution.Stdout)
	expected := normalizeOutput(req.ExpectedOutput)
	hasExpected := strings.TrimSpace(req.ExpectedOutput) != ""

	passedByExecution := !hasExpected && execution.ExitCode == 0 && !execution.TimedOut && strings.TrimSpace(execution.Stderr) == ""
	passedByExpected := hasExpected && execution.ExitCode == 0 && !execution.TimedOut && strings.TrimSpace(execution.Stderr) == "" && actual == expected
	passed := passedByExecution || passedByExpected

	progressMarked := false
	if passed {
		if err := s.course.MarkLessonProgress(req.UserID, req.CourseSlug, req.LessonSlug, true); err != nil {
			return LessonSubmissionResult{}, fmt.Errorf("mark lesson progress: %w", err)
		}
		progressMarked = true
	}

	result := LessonSubmissionResult{
		ExecuteResponse: execution,
		Passed:            passed,
		ActualOutput:      actual,
		ProgressMarked:    progressMarked,
	}

	if hasExpected {
		result.ExpectedOutput = expected
	}

	return result, nil
}
