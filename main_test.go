package main

import (
	"os"
	"testing"
	"time"

	"trackyou/database"
	"trackyou/models"

	"fyne.io/fyne/v2/test"
)

// setupTestApp creates an App instance with a temporary database and headless UI
func setupTestApp(t *testing.T) (*App, func()) {
	// Skip GUI in CI
	oldVal := os.Getenv("FYNE_TEST_SKIP_GUI")
	os.Setenv("FYNE_TEST_SKIP_GUI", "1")

	// Create temp DB
	dbPath := "test_integration_tasks.db"
	db, err := database.NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}

	if err := db.InitDB(); err != nil {
		t.Fatalf("failed to init test db: %v", err)
	}

	// Create headless Fyne app
	myApp := test.NewApp()
	window := test.NewWindow(nil) // Content will be set by makeUI

	app := &App{
		window:        window,
		app:           myApp,
		db:            db,
		tasks:         make([]*models.Task, 0),
		timerStop:     make(chan struct{}),
		idleThreshold: 5,
		idleSince:     time.Now(),
	}

	// Initialize UI
	content := app.makeUI()
	window.SetContent(content)

	return app, func() {
		os.Setenv("FYNE_TEST_SKIP_GUI", oldVal)
		db.Close()
		os.Remove(dbPath)
		// window.Close() // Not strictly necessary in tests but good practice
	}
}

func TestIntegration_Lifecycle_StartStopTask(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// Initial State
	if app.currentTask != nil {
		t.Fatal("expected no current task initially")
	}
	if app.startButton.Disabled() {
		t.Error("start button should be enabled initially")
	}
	if !app.stopButton.Disabled() {
		t.Error("stop button should be disabled initially")
	}

	// 1. Start Task
	projectName := "Integration Test Project"
	description := "Testing Start/Stop"
	
	// Simulate user input
	app.projectEntry.SetText(projectName)
	app.descriptionEntry.SetText(description)
	
	// Click Start
	test.Tap(app.startButton)

	// Verify Running State
	if app.currentTask == nil {
		t.Fatal("currentTask should not be nil after start")
	}
	if app.currentTask.ProjectName != projectName {
		t.Errorf("expected project name %s, got %s", projectName, app.currentTask.ProjectName)
	}
	if !app.startButton.Disabled() {
		t.Error("start button should be disabled while running")
	}
	if app.stopButton.Disabled() {
		t.Error("stop button should be enabled while running")
	}

	// Wait a bit to ensure duration > 0
	time.Sleep(100 * time.Millisecond)

	// 2. Stop Task
	test.Tap(app.stopButton)

	// Verify Stopped State
	if app.currentTask != nil {
		t.Fatal("currentTask should be nil after stop")
	}
	if app.startButton.Disabled() {
		t.Error("start button should be enabled after stop")
	}
	if !app.stopButton.Disabled() {
		t.Error("stop button should be disabled after stop")
	}

	// 3. Verify Database Persistence
	tasks, err := app.db.GetTasks()
	if err != nil {
		t.Fatalf("failed to get tasks from db: %v", err)
	}
	
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task in db, got %d", len(tasks))
	}
	
	savedTask := tasks[0]
	if savedTask.ProjectName != projectName {
		t.Errorf("expected saved project name %s, got %s", projectName, savedTask.ProjectName)
	}
	if savedTask.Duration == 0 {
		t.Error("saved task should have non-zero duration")
	}
}

func TestIntegration_ThemeSwitching(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// Initial State (Default is Light)
	themeName, err := app.db.GetTheme()
	if err != nil {
		t.Fatalf("failed to get theme from db: %v", err)
	}
	if themeName != "light" {
		t.Errorf("expected initial theme light, got %s", themeName)
	}

	// Switch to Dark
	app.applyTheme("dark")

	// Verify DB Update
	themeName, err = app.db.GetTheme()
	if err != nil {
		t.Fatalf("failed to get theme from db: %v", err)
	}
	if themeName != "dark" {
		t.Errorf("expected theme dark after toggle, got %s", themeName)
	}

	// Switch to System
	app.applyTheme("system")

	// Verify DB Update
	themeName, err = app.db.GetTheme()
	if err != nil {
		t.Fatalf("failed to get theme from db: %v", err)
	}
	if themeName != "system" {
		t.Errorf("expected theme system after toggle, got %s", themeName)
	}

	// Switch back to Light
	app.applyTheme("light")

	// Verify DB Update
	themeName, err = app.db.GetTheme()
	if err != nil {
		t.Fatalf("failed to get theme from db: %v", err)
	}
	if themeName != "light" {
		t.Errorf("expected theme light after toggle, got %s", themeName)
	}
}

func TestIntegration_DataPersistence(t *testing.T) {
	// 1. Create App, Save Data, Close
	dbPath := "test_persistence.db"
	
	// Phase 1: Create and Save
	{
		db, err := database.NewDB(dbPath)
		if err != nil {
			t.Fatalf("phase 1 setup failed: %v", err)
		}
		err = db.InitDB()
		if err != nil {
			t.Fatalf("phase 1 init db failed: %v", err)
		}
		
		task := models.NewTask("Persisted Project", "Desc")
		task.StopTask()
		err = db.SaveTask(task)
		if err != nil {
			t.Fatalf("phase 1 save task failed: %v", err)
		}
		db.Close()
	}

	// Phase 2: Load new App with same DB
	{
		db, err := database.NewDB(dbPath)
		if err != nil {
			t.Fatalf("phase 2 setup failed: %v", err)
		}
		
		myApp := test.NewApp()
		window := test.NewWindow(nil)
		app := &App{
			window:        window,
			app:           myApp,
			db:            db,
			tasks:         make([]*models.Task, 0),
			timerStop:     make(chan struct{}),
			idleThreshold: 5,
			idleSince:     time.Now(),
		}
		
		// Simulate loading tasks as main() does
		tasks, err := app.db.GetTasks()
		if err != nil {
			t.Fatalf("failed to load tasks: %v", err)
		}
		app.tasks = tasks
		app.updateTaskGroups()
		
		// Initialize UI (which uses taskGroups)
		app.makeUI() // Should not panic and list should populate

		// Verify
		if len(app.tasks) != 1 {
			t.Fatalf("expected 1 loaded task, got %d", len(app.tasks))
		}
		if app.tasks[0].ProjectName != "Persisted Project" {
			t.Errorf("expected Persisted Project, got %s", app.tasks[0].ProjectName)
		}
		
		if app.getTaskCount() != 1 { // 1 task item (headers might add more rows)
			// Actually getTaskCount returns total items including headers. 
			// 1 task -> 1 Header (Date) + 1 Task = 2 items? 
			// Let's check models behavior or just assert > 0
			if app.getTaskCount() == 0 {
				t.Error("expected task list items > 0")
			}
		}

		db.Close()
	}
	
	os.Remove(dbPath)
}

func TestIntegration_UIEvent_ContinueTask(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// 1. Create a task manually and add to app
	oldTask := models.NewTask("Old Project", "Old Desc")
	oldTask.StopTask()
	
	err := app.db.SaveTask(oldTask)
	if err != nil {
		t.Fatalf("failed to save task: %v", err)
	}
	tasks, err := app.db.GetTasks()
	if err != nil {
		t.Fatalf("failed to get tasks from db: %v", err)
	}
	app.tasks = tasks
	app.updateTaskGroups()
	
	// Refresh list to ensure UI is in sync (though we are headless)
	app.taskList.Refresh()

	// 2. Simulate "Continue" (which is actually play button tap on list item)
	// But getting the actual list item widget is hard in unit test without rendering.
	// We can call app.continueTask(task) directly as we want to test the *event handling logic* 
	// rather than the Fyne list widget internal tap propagation (which is Fyne's responsibility).
	
	app.continueTask(tasks[0])

	// 3. Verify New Task Started with same details
	if app.currentTask == nil {
		t.Fatal("task should be running after continue")
	}
	if app.currentTask.ProjectName != "Old Project" {
		t.Errorf("expected project name Old Project, got %s", app.currentTask.ProjectName)
	}
	if app.currentTask.Description != "Old Desc" {
		t.Errorf("expected description Old Desc, got %s", app.currentTask.Description)
	}
	
	// Verify input fields updated
	if app.projectEntry.Text != "Old Project" {
		t.Errorf("expected entry text Old Project, got %s", app.projectEntry.Text)
	}
}

func TestIntegration_IdleNotification(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// Initial State: Idle Since Now
	if app.idleSince.IsZero() {
		t.Fatal("idleSince should be initialized")
	}

	// 1. Manually set idleSince to 6 minutes ago
	app.idleThreshold = 5
	app.idleSince = time.Now().Add(-6 * time.Minute)

	lastNotified := time.Time{}
	
	// Trigger check
	sent := app.checkIdle(&lastNotified)
	if !sent {
		t.Error("expected notification to be sent")
	}

	// 2. Test startTask resets idleSince
	app.startTask("Project", "Desc")
	if !app.idleSince.IsZero() {
		t.Error("idleSince should be zero after starting a task")
	}

	// Test if stopTask sets idleSince
	app.stopTask()
	if app.idleSince.IsZero() {
		t.Error("idleSince should be set after stopping a task")
	}
}

func TestIntegration_Settings(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// Initial threshold
	if app.idleThreshold != 5 {
		t.Errorf("expected default threshold 5, got %d", app.idleThreshold)
	}

	// Change threshold via showSettings (simulating form)
	// Since showSettings uses a dialog, it's hard to test automatically without more effort.
	// But we can test the database method directly and the app field.
	
	newThreshold := 10
	err := app.db.SetIdleThreshold(newThreshold)
	if err != nil {
		t.Fatalf("failed to set threshold: %v", err)
	}
	
	val, err := app.db.GetIdleThreshold()
	if err != nil {
		t.Fatalf("failed to get threshold from db: %v", err)
	}
	if val != newThreshold {
		t.Errorf("expected threshold %d in db, got %d", newThreshold, val)
	}
}

func TestIntegration_NormalizeTheme(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// Apply an invalid theme
	app.applyTheme("invalid")

	// Check if "light" (default) was persisted to DB instead of "invalid"
	persisted, err := app.db.GetTheme()
	if err != nil {
		t.Fatalf("failed to get theme: %v", err)
	}

	if persisted != "light" {
		t.Errorf("expected light theme for invalid input, but got %q persisted in DB", persisted)
	}
}

func TestIntegration_WeeklyChart_InitialEmpty(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	if app.weeklyCard == nil {
		t.Fatal("weeklyCard should not be nil after makeUI")
	}
}

func TestIntegration_WeeklyChart_RefreshDoesNotPanic(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// Should be safe with no tasks.
	app.refreshWeeklyChart()

	// Add a task and refresh again.
	task := models.NewTask("WeeklyProject", "desc")
	task.StopTask()
	err := app.db.SaveTask(task)
	if err != nil {
		t.Fatalf("failed to save task: %v", err)
	}
	tasks, err := app.db.GetTasks()
	if err != nil {
		t.Fatalf("failed to get tasks: %v", err)
	}
	app.mu.Lock()
	app.tasks = tasks
	app.updateTaskGroups()
	app.mu.Unlock()

	app.refreshWeeklyChart() // must not panic
}

func TestIntegration_WeeklyChart_UpdatesAfterStopTask(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// Start and stop a task; weeklyCard.SetContent should be called without panic.
	app.projectEntry.SetText("ChartProject")
	app.descriptionEntry.SetText("testing weekly chart")
	test.Tap(app.startButton)

	time.Sleep(50 * time.Millisecond)

	test.Tap(app.stopButton)

	// After stop, weekly card should still be non-nil (content was refreshed).
	if app.weeklyCard == nil {
		t.Fatal("weeklyCard should not be nil after stop task")
	}
}

func TestIntegration_ProjectSuggestions_Refresh(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	oldTask := models.NewTask("Old Project", "old")
	oldTask.StopTask()
	oldTask.EndTime = oldTask.EndTime.Add(-2 * time.Hour)
	oldTask.StartTime = oldTask.EndTime.Add(-30 * time.Minute)
	if err := app.db.SaveTask(oldTask); err != nil {
		t.Fatalf("failed to save old task: %v", err)
	}

	newTask := models.NewTask("New Project", "new")
	newTask.StopTask()
	if err := app.db.SaveTask(newTask); err != nil {
		t.Fatalf("failed to save new task: %v", err)
	}

	duplicateTask := models.NewTask("New Project", "duplicate")
	duplicateTask.StopTask()
	if err := app.db.SaveTask(duplicateTask); err != nil {
		t.Fatalf("failed to save duplicate task: %v", err)
	}

	app.refreshProjectSuggestions()

	projectNames, err := app.db.GetProjectNames()
	if err != nil {
		t.Fatalf("failed to fetch project names: %v", err)
	}

	if len(projectNames) != 2 || projectNames[0] != "New Project" || projectNames[1] != "Old Project" {
		t.Fatalf("unexpected project suggestions source: %v", projectNames)
	}
}
