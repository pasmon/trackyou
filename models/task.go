package models

import "time"

// Task represents a time tracking task
type Task struct {
	ID          int64
	ProjectName string
	Description string
	StartTime   time.Time
	EndTime     time.Time
	Duration    time.Duration
}

// NewTask creates a new task with the current time as start time
func NewTask(projectName, description string) *Task {
	now := time.Now()
	return &Task{
		ProjectName: projectName,
		Description: description,
		StartTime:   now,
		EndTime:     now,
		Duration:    0,
	}
}

// StopTask marks the task as completed and calculates the duration
func (t *Task) StopTask() {
	t.EndTime = time.Now()
	t.Duration = t.EndTime.Sub(t.StartTime)
}

// UpdateDuration updates the task duration based on start and end times
func (t *Task) UpdateDuration() {
	t.Duration = t.EndTime.Sub(t.StartTime)
}
