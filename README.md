# Pipeline

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

Pipeline is a powerful workflow execution engine that supports local execution and service deployment. It provides flexible configuration, rich execution engines, and a complete Web Console and REST API.

## ✨ Features

- 🚀 **Multiple Execution Modes**: Support local execution, Server mode, and Client mode
- 🐳 **Multiple Execution Engines**: Support host, docker, ssh, idp and other execution engines
- 📦 **Native Service Deployment**: docker-compose / docker-swarm / kubernetes from raw definitions, executed with Go SDKs
- 📊 **Web Console**: Complete web interface for Pipeline management and monitoring
- 🔄 **Queue System**: Built-in queue system with concurrency control and task management
- 📝 **Complete Logging**: Detailed execution logs and error information
- 🔧 **Flexible Configuration**: Support YAML configuration files and command-line parameters
- 🔌 **Plugin System**: Support custom plugins to extend functionality
- 🌐 **Service-oriented**: Support remote execution via WebSocket and REST API

## 🚀 Quick Start

### Installation

#### Build from Source

```bash
git clone https://github.com/go-idp/pipeline.git
cd pipeline
go build -o pipeline cmd/pipeline/main.go
```

#### Install with Go

```bash
go install github.com/go-idp/pipeline/cmd/pipeline@latest
```

#### Use Docker

```bash
docker pull ghcr.io/go-idp/pipeline:latest
```

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

**Documentation**: [Run Command Documentation](https://go-idp.github.io/pipeline/commands/run)

### 2. Web Console (Web Mode)

Start the full management console: the embedded frontend (React + TypeScript)
is served together with the server backend. Build the frontend first:

```bash
cd web/ui && pnpm install && pnpm build   # build the frontend once
go build -o pipeline ./cmd/pipeline
```

```bash
# Start the console
pipeline web

# Access Web Console
open http://localhost:8080
```

**Documentation**: [Web Command Documentation](https://go-idp.github.io/pipeline/commands/web)

### 3. Server Mode (API Only)

Start a lightweight API-only service (REST API + WebSocket + queue), without
the embedded frontend:

```bash
pipeline server
```

**Documentation**: [Server Command Documentation](https://go-idp.github.io/pipeline/commands/server)

### 4. Client Mode

Connect to a Pipeline Server and execute Pipeline (WebSocket default path is
`/ws`):

```bash
pipeline client -c pipeline.yaml -s ws://localhost:8080/ws
```

**Documentation**: [Client Command Documentation](https://go-idp.github.io/pipeline/commands/client)

## 📚 Documentation

- **[Documentation Site](https://go-idp.github.io/pipeline/)** - Complete documentation with guides, API reference, and examples
- **[Getting Started](https://go-idp.github.io/pipeline/guide/)** - Installation and quick start guide
- **[Commands](https://go-idp.github.io/pipeline/commands/)** - Command reference documentation
- **[Architecture](https://go-idp.github.io/pipeline/architecture/)** - System architecture and design
- **[Best Practices](https://go-idp.github.io/pipeline/best-practices/)** - Usage recommendations

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

## 🔧 Configuration Examples

### Basic Configuration

```yaml
name: Build Application

stages:
  - name: checkout
    jobs:
      - name: checkout
        steps:
          - name: git-clone
            command: git clone https://github.com/user/repo.git .

  - name: build
    jobs:
      - name: build
        steps:
          - name: build-app
            image: golang:1.20
            command: go build -o app ./cmd/app
```

### Using Docker

```yaml
name: Docker Build

stages:
  - name: build
    jobs:
      - name: build-image
        steps:
          - name: build
            image: docker:latest
            command: docker build -t myapp:latest .
```

### Using Plugins

```yaml
name: Plugin Example

stages:
  - name: deploy
    jobs:
      - name: deploy
        steps:
          - name: deploy-step
            plugin:
              image: my-plugin:latest
              settings:
                token: ${GITHUB_TOKEN}
```

More examples can be found in the [examples](./examples/) directory (including `step-service-docker-compose.yaml`, `step-service-docker-swarm.yaml`, `step-service-kubernetes.yaml`, and `service-deploy.yml`).

## 🌟 Key Features

### Web Console (`pipeline web`)

The `pipeline web` command provides a complete management console
(Linear-style monochrome UI, `⌘K` command palette, zh/en i18n):

- 📊 **Overview**: run trends, success rate, status distribution
- 🧬 **Pipelines**: YAML-first definition CRUD with CodeMirror editor + visual preview
- 📈 **Runs**: run list/detail with stage timeline, stage/job/step tree and
  real-time logs; cancel / rerun / export
- 📊 **Queue Monitoring**: real-time queue status and concurrency
- ⚙️ **System Settings**: concurrency, timeout, executor, environment allowlist

The `pipeline server` command remains a lightweight API-only service.

On startup the server injects **built-in templates** (CI / release / deploy /
docs / nightly) and a few **demo runs** (idempotent, only when storage is
empty), so the console is immediately usable — see
[`web` command documentation](https://go-idp.github.io/pipeline/commands/web).

### Queue System

- **Concurrency Control**: Configurable maximum concurrent execution
- **Auto Execution**: Queue automatically detects and executes pending Pipelines
- **Status Management**: Complete Pipeline status tracking (pending, running, succeeded, failed)
- **Task Cancellation**: Support canceling tasks in the queue

### Error Handling

- **Workdir Preservation**: Preserve workdir on failure for debugging
- **Detailed Logs**: Output detailed error information and debugging hints
- **Status Tracking**: Complete execution status and error information recording

## 🛠️ Development

### Run Tests

```bash
go test ./...
```

### Build

```bash
go build -o pipeline cmd/pipeline/main.go
```

### Run Examples

```bash
# Run basic example
pipeline run -c examples/basic.yml

# Run Docker example (requires Docker)
pipeline run -c examples/docker.yaml
```

## 📦 Project Structure

```
pipeline/
├── cmd/pipeline/          # CLI entry point
│   └── commands/          # Command implementations
│       ├── run.go         # run command
│       ├── server.go      # server command (API only)
│       ├── web.go         # web command (embedded frontend + server)
│       └── client.go      # client command
├── event/                 # Run state events (stage/job/step observer)
├── svc/                   # Service layer
│   ├── server/            # Server implementation
│   │   ├── server.go      # Server main logic
│   │   ├── queue.go       # Queue system
│   │   ├── store.go       # Storage system
│   │   ├── api_web.go     # Management API extensions
│   │   └── web.go         # SPA static serving
│   └── client/            # Client implementation
├── web/                   # Web console
│   ├── embed.go           # go:embed of web/ui/dist
│   └── ui/                # Frontend (pnpm + Vite + React + TypeScript)
├── examples/              # Example configurations
├── docs/                  # Documentation (VitePress)
└── *.go                   # Core code
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## 🔗 Links

- **GitHub**: https://github.com/go-idp/pipeline
- **Documentation**: https://go-idp.github.io/pipeline/
- **Examples**: [examples/](./examples/)

## 💡 Use Cases

- **CI/CD**: As a CI/CD pipeline execution engine
- **Automation Tasks**: Execute various automation tasks and scripts
- **Build System**: As a build and deployment system
- **Task Scheduling**: As a task scheduling and execution platform

---

**Start using Pipeline to make workflow execution simpler!** 🚀
