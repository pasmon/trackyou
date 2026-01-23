package models

import (
	"testing"
	"time"
)

func TestNewTask(t *testing.T) {
	projectName := "Test Project"
	description := "Test Description"
	task := NewTask(projectName, description)

	if task.ProjectName != projectName {
		t.Errorf("expected ProjectName %s, got %s", projectName, task.ProjectName)
	}
	if task.Description != description {
		t.Errorf("expected Description %s, got %s", description, task.Description)
	}
	if task.StartTime.IsZero() {
		t.Error("expected StartTime to be set")
	}
	if task.Duration != 0 {
		t.Errorf("expected initial Duration 0, got %v", task.Duration)
	}
}

func TestStopTask(t *testing.T) {
	task := NewTask("Test", "Test")
	// Simulate some time passing
	task.StartTime = time.Now().Add(-10 * time.Minute)
	task.StopTask()

	if task.EndTime.IsZero() {
		t.Error("expected EndTime to be set")
	}
	
	// Duration should be approximately 10 minutes
	expectedDuration := 10 * time.Minute
	diff := task.Duration - expectedDuration
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Second {
		t.Errorf("expected duration around %v, got %v", expectedDuration, task.Duration)
	}
}

func TestUpdateDuration(t *testing.T) {
	task := NewTask("Test", "Test")
	task.StartTime = time.Now().Add(-5 * time.Minute)
	task.EndTime = time.Now()
	task.UpdateDuration()

	expectedDuration := task.EndTime.Sub(task.StartTime)
	if task.Duration != expectedDuration {
		t.Errorf("expected duration %v, got %v", expectedDuration, task.Duration)
	}
}
