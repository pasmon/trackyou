package models

import (
	"testing"
	"time"
)

// ── ComputeWeeklySummaries tests ────────────────────────────────────────────

func TestComputeWeeklySummaries_Empty(t *testing.T) {
	result := ComputeWeeklySummaries(nil, time.Now(), 7)
	if result != nil {
		t.Errorf("expected nil for no tasks, got %v", result)
	}
}

func TestComputeWeeklySummaries_SingleProject(t *testing.T) {
	now := time.Now()
	tasks := []*Task{
		{
			ProjectName: "Alpha",
			StartTime:   now.Add(-2 * time.Hour),
			Duration:    2 * time.Hour,
		},
	}
	summaries := ComputeWeeklySummaries(tasks, now, 7)
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	if summaries[0].ProjectName != "Alpha" {
		t.Errorf("expected Alpha, got %s", summaries[0].ProjectName)
	}
	if summaries[0].Duration != 2*time.Hour {
		t.Errorf("expected 2h, got %v", summaries[0].Duration)
	}
	if summaries[0].Percentage != 1.0 {
		t.Errorf("expected percentage 1.0, got %f", summaries[0].Percentage)
	}
}

func TestComputeWeeklySummaries_MultipleProjects_SortedByDurationDesc(t *testing.T) {
	now := time.Now()
	tasks := []*Task{
		{ProjectName: "B", StartTime: now.Add(-1 * time.Hour), Duration: 1 * time.Hour},
		{ProjectName: "A", StartTime: now.Add(-3 * time.Hour), Duration: 3 * time.Hour},
		{ProjectName: "C", StartTime: now.Add(-2 * time.Hour), Duration: 2 * time.Hour},
	}
	summaries := ComputeWeeklySummaries(tasks, now, 7)
	if len(summaries) != 3 {
		t.Fatalf("expected 3 summaries, got %d", len(summaries))
	}
	if summaries[0].ProjectName != "A" || summaries[1].ProjectName != "C" || summaries[2].ProjectName != "B" {
		t.Errorf("unexpected order: %v, %v, %v", summaries[0].ProjectName, summaries[1].ProjectName, summaries[2].ProjectName)
	}
	if summaries[0].Percentage != 1.0 {
		t.Errorf("expected top project percentage 1.0, got %f", summaries[0].Percentage)
	}
}

func TestComputeWeeklySummaries_TieBreakerByName(t *testing.T) {
	now := time.Now()
	tasks := []*Task{
		{ProjectName: "Zebra", StartTime: now.Add(-1 * time.Hour), Duration: time.Hour},
		{ProjectName: "Alpha", StartTime: now.Add(-1 * time.Hour), Duration: time.Hour},
	}
	summaries := ComputeWeeklySummaries(tasks, now, 7)
	if len(summaries) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(summaries))
	}
	if summaries[0].ProjectName != "Alpha" {
		t.Errorf("expected Alpha first (tiebreaker by name), got %s", summaries[0].ProjectName)
	}
}

func TestComputeWeeklySummaries_TaskOutsideWindow_Excluded(t *testing.T) {
	now := time.Now()
	// Task that finished 8 days ago — entirely outside a 7-day window.
	old := now.AddDate(0, 0, -8)
	tasks := []*Task{
		{ProjectName: "Old", StartTime: old, Duration: time.Hour},
	}
	summaries := ComputeWeeklySummaries(tasks, now, 7)
	if summaries != nil {
		t.Errorf("expected nil (task outside window), got %v", summaries)
	}
}

func TestComputeWeeklySummaries_TaskCrossesWindowStart_Clipped(t *testing.T) {
	now := time.Now()
	y, m, d := now.Date()
	loc := now.Location()
	windowStart := time.Date(y, m, d-6, 0, 0, 0, 0, loc) // 7 days ago midnight

	// Task starts 1 hour before window, ends 1 hour after window start.
	taskStart := windowStart.Add(-1 * time.Hour)
	tasks := []*Task{
		{ProjectName: "CrossBoundary", StartTime: taskStart, Duration: 2 * time.Hour},
	}
	summaries := ComputeWeeklySummaries(tasks, now, 7)
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	// Only the 1 hour after window start should be counted.
	if summaries[0].Duration != time.Hour {
		t.Errorf("expected clipped duration 1h, got %v", summaries[0].Duration)
	}
}

func TestComputeWeeklySummaries_AggregatesAcrossDays(t *testing.T) {
	now := time.Now()
	tasks := []*Task{
		{ProjectName: "X", StartTime: now.Add(-24 * time.Hour), Duration: 30 * time.Minute},
		{ProjectName: "X", StartTime: now.Add(-48 * time.Hour), Duration: 30 * time.Minute},
	}
	summaries := ComputeWeeklySummaries(tasks, now, 7)
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary for same project, got %d", len(summaries))
	}
	if summaries[0].Duration != time.Hour {
		t.Errorf("expected 1h aggregated, got %v", summaries[0].Duration)
	}
}

func TestComputeWeeklySummaries_Timezone(t *testing.T) {
	loc := time.FixedZone("TZ-5", -5*60*60)
	now := time.Date(2024, 10, 10, 12, 0, 0, 0, loc)
	tasks := []*Task{
		{
			ProjectName: "TZProject",
			StartTime:   time.Date(2024, 10, 9, 10, 0, 0, 0, loc),
			Duration:    time.Hour,
		},
	}
	summaries := ComputeWeeklySummaries(tasks, now, 7)
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	if summaries[0].ProjectName != "TZProject" {
		t.Errorf("unexpected project: %s", summaries[0].ProjectName)
	}
}

func TestComputeWeeklySummaries_WindowDaysOneMeansToday(t *testing.T) {
	now := time.Now()
	y, m, d := now.Date()
	loc := now.Location()
	startOfToday := time.Date(y, m, d, 0, 0, 0, 0, loc)

	tasks := []*Task{
		// Within today — should be included.
		{ProjectName: "Today", StartTime: startOfToday.Add(time.Hour), Duration: time.Hour},
		// Yesterday — should be excluded.
		{ProjectName: "Yesterday", StartTime: startOfToday.Add(-2 * time.Hour), Duration: time.Hour},
	}
	summaries := ComputeWeeklySummaries(tasks, now, 1)
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary (today only), got %d", len(summaries))
	}
	if summaries[0].ProjectName != "Today" {
		t.Errorf("expected Today project, got %s", summaries[0].ProjectName)
	}
}

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

func TestGroupTasksByDate_Timezone(t *testing.T) {
	loc := time.FixedZone("Custom", -5*60*60)

	// Create a task at 11 PM in custom timezone
	startTime := time.Date(2024, 10, 10, 23, 0, 0, 0, loc)
	tasks := []*Task{
		{
			ProjectName: "P1",
			StartTime:   startTime,
			Duration:    time.Hour,
		},
	}

	groups := GroupTasksByDate(tasks)

	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}

	groupDate := groups[0].Date
	if groupDate.Location().String() != loc.String() {
		t.Errorf("expected group date location %v, got %v", loc, groupDate.Location())
	}

	y, m, d := groupDate.Date()
	if y != 2024 || m != 10 || d != 10 {
		t.Errorf("expected date 2024-10-10, got %d-%d-%d", y, m, d)
	}
}
