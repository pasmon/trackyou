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

func GroupTasksByDate(tasks []*Task) []TaskGroup {
	// Create a map to group tasks by date
	groups := make(map[string][]*Task)
	for _, task := range tasks {
		date := task.StartTime.Format("2006-01-02")
		groups[date] = append(groups[date], task)
	}

	// Convert map to slice and sort by date
	var taskGroups []TaskGroup
	for dateStr, tasksInGroup := range groups {
		date, _ := time.Parse("2006-01-02", dateStr)
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

// GetTaskItemData returns structured data for a list item
func GetTaskItemData(groups []TaskGroup, id int) (title, subtitle string, itemType ItemType) {
	currentIndex := 0

	for _, group := range groups {
		// Date header
		if currentIndex == id {
			var totalDuration time.Duration
			for _, task := range group.Tasks {
				totalDuration += task.Duration
			}
			return group.Date.Format("Monday, January 2"), 
				fmt.Sprintf("Total: %v", totalDuration.Round(time.Second)), 
				ItemTypeHeader
		}
		currentIndex++

		// Project summaries
		if id < currentIndex+len(group.ProjectSummaries) {
			summary := group.ProjectSummaries[id-currentIndex]
			return summary.Name, 
				fmt.Sprintf("%v", summary.Duration.Round(time.Second)), 
				ItemTypeSummary
		}
		currentIndex += len(group.ProjectSummaries)

		// Tasks
		if id < currentIndex+len(group.Tasks) {
			task := group.Tasks[id-currentIndex]
			return task.ProjectName, 
				fmt.Sprintf("%s (%v)", task.Description, task.Duration.Round(time.Second)), 
				ItemTypeTask
		}
		currentIndex += len(group.Tasks)
	}

	return "", "", ItemTypeHeader
}

// Deprecated: Use GetTaskItemData instead (kept for now to avoid breaking if not updated)
func GetTaskItemInfo(groups []TaskGroup, id int) (text string, isHeader bool) {
	title, subtitle, itemType := GetTaskItemData(groups, id)
	if itemType == ItemTypeHeader {
		return fmt.Sprintf("=== %s (%s) ===", title, subtitle), true
	} else if itemType == ItemTypeSummary {
		return fmt.Sprintf("  Total %s: %s", title, subtitle), true
	}
	return fmt.Sprintf("    %s - %s", title, subtitle), false
}

func GetTotalItemCount(groups []TaskGroup) int {
	count := 0
	for _, group := range groups {
		count++ // Date header
		count += len(group.ProjectSummaries)
		count += len(group.Tasks)
	}
	return count
}

func GetTaskByListItemID(groups []TaskGroup, id int) *Task {
	currentIndex := 0

	for _, group := range groups {
		// Skip date header
		if currentIndex == id {
			return nil
		}
		currentIndex++

		// Skip project summaries
		if id < currentIndex+len(group.ProjectSummaries) {
			return nil
		}
		currentIndex += len(group.ProjectSummaries)

		// Check tasks
		if id < currentIndex+len(group.Tasks) {
			return group.Tasks[id-currentIndex]
		}
		currentIndex += len(group.Tasks)
	}

	return nil
}
