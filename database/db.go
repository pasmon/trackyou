package database

import (
	"database/sql"
	"time"
	"trackyou/models"

	_ "github.com/mattn/go-sqlite3"
)

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
