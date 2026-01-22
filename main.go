package main

import (
	"fmt"
	"image/color"
	"os"
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
	currentTask *models.Task
	
	// UI Components
	timerLabel       *widget.Label
	timerStop        chan struct{}
	projectEntry     *widget.Entry
	descriptionEntry *widget.Entry
	startButton      *widget.Button
	stopButton       *widget.Button
	recordingIcon    *canvas.Circle
	recordingAnim    *fyne.Animation
}

func (a *App) updateTaskGroups() {
	a.taskGroups = models.GroupTasksByDate(a.tasks)
}

func (a *App) updateTimer() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if a.currentTask != nil {
				duration := time.Since(a.currentTask.StartTime)
				a.timerLabel.SetText(fmt.Sprintf("%v", duration.Round(time.Second)))
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
	return models.GetTaskItemData(a.taskGroups, int(id))
}

func (a *App) getTaskCount() int {
	return models.GetTotalItemCount(a.taskGroups)
}

func (a *App) getTask(id widget.ListItemID) *models.Task {
	return models.GetTaskByListItemID(a.taskGroups, int(id))
}

func (a *App) startTask(projectName, description string) {
	if a.currentTask != nil {
		if os.Getenv("FYNE_TEST_SKIP_GUI") == "" {
			dialog.ShowInformation("Error", "A task is already running", a.window)
		}
		return
	}

	if projectName == "" {
		a.showDialogError(fmt.Errorf("project name is required"))
		return
	}

	a.currentTask = models.NewTask(projectName, description)
	a.updateButtonsState(true)
	a.timerLabel.SetText("Starting...")
	
	// Sync entries
	a.projectEntry.SetText(projectName)
	a.descriptionEntry.SetText(description)

	// Start Animation
	if a.recordingAnim != nil {
		a.recordingAnim.Start()
	}
	if a.recordingIcon != nil {
		a.recordingIcon.Show()
	}

	go a.updateTimer()
}

func (a *App) stopTask() {
	if a.currentTask == nil {
		return
	}

	a.currentTask.StopTask()
	if err := a.db.SaveTask(a.currentTask); err != nil {
		a.showDialogError(err)
		return
	}

	// Prepend
	a.tasks = append([]*models.Task{a.currentTask}, a.tasks...)
	a.updateTaskGroups()
	
	a.currentTask = nil
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
	
	// Stop Animation
	if a.recordingAnim != nil {
		a.recordingAnim.Stop()
	}
	if a.recordingIcon != nil {
		a.recordingIcon.Hide()
	}
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
	
	// Recording Animation (Pulse Opacity)
	a.recordingAnim = fyne.NewAnimation(2*time.Second, func(p float32) {
		// Pulse alpha from 0.2 to 1.0 and back
		alpha := uint8(255 * (0.5 + 0.5*p)) // simplify: just fade in?
		// Triangle wave
		if p > 0.5 {
			p = 1.0 - p
		}
		// p is 0 -> 0.5 -> 0
		alpha = uint8(255 * (0.3 + 1.4*p)) // 0.3 to 1.0 approx
		c := color.RGBA{R: 255, G: 0, B: 0, A: alpha}
		a.recordingIcon.FillColor = c
		a.recordingIcon.Refresh()
	})
	a.recordingAnim.RepeatCount = fyne.AnimationRepeatForever
	a.recordingAnim.AutoReverse = false

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
	db, err := database.NewDB("tasks.db")
	if err != nil {
		dialog.ShowError(err, window)
		return
	}
	if err := db.InitDB(); err != nil {
		dialog.ShowError(err, window)
		return
	}

	application := &App{
		window:    window,
		app:       myApp,
		db:        db,
		tasks:     make([]*models.Task, 0),
		timerStop: make(chan struct{}),
	}

	// Load Theme
	savedTheme, err := db.GetTheme()
	if err != nil {
		dialog.ShowError(err, window)
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
		dialog.ShowError(err, window)
		return
	}
	application.tasks = tasks
	application.updateTaskGroups()

	window.SetContent(mainContent)
	window.Resize(fyne.NewSize(500, 700)) // Portrait mobile-ish size
	window.ShowAndRun()
}
