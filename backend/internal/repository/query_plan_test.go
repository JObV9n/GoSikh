package repository

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"learning-platform/internal/database"
)

func setupQueryPlanDB(t *testing.T) *database.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "query-plan.db")
	db, err := database.New(database.Config{Path: dbPath})
	if err != nil {
		t.Fatalf("create db: %v", err)
	}

	if err := db.Migrate(); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	if _, err := db.Exec("INSERT INTO courses (slug, title, description) VALUES ('go-basics', 'Go Basics', 'Intro')"); err != nil {
		t.Fatalf("seed course: %v", err)
	}
	if _, err := db.Exec("INSERT INTO users (email, password_hash) VALUES ('bench@example.com', 'hash')"); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if _, err := db.Exec("INSERT INTO lessons (course_id, slug, title, content, position) VALUES (1, 'hello-go', 'Hello Go', 'content', 1)"); err != nil {
		t.Fatalf("seed lesson 1: %v", err)
	}
	if _, err := db.Exec("INSERT INTO lessons (course_id, slug, title, content, position) VALUES (1, 'variables', 'Variables', 'content', 2)"); err != nil {
		t.Fatalf("seed lesson 2: %v", err)
	}
	if _, err := db.Exec("INSERT INTO progress (user_id, lesson_id, completed) VALUES (1, 1, 1)"); err != nil {
		t.Fatalf("seed progress: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

func explainPlanDetails(t *testing.T, db *sql.DB, query string, args ...any) []string {
	t.Helper()

	rows, err := db.Query("EXPLAIN QUERY PLAN "+query, args...)
	if err != nil {
		t.Fatalf("explain query plan: %v", err)
	}
	defer rows.Close()

	details := make([]string, 0)
	for rows.Next() {
		var id, parent, notused int
		var detail string
		if err := rows.Scan(&id, &parent, &notused, &detail); err != nil {
			t.Fatalf("scan explain row: %v", err)
		}
		details = append(details, strings.ToLower(detail))
	}

	if err := rows.Err(); err != nil {
		t.Fatalf("iterate explain rows: %v", err)
	}

	return details
}

func TestQueryPlanCountLessonsUsesIndex(t *testing.T) {
	db := setupQueryPlanDB(t)
	details := explainPlanDetails(t, db.DB, "SELECT COUNT(*) FROM lessons WHERE course_id = ?", 1)

	joined := strings.Join(details, "\n")
	if !strings.Contains(joined, "idx_lessons_course_position") {
		t.Fatalf("expected lessons count query to use idx_lessons_course_position, got plan:\n%s", joined)
	}
}

func TestQueryPlanCourseProgressSummaryAvoidsFullProgressScan(t *testing.T) {
	db := setupQueryPlanDB(t)
	details := explainPlanDetails(t, db.DB, `
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
`, 1)

	joined := strings.Join(details, "\n")
	if strings.Contains(joined, "scan p") {
		t.Fatalf("expected progress join to avoid full table scan, got plan:\n%s", joined)
	}
	if !strings.Contains(joined, "idx_progress_user_lesson_completed") &&
		!strings.Contains(joined, "idx_progress_user_completed") &&
		!strings.Contains(joined, "idx_progress_user") &&
		!strings.Contains(joined, "sqlite_autoindex_progress_1") {
		t.Fatalf("expected progress join to use a progress user index, got plan:\n%s", joined)
	}
}
