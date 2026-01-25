# System Architecture

Pipeline is a lightweight CI/CD pipeline execution engine written in Go, similar to Jenkins and GitLab CI, but more lightweight and flexible. It supports multiple execution engines, plugin systems, service orchestration, and more.

**Version**: 1.6.0  
**Language**: Go 1.22.1

## Hierarchy Structure

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

## Execution Engines

Pipeline supports multiple execution engines:

### Host Engine (Default)

Execute commands directly on the local host.

```yaml
steps:
  - name: example
    command: echo "hello"
    # engine: host (default)
```

### Docker Engine

Execute commands in Docker containers. If `image` is specified, Docker engine is automatically used.

```yaml
steps:
  - name: example
    image: alpine:latest
    command: echo "hello"
    # engine: docker (automatic)
```

### SSH Engine

Execute commands on remote servers via SSH.

```yaml
steps:
  - name: example
    engine: ssh://username:password@host:port
    command: echo "hello"
```

### IDP Engine

Execute commands by connecting to remote IDP Agent via WebSocket.

```yaml
steps:
  - name: example
    engine: idp://client_id:client_secret@host:port
    command: echo "hello"
```

## Plugin System

Pipeline supports a plugin mechanism that encapsulates reusable functionality through Docker images.

### Plugin Configuration

```yaml
steps:
  - name: checkout
    plugin:
      image: ghcr.io/go-idp/pipeline-checkout:latest
      settings:
        repository: https://github.com/user/repo.git
        branch: main
        token: ${GITHUB_TOKEN}
```

## Service Orchestration

Pipeline supports service orchestration, allowing you to start dependent services (such as databases, caches, etc.) for Pipeline use.

### Docker Compose Service

```yaml
steps:
  - name: test-with-db
    service:
      type: docker-compose
      name: test-db
      version: v1
      config: |
        version: '3'
        services:
          postgres:
            image: postgres:13
            environment:
              POSTGRES_PASSWORD: test
            ports:
              - "5432:5432"
    command: |
      # Test code can connect to postgres:5432
      pytest tests/
```

## Configuration Inheritance

Configuration is inherited in the following hierarchy: **Pipeline → Stage → Job → Step**

Each level inherits configuration from its parent, with child-level configuration taking higher priority.
