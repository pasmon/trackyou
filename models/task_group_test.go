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

func TestFlattenTaskGroups(t *testing.T) {
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

	items := FlattenTaskGroups(groups)

	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}

	// Index 0: Header
	if items[0].Type != ItemTypeHeader {
		t.Error("expected index 0 to be header")
	}
	if items[0].Title != "Monday, January 1" {
		t.Errorf("expected title 'Monday, January 1', got '%s'", items[0].Title)
	}
	if items[0].Subtitle != "Total: 1h0m0s" {
		t.Errorf("expected subtitle 'Total: 1h0m0s', got '%s'", items[0].Subtitle)
	}

	// Index 1: Project Summary (P1)
	if items[1].Type != ItemTypeSummary {
		t.Error("expected index 1 to be summary")
	}
	if items[1].Title != "P1" {
		t.Errorf("expected title 'P1', got '%s'", items[1].Title)
	}

	// Index 2: Task
	if items[2].Type != ItemTypeTask {
		t.Error("expected index 2 to be task")
	}
	if items[2].Title != "P1" {
		t.Errorf("expected title 'P1', got '%s'", items[2].Title)
	}
	expectedSub := "D1 (1h0m0s)"
	if items[2].Subtitle != expectedSub {
		t.Errorf("expected subtitle '%s', got '%s'", expectedSub, items[2].Subtitle)
	}
}