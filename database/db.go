package database

import (
	"database/sql"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"trackyou/models"

	_ "github.com/mattn/go-sqlite3"
)

// GetDefaultDBPath returns the platform-specific default path for the database file.
// It ensures the directory structure exists.
func GetDefaultDBPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	appDir := filepath.Join(configDir, "TrackYou")
	if err := os.MkdirAll(appDir, 0700); err != nil {
		return "", err
	}

	return filepath.Join(appDir, "tasks.db"), nil
}

type DB struct {
	*sql.DB
}

// NewDB creates a new database connection
func NewDB(dbPath string) (*DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

// InitDB creates the necessary tables if they don't exist
func (db *DB) InitDB() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_name TEXT NOT NULL,
			description TEXT,
			start_time DATETIME NOT NULL,
			end_time DATETIME NOT NULL,
			duration INTEGER NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS preferences (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}

	// Set default theme if not exists
	_, err := db.Exec(`
		INSERT OR IGNORE INTO preferences (key, value) 
		VALUES ('theme', 'light')
	`)
	if err != nil {
		return err
	}

	// Set default idle threshold if not exists (5 minutes)
	_, err = db.Exec(`
		INSERT OR IGNORE INTO preferences (key, value) 
		VALUES ('idle_threshold', '5')
	`)
	if err != nil {
		return err
	}

	// Set default workday length if not exists (8.0 hours)
	_, err = db.Exec(`
		INSERT OR IGNORE INTO preferences (key, value) 
		VALUES ('workday_length', '8.0')
	`)
	return err
}

// GetWorkdayLength retrieves the workday length preference in hours
func (db *DB) GetWorkdayLength() (float64, error) {
	var length string
	err := db.QueryRow("SELECT value FROM preferences WHERE key = 'workday_length'").Scan(&length)
	if err != nil {
		if err == sql.ErrNoRows {
			return 8.0, nil
		}
		return 8.0, err
	}
	val, err := strconv.ParseFloat(length, 64)
	if err != nil || val <= 0 || math.IsNaN(val) || math.IsInf(val, 0) {
		return 8.0, nil
	}
	return val, nil
}

// SetWorkdayLength saves the workday length preference in hours
func (db *DB) SetWorkdayLength(hours float64) error {
	if hours <= 0 || math.IsNaN(hours) || math.IsInf(hours, 0) {
		return fmt.Errorf("workday length must be a finite number > 0")
	}
	query := `
	INSERT OR REPLACE INTO preferences (key, value)
	VALUES ('workday_length', ?)`
	_, err := db.Exec(query, strconv.FormatFloat(hours, 'f', 2, 64))
	return err
}

// SaveTask saves a task to the database
func (db *DB) SaveTask(task *models.Task) error {
	query := `
	INSERT INTO tasks (project_name, description, start_time, end_time, duration)
	VALUES (?, ?, ?, ?, ?)`

	_, err := db.Exec(query,
		task.ProjectName,
		task.Description,
		task.StartTime,
		task.EndTime,
		task.Duration.Nanoseconds())
	return err
}

// GetTasks retrieves all tasks from the database
func (db *DB) GetTasks() ([]*models.Task, error) {
	query := `SELECT id, project_name, description, start_time, end_time, duration FROM tasks`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		var duration int64
		err := rows.Scan(
			&task.ID,
			&task.ProjectName,
			&task.Description,
			&task.StartTime,
			&task.EndTime,
			&duration,
		)
		if err != nil {
			return nil, err
		}
		task.Duration = time.Duration(duration)
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// GetProjectNames retrieves distinct historical project names, newest first.
func (db *DB) GetProjectNames() ([]string, error) {
	query := `
	SELECT project_name
	FROM tasks
	WHERE project_name IS NOT NULL AND project_name <> ''
	GROUP BY project_name
	ORDER BY MAX(end_time) DESC`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projectNames := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		projectNames = append(projectNames, name)
	}
	return projectNames, nil
}

// UpdateTask updates an existing task in the database
func (db *DB) UpdateTask(task *models.Task) error {
	query := `
	UPDATE tasks 
	SET project_name = ?, description = ?, start_time = ?, end_time = ?, duration = ?
	WHERE id = ?`

	_, err := db.Exec(query,
		task.ProjectName,
		task.Description,
		task.StartTime,
		task.EndTime,
		task.Duration.Nanoseconds(),
		task.ID)
	return err
}

// DeleteTask deletes a task from the database
func (db *DB) DeleteTask(id int64) error {
	query := `DELETE FROM tasks WHERE id = ?`
	_, err := db.Exec(query, id)
	return err
}

// GetTheme retrieves the current theme preference
func (db *DB) GetTheme() (string, error) {
	var theme string
	err := db.QueryRow("SELECT value FROM preferences WHERE key = 'theme'").Scan(&theme)
	if err != nil {
		return "light", err
	}
	return theme, nil
}

// SetTheme saves the theme preference
func (db *DB) SetTheme(theme string) error {
	query := `
	INSERT OR REPLACE INTO preferences (key, value)
	VALUES ('theme', ?)`
	_, err := db.Exec(query, theme)
	return err
}

// GetIdleThreshold retrieves the idle threshold preference (in minutes)
func (db *DB) GetIdleThreshold() (int, error) {
	var threshold string
	err := db.QueryRow("SELECT value FROM preferences WHERE key = 'idle_threshold'").Scan(&threshold)
	if err != nil {
		if err == sql.ErrNoRows {
			return 5, nil
		}
		return 5, err
	}
	val, err := strconv.Atoi(threshold)
	if err != nil || val < 1 {
		return 5, nil
	}
	return val, nil
}

// SetIdleThreshold saves the idle threshold preference
func (db *DB) SetIdleThreshold(minutes int) error {
	if minutes < 1 {
		return fmt.Errorf("idle threshold must be >= 1")
	}
	query := `
	INSERT OR REPLACE INTO preferences (key, value)
	VALUES ('idle_threshold', ?)`
	_, err := db.Exec(query, strconv.Itoa(minutes))
	return err
}
