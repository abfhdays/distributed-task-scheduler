package models

type TaskStatus string

const (
    StatusPending   TaskStatus = "PENDING"   // Waiting for dependencies
    StatusReady     TaskStatus = "READY"     // Dependencies satisfied, can run
    StatusQueued    TaskStatus = "QUEUED"    // Added to worker queue
    StatusRunning   TaskStatus = "RUNNING"   // Currently executing
    StatusSuccess   TaskStatus = "SUCCESS"   // Completed successfully
    StatusFailed    TaskStatus = "FAILED"    // Failed execution
    StatusRetrying  TaskStatus = "RETRYING"  // Failed, will retry
)