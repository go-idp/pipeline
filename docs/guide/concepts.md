# Core Concepts

Understanding Pipeline's core concepts helps you use and understand Pipeline better.

## Pipeline

Pipeline is the top-level execution unit that contains multiple Stages. A Pipeline represents a complete workflow.

```yaml
name: My Pipeline
stages:
  - name: stage1
    jobs: [...]
```

## Stage

Stage is an execution phase of a Pipeline that can contain multiple Jobs, supporting parallel or serial execution.

```yaml
stages:
  - name: build
    run_mode: parallel  # parallel or serial
    jobs: [...]
```

### Stage Execution Modes

- **parallel**: All Jobs execute simultaneously, suitable for independent tasks
- **serial**: Jobs execute sequentially, suitable for tasks with dependencies

## Job

Job is a task unit in a Stage that contains multiple Steps. Steps execute sequentially.

```yaml
jobs:
  - name: build-job
    steps: [...]
```

## Step

Step is the smallest execution unit that executes specific commands or operations.

```yaml
steps:
  - name: compile
    command: make build
    image: golang:1.20
```

## Execution Flow

```
Pipeline.prepare() - Preparation Phase
  ├── Create working directory
  ├── Set environment variables
  ├── Add pre/post hooks
  └── Initialize all Stages

Pipeline.Run() - Execution Phase
  └── Iterate through all Stages
      └── Stage.Run()
          └── Execute Jobs based on RunMode
              ├── serial: Execute serially
              └── parallel: Execute in parallel
                  └── Job.Run()
                      └── Execute all Steps sequentially
                          └── Step.Run()
                              └── Execute commands based on Engine

Pipeline.clean() - Cleanup Phase
  └── Clean working directory
```

## Error Handling

### Failure Behavior

- Any Step failure causes the entire Job to fail
- Any Job failure causes the entire Stage to fail
- Any Stage failure causes the entire Pipeline to fail
- Execution stops immediately after failure, subsequent steps are not executed

## Configuration Inheritance

Configuration is inherited in the following hierarchy: **Pipeline → Stage → Job → Step**

Each level inherits configuration from its parent, with child-level configuration taking higher priority.

## Best Practices

### 1. Use Parallel Execution for Efficiency

For independent tasks, use parallel execution:

```yaml
stages:
  - name: build
    run_mode: parallel
    jobs:
      - name: build-frontend
        steps: [...]
      - name: build-backend
        steps: [...]
```

### 2. Use Serial Execution for Dependencies

For tasks with dependencies, use serial execution:

```yaml
stages:
  - name: build
    run_mode: serial
    jobs:
      - name: build-step1
        steps: [...]
      - name: build-step2
        steps: [...]  # Depends on step1
```
