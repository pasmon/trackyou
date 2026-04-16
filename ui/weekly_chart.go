package ui

import (
	"fmt"
	"time"

	"trackyou/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// MakeWeeklyChartContent returns a visual breakdown of hours per project.
// When summaries is empty it returns a centred empty-state label.
func MakeWeeklyChartContent(summaries []models.WeeklySummary) fyne.CanvasObject {
	if len(summaries) == 0 {
		lbl := widget.NewLabel("No tracked time this week.")
		lbl.Importance = widget.LowImportance
		lbl.Alignment = fyne.TextAlignCenter
		return container.NewCenter(lbl)
	}

	rows := make([]fyne.CanvasObject, 0, len(summaries))
	for _, s := range summaries {
		nameLabel := widget.NewLabel(s.ProjectName)
		nameLabel.Truncation = fyne.TextTruncateEllipsis

		bar := widget.NewProgressBar()
		bar.Min = 0
		bar.Max = 1.0
		bar.SetValue(s.Percentage)
		bar.TextFormatter = func() string { return "" }

		durLabel := widget.NewLabel(formatWeeklyDuration(s.Duration))
		durLabel.Alignment = fyne.TextAlignTrailing

		row := container.NewBorder(nil, nil, nameLabel, durLabel, bar)
		rows = append(rows, row)
	}

	return container.NewVBox(rows...)
}

// formatWeeklyDuration formats a duration as "Xh Ym" or "Ym" for display in
// the weekly chart.
func formatWeeklyDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}
