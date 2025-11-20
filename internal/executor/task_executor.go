package executor

import (
    "bytes"
    "context"
    "fmt"
    "os/exec"
    "time"
    "distributed-task-scheduler-go/internal/models"
)

type TaskExecutor struct {
    timeout time.Duration
}

func NewTaskExecutor(timeout time.Duration) *TaskExecutor {
    return &TaskExecutor{
        timeout: timeout,
    }
}

func (e *TaskExecutor) Execute(task *models.Task) error {
    ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
    defer cancel()
    
    var cmd *exec.Cmd
    switch task.Type {
    case models.TaskTypeBash:
        cmd = exec.CommandContext(ctx, "bash", "-c", task.Command)
    case models.TaskTypePython:
        cmd = exec.CommandContext(ctx, "python3", "-c", task.Command)
    default:
        return fmt.Errorf("unsupported task type: %s", task.Type)
    }
    
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    
    task.StartTime = time.Now()
    err := cmd.Run()
    task.EndTime = time.Now()
    
    task.Output = stdout.String()
    if stderr.Len() > 0 {
        task.Output += "\nSTDERR:\n" + stderr.String()
    }
    
    if err != nil {
        task.Error = err
        return err
    }
    
    return nil
}