package ui

import (
	"fmt"
	"strings"
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
		nameLabel.TextStyle = fyne.TextStyle{Bold: true}

		durLabel := widget.NewLabel(formatWeeklyDuration(s.Duration))
		durLabel.Alignment = fyne.TextAlignTrailing

		dailyLabel := widget.NewLabel(formatDailyDurations(s.DailyDurations))
		dailyLabel.Importance = widget.LowImportance
		dailyLabel.Wrapping = fyne.TextWrapWord

		row := container.NewVBox(
			container.NewBorder(nil, nil, nameLabel, durLabel, nil),
			dailyLabel,
		)
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

func formatDailyDurations(daily [7]time.Duration) string {
	labels := [7]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	firstLineParts := make([]string, 0, 4)
	secondLineParts := make([]string, 0, 3)
	for i := range daily {
		part := fmt.Sprintf("%s: %s", labels[i], formatWeeklyDuration(daily[i]))
		if i < 4 {
			firstLineParts = append(firstLineParts, part)
			continue
		}
		secondLineParts = append(secondLineParts, part)
	}
	return strings.Join(firstLineParts, "  |  ") + "\n" + strings.Join(secondLineParts, "  |  ")
}
