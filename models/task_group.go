package models

import (
	"fmt"
	"sort"
	"time"
)

// WeeklySummary holds the total tracked duration for a project over a time window.
type WeeklySummary struct {
	ProjectName    string
	Duration       time.Duration
	DailyDurations [7]time.Duration // Monday (index 0) through Sunday (index 6)
	Percentage     float64          // fraction of the largest project's duration (0.0–1.0)
}

// StartOfCurrentWeek returns midnight on the Monday of the week that contains
// now, using now's timezone.
func StartOfCurrentWeek(now time.Time) time.Time {
	y, m, d := now.Date()
	wd := int(now.Weekday()) // Sunday=0, Monday=1, …, Saturday=6
	if wd == 0 {
		wd = 7 // treat Sunday as day 7 so Monday is always day 1
	}
	return time.Date(y, m, d-(wd-1), 0, 0, 0, 0, now.Location())
}

// ComputeWeeklySummaries aggregates completed task durations per project for
// the window [windowStart … now], clipping each task's duration to that range.
// Returns summaries sorted by duration descending, name ascending as a
// tiebreaker.
func ComputeWeeklySummaries(tasks []*Task, now time.Time, windowStart time.Time) []WeeklySummary {
	windowEnd := now

	projectSummaries := make(map[string]*WeeklySummary)
	for _, task := range tasks {
		taskStart := task.StartTime
		taskEnd := task.StartTime.Add(task.Duration)

		start := taskStart
		if start.Before(windowStart) {
			start = windowStart
		}
		end := taskEnd
		if end.After(windowEnd) {
			end = windowEnd
		}
		if !end.After(start) {
			continue
		}

		summary, ok := projectSummaries[task.ProjectName]
		if !ok {
			summary = &WeeklySummary{ProjectName: task.ProjectName}
			projectSummaries[task.ProjectName] = summary
		}

		// Split each clipped task overlap into day-sized segments so each segment
		// can be accumulated into the correct Monday–Sunday bucket.
		segmentDayStart := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
		for segmentDayStart.Before(end) {
			nextDay := segmentDayStart.AddDate(0, 0, 1)
			segmentStart := start
			if segmentStart.Before(segmentDayStart) {
				segmentStart = segmentDayStart
			}
			segmentEnd := end
			if segmentEnd.After(nextDay) {
				segmentEnd = nextDay
			}
			if segmentEnd.After(segmentStart) {
				segmentDuration := segmentEnd.Sub(segmentStart)
				if dayIdx := weekDayIndex(segmentDayStart, windowStart); dayIdx >= 0 {
					summary.DailyDurations[dayIdx] += segmentDuration
				}
				summary.Duration += segmentDuration
			}
			segmentDayStart = nextDay
		}
	}

	if len(projectSummaries) == 0 {
		return nil
	}

	summaries := make([]WeeklySummary, 0, len(projectSummaries))
	var maxDuration time.Duration
	for _, summary := range projectSummaries {
		summaries = append(summaries, *summary)
		if summary.Duration > maxDuration {
			maxDuration = summary.Duration
		}
	}

	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].Duration != summaries[j].Duration {
			return summaries[i].Duration > summaries[j].Duration
		}
		return summaries[i].ProjectName < summaries[j].ProjectName
	})

	if maxDuration > 0 {
		for i := range summaries {
			summaries[i].Percentage = float64(summaries[i].Duration) / float64(maxDuration)
		}
	}

	return summaries
}

func weekDayIndex(dayStart time.Time, weekStart time.Time) int {
	loc := weekStart.Location()
	dayStart = dayStart.In(loc)
	dayYear, dayMonth, dayDate := dayStart.Date()
	for i := 0; i < 7; i++ {
		candidate := weekStart.AddDate(0, 0, i)
		cy, cm, cd := candidate.Date()
		if dayYear == cy && dayMonth == cm && dayDate == cd {
			return i
		}
	}
	return -1
}

type TaskGroup struct {
	Date  time.Time
	Tasks []*Task
}

type ItemType int

const (
	ItemTypeHeader ItemType = iota
	ItemTypeTask
)

type FlatListItem struct {
	Type     ItemType
	Title    string
	Subtitle string
	Task     *Task
}

func FlattenTaskGroups(groups []TaskGroup) []FlatListItem {
	// Pre-calculate capacity: 1 header + tasks per group
	capacity := 0
	for _, group := range groups {
		capacity += 1 + len(group.Tasks)
	}
	items := make([]FlatListItem, 0, capacity)
	for _, group := range groups {
		// Date header
		var totalDuration time.Duration
		for _, task := range group.Tasks {
			totalDuration += task.Duration
		}
		items = append(items, FlatListItem{
			Type:     ItemTypeHeader,
			Title:    group.Date.Format("Monday, January 2"),
			Subtitle: fmt.Sprintf("Total: %v", totalDuration.Round(time.Second)),
		})

		// Tasks
		for _, task := range group.Tasks {
			items = append(items, FlatListItem{
				Type:     ItemTypeTask,
				Title:    task.ProjectName,
				Subtitle: fmt.Sprintf("%s (%v)", task.Description, task.Duration.Round(time.Second)),
				Task:     task,
			})
		}
	}
	return items
}

func GroupTasksByDate(tasks []*Task) []TaskGroup {
	// Create a map to group tasks by date
	type dateKey struct {
		y int
		m time.Month
		d int
	}
	groups := make(map[dateKey][]*Task)
	for _, task := range tasks {
		y, m, d := task.StartTime.Date()
		key := dateKey{y, m, d}
		groups[key] = append(groups[key], task)
	}

	// Convert map to slice and sort by date
	var taskGroups []TaskGroup
	for key, tasksInGroup := range groups {
		// Use the location of the first task to avoid timezone drift
		loc := tasksInGroup[0].StartTime.Location()
		date := time.Date(key.y, key.m, key.d, 0, 0, 0, 0, loc)
		// Sort tasks within each group by start time
		sort.Slice(tasksInGroup, func(i, j int) bool {
			return tasksInGroup[i].StartTime.After(tasksInGroup[j].StartTime)
		})

		taskGroups = append(taskGroups, TaskGroup{
			Date:  date,
			Tasks: tasksInGroup,
		})
	}

	// Sort groups by date (newest first)
	sort.Slice(taskGroups, func(i, j int) bool {
		return taskGroups[i].Date.After(taskGroups[j].Date)
	})

	return taskGroups
}
