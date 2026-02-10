package models

import (
	"fmt"
	"sort"
	"time"
)

type ProjectSummary struct {
	Name     string
	Duration time.Duration
}

type TaskGroup struct {
	Date             time.Time
	Tasks            []*Task
	ProjectSummaries []ProjectSummary
}

type ItemType int

const (
	ItemTypeHeader ItemType = iota
	ItemTypeSummary
	ItemTypeTask
)

type FlatListItem struct {
	Type     ItemType
	Title    string
	Subtitle string
	Task     *Task
}

func FlattenTaskGroups(groups []TaskGroup) []FlatListItem {
	// Pre-calculate capacity: 1 header + summaries + tasks per group
	capacity := 0
	for _, group := range groups {
		capacity += 1 + len(group.ProjectSummaries) + len(group.Tasks)
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

		// Project summaries
		for _, summary := range group.ProjectSummaries {
			items = append(items, FlatListItem{
				Type:     ItemTypeSummary,
				Title:    summary.Name,
				Subtitle: fmt.Sprintf("%v", summary.Duration.Round(time.Second)),
			})
		}

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

		// Calculate project summaries
		projectDurations := make(map[string]time.Duration)
		for _, task := range tasksInGroup {
			projectDurations[task.ProjectName] += task.Duration
		}

		var summaries []ProjectSummary
		for name, duration := range projectDurations {
			summaries = append(summaries, ProjectSummary{Name: name, Duration: duration})
		}
		// Sort summaries by name
		sort.Slice(summaries, func(i, j int) bool {
			return summaries[i].Name < summaries[j].Name
		})

		taskGroups = append(taskGroups, TaskGroup{
			Date:             date,
			Tasks:            tasksInGroup,
			ProjectSummaries: summaries,
		})
	}

	// Sort groups by date (newest first)
	sort.Slice(taskGroups, func(i, j int) bool {
		return taskGroups[i].Date.After(taskGroups[j].Date)
	})

	return taskGroups
}
