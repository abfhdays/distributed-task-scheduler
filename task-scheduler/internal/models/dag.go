package models

import "fmt"

type DAG struct {
    Name  string           // DAG name
    Tasks map[string]*Task // Map of task name -> Task
}

func NewDAG(name string) *DAG {
    return &DAG{
        Name:  name,
        Tasks: make(map[string]*Task),
    }
}

func (d *DAG) AddTask(task *Task) error {
    if _, exists := d.Tasks[task.Name]; exists {
        return fmt.Errorf("task %s already exists", task.Name)
    }
    d.Tasks[task.Name] = task
    return nil
}

func (d *DAG) GetTask(name string) (*Task, error) {
    task, exists := d.Tasks[name]
    if !exists {
        return nil, fmt.Errorf("task %s not found", name)
    }
    return task, nil
}

// Validate checks for cyclic dependencies
func (d *DAG) Validate() error {
    // Check all dependencies exist
    for _, task := range d.Tasks {
        for _, dep := range task.Dependencies {
            if _, exists := d.Tasks[dep]; !exists {
                return fmt.Errorf("task %s depends on non-existent task %s", task.Name, dep)
            }
        }
    }
    
    // Check for cycles using DFS
    visited := make(map[string]bool)
    recStack := make(map[string]bool)
    
    var hasCycle func(string) bool
    hasCycle = func(taskName string) bool {
        visited[taskName] = true
        recStack[taskName] = true
        
        task := d.Tasks[taskName]
        for _, dep := range task.Dependencies {
            if !visited[dep] {
                if hasCycle(dep) {
                    return true
                }
            } else if recStack[dep] {
                return true
            }
        }
        
        recStack[taskName] = false
        return false
    }
    
    for taskName := range d.Tasks {
        if !visited[taskName] {
            if hasCycle(taskName) {
                return fmt.Errorf("cycle detected in DAG")
            }
        }
    }
    
    return nil
}