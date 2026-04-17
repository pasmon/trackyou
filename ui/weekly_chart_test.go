package ui

import (
	"testing"
	"time"
)

func TestFormatWeeklyDuration(t *testing.T) {
	if got := formatWeeklyDuration(45 * time.Minute); got != "45m" {
		t.Fatalf("expected 45m, got %q", got)
	}
	if got := formatWeeklyDuration(2*time.Hour + 5*time.Minute); got != "2h 5m" {
		t.Fatalf("expected 2h 5m, got %q", got)
	}
}

func TestFormatDailyDurations(t *testing.T) {
	daily := [7]time.Duration{
		time.Hour,
		2 * time.Hour,
		0,
		15 * time.Minute,
		30 * time.Minute,
		0,
		3*time.Hour + 10*time.Minute,
	}

	expected := "Mon: 1h 0m  |  Tue: 2h 0m  |  Wed: 0m  |  Thu: 15m\nFri: 30m  |  Sat: 0m  |  Sun: 3h 10m"
	if got := formatDailyDurations(daily); got != expected {
		t.Fatalf("unexpected daily string:\nexpected: %q\ngot:      %q", expected, got)
	}
}
