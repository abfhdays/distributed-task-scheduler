package coordinator

import (
    "fmt"
    "log"
    "sync"
    "time"
    "github.com/aarushghosh/task-scheduler/internal/models"
    "github.com/aarushghosh/task-scheduler/internal/queue"
    "github.com/aarushghosh/task-scheduler/internal/executor"
    "github.com/aarushghosh/task-scheduler/internal/resolver"
)

type Coordinator struct {
    dag         *models.DAG
    taskQueue   *queue.TaskQueue
    workers     []*executor.Worker
    results     chan *models.Task
    numWorkers  int
    mu          sync.Mutex
    completed   int
    failed      int
}

func NewCoordinator(dag *models.DAG, numWorkers int) *Coordinator {
    return &Coordinator{
        dag:        dag,
        taskQueue:  queue.NewTaskQueue(1000),
        workers:    make([]*executor.Worker, numWorkers),
        results:    make(chan *models.Task, 100),
        numWorkers: numWorkers,
        completed:  0,
        failed:     0,
    }
}

func (c *Coordinator) Start() error {
    log.Printf("Starting coordinator with %d workers", c.numWorkers)
    
    // Start workers
    for i := 0; i < c.numWorkers; i++ {
        worker := executor.NewWorker(i, c.taskQueue, c.results)
        c.workers[i] = worker
        worker.Start()
    }
    
    // Start result processor
    go c.processResults()
    
    // Initial dispatch
    c.dispatchReadyTasks()
    
    // Wait for completion
    return c.waitForCompletion()
}

func (c *Coordinator) dispatchReadyTasks() {
    readyTasks := resolver.GetReadyTasks(c.dag)
    
    for _, task := range readyTasks {
        task.Status = models.StatusQueued
        if err := c.taskQueue.Enqueue(task); err != nil {
            log.Printf("Failed to enqueue task %s: %v", task.Name, err)
        } else {
            log.Printf("Dispatched task: %s", task.Name)
        }
    }
}

func (c *Coordinator) processResults() {
    for task := range c.results {
        c.mu.Lock()
        
        if task.Status == models.StatusSuccess {
            c.completed++
            log.Printf("Task %s completed (%d/%d)", task.Name, c.completed, len(c.dag.Tasks))
            
            // Dispatch newly ready tasks
            c.dispatchReadyTasks()
        } else if task.Status == models.StatusFailed {
            c.failed++
            log.Printf("Task %s failed permanently (%d failures)", task.Name, c.failed)
        }
        
        c.mu.Unlock()
    }
}

func (c *Coordinator) waitForCompletion() error {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    timeout := time.After(10 * time.Minute)
    
    for {
        select {
        case <-timeout:
            return fmt.Errorf("execution timeout")
        case <-ticker.C:
            c.mu.Lock()
            total := c.completed + c.failed
            
            if total == len(c.dag.Tasks) {
                c.mu.Unlock()
                
                // Shutdown
                c.taskQueue.Close()
                for _, worker := range c.workers {
                    worker.Stop()
                }
                close(c.results)
                
                if c.failed > 0 {
                    return fmt.Errorf("DAG completed with %d failures", c.failed)
                }
                
                log.Printf("DAG completed successfully!")
                return nil
            }
            c.mu.Unlock()
        }
    }
}

func (c *Coordinator) PrintStatus() {
    fmt.Println("\n=== DAG Execution Status ===")
    for _, task := range c.dag.Tasks {
        duration := task.EndTime.Sub(task.StartTime)
        fmt.Printf("Task: %-15s Status: %-10s Duration: %v\n", 
            task.Name, task.Status, duration)
    }
    fmt.Printf("\nTotal: %d tasks | Completed: %d | Failed: %d\n", 
        len(c.dag.Tasks), c.completed, c.failed)
}