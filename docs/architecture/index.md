# Architecture Overview

Pipeline is a lightweight CI/CD pipeline execution engine written in Go, similar to Jenkins and GitLab CI, but more lightweight and flexible. It supports multiple execution engines, plugin systems, service orchestration, and more.

## Core Architecture

Pipeline uses a four-layer architecture model:

```
Pipeline (管道)
  └── Stage (阶段)
      └── Job (任务)
          └── Step (步骤)
```

- **Pipeline**: The container for the entire pipeline, containing multiple stages
- **Stage**: A phase of the pipeline that can execute multiple tasks serially or in parallel
- **Job**: A task unit containing multiple steps, executed sequentially
- **Step**: The smallest execution unit that executes specific commands or operations

## Execution Flow

```
1. Pipeline.prepare() - Preparation Phase
   ├── Create working directory
   ├── Set environment variables
   ├── Add pre/post hooks
   └── Initialize all Stages

2. Pipeline.Run() - Execution Phase
   └── Iterate through all Stages
       └── Stage.Run()
           └── Execute Jobs based on RunMode
               ├── serial: Execute serially
               └── parallel: Execute in parallel
                   └── Job.Run()
                       └── Execute all Steps sequentially
                           └── Step.Run()
                               └── Execute commands based on Engine

3. Pipeline.clean() - Cleanup Phase
   └── Clean working directory
```

## Core Modules

### Pipeline Module

**Main Files**:
- `pipeline.go`: Pipeline structure definition and basic methods
- `run.go`: Pipeline execution logic
- `state.go`: Pipeline state management

**Core Features**:
- Pipeline lifecycle management
- Environment variable management
- Working directory management
- Pre/Post hook support
- Timeout control (default 86400 seconds)

### Stage Module

**Main Files**:
- `stage.go`: Stage structure definition
- `run.go`: Stage execution logic
- `setup.go`: Stage initialization
- `state.go`: Stage state management
- `constants.go`: Run mode constants

**Core Features**:
- Stage-level task orchestration
- Support for serial and parallel execution modes
- Environment variable and working directory inheritance

### Job Module

**Main Files**:
- `job.go`: Job structure definition
- `run.go`: Job execution logic
- `setup.go`: Job initialization
- `state.go`: Job state management

**Core Features**:
- Job-level step management
- Steps executed sequentially
- Environment variable and working directory inheritance

### Step Module

**Main Files**:
- `step.go`: Step structure definition
- `run.go`: Step execution logic
- `setup.go`: Step initialization
- `state.go`: Step state management
- `plugin.go`: Plugin support
- `service.go`: Service orchestration support
- `init.go`: Execution engine registration

**Core Features**:
- Command execution
- Multiple execution engine support
- Plugin system
- Language runtime support
- Service orchestration support

## Execution Engines

Pipeline supports multiple execution engines:

- **Host**: Execute commands directly on the local host (default)
- **Docker**: Execute commands in Docker containers
- **SSH**: Execute commands on remote servers via SSH
- **IDP**: Execute commands by connecting to remote IDP Agent via WebSocket

## Plugin System

Pipeline supports a plugin mechanism that encapsulates reusable functionality through Docker images. Plugins can:
- Encapsulate complex operation logic
- Provide language runtime support
- Integrate third-party services

## Service Orchestration

Pipeline supports service orchestration, allowing you to start dependent services (such as databases, caches, etc.) for Pipeline use.

## Configuration Inheritance

Configuration is inherited in the following hierarchy: **Pipeline → Stage → Job → Step**

Each level inherits configuration from its parent, with child-level configuration taking higher priority.

## More Information

- [System Architecture](./system.md) - Detailed system architecture
- [Error Handling](./error-handling.md) - Error handling mechanisms
- [Performance Optimization](./optimization.md) - Performance optimization recommendations
