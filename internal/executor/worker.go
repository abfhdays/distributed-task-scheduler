package executor

import (
    "log"
    "time"
    "distributed-task-scheduler-go/internal/models"
    "distributed-task-scheduler-go/internal/queue"
)

type Worker struct {
    ID       int
    queue    *queue.TaskQueue
    executor *TaskExecutor
    results  chan *models.Task
    done     chan bool
}

func NewWorker(id int, taskQueue *queue.TaskQueue, results chan *models.Task) *Worker {
    return &Worker{
        ID:       id,
        queue:    taskQueue,
        executor: NewTaskExecutor(5 * time.Minute),
        results:  results,
        done:     make(chan bool),
    }
}

func (w *Worker) Start() {
    go func() {
        log.Printf("[Worker %d] Started", w.ID)
        
        for {
            select {
            case <-w.done:
                log.Printf("[Worker %d] Shutting down", w.ID)
                return
            default:
                // Pull task from queue (blocking)
                task, ok := w.queue.Dequeue()
                if !ok {
                    // Queue closed
                    return
                }
                
                w.executeTask(task)
            }
        }
    }()
}

func (w *Worker) executeTask(task *models.Task) {
    log.Printf("[Worker %d] Executing task: %s", w.ID, task.Name)
    
    task.Status = models.StatusRunning
    
    err := w.executor.Execute(task)
    
    if err != nil {
        log.Printf("[Worker %d] Task %s failed: %v", w.ID, task.Name, err)
        
        if task.RetryCount < task.MaxRetries {
            task.RetryCount++
            task.Status = models.StatusRetrying
            log.Printf("[Worker %d] Retrying task %s (attempt %d/%d)", 
                w.ID, task.Name, task.RetryCount, task.MaxRetries)
            
            // Exponential backoff
            backoff := time.Duration(task.RetryCount) * time.Second
            time.Sleep(backoff)
            
            // Re-enqueue for retry
            w.queue.Enqueue(task)
            return
        }
        
        task.Status = models.StatusFailed
    } else {
        task.Status = models.StatusSuccess
        log.Printf("[Worker %d] Task %s completed successfully", w.ID, task.Name)
    }
    
    // Send result back to coordinator
    w.results <- task
}

func (w *Worker) Stop() {
    w.done <- true
}