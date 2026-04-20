package service

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"learning-platform/sandbox"
)

var (
	ErrUnsupportedLanguage = errors.New("unsupported language")
	ErrEmptyCode           = errors.New("code is required")
)

type ExecuteRequest struct {
	Language string `json:"language"`
	Code     string `json:"code"`
	Input    string `json:"input,omitempty"`
}

type ExecuteResponse struct {
	Language    string `json:"language"`
	Stdout      string `json:"stdout"`
	Stderr      string `json:"stderr"`
	ExitCode    int    `json:"exit_code"`
	TimedOut    bool   `json:"timed_out"`
	Runner      string `json:"runner"`
	RunnerImage string `json:"runner_image,omitempty"`
	Sandboxed   bool   `json:"sandboxed"`
}

type ExecuteService struct {
	supported map[string]struct{}
	executor  *sandbox.Executor
}

func NewExecuteService() *ExecuteService {
	timeout := 20 * time.Second
	if timeoutOverride := strings.TrimSpace(os.Getenv("EXECUTION_TIMEOUT")); timeoutOverride != "" {
		if parsed, err := time.ParseDuration(timeoutOverride); err == nil {
			timeout = parsed
		}
	}

	disableDocker := strings.EqualFold(strings.TrimSpace(os.Getenv("EXECUTION_DISABLE_DOCKER")), "true")
	runnersDir := strings.TrimSpace(os.Getenv("RUNNERS_DIR"))

	executor := sandbox.NewExecutor(sandbox.ExecutorConfig{
		Timeout:       timeout,
		MaxOutput:     1024 * 1024, // 1MB
		DisableDocker: disableDocker,
		RunnersDir:    runnersDir,
	})

	return &ExecuteService{
		supported: map[string]struct{}{
			"go":         {},
			"javascript": {},
			"python":     {},
		},
		executor: executor,
	}
}

func (s *ExecuteService) Execute(ctx context.Context, req ExecuteRequest) (ExecuteResponse, error) {
	language := strings.ToLower(strings.TrimSpace(req.Language))
	if _, ok := s.supported[language]; !ok {
		return ExecuteResponse{}, ErrUnsupportedLanguage
	}

	code := strings.TrimSpace(req.Code)
	if code == "" {
		return ExecuteResponse{}, ErrEmptyCode
	}

	// Execute code based on language
	var result *sandbox.ExecutionResult
	var err error

	switch language {
	case "go":
		result, err = s.executor.ExecuteGo(ctx, code)
	case "javascript":
		result, err = s.executor.ExecuteJavaScript(ctx, code)
	case "python":
		result, err = s.executor.ExecutePython(ctx, code)
	default:
		return ExecuteResponse{}, ErrUnsupportedLanguage
	}

	if err != nil {
		return ExecuteResponse{}, err
	}

	// Convert sandbox result to service response
	return ExecuteResponse{
		Language:    language,
		Stdout:      result.Stdout,
		Stderr:      result.Stderr,
		ExitCode:    result.ExitCode,
		TimedOut:    result.TimedOut,
		Runner:      result.Runner,
		RunnerImage: result.RunnerImage,
		Sandboxed:   result.Sandboxed,
	}, nil
}
