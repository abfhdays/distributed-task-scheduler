package models

import (
    "time"
    "github.com/google/uuid"
)

type TaskType string

const (
    TaskTypeBash   TaskType = "bash"
    TaskTypePython TaskType = "python"
)

type Task struct {
    ID           string        // Unique identifier
    Name         string        // Human-readable name
    Type         TaskType      // bash or python
    Command      string        // Command to execute
    Dependencies []string      // Task names this depends on
    Status       TaskStatus    // Current status
    RetryCount   int           // Current retry attempt
    MaxRetries   int           // Maximum retry attempts
    StartTime    time.Time     // When execution started
    EndTime      time.Time     // When execution finished
    Output       string        // stdout/stderr from execution
    Error        error         // Error if failed
}

func NewTask(name string, taskType TaskType, command string, deps []string, maxRetries int) *Task {
    return &Task{
        ID:           uuid.New().String(),
        Name:         name,
        Type:         taskType,
        Command:      command,
        Dependencies: deps,
        Status:       StatusPending,
        MaxRetries:   maxRetries,
        RetryCount:   0,
    }
}