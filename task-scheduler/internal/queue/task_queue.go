package queue

import (
    "sync"
    "github.com/aarushghosh/task-scheduler/internal/models"
)

type TaskQueue struct {
    tasks chan *models.Task
    mu    sync.Mutex
    size  int
}

func NewTaskQueue(bufferSize int) *TaskQueue {
    return &TaskQueue{
        tasks: make(chan *models.Task, bufferSize),
        size:  0,
    }
}

func (q *TaskQueue) Enqueue(task *models.Task) error {
    q.mu.Lock()
    q.size++
    q.mu.Unlock()
    
    select {
    case q.tasks <- task:
        return nil
    default:
        q.mu.Lock()
        q.size--
        q.mu.Unlock()
        return fmt.Errorf("queue is full")
    }
}

func (q *TaskQueue) Dequeue() (*models.Task, bool) {
    task, ok := <-q.tasks
    if ok {
        q.mu.Lock()
        q.size--
        q.mu.Unlock()
    }
    return task, ok
}

func (q *TaskQueue) Close() {
    close(q.tasks)
}

func (q *TaskQueue) Size() int {
    q.mu.Lock()
    defer q.mu.Unlock()
    return q.size
}