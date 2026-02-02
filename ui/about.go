package ui

import (
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ShowAboutWindow displays the application information in a new window
func ShowAboutWindow(app fyne.App, version, buildDate, commit string) {
	w := app.NewWindow("About TrackYou")

	// 1. App Icon
	// Using Fyne logo as placeholder. In a real app, use resourceIconPng or similar if available.
	icon := canvas.NewImageFromResource(theme.FyneLogo())
	icon.FillMode = canvas.ImageFillContain
	icon.SetMinSize(fyne.NewSize(80, 80))

	// 2. App Name & Description
	title := widget.NewLabel("TrackYou")
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	desc := widget.NewLabel("Minimalist Time Tracker\nCross-platform & Open Source")
	desc.Alignment = fyne.TextAlignCenter
	desc.TextStyle = fyne.TextStyle{Italic: true}

	// 3. Version Info Grid
	// We want right-aligned labels and left-aligned values for a clean look

	versionLabel := widget.NewLabel("Version")
	versionLabel.TextStyle = fyne.TextStyle{Bold: true}
	versionLabel.Alignment = fyne.TextAlignTrailing

	versionValue := widget.NewLabel(version)

	buildLabel := widget.NewLabel("Build")
	buildLabel.TextStyle = fyne.TextStyle{Bold: true}
	buildLabel.Alignment = fyne.TextAlignTrailing

	buildValue := widget.NewLabel(buildDate)

	commitLabel := widget.NewLabel("Commit")
	commitLabel.TextStyle = fyne.TextStyle{Bold: true}
	commitLabel.Alignment = fyne.TextAlignTrailing

	// Truncate commit hash if it's long
	if len(commit) > 8 {
		commit = commit[:8]
	}
	commitValue := widget.NewLabel(commit)

	infoGrid := container.NewGridWithColumns(2,
		versionLabel, versionValue,
		buildLabel, buildValue,
		commitLabel, commitValue,
	)

	// 4. Buttons
	docsBtn := widget.NewButton("Docs", func() {
		u, _ := url.Parse("https://github.com/pasmon/trackyou#readme")
		app.OpenURL(u)
	})

	githubBtn := widget.NewButton("GitHub", func() {
		u, _ := url.Parse("https://github.com/pasmon/trackyou")
		app.OpenURL(u)
	})

	closeBtn := widget.NewButton("Close", func() {
		w.Close()
	})

	// Action buttons row
	actionButtons := container.NewHBox(layout.NewSpacer(), docsBtn, githubBtn, layout.NewSpacer())

	// Close button row
	closeRow := container.NewHBox(layout.NewSpacer(), closeBtn, layout.NewSpacer())

	// Compose Content
	// Vertical box with spacing
	content := container.NewVBox(
		container.NewPadded(container.NewCenter(icon)),
		title,
		desc,
		widget.NewSeparator(),
		container.NewCenter(infoGrid), // Center the grid
		widget.NewSeparator(),
		container.NewPadded(actionButtons),
		container.NewPadded(closeRow),
	)

	w.SetContent(container.NewPadded(content))
	w.Resize(fyne.NewSize(300, 400))
	w.Show()
}