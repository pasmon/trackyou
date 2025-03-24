package main

import (
	"fmt"
	"sort"
	"time"

	"trackyou/database"
	"trackyou/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type App struct {
	window      fyne.Window
	app         fyne.App
	db          *database.DB
	taskList    *widget.List
	tasks       []*models.Task
	currentTask *models.Task
	timerLabel  *widget.Label
	timerStop   chan struct{}
	themeCheck  *widget.Check
}

type TaskGroup struct {
	Date  time.Time
	Tasks []*models.Task
}

func (a *App) updateTimer() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if a.currentTask != nil {
				duration := time.Since(a.currentTask.StartTime)
				a.timerLabel.SetText(fmt.Sprintf("Current Task Duration: %v", duration.Round(time.Second)))
			}
		case <-a.timerStop:
			return
		}
	}
}

func (a *App) toggleTheme(isDark bool) {
	if isDark {
		a.app.Settings().SetTheme(theme.DarkTheme())
	} else {
		a.app.Settings().SetTheme(theme.LightTheme())
	}
	if err := a.db.SetTheme(themeName(isDark)); err != nil {
		dialog.ShowError(err, a.window)
	}
}

func themeName(isDark bool) string {
	if isDark {
		return "dark"
	}
	return "light"
}

func (a *App) groupTasksByDate() []TaskGroup {
	// Create a map to group tasks by date
	groups := make(map[string][]*models.Task)
	for _, task := range a.tasks {
		date := task.StartTime.Format("2006-01-02")
		groups[date] = append(groups[date], task)
	}

	// Convert map to slice and sort by date
	var taskGroups []TaskGroup
	for dateStr, tasks := range groups {
		date, _ := time.Parse("2006-01-02", dateStr)
		// Sort tasks within each group by start time
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].StartTime.After(tasks[j].StartTime)
		})
		taskGroups = append(taskGroups, TaskGroup{Date: date, Tasks: tasks})
	}

	// Sort groups by date (newest first)
	sort.Slice(taskGroups, func(i, j int) bool {
		return taskGroups[i].Date.After(taskGroups[j].Date)
	})

	return taskGroups
}

func (a *App) getTaskItem(id widget.ListItemID) (string, bool) {
	groups := a.groupTasksByDate()
	currentIndex := 0

	for _, group := range groups {
		// Add date header
		if currentIndex == id {
			// Calculate total duration for the day
			var totalDuration time.Duration
			for _, task := range group.Tasks {
				totalDuration += task.Duration
			}
			return fmt.Sprintf("=== %s (Total: %v) ===",
				group.Date.Format("Monday, January 2, 2006"),
				totalDuration.Round(time.Second)), true
		}
		currentIndex++

		// Add tasks
		for _, task := range group.Tasks {
			if currentIndex == id {
				return fmt.Sprintf("  %s - %s (Duration: %v)",
					task.ProjectName,
					task.Description,
					task.Duration.Round(time.Second)), false
			}
			currentIndex++
		}
	}

	return "", false
}

func (a *App) getTaskCount() int {
	groups := a.groupTasksByDate()
	count := 0
	for _, group := range groups {
		count++ // Date header
		count += len(group.Tasks)
	}
	return count
}

func main() {
	myApp := app.New()
	window := myApp.NewWindow("Project Time Tracker")

	// Initialize database
	db, err := database.NewDB("tasks.db")
	if err != nil {
		dialog.ShowError(err, window)
		return
	}

	if err := db.InitDB(); err != nil {
		dialog.ShowError(err, window)
		return
	}

	app := &App{
		window:    window,
		app:       myApp,
		db:        db,
		tasks:     make([]*models.Task, 0),
		timerStop: make(chan struct{}),
	}

	// Load theme preference
	savedTheme, err := db.GetTheme()
	if err != nil {
		dialog.ShowError(err, window)
		return
	}
	isDark := savedTheme == "dark"
	if isDark {
		myApp.Settings().SetTheme(theme.DarkTheme())
	}

	// Create UI elements
	projectEntry := widget.NewEntry()
	projectEntry.SetPlaceHolder("Project Name")

	descriptionEntry := widget.NewEntry()
	descriptionEntry.SetPlaceHolder("Task Description")

	// Create timer label
	app.timerLabel = widget.NewLabel("No task running")
	app.timerLabel.Alignment = fyne.TextAlignCenter

	// Create theme checkbox
	app.themeCheck = widget.NewCheck("Dark Mode", func(isDark bool) {
		app.toggleTheme(isDark)
	})
	app.themeCheck.SetChecked(isDark)

	// Declare buttons first
	startButton := widget.NewButton("Start Task", nil)
	stopButton := widget.NewButton("Stop Task", nil)
	stopButton.Disable()

	// Set button callbacks
	startButton.OnTapped = func() {
		if app.currentTask != nil {
			dialog.ShowInformation("Error", "A task is already running", window)
			return
		}

		projectName := projectEntry.Text
		description := descriptionEntry.Text

		if projectName == "" {
			dialog.ShowError(fmt.Errorf("project name is required"), window)
			return
		}

		app.currentTask = models.NewTask(projectName, description)
		startButton.Disable()
		stopButton.Enable()
		app.timerLabel.SetText("Starting...")
		go app.updateTimer()
	}

	stopButton.OnTapped = func() {
		if app.currentTask == nil {
			dialog.ShowInformation("Error", "No task is running", window)
			return
		}

		app.currentTask.StopTask()
		if err := app.db.SaveTask(app.currentTask); err != nil {
			dialog.ShowError(err, window)
			return
		}

		// Prepend the new task to the beginning of the slice
		app.tasks = append([]*models.Task{app.currentTask}, app.tasks...)
		app.currentTask = nil
		startButton.Enable()
		stopButton.Disable()
		app.taskList.Refresh()
		app.timerStop <- struct{}{}
		app.timerLabel.SetText("No task running")
	}

	// Create task list
	app.taskList = widget.NewList(
		app.getTaskCount,
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			text, isHeader := app.getTaskItem(id)
			label := item.(*widget.Label)
			label.SetText(text)
			if isHeader {
				label.TextStyle = fyne.TextStyle{Bold: true}
			} else {
				label.TextStyle = fyne.TextStyle{}
			}
		},
	)

	// Load existing tasks
	tasks, err := app.db.GetTasks()
	if err != nil {
		dialog.ShowError(err, window)
		return
	}
	app.tasks = tasks

	// Create layout
	content := container.NewBorder(
		container.NewVBox(
			container.NewHBox(
				app.themeCheck,
			),
			projectEntry,
			descriptionEntry,
			app.timerLabel,
			container.NewHBox(startButton, stopButton),
		),
		nil, nil, nil,
		app.taskList,
	)

	window.SetContent(content)
	window.Resize(fyne.NewSize(600, 400))
	window.ShowAndRun()
}
