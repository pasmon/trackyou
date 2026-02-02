package ui

import (
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ShowAboutDialog displays the application information dialog
func ShowAboutDialog(app fyne.App, parent fyne.Window, version, buildDate, commit string) {
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

	buttons := container.NewHBox(layout.NewSpacer(), docsBtn, githubBtn, layout.NewSpacer())

	// Compose Content
	// Vertical box with spacing
	content := container.NewVBox(
		container.NewPadded(container.NewCenter(icon)),
		title,
		desc,
		widget.NewSeparator(),
		container.NewCenter(infoGrid), // Center the grid
		widget.NewSeparator(),
		container.NewPadded(buttons),
	)

	// Create and show the custom dialog
	d := dialog.NewCustom("About TrackYou", "Close", content, parent)
	// d.Resize(fyne.NewSize(300, 400)) // Let it size itself or set min size
	d.Show()
}
