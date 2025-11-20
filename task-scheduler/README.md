Distributed Task Scheduler 
- Coordinator-worker with multiprocessing
- Task queue (Queue → gRPC later)
- DAG parsing and topological sort
- Distributed workers (Process → actual distributed later)
- Fault tolerance and retries
- CLI interface

task-scheduler/
├── cmd/
│   └── scheduler/
│       └── main.go                 # CLI entry point
├── internal/
│   ├── models/
│   │   ├── dag.go                  # DAG data structures
│   │   ├── task.go                 # Task definitions
│   │   └── status.go               # Task status types
│   ├── parser/
│   │   └── yaml_parser.go          # Parse YAML to DAG
│   ├── resolver/
│   │   └── topological.go          # Dependency resolution
│   ├── executor/
│   │   ├── worker.go               # Worker implementation
│   │   └── task_executor.go        # Execute bash/python tasks
│   ├── coordinator/
│   │   ├── coordinator.go          # Main coordinator logic
│   │   └── scheduler.go            # Task scheduling
│   └── queue/
│       └── task_queue.go           # Thread-safe task queue
├── examples/
│   ├── simple_dag.yaml
│   ├── parallel_dag.yaml
│   └── ml_pipeline.yaml
├── go.mod
├── go.sum
└── README.md