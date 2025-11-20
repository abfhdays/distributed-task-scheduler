package parser

import (
    "os"
    "gopkg.in/yaml.v3"
    "github.com/aarushghosh/task-scheduler/internal/models"
)

type YAMLTask struct {
    Name         string   `yaml:"name"`
    Type         string   `yaml:"type"`
    Command      string   `yaml:"command"`
    Dependencies []string `yaml:"dependencies"`
    MaxRetries   int      `yaml:"max_retries"`
}

type YAMDAG struct {
    Name  string     `yaml:"name"`
    Tasks []YAMLTask `yaml:"tasks"`
}

func ParseYAML(filepath string) (*models.DAG, error) {
    // Read file
    data, err := os.ReadFile(filepath)
    if err != nil {
        return nil, err
    }
    
    // Parse YAML
    var yamlDAG YAMDAG
    if err := yaml.Unmarshal(data, &yamlDAG); err != nil {
        return nil, err
    }
    
    // Convert to DAG
    dag := models.NewDAG(yamlDAG.Name)
    for _, yamlTask := range yamlDAG.Tasks {
        task := models.NewTask(
            yamlTask.Name,
            models.TaskType(yamlTask.Type),
            yamlTask.Command,
            yamlTask.Dependencies,
            yamlTask.MaxRetries,
        )
        if err := dag.AddTask(task); err != nil {
            return nil, err
        }
    }
    
    // Validate DAG
    if err := dag.Validate(); err != nil {
        return nil, err
    }
    
    return dag, nil
}