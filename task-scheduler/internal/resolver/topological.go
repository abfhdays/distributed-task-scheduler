package resolver

import (
    "fmt"
    "github.com/aarushghosh/task-scheduler/internal/models"
)

// TopologicalSort returns tasks in executable order
func TopologicalSort(dag *models.DAG) ([]*models.Task, error) {
    // Calculate in-degrees (number of dependencies)
    inDegree := make(map[string]int)
    for name := range dag.Tasks {
        inDegree[name] = 0
    }
    for _, task := range dag.Tasks {
        for _, dep := range task.Dependencies {
            inDegree[task.Name]++
        }
    }
    
    // Queue of tasks with no dependencies
    queue := make([]*models.Task, 0)
    for name, degree := range inDegree {
        if degree == 0 {
            queue = append(queue, dag.Tasks[name])
        }
    }
    
    // Process queue
    result := make([]*models.Task, 0, len(dag.Tasks))
    for len(queue) > 0 {
        // Dequeue
        current := queue[0]
        queue = queue[1:]
        result = append(result, current)
        
        // Find tasks that depend on current
        for _, task := range dag.Tasks {
            for _, dep := range task.Dependencies {
                if dep == current.Name {
                    inDegree[task.Name]--
                    if inDegree[task.Name] == 0 {
                        queue = append(queue, task)
                    }
                }
            }
        }
    }
    
    // Check if all tasks were processed (no cycles)
    if len(result) != len(dag.Tasks) {
        return nil, fmt.Errorf("cycle detected in DAG")
    }
    
    return result, nil
}

// GetReadyTasks returns tasks whose dependencies are all satisfied
func GetReadyTasks(dag *models.DAG) []*models.Task {
    ready := make([]*models.Task, 0)
    
    for _, task := range dag.Tasks {
        if task.Status != models.StatusPending {
            continue
        }
        
        // Check if all dependencies are complete
        allDepsComplete := true
        for _, depName := range task.Dependencies {
            dep := dag.Tasks[depName]
            if dep.Status != models.StatusSuccess {
                allDepsComplete = false
                break
            }
        }
        
        if allDepsComplete {
            ready = append(ready, task)
        }
    }
    
    return ready
}