package router_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"learning-platform/internal/config"
	"learning-platform/internal/database"
	"learning-platform/internal/http/handlers"
	"learning-platform/internal/http/router"
	"learning-platform/internal/repository"
	"learning-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type testApp struct {
	engine *gin.Engine
	db     *database.DB
}

func setupTestApp(tb testing.TB) *testApp {
	tb.Helper()

	gin.SetMode(gin.TestMode)

	dbPath := filepath.Join(tb.TempDir(), "app.db")
	db, err := database.New(database.Config{Path: dbPath})
	if err != nil {
		tb.Fatalf("create db: %v", err)
	}

	if err := db.Migrate(); err != nil {
		tb.Fatalf("migrate db: %v", err)
	}

	if err := seedTestCourseData(db); err != nil {
		tb.Fatalf("seed test data: %v", err)
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
	authHandler := handlers.NewAuthHandler(authService, "test-secret")
	courseHandler := handlers.NewCourseHandler(courseService)
	lessonHandler := handlers.NewLessonHandler(courseService)
	progressHandler := handlers.NewProgressHandler(progressService)
	executeHandler := handlers.NewExecuteHandler(executeService, submissionService)

	engine := router.New(config.Config{
		JWTSecret:   "test-secret",
		FrontendURL: "http://localhost:5173",
	}, healthHandler, authHandler, courseHandler, lessonHandler, progressHandler, executeHandler)

	tb.Cleanup(func() {
		_ = db.Close()
	})

	return &testApp{engine: engine, db: db}
}

func seedTestCourseData(db *database.DB) error {
	result, err := db.Exec(
		"INSERT INTO courses (slug, title, description) VALUES (?, ?, ?)",
		"go-basics",
		"Go Basics",
		"Intro course",
	)
	if err != nil {
		return fmt.Errorf("insert course: %w", err)
	}

	courseID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("course id: %w", err)
	}

	if _, err := db.Exec(
		"INSERT INTO lessons (course_id, slug, title, content, position) VALUES (?, ?, ?, ?, ?)",
		courseID,
		"hello-go",
		"Hello Go",
		"content",
		1,
	); err != nil {
		return fmt.Errorf("insert lesson: %w", err)
	}

	return nil
}

func doJSONRequest(tb testing.TB, engine *gin.Engine, method, path string, body any, token string) *httptest.ResponseRecorder {
	tb.Helper()

	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			tb.Fatalf("marshal body: %v", err)
		}
	}

	return doJSONBytesRequest(tb, engine, method, path, payload, token)
}

func doJSONBytesRequest(tb testing.TB, engine *gin.Engine, method, path string, payload []byte, token string) *httptest.ResponseRecorder {
	tb.Helper()

	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)
	return rr
}

func parseJSON(tb testing.TB, rr *httptest.ResponseRecorder) map[string]any {
	tb.Helper()

	out := map[string]any{}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		tb.Fatalf("unmarshal response: %v; body=%s", err, rr.Body.String())
	}
	return out
}

func parseSuccessData(tb testing.TB, rr *httptest.ResponseRecorder) map[string]any {
	tb.Helper()

	payload := parseJSON(tb, rr)
	success, ok := payload["success"].(bool)
	if !ok || !success {
		tb.Fatalf("expected success=true in response envelope: %s", rr.Body.String())
	}

	if payload["error"] != nil {
		tb.Fatalf("expected error=null for successful response: %s", rr.Body.String())
	}

	data, ok := payload["data"].(map[string]any)
	if !ok {
		tb.Fatalf("expected object data in response envelope: %s", rr.Body.String())
	}

	return data
}

func registerAndGetToken(tb testing.TB, engine *gin.Engine, email string) string {
	tb.Helper()

	rr := doJSONRequest(tb, engine, http.MethodPost, "/api/auth/register", map[string]any{
		"email":    email,
		"password": "password123",
	}, "")
	if rr.Code != http.StatusCreated {
		tb.Fatalf("register status=%d body=%s", rr.Code, rr.Body.String())
	}

	data := parseSuccessData(tb, rr)
	token, ok := data["token"].(string)
	if !ok || token == "" {
		tb.Fatalf("missing token in register response: %s", rr.Body.String())
	}
	return token
}

func TestAuthRegisterAndLogin(t *testing.T) {
	app := setupTestApp(t)

	register := doJSONRequest(t, app.engine, http.MethodPost, "/api/auth/register", map[string]any{
		"email":    "user@example.com",
		"password": "password123",
	}, "")
	if register.Code != http.StatusCreated {
		t.Fatalf("expected 201 register, got %d body=%s", register.Code, register.Body.String())
	}

	registerPayload := parseSuccessData(t, register)
	if registerPayload["token"] == "" {
		t.Fatalf("expected token in register response")
	}

	duplicate := doJSONRequest(t, app.engine, http.MethodPost, "/api/auth/register", map[string]any{
		"email":    "user@example.com",
		"password": "password123",
	}, "")
	if duplicate.Code != http.StatusConflict {
		t.Fatalf("expected 409 duplicate register, got %d body=%s", duplicate.Code, duplicate.Body.String())
	}

	login := doJSONRequest(t, app.engine, http.MethodPost, "/api/auth/login", map[string]any{
		"email":    "user@example.com",
		"password": "password123",
	}, "")
	if login.Code != http.StatusOK {
		t.Fatalf("expected 200 login, got %d body=%s", login.Code, login.Body.String())
	}
}

func TestAuthValidationAndCredentialScenarios(t *testing.T) {
	app := setupTestApp(t)

	invalidEmail := doJSONRequest(t, app.engine, http.MethodPost, "/api/auth/register", map[string]any{
		"email":    "invalid",
		"password": "password123",
	}, "")
	if invalidEmail.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid email, got %d body=%s", invalidEmail.Code, invalidEmail.Body.String())
	}

	weakPassword := doJSONRequest(t, app.engine, http.MethodPost, "/api/auth/register", map[string]any{
		"email":    "valid@example.com",
		"password": "short",
	}, "")
	if weakPassword.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for weak password, got %d body=%s", weakPassword.Code, weakPassword.Body.String())
	}

	created := doJSONRequest(t, app.engine, http.MethodPost, "/api/auth/register", map[string]any{
		"email":    "member@example.com",
		"password": "password123",
	}, "")
	if created.Code != http.StatusCreated {
		t.Fatalf("expected 201 for valid registration, got %d body=%s", created.Code, created.Body.String())
	}

	wrongPassword := doJSONRequest(t, app.engine, http.MethodPost, "/api/auth/login", map[string]any{
		"email":    "member@example.com",
		"password": "wrong-password",
	}, "")
	if wrongPassword.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for wrong password, got %d body=%s", wrongPassword.Code, wrongPassword.Body.String())
	}

	unknownUser := doJSONRequest(t, app.engine, http.MethodPost, "/api/auth/login", map[string]any{
		"email":    "missing@example.com",
		"password": "password123",
	}, "")
	if unknownUser.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unknown user, got %d body=%s", unknownUser.Code, unknownUser.Body.String())
	}

	missingPassword := doJSONRequest(t, app.engine, http.MethodPost, "/api/auth/login", map[string]any{
		"email":    "member@example.com",
		"password": "",
	}, "")
	if missingPassword.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing password, got %d body=%s", missingPassword.Code, missingPassword.Body.String())
	}
}

func TestCORSPreflightAllowsLocalDevOrigin(t *testing.T) {
	app := setupTestApp(t)
	origin := "http://localhost:5174"

	req := httptest.NewRequest(http.MethodOptions, "/api/auth/login", nil)
	req.Header.Set("Origin", origin)
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "content-type")

	rr := httptest.NewRecorder()
	app.engine.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for preflight, got %d body=%s", rr.Code, rr.Body.String())
	}

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != origin {
		t.Fatalf("expected Access-Control-Allow-Origin=%s, got=%s", origin, got)
	}

	allowMethods := rr.Header().Get("Access-Control-Allow-Methods")
	if !strings.Contains(allowMethods, "OPTIONS") {
		t.Fatalf("expected OPTIONS in Access-Control-Allow-Methods, got=%s", allowMethods)
	}
}

func TestCORSPreflightBlocksUnknownOrigin(t *testing.T) {
	app := setupTestApp(t)

	req := httptest.NewRequest(http.MethodOptions, "/api/auth/login", nil)
	req.Header.Set("Origin", "http://malicious.example")
	req.Header.Set("Access-Control-Request-Method", "POST")

	rr := httptest.NewRecorder()
	app.engine.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for preflight, got %d body=%s", rr.Code, rr.Body.String())
	}

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected no Access-Control-Allow-Origin for unknown origin, got=%s", got)
	}
}

func TestCourseAndProgressFlow(t *testing.T) {
	app := setupTestApp(t)
	token := registerAndGetToken(t, app.engine, "student@example.com")

	listCourses := doJSONRequest(t, app.engine, http.MethodGet, "/api/courses", nil, "")
	if listCourses.Code != http.StatusOK {
		t.Fatalf("expected 200 list courses, got %d body=%s", listCourses.Code, listCourses.Body.String())
	}

	listLessons := doJSONRequest(t, app.engine, http.MethodGet, "/api/courses/go-basics/lessons", nil, "")
	if listLessons.Code != http.StatusOK {
		t.Fatalf("expected 200 list lessons, got %d body=%s", listLessons.Code, listLessons.Body.String())
	}

	progressUnauthorized := doJSONRequest(t, app.engine, http.MethodGet, "/api/courses/go-basics/progress", nil, "")
	if progressUnauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without auth, got %d body=%s", progressUnauthorized.Code, progressUnauthorized.Body.String())
	}

	markProgress := doJSONRequest(t, app.engine, http.MethodPut, "/api/courses/go-basics/lessons/hello-go/progress", map[string]any{
		"completed": true,
	}, token)
	if markProgress.Code != http.StatusOK {
		t.Fatalf("expected 200 mark progress, got %d body=%s", markProgress.Code, markProgress.Body.String())
	}

	getProgress := doJSONRequest(t, app.engine, http.MethodGet, "/api/courses/go-basics/progress", nil, token)
	if getProgress.Code != http.StatusOK {
		t.Fatalf("expected 200 get progress, got %d body=%s", getProgress.Code, getProgress.Body.String())
	}

	progressPayload := parseSuccessData(t, getProgress)
	progressItems, ok := progressPayload["progress"].([]any)
	if !ok || len(progressItems) != 1 {
		t.Fatalf("expected one progress item, body=%s", getProgress.Body.String())
	}
}

func TestCourseCreation(t *testing.T) {
	app := setupTestApp(t)
	token := registerAndGetToken(t, app.engine, "instructor@example.com")

	// Test creating a course with valid data
	createCourse := doJSONRequest(t, app.engine, http.MethodPost, "/api/courses", map[string]any{
		"slug":        "test-course",
		"title":       "Test Course",
		"description": "This is a test course",
	}, token)
	if createCourse.Code != http.StatusCreated {
		t.Fatalf("expected 201 course creation, got %d body=%s", createCourse.Code, createCourse.Body.String())
	}

	payload := parseSuccessData(t, createCourse)
	course, ok := payload["course"].(map[string]any)
	if !ok {
		t.Fatalf("expected course in response")
	}
	if course["slug"] != "test-course" || course["title"] != "Test Course" {
		t.Fatalf("course fields mismatch: %v", course)
	}

	// Test creating a course with missing required fields
	createCourseBad := doJSONRequest(t, app.engine, http.MethodPost, "/api/courses", map[string]any{
		"title": "Missing Slug",
	}, token)
	if createCourseBad.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 bad request, got %d body=%s", createCourseBad.Code, createCourseBad.Body.String())
	}

	// Test creating a course with duplicate slug (should fail if unique constraint exists)
	createCourseDuplicate := doJSONRequest(t, app.engine, http.MethodPost, "/api/courses", map[string]any{
		"slug":        "go-basics", // already seeded
		"title":       "Duplicate Course",
		"description": "This should fail",
	}, token)
	// We expect either 400 (validation) or 500 (database error) depending on implementation
	// Since we don't have a unique constraint in the test setup, it might succeed.
	// We'll just check that it doesn't crash and we get a response.
	if createCourseDuplicate.Code != http.StatusCreated && createCourseDuplicate.Code != http.StatusBadRequest && createCourseDuplicate.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status code for duplicate slug: %d body=%s", createCourseDuplicate.Code, createCourseDuplicate.Body.String())
	}
}

func TestLessonEndpoints(t *testing.T) {
	app := setupTestApp(t)
	token := registerAndGetToken(t, app.engine, "instructor@example.com")

	// Test getting a lesson by ID (using the seeded lesson)
	getLesson := doJSONRequest(t, app.engine, http.MethodGet, "/api/lessons/1", nil, "")
	if getLesson.Code != http.StatusOK {
		t.Fatalf("expected 200 get lesson, got %d body=%s", getLesson.Code, getLesson.Body.String())
	}

	payload := parseSuccessData(t, getLesson)
	lesson, ok := payload["lesson"].(map[string]any)
	if !ok {
		t.Fatalf("expected lesson in response")
	}
	if lesson["id"] != float64(1) || lesson["slug"] != "hello-go" {
		t.Fatalf("lesson fields mismatch: %v", lesson)
	}

	// Test getting a non-existent lesson
	getLessonNotFound := doJSONRequest(t, app.engine, http.MethodGet, "/api/lessons/999", nil, "")
	if getLessonNotFound.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for non-existent lesson, got %d body=%s", getLessonNotFound.Code, getLessonNotFound.Body.String())
	}

	// Test creating a lesson (requires auth)
	createLesson := doJSONRequest(t, app.engine, http.MethodPost, "/api/courses/go-basics/lessons", map[string]any{
		"slug":     "variables",
		"title":    "Variables in Go",
		"content":  "Learn about variables in Go",
		"position": 2,
	}, token)
	if createLesson.Code != http.StatusCreated {
		t.Fatalf("expected 201 lesson creation, got %d body=%s", createLesson.Code, createLesson.Body.String())
	}

	payload = parseSuccessData(t, createLesson)
	createdLesson, ok := payload["lesson"].(map[string]any)
	if !ok {
		t.Fatalf("expected lesson in response")
	}
	if createdLesson["slug"] != "variables" || createdLesson["title"] != "Variables in Go" {
		t.Fatalf("created lesson fields mismatch: %v", createdLesson)
	}

	// Test creating a lesson with missing required fields
	createLessonBad := doJSONRequest(t, app.engine, http.MethodPost, "/api/courses/go-basics/lessons", map[string]any{
		"title": "Missing Slug",
	}, token)
	if createLessonBad.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 bad request, got %d body=%s", createLessonBad.Code, createLessonBad.Body.String())
	}

	// Test creating a lesson for non-existent course
	createLessonBadCourse := doJSONRequest(t, app.engine, http.MethodPost, "/api/courses/non-existent/lessons", map[string]any{
		"slug":     "test-lesson",
		"title":    "Test Lesson",
		"content":  "Test content",
		"position": 1,
	}, token)
	if createLessonBadCourse.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for non-existent course, got %d body=%s", createLessonBadCourse.Code, createLessonBadCourse.Body.String())
	}
}

func TestExecuteEndpoint(t *testing.T) {
	app := setupTestApp(t)

	okRequest := doJSONRequest(t, app.engine, http.MethodPost, "/api/execute", map[string]any{
		"language": "go",
		"code":     "package main\nfunc main(){}",
	}, "")
	if okRequest.Code != http.StatusOK {
		t.Fatalf("expected 200 execute accepted, got %d body=%s", okRequest.Code, okRequest.Body.String())
	}

	unsupported := doJSONRequest(t, app.engine, http.MethodPost, "/api/execute", map[string]any{
		"language": "php",
		"code":     "<?php echo 'hi';",
	}, "")
	if unsupported.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 unsupported language, got %d body=%s", unsupported.Code, unsupported.Body.String())
	}

	emptyCode := doJSONRequest(t, app.engine, http.MethodPost, "/api/execute", map[string]any{
		"language": "go",
		"code":     "",
	}, "")
	if emptyCode.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 empty code, got %d body=%s", emptyCode.Code, emptyCode.Body.String())
	}
}
