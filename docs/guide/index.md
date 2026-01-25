# Introduction

Pipeline is a powerful workflow execution engine that supports local execution and service deployment. It provides flexible configuration, rich execution engines, and a complete Web Console and REST API.

## ✨ Features

- 🚀 **Multiple Execution Modes**: Support local execution, Server mode, and Client mode
- 🐳 **Multiple Execution Engines**: Support host, docker, ssh, idp and other execution engines
- 📊 **Web Console**: Complete web interface for Pipeline management and monitoring
- 🔄 **Queue System**: Built-in queue system with concurrency control and task management
- 📝 **Complete Logging**: Detailed execution logs and error information
- 🔧 **Flexible Configuration**: Support YAML configuration files and command-line parameters
- 🔌 **Plugin System**: Support custom plugins to extend functionality
- 🌐 **Service-oriented**: Support remote execution via WebSocket and REST API

## 🎯 Core Concepts

### Pipeline

Pipeline is the top-level execution unit that contains multiple Stages.

```yaml
name: My Pipeline
stages:
  - name: stage1
    jobs: [...]
```

### Stage

Stage is an execution phase of a Pipeline that can contain multiple Jobs, supporting parallel or serial execution.

```yaml
stages:
  - name: build
    run_mode: parallel  # parallel or serial
    jobs: [...]
```

### Job

Job is a task unit in a Stage that contains multiple Steps.

```yaml
jobs:
  - name: build-job
    steps: [...]
```

### Step

Step is the smallest execution unit that executes specific commands or operations.

```yaml
steps:
  - name: compile
    command: make build
    image: golang:1.20
```

## 🚀 Quick Start

### Your First Pipeline

1. **Create a configuration file** `.pipeline.yaml`:

```yaml
name: My First Pipeline

stages:
  - name: build
    jobs:
      - name: build-job
        steps:
          - name: hello
            command: echo "Hello, Pipeline!"
```

2. **Run the Pipeline**:

```bash
pipeline run
```

## 📖 Usage

### 1. Local Execution

Execute Pipeline directly on your local machine:

```bash
pipeline run -c pipeline.yaml
```

### 2. Server Mode

Start a Pipeline service that provides Web Console and REST API:

```bash
# Start the server
pipeline server

# Access Web Console
open http://localhost:8080/console
```

### 3. Client Mode

Connect to a Pipeline Server and execute Pipeline:

```bash
pipeline client -c pipeline.yaml -s ws://localhost:8080
```

## 💡 Use Cases

- **CI/CD**: As a CI/CD pipeline execution engine
- **Automation Tasks**: Execute various automation tasks and scripts
- **Build System**: As a build and deployment system
- **Task Scheduling**: As a task scheduling and execution platform

---

**Start using Pipeline to make workflow execution simpler!** 🚀
