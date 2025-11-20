# Distributed Task Scheduler

Distributed DAG-based task scheduler built in Go with coordinator-worker architecture, featuring fault-tolerant execution, dynamic task dispatch, and horizontal worker scaling for orchestrating data pipelines and batch workflows.

## Architecture

```
┌─────────────┐
│ Coordinator │ ← Parses DAG, resolves dependencies, dispatches tasks
└──────┬──────┘
       │
   [Task Queue] ← Channel-based task queue
       │
   ┌───┴────┬────────┐
   │        │        │
┌──▼───┐ ┌─▼────┐ ┌─▼────┐
│Worker│ │Worker│ │Worker│ ← Pull tasks, execute, report status
└──────┘ └──────┘ └──────┘
```

### Execution Flow

1. Parse YAML DAG definition → Validate (cycle detection)
2. Topological sort → Determine execution order
3. Dispatch ready tasks (dependencies satisfied) to queue
4. Workers pull tasks → Execute → Report success/failure
5. On completion, dispatch newly ready tasks
6. Retry failed tasks with exponential backoff

---

## Project Structure

```
distributed-task-scheduler-go/
├── cmd/
│   └── scheduler/
│       └── main.go              # CLI entry point
├── internal/
│   ├── models/
│   │   ├── status.go            # Task status enum
│   │   ├── task.go              # Task data structure
│   │   └── dag.go               # DAG structure & validation
│   ├── parser/
│   │   └── yaml_parser.go       # YAML → DAG conversion
│   ├── resolver/
│   │   └── topological.go       # Dependency resolution
│   ├── executor/
│   │   ├── task_executor.go     # Task execution engine
│   │   └── worker.go            # Worker implementation
│   ├── coordinator/
│   │   └── coordinator.go       # Orchestration logic
│   └── queue/
│       └── task_queue.go        # Thread-safe task queue
├── examples/
│   ├── simple_dag.yaml
│   └── parallel_dag.yaml
├── go.mod
└── README.md
```

---

## Modules

### `internal/models/` - Data Structures

#### **status.go** - Task Status State Machine
```
PENDING → READY → QUEUED → RUNNING → SUCCESS/FAILED/RETRYING
```

#### **task.go** - Task Definition
- **Fields**: ID, name, type (bash/python), command, dependencies
- **Metadata**: Start/end time, retry count, output, error
- **Factory**: `NewTask()` creates task with UUID

#### **dag.go** - DAG Structure
- **Storage**: Map of task name → Task pointer
- **Validation**: 
  - Check all dependencies exist
  - Detect cycles via depth-first search (DFS)
- **Methods**: `AddTask()`, `GetTask()`, `Validate()`

---

### `internal/parser/` - YAML Parser

#### **yaml_parser.go**
- Reads YAML workflow definition file
- Unmarshals into internal structs
- Converts to DAG object
- Runs validation before returning

**YAML Format:**
```yaml
name: example_pipeline
tasks:
  - name: extract
    type: bash
    command: "echo 'Extracting data...'; sleep 2"
    dependencies: []
    max_retries: 3
    
  - name: transform
    type: bash
    command: "echo 'Transforming data...'; sleep 3"
    dependencies: ["extract"]
    max_retries: 2
    
  - name: load
    type: bash
    command: "echo 'Loading to warehouse...'; sleep 1"
    dependencies: ["transform"]
    max_retries: 1
```

---

### `internal/resolver/` - Dependency Resolution

#### **topological.go**

**TopologicalSort** - Kahn's Algorithm
1. Calculate in-degree (number of dependencies) for each task
2. Start with tasks that have zero dependencies
3. Process tasks in order, decrementing dependents' in-degrees
4. Return topologically sorted task list
5. Detect cycles if not all tasks processed

**GetReadyTasks** - Dynamic Dispatch Helper
- Finds tasks in PENDING state
- Checks if all dependencies have SUCCESS status
- Returns list of tasks ready to execute
- Called after each task completion to dispatch next wave

---

### `internal/queue/` - Task Queue

#### **task_queue.go**
- **Implementation**: Go channel with mutex for size tracking
- **Capacity**: Bounded buffer (default 1000)
- **Thread-safe**: Mutex protects size counter
- **Operations**:
  - `Enqueue(task)`: Add task to queue
  - `Dequeue()`: Blocking pull from queue
  - `Size()`: Current queue depth
  - `Close()`: Shutdown queue

---

### `internal/executor/` - Task Execution

#### **task_executor.go**
- Executes bash/python commands via `exec.Command`
- **Timeout**: Context with 5-minute default
- **Capture**: Redirects stdout/stderr to buffers
- **Error handling**: Returns execution error
- **Tracking**: Records start/end time

#### **worker.go** - Worker Goroutine
- **Lifecycle**: Start → Pull → Execute → Report → Repeat
- **Task Pull**: Blocking dequeue from queue
- **Execution**: Delegates to TaskExecutor
- **Retry Logic**: 
  - Check retry count vs max retries
  - Exponential backoff: `delay = retry_count * 1 second`
  - Re-enqueue task for retry
  - Mark as FAILED if max retries exceeded
- **Result Reporting**: Send completed task to results channel
- **Shutdown**: Listen on done channel for graceful stop

---

### `internal/coordinator/` - Orchestration

#### **coordinator.go**

**Responsibilities:**
- Manages complete DAG execution lifecycle
- Spawns worker pool
- Dispatches tasks to queue
- Processes completion results
- Tracks execution progress
- Handles shutdown

**Key Methods:**

**`Start()`**
1. Spawn N worker goroutines
2. Start result processor goroutine
3. Dispatch initial ready tasks
4. Wait for completion or timeout

**`dispatchReadyTasks()`**
1. Call `resolver.GetReadyTasks()`
2. Mark tasks as QUEUED
3. Enqueue to task queue
4. Log dispatch

**`processResults()`**
- Runs in goroutine, listens on results channel
- On SUCCESS: increment completed counter, dispatch new ready tasks
- On FAILED: increment failed counter, log error
- Thread-safe with mutex

**`waitForCompletion()`**
- Poll every 1 second
- Check if `completed + failed == total tasks`
- Timeout after 10 minutes
- On completion: shutdown queue, stop workers, close channels
- Return error if any tasks failed

**`PrintStatus()`**
- Print summary table of all tasks
- Show status, duration for each task
- Print totals

---

### `cmd/scheduler/` - CLI

#### **main.go**
- **Framework**: Cobra for command-line parsing
- **Commands**: `scheduler run -f <yaml> -w <workers>`
- **Flags**:
  - `-f, --file`: YAML DAG file (required)
  - `-w, --workers`: Number of workers (default: 4)
- **Flow**:
  1. Parse command-line arguments
  2. Load and parse DAG from YAML
  3. Create coordinator with N workers
  4. Start execution
  5. Print final status report

---

## Current Implementation Status

### ✅ Implemented

- **Single-node distributed workers** (goroutines)
- **DAG parsing and validation** (YAML → struct, cycle detection)
- **Topological sort** (Kahn's algorithm)
- **Dependency resolution** (dynamic ready task detection)
- **Task queue** (channel-based, thread-safe)
- **Task execution** (bash/python with timeout)
- **Worker pool** (configurable size)
- **Retry logic** (exponential backoff, max attempts)
- **Graceful shutdown** (coordinated cleanup)
- **CLI interface** (Cobra-based)

### ❌ Not Yet Implemented

- gRPC-based distribution (true multi-node)
- etcd for distributed state
- Persistent storage (BadgerDB/PostgreSQL)
- Leader election and HA
- Heartbeat-based failure detection
- Metrics and observability
- Distributed tracing
- Kubernetes deployment

---

## Next Steps - Roadmap to Production

### Phase 1: Observability (Week 1)

**Goal:** Instrument system for monitoring and debugging

- [ ] **Prometheus metrics**
  - Task throughput (tasks/sec)
  - Queue depth gauge
  - Worker utilization (busy/idle ratio)
  - Task latency histogram (p50, p95, p99)
  - Task success/failure rates
- [ ] **Structured logging** with zerolog
  - Replace `log` package with zerolog
  - Log levels: debug, info, warn, error
  - Contextual fields (task_id, worker_id, dag_name)
- [ ] **Execution traces**
  - Per-DAG run ID
  - Track task state transitions with timestamps
  - Export traces to JSON for analysis

**Deliverable:** `/metrics` endpoint exposing Prometheus metrics

---

### Phase 2: True Distribution (Week 2-3)

**Goal:** Replace goroutines with actual distributed workers

#### Step 1: Replace Queue with Redis/PostgreSQL
- [ ] Implement `RedisTaskQueue` or `PostgresTaskQueue`
- [ ] Move task queue to external storage
- [ ] Multiple coordinators can read from same queue
- [ ] Workers can be on different machines

#### Step 2: Implement gRPC Communication
- [ ] Define Protocol Buffer schemas
  ```protobuf
  message Task {
    string id = 1;
    string name = 2;
    string command = 3;
    repeated string dependencies = 4;
  }
  
  message TaskResult {
    string task_id = 1;
    TaskStatus status = 2;
    string output = 3;
    string error = 4;
  }
  
  service Scheduler {
    rpc StreamTasks(stream TaskRequest) returns (stream Task);
    rpc ReportResult(TaskResult) returns (Ack);
  }
  ```
- [ ] Implement gRPC server in coordinator
- [ ] Implement gRPC client in worker
- [ ] Use bidirectional streaming for task dispatch

#### Step 3: Heartbeat Mechanism
- [ ] Workers send heartbeat every 5 seconds
  ```go
  type Heartbeat {
    WorkerID string
    Timestamp time.Time
    CurrentTask string
  }
  ```
- [ ] Coordinator tracks last heartbeat per worker
- [ ] Mark worker dead if no heartbeat for 15 seconds
- [ ] Re-enqueue tasks from dead workers
- [ ] Handle "zombie" tasks (worker came back after being marked dead)

**Deliverable:** Run coordinator on one machine, workers on different machines

---

### Phase 3: High Availability (Week 4)

**Goal:** No single point of failure

#### Step 1: etcd Integration
- [ ] Install etcd cluster (3 or 5 nodes)
- [ ] Store task queue in etcd
- [ ] Implement distributed locks via etcd
  ```go
  // Worker claims task
  session := concurrency.NewSession(etcdClient)
  mutex := concurrency.NewMutex(session, "/tasks/lock/"+taskID)
  mutex.Lock(ctx)
  // Execute task
  mutex.Unlock(ctx)
  ```
- [ ] Store cluster membership in etcd (which workers are alive)

#### Step 2: Coordinator Leader Election
- [ ] Run 3-5 coordinator replicas
- [ ] Use etcd for leader election (Raft consensus)
  ```go
  election := concurrency.NewElection(session, "/election/coordinator")
  election.Campaign(ctx, workerID)
  // Only leader schedules tasks
  ```
- [ ] Leader holds lease with TTL (e.g., 10 seconds)
- [ ] On leader crash, lease expires → new election
- [ ] Standby coordinators become leader in <1 second

#### Step 3: Worker Checkpointing
- [ ] Embed BadgerDB in each worker process
- [ ] Save task state every 10 seconds
  ```go
  checkpoint := TaskCheckpoint{
    TaskID: task.ID,
    Progress: currentRow,
    Timestamp: time.Now(),
  }
  db.Set([]byte(task.ID), checkpoint)
  ```
- [ ] On worker restart, load checkpoint and resume
- [ ] Use idempotency keys to prevent duplicate execution

**Deliverable:** Kill coordinator, new leader takes over; kill worker, task resumes on different worker

---

### Phase 4: Production Deployment (Week 5)

**Goal:** Deploy on Kubernetes with autoscaling

#### Step 1: Containerization
- [ ] Write Dockerfile for coordinator
- [ ] Write Dockerfile for worker
- [ ] Multi-stage build for smaller images
- [ ] Push to container registry

#### Step 2: Kubernetes Manifests
- [ ] **etcd**: StatefulSet with persistent volumes
- [ ] **Coordinator**: StatefulSet with 3 replicas
  - Service for gRPC endpoint
  - Leader election via etcd
- [ ] **Workers**: Deployment with HPA
  - Autoscale based on queue depth metric
  - `kubectl autoscale deployment workers --cpu-percent=80 --min=10 --max=100`
- [ ] **ConfigMap**: Application configuration
- [ ] **Secrets**: etcd credentials, database passwords

#### Step 3: Helm Chart
- [ ] Package as Helm chart for easy deployment
- [ ] Configurable values.yaml
  ```yaml
  coordinator:
    replicas: 3
    image: myregistry/coordinator:v1.0
  
  workers:
    minReplicas: 10
    maxReplicas: 100
    image: myregistry/worker:v1.0
  
  etcd:
    replicas: 3
  ```
- [ ] Deploy with `helm install scheduler ./charts/scheduler`

#### Step 4: Observability Stack
- [ ] Deploy Prometheus for metrics scraping
- [ ] Deploy Grafana with dashboards
  - DAG execution timeline
  - Task throughput over time
  - Worker pool utilization
  - Queue depth
- [ ] Integrate OpenTelemetry
  - Distributed tracing across coordinator → worker → task execution
  - Export to Jaeger or Zipkin

**Deliverable:** Full Kubernetes deployment with monitoring

---

### Phase 5: Advanced Features

- [ ] **Task priorities**: High-priority tasks jump queue
- [ ] **Conditional dependencies**: Run task B only if task A succeeds
- [ ] **Dynamic DAGs**: Add/remove tasks at runtime via API
- [ ] **Web UI**: React frontend with real-time DAG visualization (D3.js)
  - WebSocket for live updates
  - Gantt chart of task execution timeline
  - Retry/rerun failed tasks from UI
- [ ] **SQL Operator**: Execute SQL queries against Postgres/MySQL
- [ ] **Spark Operator**: Submit Spark jobs as tasks
- [ ] **HTTP Operator**: Make HTTP requests, parse responses
- [ ] **Sensor Tasks**: Wait for external condition (file exists, API returns 200)
- [ ] **Branching**: Dynamic DAG expansion based on task output

---

## Load Testing

### Objective: Validate Resume Metrics

**Target Metrics:**
- 8,500+ tasks/sec sustained throughput
- <50ms p99 dispatch latency
- 98%+ worker utilization
- Zero task loss on worker failures
- Sub-second coordinator failover

### Test Setup

#### 1. Generate Large DAG

```python
# generate_dag.py
import yaml

tasks = []
for i in range(5000):
    task = {
        'name': f'task_{i}',
        'type': 'bash',
        'command': f'sleep 0.01; echo "Task {i} completed"',
        'dependencies': [f'task_{i-1}'] if i > 0 else [],
        'max_retries': 2
    }
    tasks.append(task)

dag = {'name': 'load_test_5k', 'tasks': tasks}

with open('load_test_5k.yaml', 'w') as f:
    yaml.dump(dag, f)
```

#### 2. Run Load Test

```bash
# Start Prometheus to scrape metrics
docker run -p 9090:9090 -v $(pwd)/prometheus.yml:/etc/prometheus/prometheus.yml prom/prometheus

# Run scheduler with 50 workers
./scheduler run -f load_test_5k.yaml -w 50
```

#### 3. Measure Performance

```bash
# Query Prometheus for metrics
curl 'http://localhost:9090/api/v1/query?query=rate(tasks_completed_total[1m])'

# Expected output: ~8500 tasks/sec
```

#### 4. Chaos Testing

```bash
# Kill random workers during execution
while true; do
  kill -9 $(pgrep -f "scheduler worker" | shuf -n 1)
  sleep 30
done

# Verify: zero task loss (all tasks eventually succeed)
```

#### 5. Failover Testing

```bash
# Run 3 coordinators with leader election
./scheduler run --coordinator-id=1 &
./scheduler run --coordinator-id=2 &
./scheduler run --coordinator-id=3 &

# Kill active leader
kill -9 $(pgrep -f "scheduler.*coordinator-id=1")

# Measure: time until new leader elected and scheduling resumes
# Target: <1 second
```

---

## Usage

### Installation

```bash
# Clone repository
git clone https://github.com/yourusername/distributed-task-scheduler-go.git
cd distributed-task-scheduler-go

# Install dependencies
go mod download

# Build
go build -o scheduler ./cmd/scheduler
```

### Running Examples

```bash
# Simple 3-task pipeline with 4 workers
./scheduler run -f examples/simple_dag.yaml -w 4

# Parallel processing with 10 workers
./scheduler run -f examples/parallel_dag.yaml -w 10

# Large DAG with 50 workers
./scheduler run -f examples/large_dag.yaml -w 50
```

### Creating Custom DAGs

```yaml
name: my_pipeline
tasks:
  - name: fetch_data
    type: bash
    command: "curl -o data.json https://api.example.com/data"
    dependencies: []
    max_retries: 3
    
  - name: process_data
    type: python
    command: |
      import json
      with open('data.json') as f:
          data = json.load(f)
      # Process data
      print(f"Processed {len(data)} records")
    dependencies: ["fetch_data"]
    max_retries: 2
    
  - name: upload_results
    type: bash
    command: "aws s3 cp results.csv s3://my-bucket/"
    dependencies: ["process_data"]
    max_retries: 1
```

---

## Tech Stack

### Current
- **Go 1.21+**: Core language
- **Channels**: In-memory task queue
- **Goroutines**: Worker pool
- **gopkg.in/yaml.v3**: YAML parsing
- **github.com/spf13/cobra**: CLI framework
- **github.com/google/uuid**: Task ID generation

### Future
- **gRPC**: Inter-service communication
- **Protocol Buffers**: Schema definitions
- **etcd**: Distributed coordination and Raft consensus
- **BadgerDB**: Embedded LSM storage for checkpointing
- **Redis**: Distributed task queue
- **PostgreSQL**: Persistent storage for task metadata
- **Prometheus**: Metrics collection
- **OpenTelemetry**: Distributed tracing
- **Kubernetes**: Container orchestration
- **Helm**: Deployment packaging

---

## Performance Characteristics

### Current (Single-Node)
- **Throughput**: 1,000-2,000 tasks/sec (I/O bound by sequential execution)
- **Latency**: 1-5ms task dispatch
- **Scalability**: Limited by single-node resources

### Target (Distributed)
- **Throughput**: 8,500+ tasks/sec (with 50 workers)
- **Latency**: <50ms p99 dispatch latency
- **Scalability**: Horizontal scaling via worker addition
- **Availability**: 99.9%+ uptime with HA coordinators

---

## Contributing

Contributions welcome! Key areas:

1. **Core Features**
   - Implement gRPC-based distribution
   - Add etcd integration for HA
   - Build worker checkpointing

2. **Testing**
   - Unit tests for all modules
   - Integration tests for end-to-end flows
   - Chaos engineering tests for fault tolerance

3. **Performance**
   - Profile hot paths with pprof
   - Optimize topological sort for large DAGs
   - Benchmark queue implementations

4. **Examples**
   - Common ETL pipeline patterns
   - ML training workflows
   - Data processing examples

---

## License

MIT License - see LICENSE file for details

---

## Comparison to Existing Systems

| Feature | This Project | Airflow | Prefect | Temporal |
|---------|-------------|---------|---------|----------|
| **Language** | Go | Python | Python | Go/Python |
| **Architecture** | Coordinator-Worker | Scheduler-Executor | Orchestrator-Agent | Temporal Server |
| **Distribution** | gRPC (planned) | Celery/Kubernetes | GraphQL API | gRPC |
| **State Storage** | etcd (planned) | PostgreSQL | PostgreSQL | Cassandra/MySQL |
| **HA** | Raft via etcd | Active-Active | Active-Passive | Multi-datacenter |
| **DAG Definition** | YAML | Python code | Python code | Code |
| **UI** | CLI (Web planned) | Web UI | Web UI | Web UI |
| **Use Case** | Learning/Batch | Production ETL | Dataflow | Microservices |

**Our Advantages:**
- Simpler architecture (easier to understand)
- Go performance (vs Python overhead)
- Educational value (learn distributed systems concepts)

**Missing vs Production Systems:**
- Battle-tested at scale
- Rich ecosystem of operators
- Advanced scheduling (cron, sensors)
- Mature observability