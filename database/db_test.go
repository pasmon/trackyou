package database

import (
	"math"
	"os"
	"testing"
	"trackyou/models"
)

func setupTestDB(t *testing.T) (*DB, func()) {
	dbPath := "test_tasks.db"
	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}

	err = db.InitDB()
	if err != nil {
		t.Fatalf("failed to init test db: %v", err)
	}

	return db, func() {
		db.Close()
		os.Remove(dbPath)
	}
}

func TestDB_SaveAndGetTasks(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	task := models.NewTask("Project 1", "Description 1")
	task.StopTask()

	err := db.SaveTask(task)
	if err != nil {
		t.Fatalf("failed to save task: %v", err)
	}

	tasks, err := db.GetTasks()
	if err != nil {
		t.Fatalf("failed to get tasks: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}

	if tasks[0].ProjectName != "Project 1" {
		t.Errorf("expected ProjectName Project 1, got %s", tasks[0].ProjectName)
	}
}

func TestDB_UpdateTask(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	task := models.NewTask("Project 1", "Description 1")
	err := db.SaveTask(task)
	if err != nil {
		t.Fatalf("failed to save task: %v", err)
	}

	tasks, _ := db.GetTasks()
	savedTask := tasks[0]
	savedTask.ProjectName = "Updated Project"
	savedTask.Description = "Updated Description"

	err = db.UpdateTask(savedTask)
	if err != nil {
		t.Fatalf("failed to update task: %v", err)
	}

	tasks, _ = db.GetTasks()
	if tasks[0].ProjectName != "Updated Project" {
		t.Errorf("expected ProjectName Updated Project, got %s", tasks[0].ProjectName)
	}
}

func TestDB_DeleteTask(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	task := models.NewTask("Project 1", "Description 1")
	db.SaveTask(task)

	tasks, _ := db.GetTasks()
	id := tasks[0].ID

	err := db.DeleteTask(id)
	if err != nil {
		t.Fatalf("failed to delete task: %v", err)
	}

	tasks, _ = db.GetTasks()
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(tasks))
	}
}

func TestDB_ThemePreferences(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Default theme should be light
	theme, err := db.GetTheme()
	if err != nil {
		t.Fatalf("failed to get theme: %v", err)
	}
	if theme != "light" {
		t.Errorf("expected default theme light, got %s", theme)
	}

	err = db.SetTheme("dark")
	if err != nil {
		t.Fatalf("failed to set theme: %v", err)
	}

	theme, _ = db.GetTheme()
	if theme != "dark" {
		t.Errorf("expected theme dark, got %s", theme)
	}
}

func TestDB_IdleThreshold(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Default should be 5
	threshold, err := db.GetIdleThreshold()
	if err != nil {
		t.Fatalf("failed to get default threshold: %v", err)
	}
	if threshold != 5 {
		t.Errorf("expected default threshold 5, got %d", threshold)
	}

	// Valid set
	err = db.SetIdleThreshold(10)
	if err != nil {
		t.Fatalf("failed to set threshold: %v", err)
	}
	threshold, _ = db.GetIdleThreshold()
	if threshold != 10 {
		t.Errorf("expected threshold 10, got %d", threshold)
	}

	// Invalid set (too low)
	err = db.SetIdleThreshold(0)
	if err == nil {
		t.Error("expected error when setting threshold to 0, got nil")
	}
	err = db.SetIdleThreshold(-5)
	if err == nil {
		t.Error("expected error when setting threshold to -5, got nil")
	}

	// Manual database entry with invalid value should return default 5
	_, err = db.Exec("INSERT OR REPLACE INTO preferences (key, value) VALUES ('idle_threshold', '0')")
	if err != nil {
		t.Fatalf("failed to insert invalid threshold: %v", err)
	}
	threshold, _ = db.GetIdleThreshold()
	if threshold != 5 {
		t.Errorf("expected default 5 for invalid DB value 0, got %d", threshold)
	}

	_, err = db.Exec("INSERT OR REPLACE INTO preferences (key, value) VALUES ('idle_threshold', 'invalid')")
	if err != nil {
		t.Fatalf("failed to insert non-numeric threshold: %v", err)
	}
	threshold, _ = db.GetIdleThreshold()
	if threshold != 5 {
		t.Errorf("expected default 5 for non-numeric DB value, got %d", threshold)
	}
}

func TestDB_WorkdayLength(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Default value
	val, err := db.GetWorkdayLength()
	if err != nil {
		t.Errorf("GetWorkdayLength failed: %v", err)
	}
	if val != 8.0 {
		t.Errorf("Expected default 8.0, got %f", val)
	}

	// Set value
	if err := db.SetWorkdayLength(7.5); err != nil {
		t.Errorf("SetWorkdayLength failed: %v", err)
	}

	val, err = db.GetWorkdayLength()
	if err != nil {
		t.Errorf("GetWorkdayLength failed after set: %v", err)
	}
	if val != 7.5 {
		t.Errorf("Expected 7.5, got %f", val)
	}

	// Invalid value
	if err := db.SetWorkdayLength(-1.0); err == nil {
		t.Error("Expected error for negative workday length, got nil")
	}

	// Manual invalid values
	_, err = db.Exec("INSERT OR REPLACE INTO preferences (key, value) VALUES ('workday_length', '0.0')")
	if err != nil {
		t.Fatalf("failed to insert invalid goal: %v", err)
	}
	val, _ = db.GetWorkdayLength()
	if val != 8.0 {
		t.Errorf("Expected default 8.0 for 0.0 DB value, got %f", val)
	}

	_, err = db.Exec("INSERT OR REPLACE INTO preferences (key, value) VALUES ('workday_length', 'invalid')")
	if err != nil {
		t.Fatalf("failed to insert non-numeric goal: %v", err)
	}
	val, _ = db.GetWorkdayLength()
	if val != 8.0 {
		t.Errorf("Expected default 8.0 for non-numeric DB value, got %f", val)
	}

	// Non-finite values
	_, err = db.Exec("INSERT OR REPLACE INTO preferences (key, value) VALUES ('workday_length', 'NaN')")
	if err != nil {
		t.Fatalf("failed to insert NaN goal: %v", err)
	}
	val, _ = db.GetWorkdayLength()
	if val != 8.0 {
		t.Errorf("Expected default 8.0 for NaN DB value, got %f", val)
	}

	_, err = db.Exec("INSERT OR REPLACE INTO preferences (key, value) VALUES ('workday_length', 'Inf')")
	if err != nil {
		t.Fatalf("failed to insert Inf goal: %v", err)
	}
	val, _ = db.GetWorkdayLength()
	if val != 8.0 {
		t.Errorf("Expected default 8.0 for Inf DB value, got %f", val)
	}

	if err := db.SetWorkdayLength(math.NaN()); err == nil {
		t.Error("Expected error for NaN workday length, got nil")
	}
	if err := db.SetWorkdayLength(math.Inf(1)); err == nil {
		t.Error("Expected error for Inf workday length, got nil")
	}
}
