package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/exec"
)

// Example 1: Hardcoded credentials
const (
	dbUser     = "admin"
	dbPassword = "super_secret_password_123"
	dbHost     = "localhost"
)

// Example 2: SQL Injection vulnerability
func getUserByID(db *sql.DB, userID string) error {
	query := fmt.Sprintf("SELECT * FROM users WHERE id = '%s'", userID)
	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()
	return nil
}

// Example 3: Command injection vulnerability
func executeCommand(w http.ResponseWriter, r *http.Request) {
	command := r.URL.Query().Get("cmd")
	output, err := exec.Command("sh", "-c", command).Output()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(output)
}

// Example 4: Missing input validation
func readFile(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("file")
	// Path traversal vulnerability - no validation
	data, err := os.ReadFile(filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(data)
}
