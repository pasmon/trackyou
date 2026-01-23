package database

import (
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
