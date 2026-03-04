package main

import (
	"context"
	"fmt"
	"image/color"
	"os"
	"strconv"
	"sync"
	"time"

	"trackyou/database"
	"trackyou/models"
	"trackyou/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type App struct {
	window      fyne.Window
	app         fyne.App
	db          *database.DB
	taskList    *widget.List
	tasks       []*models.Task
	taskGroups  []models.TaskGroup
	flatItems   []models.FlatListItem
	currentTask *models.Task

	mu            sync.RWMutex
	idleThreshold int
	idleSince     time.Time
	idleCtx       context.Context
	idleCancel    context.CancelFunc

	// UI Components
	timerLabel       *widget.Label
	timerStop        chan struct{}
	projectEntry     *widget.Entry
	descriptionEntry *widget.Entry
	startButton      *widget.Button
	stopButton       *widget.Button
	recordingIcon    *canvas.Circle
}

func (a *App) updateTaskGroups() {
	a.taskGroups = models.GroupTasksByDate(a.tasks)
	a.flatItems = models.FlattenTaskGroups(a.taskGroups)
}

func (a *App) updateTimer() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	blink := false
	for {
		select {
		case <-ticker.C:
			a.mu.RLock()
			task := a.currentTask
			a.mu.RUnlock()
			if task != nil {
				duration := time.Since(task.StartTime)
				blink = !blink
				fyne.Do(func() {
					a.timerLabel.SetText(fmt.Sprintf("%v", duration.Round(time.Second)))
					if blink {
						a.recordingIcon.FillColor = color.RGBA{R: 255, G: 0, B: 0, A: 255}
					} else {
						a.recordingIcon.FillColor = color.RGBA{R: 255, G: 0, B: 0, A: 100}
					}
					a.recordingIcon.Refresh()
				})
			}
		case <-a.timerStop:
			return
		}
	}
}

func (a *App) showDialogError(err error) {
	if os.Getenv("FYNE_TEST_SKIP_GUI") != "" {
		fmt.Printf("Error (skipped dialog): %v\n", err)
		return
	}
	dialog.ShowError(err, a.window)
}

func (a *App) toggleTheme(isDark bool) {
	if isDark {
		a.app.Settings().SetTheme(ui.NewMaterialTheme(theme.VariantDark))
	} else {
		a.app.Settings().SetTheme(ui.NewMaterialTheme(theme.VariantLight))
	}
	if err := a.db.SetTheme(themeName(isDark)); err != nil {
		a.showDialogError(err)
	}
}

func themeName(isDark bool) string {
	if isDark {
		return "dark"
	}
	return "light"
}

func (a *App) getTaskItem(id widget.ListItemID) (title, subtitle string, itemType models.ItemType) {
	if id < 0 || id >= len(a.flatItems) {
		return "", "", models.ItemTypeHeader
	}
	item := a.flatItems[id]
	return item.Title, item.Subtitle, item.Type
}

func (a *App) getTaskCount() int {
	return len(a.flatItems)
}

func (a *App) getTask(id widget.ListItemID) *models.Task {
	if id < 0 || id >= len(a.flatItems) {
		return nil
	}
	return a.flatItems[id].Task
}

func (a *App) startTask(projectName, description string) {
	a.mu.Lock()
	if a.currentTask != nil {
		a.mu.Unlock()
		if os.Getenv("FYNE_TEST_SKIP_GUI") == "" {
			dialog.ShowInformation("Error", "A task is already running", a.window)
		}
		return
	}

	if projectName == "" {
		a.mu.Unlock()
		a.showDialogError(fmt.Errorf("project name is required"))
		return
	}

	a.currentTask = models.NewTask(projectName, description)
	a.idleSince = time.Time{}
	a.mu.Unlock()

	a.updateButtonsState(true)
	a.timerLabel.SetText("Starting...")

	// Sync entries
	a.projectEntry.SetText(projectName)
	a.descriptionEntry.SetText(description)

	if a.recordingIcon != nil {
		a.recordingIcon.Show()
	}

	go a.updateTimer()
}

func (a *App) stopTask() {
	a.mu.Lock()
	if a.currentTask == nil {
		a.mu.Unlock()
		return
	}

	a.currentTask.StopTask()
	task := a.currentTask
	a.currentTask = nil
	a.idleSince = time.Now().Round(0)
	a.mu.Unlock()

	if err := a.db.SaveTask(task); err != nil {
		a.showDialogError(err)
		return
	}

	// Prepend
	a.tasks = append([]*models.Task{task}, a.tasks...)
	a.updateTaskGroups()

	a.updateButtonsState(false)

	if a.taskList != nil {
		a.taskList.Refresh()
	}

	select {
	case a.timerStop <- struct{}{}:
	default:
	}

	if a.timerLabel != nil {
		a.timerLabel.SetText("Ready")
	}

	if a.recordingIcon != nil {
		a.recordingIcon.Hide()
	}
}

func (a *App) monitorIdle(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	lastNotified := time.Time{}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if a.checkIdle(&lastNotified) {
				lastNotified = time.Now().Round(0)
			}
		}
	}
}

// checkIdle checks if the app is idle and sends a notification if needed.
// Returns true if a notification was sent.
func (a *App) checkIdle(lastNotified *time.Time) bool {
	a.mu.RLock()
	task := a.currentTask
	idleSince := a.idleSince
	threshold := time.Duration(a.idleThreshold) * time.Minute
	a.mu.RUnlock()

	if task == nil && !idleSince.IsZero() {
		idleDuration := time.Since(idleSince)

		rearmInterval := threshold
		if rearmInterval < 5*time.Minute {
			rearmInterval = 5 * time.Minute
		}

		if idleDuration >= threshold && time.Since(*lastNotified) >= rearmInterval {
			a.app.SendNotification(fyne.NewNotification(
				"TrackYou",
				fmt.Sprintf("You've been idle for %d minutes. Don't forget to start a task!", int(idleDuration.Minutes())),
			))
			return true
		}
	}
	return false
}

func (a *App) updateButtonsState(running bool) {
	if a.startButton == nil || a.stopButton == nil || a.projectEntry == nil || a.descriptionEntry == nil {
		return
	}
	if running {
		a.startButton.Disable()
		a.stopButton.Enable()
		a.projectEntry.Disable()
		a.descriptionEntry.Disable()
	} else {
		a.startButton.Enable()
		a.stopButton.Disable()
		a.projectEntry.Enable()
		a.descriptionEntry.Enable()
	}
}

func (a *App) continueTask(task *models.Task) {
	a.startTask(task.ProjectName, task.Description)
}

func (a *App) showSettings() {
	a.mu.RLock()
	currentThreshold := a.idleThreshold
	a.mu.RUnlock()

	thresholdEntry := widget.NewEntry()
	thresholdEntry.SetText(fmt.Sprintf("%d", currentThreshold))

	items := []*widget.FormItem{
		widget.NewFormItem("Idle Threshold (min)", thresholdEntry),
	}

	dialog.ShowForm("Settings", "Save", "Cancel", items, func(confirmed bool) {
		if confirmed {
			val, err := strconv.Atoi(thresholdEntry.Text)
			if err != nil || val < 1 {
				a.showDialogError(fmt.Errorf("invalid threshold value"))
				return
			}
			a.mu.Lock()
			a.idleThreshold = val
			a.mu.Unlock()
			if err := a.db.SetIdleThreshold(val); err != nil {
				a.showDialogError(err)
			}
		}
	}, a.window)
}

func (a *App) makeUI() fyne.CanvasObject {
	// Top Bar: Theme Toggle
	themeCheck := widget.NewCheck("Dark Mode", func(checked bool) {
		a.toggleTheme(checked)
	})
	// Initialize check state based on current theme
	currentTheme, _ := a.db.GetTheme()
	themeCheck.SetChecked(currentTheme == "dark")

	topBar := container.NewHBox(layout.NewSpacer(), themeCheck)

	// Input Area
	a.projectEntry = widget.NewEntry()
	a.projectEntry.SetPlaceHolder("Project")
	a.descriptionEntry = widget.NewEntry()
	a.descriptionEntry.SetPlaceHolder("What are you working on?")

	// Timer Area
	a.timerLabel = widget.NewLabel("Ready")
	a.timerLabel.TextStyle = fyne.TextStyle{Bold: true}
	a.timerLabel.Alignment = fyne.TextAlignCenter

	// Recording Icon (Red Circle)
	a.recordingIcon = canvas.NewCircle(color.RGBA{R: 255, G: 0, B: 0, A: 255})
	a.recordingIcon.Resize(fyne.NewSize(12, 12))
	a.recordingIcon.Hide()

	timerContainer := container.NewHBox(
		layout.NewSpacer(),
		container.NewCenter(a.recordingIcon),
		a.timerLabel,
		layout.NewSpacer(),
	)

	// Buttons
	a.startButton = widget.NewButtonWithIcon("Start", theme.MediaPlayIcon(), func() {
		a.startTask(a.projectEntry.Text, a.descriptionEntry.Text)
	})
	a.startButton.Importance = widget.HighImportance

	a.stopButton = widget.NewButtonWithIcon("Stop", theme.MediaStopIcon(), func() {
		a.stopTask()
	})
	a.stopButton.Importance = widget.DangerImportance
	a.stopButton.Disable()

	inputContainer := container.NewVBox(
		a.projectEntry,
		a.descriptionEntry,
		layout.NewSpacer(),
		timerContainer,
		layout.NewSpacer(),
		container.NewGridWithColumns(2, a.startButton, a.stopButton),
	)

	inputCard := widget.NewCard("New Task", "", container.NewPadded(inputContainer))

	// Task List
	a.taskList = widget.NewList(
		a.getTaskCount,
		func() fyne.CanvasObject {
			// Template item
			title := widget.NewLabel("Title")
			title.TextStyle = fyne.TextStyle{Bold: true}
			title.Truncation = fyne.TextTruncateEllipsis

			subtitle := widget.NewLabel("Subtitle")
			subtitle.Truncation = fyne.TextTruncateEllipsis
			subtitle.Importance = widget.LowImportance // Greyish

			icon := widget.NewIcon(theme.ContentPasteIcon())

			playBtn := widget.NewButtonWithIcon("", theme.MediaPlayIcon(), nil)
			playBtn.Importance = widget.LowImportance

			textContainer := container.NewVBox(title, subtitle)

			// Layout: Icon | Text | Button
			content := container.NewBorder(nil, nil, icon, playBtn, textContainer)

			// Add a background or Card wrapper?
			// A Card wrapper creates a nice separated look.
			card := widget.NewCard("", "", content)
			return card
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			card := item.(*widget.Card)
			content := card.Content.(*fyne.Container)

			var icon *widget.Icon
			var playBtn *widget.Button
			var textContainer *fyne.Container

			// Robustly find components by type
			for _, obj := range content.Objects {
				switch o := obj.(type) {
				case *widget.Icon:
					icon = o
				case *widget.Button:
					playBtn = o
				case *fyne.Container:
					textContainer = o
				}
			}

			// Ensure we found them (optional safety check, but cleaner than panic)
			if icon == nil || playBtn == nil || textContainer == nil {
				return
			}

			title := textContainer.Objects[0].(*widget.Label)
			subtitle := textContainer.Objects[1].(*widget.Label)

			titleText, subtitleText, itemType := a.getTaskItem(id)

			title.SetText(titleText)
			subtitle.SetText(subtitleText)

			switch itemType {
			case models.ItemTypeHeader:
				icon.SetResource(theme.HistoryIcon()) // Calendar icon not standard? History is close.
				playBtn.Hide()
				title.TextStyle = fyne.TextStyle{Bold: true}
				subtitle.TextStyle = fyne.TextStyle{Bold: true}

			case models.ItemTypeSummary:
				icon.SetResource(theme.FolderIcon())
				playBtn.Hide()
				title.TextStyle = fyne.TextStyle{Bold: false} // Normal
				subtitle.TextStyle = fyne.TextStyle{Italic: true}

			case models.ItemTypeTask:
				icon.SetResource(theme.DocumentIcon()) // Task icon
				playBtn.Show()
				playBtn.OnTapped = func() {
					task := a.getTask(id)
					if task != nil {
						a.continueTask(task)
					}
				}
				title.TextStyle = fyne.TextStyle{Bold: true}
				subtitle.TextStyle = fyne.TextStyle{}
			}

			// Refresh the card layout
			card.Refresh()
		},
	)

	mainContent := container.NewBorder(
		container.NewVBox(topBar, inputCard), // Top
		nil, nil, nil,
		container.NewPadded(a.taskList), // Center
	)

	return mainContent
}

func main() {
	myApp := app.New()
	window := myApp.NewWindow("TrackYou")

	// Initialize DB
	dbPath, err := database.GetDefaultDBPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get database path: %v\n", err)
		return
	}
	db, err := database.NewDB(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		return
	}
	if err := db.InitDB(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize database: %v\n", err)
		return
	}

	// Load Idle Threshold
	idleThreshold, err := db.GetIdleThreshold()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load idle threshold: %v\n", err)
		idleThreshold = 5
	}

	idleCtx, idleCancel := context.WithCancel(context.Background())

	application := &App{
		window:        window,
		app:           myApp,
		db:            db,
		tasks:         make([]*models.Task, 0),
		timerStop:     make(chan struct{}),
		idleThreshold: idleThreshold,
		idleSince:     time.Now().Round(0), // Assume idle from start
		idleCtx:       idleCtx,
		idleCancel:    idleCancel,
	}

	go application.monitorIdle(idleCtx)

	// Load Theme
	savedTheme, err := db.GetTheme()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load theme preference: %v\n", err)
		return
	}
	isDark := savedTheme == "dark"
	if isDark {
		myApp.Settings().SetTheme(ui.NewMaterialTheme(theme.VariantDark))
	} else {
		myApp.Settings().SetTheme(ui.NewMaterialTheme(theme.VariantLight))
	}

	// --- UI Construction ---
	mainContent := application.makeUI()

	// Load Tasks
	tasks, err := application.db.GetTasks()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load tasks: %v\n", err)
		return
	}
	application.tasks = tasks
	application.updateTaskGroups()

	// --- Menu Construction ---
	settingsMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("Settings", func() {
			application.showSettings()
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", func() {
			application.idleCancel()
			myApp.Quit()
		}),
	)
	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("About", func() {
			ui.ShowAboutWindow(myApp, version, date, commit)
		}),
	)
	mainMenu := fyne.NewMainMenu(settingsMenu, helpMenu)
	window.SetMainMenu(mainMenu)

	window.SetContent(mainContent)
	window.Resize(fyne.NewSize(500, 700)) // Portrait mobile-ish size
	window.ShowAndRun()
	application.idleCancel()
}
