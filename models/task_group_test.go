package models

import (
	"testing"
	"time"
)

// ── ComputeWeeklySummaries tests ────────────────────────────────────────────

func TestComputeWeeklySummaries_Empty(t *testing.T) {
	now := time.Now()
	result := ComputeWeeklySummaries(nil, now, StartOfCurrentWeek(now))
	if result != nil {
		t.Errorf("expected nil for no tasks, got %v", result)
	}
}

func TestComputeWeeklySummaries_SingleProject(t *testing.T) {
	now := time.Now()
	windowStart := StartOfCurrentWeek(now)
	tasks := []*Task{
		{
			ProjectName: "Alpha",
			StartTime:   now.Add(-2 * time.Hour),
			Duration:    2 * time.Hour,
		},
	}
	summaries := ComputeWeeklySummaries(tasks, now, windowStart)
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
	windowStart := StartOfCurrentWeek(now)
	tasks := []*Task{
		{ProjectName: "B", StartTime: now.Add(-1 * time.Hour), Duration: 1 * time.Hour},
		{ProjectName: "A", StartTime: now.Add(-3 * time.Hour), Duration: 3 * time.Hour},
		{ProjectName: "C", StartTime: now.Add(-2 * time.Hour), Duration: 2 * time.Hour},
	}
	summaries := ComputeWeeklySummaries(tasks, now, windowStart)
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
	windowStart := StartOfCurrentWeek(now)
	tasks := []*Task{
		{ProjectName: "Zebra", StartTime: now.Add(-1 * time.Hour), Duration: time.Hour},
		{ProjectName: "Alpha", StartTime: now.Add(-1 * time.Hour), Duration: time.Hour},
	}
	summaries := ComputeWeeklySummaries(tasks, now, windowStart)
	if len(summaries) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(summaries))
	}
	if summaries[0].ProjectName != "Alpha" {
		t.Errorf("expected Alpha first (tiebreaker by name), got %s", summaries[0].ProjectName)
	}
}

func TestComputeWeeklySummaries_TaskOutsideWindow_Excluded(t *testing.T) {
	now := time.Now()
	// Task that ended before the explicit window start — should be excluded.
	windowStart := now.Add(-48 * time.Hour)
	old := now.Add(-72 * time.Hour)
	tasks := []*Task{
		{ProjectName: "Old", StartTime: old, Duration: time.Hour},
	}
	summaries := ComputeWeeklySummaries(tasks, now, windowStart)
	if summaries != nil {
		t.Errorf("expected nil (task outside window), got %v", summaries)
	}
}

func TestComputeWeeklySummaries_TaskCrossesWindowStart_Clipped(t *testing.T) {
	now := time.Now()
	windowStart := now.Add(-6 * time.Hour)

	// Task starts 1 hour before window, ends 1 hour after window start.
	taskStart := windowStart.Add(-1 * time.Hour)
	tasks := []*Task{
		{ProjectName: "CrossBoundary", StartTime: taskStart, Duration: 2 * time.Hour},
	}
	summaries := ComputeWeeklySummaries(tasks, now, windowStart)
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
	windowStart := now.Add(-72 * time.Hour)
	task1Start := now.Add(-24 * time.Hour)
	task2Start := now.Add(-48 * time.Hour)
	tasks := []*Task{
		{ProjectName: "X", StartTime: task1Start, Duration: 30 * time.Minute},
		{ProjectName: "X", StartTime: task2Start, Duration: 30 * time.Minute},
	}
	summaries := ComputeWeeklySummaries(tasks, now, windowStart)
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary for same project, got %d", len(summaries))
	}
	if summaries[0].Duration != time.Hour {
		t.Errorf("expected 1h aggregated, got %v", summaries[0].Duration)
	}
	idx1 := weekDayIndex(time.Date(task1Start.Year(), task1Start.Month(), task1Start.Day(), 0, 0, 0, 0, task1Start.Location()), windowStart)
	idx2 := weekDayIndex(time.Date(task2Start.Year(), task2Start.Month(), task2Start.Day(), 0, 0, 0, 0, task2Start.Location()), windowStart)
	if idx1 < 0 || idx2 < 0 {
		t.Fatalf("expected both tasks to be inside the test window; got idx1=%d idx2=%d", idx1, idx2)
	}
	if summaries[0].DailyDurations[idx1] != 30*time.Minute {
		t.Errorf("expected first task day bucket to have 30m, got %v", summaries[0].DailyDurations[idx1])
	}
	if summaries[0].DailyDurations[idx2] != 30*time.Minute {
		t.Errorf("expected second task day bucket to have 30m, got %v", summaries[0].DailyDurations[idx2])
	}
}

func TestComputeWeeklySummaries_Timezone(t *testing.T) {
	loc := time.FixedZone("TZ-5", -5*60*60)
	now := time.Date(2024, 10, 10, 12, 0, 0, 0, loc)
	windowStart := StartOfCurrentWeek(now)
	tasks := []*Task{
		{
			ProjectName: "TZProject",
			StartTime:   time.Date(2024, 10, 9, 10, 0, 0, 0, loc),
			Duration:    time.Hour,
		},
	}
	summaries := ComputeWeeklySummaries(tasks, now, windowStart)
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	if summaries[0].ProjectName != "TZProject" {
		t.Errorf("unexpected project: %s", summaries[0].ProjectName)
	}
}

func TestComputeWeeklySummaries_OnlyTodayWindow(t *testing.T) {
	now := time.Now()
	y, m, d := now.Date()
	startOfToday := time.Date(y, m, d, 0, 0, 0, 0, now.Location())

	tasks := []*Task{
		// Within today — should be included.
		{ProjectName: "Today", StartTime: startOfToday.Add(time.Hour), Duration: time.Hour},
		// Yesterday — should be excluded.
		{ProjectName: "Yesterday", StartTime: startOfToday.Add(-2 * time.Hour), Duration: time.Hour},
	}
	summaries := ComputeWeeklySummaries(tasks, now, startOfToday)
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary (today only), got %d", len(summaries))
	}
	if summaries[0].ProjectName != "Today" {
		t.Errorf("expected Today project, got %s", summaries[0].ProjectName)
	}
}

// ── StartOfCurrentWeek tests ─────────────────────────────────────────────────

func TestStartOfCurrentWeek_Monday(t *testing.T) {
	// 2024-10-07 is a Monday
	loc := time.UTC
	monday := time.Date(2024, 10, 7, 14, 30, 0, 0, loc)
	got := StartOfCurrentWeek(monday)
	want := time.Date(2024, 10, 7, 0, 0, 0, 0, loc)
	if !got.Equal(want) {
		t.Errorf("Monday: expected %v, got %v", want, got)
	}
}

func TestStartOfCurrentWeek_Wednesday(t *testing.T) {
	// 2024-10-09 is a Wednesday — Monday of that week is 2024-10-07
	loc := time.UTC
	wednesday := time.Date(2024, 10, 9, 9, 0, 0, 0, loc)
	got := StartOfCurrentWeek(wednesday)
	want := time.Date(2024, 10, 7, 0, 0, 0, 0, loc)
	if !got.Equal(want) {
		t.Errorf("Wednesday: expected %v, got %v", want, got)
	}
}

func TestStartOfCurrentWeek_Sunday(t *testing.T) {
	// 2024-10-13 is a Sunday — still in the week Mon 2024-10-07 … Sun 2024-10-13
	loc := time.UTC
	sunday := time.Date(2024, 10, 13, 23, 59, 0, 0, loc)
	got := StartOfCurrentWeek(sunday)
	want := time.Date(2024, 10, 7, 0, 0, 0, 0, loc)
	if !got.Equal(want) {
		t.Errorf("Sunday: expected %v, got %v", want, got)
	}
}

func TestStartOfCurrentWeek_Saturday(t *testing.T) {
	// 2024-10-12 is a Saturday — Monday of that week is 2024-10-07
	loc := time.UTC
	saturday := time.Date(2024, 10, 12, 0, 0, 0, 0, loc)
	got := StartOfCurrentWeek(saturday)
	want := time.Date(2024, 10, 7, 0, 0, 0, 0, loc)
	if !got.Equal(want) {
		t.Errorf("Saturday: expected %v, got %v", want, got)
	}
}

func TestStartOfCurrentWeek_PreservesTimezone(t *testing.T) {
	loc := time.FixedZone("UTC+2", 2*60*60)
	thursday := time.Date(2024, 10, 10, 10, 0, 0, 0, loc)
	got := StartOfCurrentWeek(thursday)
	if got.Location().String() != loc.String() {
		t.Errorf("expected timezone %v, got %v", loc, got.Location())
	}
}

// A Friday task for this calendar week is included; last week's Friday is not.
func TestComputeWeeklySummaries_CalendarWeekBoundary(t *testing.T) {
	loc := time.UTC
	// "now" is Thursday 2024-10-10 at 17:00
	now := time.Date(2024, 10, 10, 17, 0, 0, 0, loc)
	windowStart := StartOfCurrentWeek(now) // Mon 2024-10-07 00:00

	tasks := []*Task{
		// Tuesday this week — included
		{ProjectName: "ThisWeek", StartTime: time.Date(2024, 10, 8, 9, 0, 0, 0, loc), Duration: time.Hour},
		// Last Friday (previous week) — excluded
		{ProjectName: "LastWeek", StartTime: time.Date(2024, 10, 4, 9, 0, 0, 0, loc), Duration: time.Hour},
	}
	summaries := ComputeWeeklySummaries(tasks, now, windowStart)
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary (only this week), got %d: %v", len(summaries), summaries)
	}
	if summaries[0].ProjectName != "ThisWeek" {
		t.Errorf("expected ThisWeek, got %s", summaries[0].ProjectName)
	}
	if summaries[0].DailyDurations[1] != time.Hour {
		t.Errorf("expected Tuesday bucket to contain 1h, got %v", summaries[0].DailyDurations[1])
	}
}

func TestComputeWeeklySummaries_SplitsTaskAcrossMidnightBuckets(t *testing.T) {
	loc := time.UTC
	now := time.Date(2024, 10, 10, 2, 0, 0, 0, loc) // Thursday
	windowStart := StartOfCurrentWeek(now)          // Monday

	tasks := []*Task{
		{
			ProjectName: "NightShift",
			StartTime:   time.Date(2024, 10, 9, 23, 0, 0, 0, loc), // Wednesday
			Duration:    3 * time.Hour,                            // spills to Thursday
		},
	}

	summaries := ComputeWeeklySummaries(tasks, now, windowStart)
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	if summaries[0].DailyDurations[2] != time.Hour {
		t.Errorf("expected Wednesday bucket 1h, got %v", summaries[0].DailyDurations[2])
	}
	if summaries[0].DailyDurations[3] != 2*time.Hour {
		t.Errorf("expected Thursday bucket 2h, got %v", summaries[0].DailyDurations[3])
	}
}

func TestComputeWeeklySummaries_ClipsTaskAtNowForDailyBuckets(t *testing.T) {
	loc := time.UTC
	now := time.Date(2024, 10, 10, 12, 0, 0, 0, loc) // Thursday
	windowStart := StartOfCurrentWeek(now)           // Monday

	tasks := []*Task{
		{
			ProjectName: "Current",
			StartTime:   time.Date(2024, 10, 10, 10, 0, 0, 0, loc),
			Duration:    5 * time.Hour,
		},
	}

	summaries := ComputeWeeklySummaries(tasks, now, windowStart)
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	if summaries[0].Duration != 2*time.Hour {
		t.Errorf("expected duration clipped at now to 2h, got %v", summaries[0].Duration)
	}
	if summaries[0].DailyDurations[3] != 2*time.Hour {
		t.Errorf("expected Thursday bucket clipped at now to 2h, got %v", summaries[0].DailyDurations[3])
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
		},
	}

	items := FlattenTaskGroups(groups)

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
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

	// Index 1: Task
	if items[1].Type != ItemTypeTask {
		t.Error("expected index 1 to be task")
	}
	if items[1].Title != "P1" {
		t.Errorf("expected title 'P1', got '%s'", items[1].Title)
	}
	expectedSub := "D1 (1h0m0s)"
	if items[1].Subtitle != expectedSub {
		t.Errorf("expected subtitle '%s', got '%s'", expectedSub, items[1].Subtitle)
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
