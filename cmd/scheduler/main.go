package main

import (
    "fmt"
    "log"
    "os"
    "github.com/spf13/cobra"
    "distributed-task-scheduler-go/internal/parser"
    "distributed-task-scheduler-go/internal/coordinator"
)

var (
    dagFile    string
    numWorkers int
)

var rootCmd = &cobra.Command{
    Use:   "scheduler",
    Short: "Distributed task scheduler",
    Long:  "A distributed DAG-based task scheduler written in Go",
}

var runCmd = &cobra.Command{
    Use:   "run",
    Short: "Run a DAG from YAML file",
    Run: func(cmd *cobra.Command, args []string) {
        // Parse DAG
        dag, err := parser.ParseYAML(dagFile)
        if err != nil {
            log.Fatalf("Failed to parse DAG: %v", err)
        }
        
        fmt.Printf("Loaded DAG: %s with %d tasks\n", dag.Name, len(dag.Tasks))
        
        // Create coordinator
        coord := coordinator.NewCoordinator(dag, numWorkers)
        
        // Start execution
        if err := coord.Start(); err != nil {
            log.Fatalf("Execution failed: %v", err)
        }
        
        // Print final status
        coord.PrintStatus()
    },
}

func init() {
    runCmd.Flags().StringVarP(&dagFile, "file", "f", "", "DAG YAML file (required)")
    runCmd.Flags().IntVarP(&numWorkers, "workers", "w", 4, "Number of workers")
    runCmd.MarkFlagRequired("file")
    
    rootCmd.AddCommand(runCmd)
}

func main() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}