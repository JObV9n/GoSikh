package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ExecutionResult contains the result of code execution
type ExecutionResult struct {
	Language    string
	Stdout      string
	Stderr      string
	ExitCode    int
	TimedOut    bool
	Runner      string
	RunnerImage string
	Sandboxed   bool
}

// ExecutorConfig holds configuration for code execution
type ExecutorConfig struct {
	Timeout       time.Duration
	MaxOutput     int // Maximum output size in bytes
	TempDir       string
	DisableDocker bool
	RunnersDir    string
}

// Executor executes code in isolated subprocesses
type Executor struct {
	config  ExecutorConfig
	runners map[string]runnerConfig
}

type runnerConfig struct {
	language      string
	containerName string
	image         string
	fileName      string
	localRunner   string
	localArgs     func(codePath string) []string
	containerArgs func(fileName string) []string
}

const (
	baseRunnerImage = "learning-platform-runner-base:latest"
	baseRunnerDir   = "base"
)

// NewExecutor creates a new code executor
func NewExecutor(config ExecutorConfig) *Executor {
	if config.Timeout == 0 {
		config.Timeout = 2 * time.Second
	}
	if config.MaxOutput == 0 {
		config.MaxOutput = 1024 * 1024 // 1MB default
	}
	if config.TempDir == "" {
		config.TempDir = os.TempDir()
	}

	runnersDir := config.RunnersDir
	if runnersDir == "" {
		runnersDir = discoverRunnersDir()
	}
	config.RunnersDir = runnersDir

	return &Executor{
		config: config,
		runners: map[string]runnerConfig{
			"go": {
				language:      "go",
				containerName: "go",
				image:         "learning-platform-runner-go:latest",
				fileName:      "main.go",
				localRunner:   "go",
				localArgs: func(codePath string) []string {
					return []string{"run", codePath}
				},
				containerArgs: func(fileName string) []string {
					return []string{"go", "run", fileName}
				},
			},
			"javascript": {
				language:      "javascript",
				containerName: "javascript",
				image:         "learning-platform-runner-javascript:latest",
				fileName:      "main.js",
				localRunner:   "node",
				localArgs: func(codePath string) []string {
					return []string{codePath}
				},
				containerArgs: func(fileName string) []string {
					return []string{"node", fileName}
				},
			},
			"python": {
				language:      "python",
				containerName: "python",
				image:         "learning-platform-runner-python:latest",
				fileName:      "main.py",
				localRunner:   "python3",
				localArgs: func(codePath string) []string {
					return []string{codePath}
				},
				containerArgs: func(fileName string) []string {
					return []string{"python3", fileName}
				},
			},
		},
	}
}

// ExecuteJavaScript runs JavaScript code using Node.js
func (e *Executor) ExecuteJavaScript(ctx context.Context, code string) (*ExecutionResult, error) {
	return e.executeLanguage(ctx, "javascript", code)
}

// ExecutePython runs Python code using the Python interpreter
func (e *Executor) ExecutePython(ctx context.Context, code string) (*ExecutionResult, error) {
	return e.executeLanguage(ctx, "python", code)
}

// ExecuteGo runs Go code using go run on a temporary file.
func (e *Executor) ExecuteGo(ctx context.Context, code string) (*ExecutionResult, error) {
	return e.executeLanguage(ctx, "go", code)
}

func (e *Executor) executeLanguage(ctx context.Context, language, code string) (*ExecutionResult, error) {
	runner, ok := e.runners[language]
	if !ok {
		return nil, fmt.Errorf("unsupported sandbox language: %s", language)
	}

	if !e.config.DisableDocker {
		result, dockerErr := e.executeWithDockerRunner(ctx, runner, code)
		if dockerErr == nil {
			return result, nil
		}
	}

	tmpDir, codePath, err := e.writeCodeFile(runner.fileName, code)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	return e.executeWithLocalRunner(ctx, runner, codePath)
}

func (e *Executor) writeCodeFile(fileName, code string) (string, string, error) {
	tmpDir, err := os.MkdirTemp(e.config.TempDir, "sandbox-src-*")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp sandbox dir: %w", err)
	}

	// Rootless/userns docker may not be able to read 0700 host temp dirs.
	if err := os.Chmod(tmpDir, 0755); err != nil {
		_ = os.RemoveAll(tmpDir)
		return "", "", fmt.Errorf("failed to chmod sandbox dir: %w", err)
	}

	codePath := filepath.Clean(filepath.Join(tmpDir, fileName))
	if err := os.WriteFile(codePath, []byte(code), 0644); err != nil {
		_ = os.RemoveAll(tmpDir)
		return "", "", fmt.Errorf("failed to write sandbox file: %w", err)
	}

	return tmpDir, codePath, nil
}

func (e *Executor) executeWithDockerRunner(ctx context.Context, runner runnerConfig, code string) (*ExecutionResult, error) {
	if err := e.ensureDockerRunnerImage(ctx, runner); err != nil {
		return nil, err
	}

	containerCommand := fmt.Sprintf("cat > %s && %s", runner.fileName, strings.Join(runner.containerArgs(runner.fileName), " "))
	args := []string{
		"run",
		"--rm",
		"-i",
		"--network",
		"none",
		"--cpus",
		"1.00",
		"--memory",
		"256m",
		"--pids-limit",
		"64",
		"--security-opt",
		"no-new-privileges",
		"--cap-drop",
		"ALL",
		"-w",
		"/workspace",
		runner.image,
		"sh",
		"-lc",
		containerCommand,
	}

	result, err := e.executeCommand(ctx, runner.language, "docker", args, strings.TrimSpace(code)+"\n")
	if err != nil {
		return nil, err
	}

	// Docker CLI uses 125 for runtime/infrastructure failures outside user code.
	if result.ExitCode == 125 {
		return nil, fmt.Errorf("docker runtime unavailable: %s", strings.TrimSpace(result.Stderr))
	}

	result.Sandboxed = true
	result.Runner = "docker-" + runner.containerName
	result.RunnerImage = runner.image
	return result, nil
}

func (e *Executor) executeWithLocalRunner(ctx context.Context, runner runnerConfig, codePath string) (*ExecutionResult, error) {
	result, err := e.executeCommand(ctx, runner.language, runner.localRunner, runner.localArgs(filepath.Clean(codePath)), "")
	if err != nil {
		return nil, err
	}

	result.Sandboxed = false
	result.Runner = "local-" + runner.localRunner
	return result, nil
}

func (e *Executor) ensureDockerRunnerImage(ctx context.Context, runner runnerConfig) error {
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("docker executable not found: %w", err)
	}

	if err := e.ensureDockerBaseImage(ctx); err != nil {
		return err
	}

	inspectCmd := exec.CommandContext(ctx, "docker", "image", "inspect", runner.image)
	if err := inspectCmd.Run(); err == nil {
		return nil
	}

	if e.config.RunnersDir == "" {
		return fmt.Errorf("docker image %s not found and runners directory is not configured", runner.image)
	}

	dockerfile := filepath.Join(e.config.RunnersDir, runner.containerName, "Dockerfile")
	if _, err := os.Stat(dockerfile); err != nil {
		return fmt.Errorf("runner dockerfile missing for %s: %w", runner.language, err)
	}

	buildCmd := exec.CommandContext(ctx, "docker", "build", "-f", dockerfile, "-t", runner.image, filepath.Dir(dockerfile))
	out, buildErr := buildCmd.CombinedOutput()
	if buildErr != nil {
		return fmt.Errorf("failed building %s: %w: %s", runner.image, buildErr, strings.TrimSpace(string(out)))
	}

	return nil
}

func (e *Executor) ensureDockerBaseImage(ctx context.Context) error {
	inspectCmd := exec.CommandContext(ctx, "docker", "image", "inspect", baseRunnerImage)
	if err := inspectCmd.Run(); err == nil {
		return nil
	}

	if e.config.RunnersDir == "" {
		return fmt.Errorf("docker image %s not found and runners directory is not configured", baseRunnerImage)
	}

	dockerfile := filepath.Join(e.config.RunnersDir, baseRunnerDir, "Dockerfile")
	if _, err := os.Stat(dockerfile); err != nil {
		return fmt.Errorf("base runner dockerfile missing: %w", err)
	}

	buildCmd := exec.CommandContext(ctx, "docker", "build", "-f", dockerfile, "-t", baseRunnerImage, filepath.Dir(dockerfile))
	out, buildErr := buildCmd.CombinedOutput()
	if buildErr != nil {
		return fmt.Errorf("failed building %s: %w: %s", baseRunnerImage, buildErr, strings.TrimSpace(string(out)))
	}

	return nil
}

func discoverRunnersDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}

	current := wd
	for i := 0; i < 6; i++ {
		candidate := filepath.Join(current, "sandbox", "runners")
		if info, statErr := os.Stat(candidate); statErr == nil && info.IsDir() {
			return candidate
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return ""
}

// executeCommand executes code with a prepared command and captures output.
func (e *Executor) executeCommand(ctx context.Context, language, executable string, args []string, stdinData string) (*ExecutionResult, error) {
	result := &ExecutionResult{
		Language: language,
		ExitCode: 0,
	}

	// Create a context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, e.config.Timeout)
	defer cancel()

	// Create command
	cmd := exec.CommandContext(ctxWithTimeout, executable, args...)
	if stdinData != "" {
		cmd.Stdin = io.NopCloser(strings.NewReader(stdinData))
	}

	// Set up output buffers with size limits
	stdout := &LimitedBuffer{limit: e.config.MaxOutput}
	stderr := &LimitedBuffer{limit: e.config.MaxOutput}

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	// Execute the command
	err := cmd.Run()

	// Capture output
	result.Stdout = stdout.String()
	result.Stderr = stderr.String()

	// Check for timeout
	if ctxWithTimeout.Err() == context.DeadlineExceeded {
		result.TimedOut = true
		result.ExitCode = -1
		result.Stderr = "Execution timeout: code took longer than " + e.config.Timeout.String()
		return result, nil
	}

	if err != nil {
		if _, ok := err.(*exec.Error); ok {
			return nil, err
		}
	}

	// Get exit code if command failed
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.Stderr = fmt.Sprintf("Execution error: %v", err)
			result.ExitCode = -1
		}
	}

	return result, nil
}

// LimitedBuffer - stops writing after reaching a size limit
type LimitedBuffer struct {
	buf   bytes.Buffer
	limit int
}

// Write implements io.Writer with size limit enforcement
func (lb *LimitedBuffer) Write(p []byte) (n int, err error) {
	if lb.buf.Len()+len(p) > lb.limit {
		// Calculate how much we can write
		remaining := lb.limit - lb.buf.Len()
		if remaining > 0 {
			n, _ := lb.buf.Write(p[:remaining])
			return n, fmt.Errorf("output exceeds maximum size of %d bytes", lb.limit)
		}
		return 0, fmt.Errorf("output exceeds maximum size of %d bytes", lb.limit)
	}
	return lb.buf.Write(p)
}

// String returns the buffer contents as a string
func (lb *LimitedBuffer) String() string {
	return lb.buf.String()
}
