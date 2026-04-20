package main

import (
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func seedInitialData(db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("database is nil")
	}

	if err := ensureMockUser(db); err != nil {
		return err
	}

	var courseCount int
	if err := db.QueryRow("SELECT COUNT(1) FROM courses").Scan(&courseCount); err != nil {
		return fmt.Errorf("count courses: %w", err)
	}

	if courseCount > 0 {
		return nil
	}

	result, err := db.Exec(
		"INSERT INTO courses (slug, title, description) VALUES (?, ?, ?)",
		"go-basics",
		"Go Basics",
		"Start learning Go syntax, control flow, and functions.",
	)
	if err != nil {
		return fmt.Errorf("insert course: %w", err)
	}

	courseID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get inserted course id: %w", err)
	}

	if _, err := db.Exec(
		"INSERT INTO lessons (course_id, slug, title, content, position) VALUES (?, ?, ?, ?, ?)",
		courseID,
		"hello-go",
		"Hello Go",
		"Write your first Go program using package main and fmt.Println.",
		1,
	); err != nil {
		return fmt.Errorf("insert lesson hello-go: %w", err)
	}

	if _, err := db.Exec(
		"INSERT INTO lessons (course_id, slug, title, content, position) VALUES (?, ?, ?, ?, ?)",
		courseID,
		"variables",
		"Variables and Types",
		"Learn variable declarations with var and short assignment syntax.",
		2,
	); err != nil {
		return fmt.Errorf("insert lesson variables: %w", err)
	}

	return nil
}

func ensureMockUser(db *sql.DB) error {
	const mockEmail = "test@email.com"
	const mockPassword = "password"

	var existingCount int
	if err := db.QueryRow("SELECT COUNT(1) FROM users WHERE email = ?", mockEmail).Scan(&existingCount); err != nil {
		return fmt.Errorf("count mock user: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(mockPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash mock user password: %w", err)
	}

	if existingCount > 0 {
		if _, err := db.Exec("UPDATE users SET password_hash = ? WHERE email = ?", string(hash), mockEmail); err != nil {
			return fmt.Errorf("update mock user password: %w", err)
		}
		return nil
	}

	if _, err := db.Exec("INSERT INTO users (email, password_hash) VALUES (?, ?)", mockEmail, string(hash)); err != nil {
		return fmt.Errorf("insert mock user: %w", err)
	}

	return nil
}
