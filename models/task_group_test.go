package models

import (
	"testing"
	"time"
)

func TestGroupTasksByDate(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	tasks := []*Task{
		{
			ProjectName: "P1",
			StartTime:   now.Add(-1 * time.Hour),
			Duration:    time.Hour,
		},
		{
			ProjectName: "P2",
			StartTime:   now,
			Duration:    time.Hour,
		},
		{
			ProjectName: "P3",
			StartTime:   yesterday,
			Duration:    time.Hour,
		},
	}

	groups := GroupTasksByDate(tasks)

	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}

	if len(groups[0].Tasks) != 2 {
		t.Errorf("expected 2 tasks in group 0, got %d", len(groups[0].Tasks))
	}
}

func TestGetTotalItemCount(t *testing.T) {
	groups := []TaskGroup{
		{
			Tasks: []*Task{{}, {}},
			ProjectSummaries: []ProjectSummary{{}, {}},
		},
		{
			Tasks: []*Task{{}},
			ProjectSummaries: []ProjectSummary{{}},
		},
	}
	count := GetTotalItemCount(groups)
	if count != 8 {
		t.Errorf("expected 8 items, got %d", count)
	}
}

func TestGetTaskItemData(t *testing.T) {
	date := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	groups := []TaskGroup{
		{
			Date: date,
			Tasks: []*Task{
				{
					ProjectName: "P1",
					Description: "D1",
					Duration:    time.Hour,
				},
			},
			ProjectSummaries: []ProjectSummary{
				{Name: "P1", Duration: time.Hour},
			},
		},
	}

	// Index 0: Header
	title, sub, itemType := GetTaskItemData(groups, 0)
	if itemType != ItemTypeHeader {
		t.Error("expected index 0 to be header")
	}
	if title != "Monday, January 1" {
		t.Errorf("expected title 'Monday, January 1', got '%s'", title)
	}
	if sub != "Total: 1h0m0s" {
		t.Errorf("expected subtitle 'Total: 1h0m0s', got '%s'", sub)
	}

	// Index 1: Project Summary (P1)
	title, sub, itemType = GetTaskItemData(groups, 1)
	if itemType != ItemTypeSummary {
		t.Error("expected index 1 to be summary")
	}
	if title != "P1" {
		t.Errorf("expected title 'P1', got '%s'", title)
	}

	// Index 2: Task
	title, sub, itemType = GetTaskItemData(groups, 2)
	if itemType != ItemTypeTask {
		t.Error("expected index 2 to be task")
	}
	if title != "P1" {
		t.Errorf("expected title 'P1', got '%s'", title)
	}
	expectedSub := "D1 (1h0m0s)"
	if sub != expectedSub {
		t.Errorf("expected subtitle '%s', got '%s'", expectedSub, sub)
	}
}
