# server Command

The `pipeline server` command starts a lightweight **API-only** Pipeline service (REST API + WebSocket + execution queue). For the full web management console, use the [`web`](./web.md) command, which embeds the frontend and serves it together with the same backend.

## Basic Usage

```bash
pipeline server [options]
```

## Command Options

### `-p, --port`

Specify the server listening port.

- **Type**: Integer
- **Environment Variable**: `PORT`
- **Default**: `8080`

**Example**:

```bash
pipeline server -p 9090
```

### `--task-timeout`

Default execution timeout (seconds) for pipelines without an explicit `timeout`,
preventing tasks from hanging forever and occupying concurrency slots.

- **Type**: integer
- **Env vars**: `TASK_TIMEOUT`
- **Default**: `3600`
- **Note**: only applies to tasks whose YAML has no explicit `timeout`; an
  explicit timeout always wins. `0` disables the default. Server-only; the
  `run` mode is unaffected (its default stays 86400s).

**Example**:

```bash
# tasks without an explicit timeout run for at most 1 hour
pipeline server --task-timeout 3600

# disable the default timeout
pipeline server --task-timeout 0
```

### `--task-executor`

Task execution mode, controlling how isolated tasks are from the server process.

- **Type**: string (`in-process` | `subprocess`)
- **Env vars**: `TASK_EXECUTOR`
- **Default**: `in-process`
- **Note**:
  - `in-process` (default): tasks run in goroutines inside the server process
    (legacy behavior).
  - `subprocess`: each task runs in a separate `pipeline run` subprocess
    (reusing the running pipeline binary). Task crashes, panics (including
    native service SDKs) and resource usage stay inside the subprocess, so they
    cannot affect the server process. On cancellation the whole process group is
    killed (Linux/macOS) so step children are not orphaned.

**Example**:

```bash
# process-level isolation per task
pipeline server --task-executor subprocess
```

### `--max-concurrent`

Set the maximum number of concurrently executing Pipelines.

- **Type**: Integer
- **Environment Variable**: `MAX_CONCURRENT`
- **Default**: `2`
- **Description**: Controls the number of Pipelines executing simultaneously; Pipelines exceeding this number will enter the queue

**Example**:

```bash
pipeline server --max-concurrent 5
```

## Features

### REST API

The server provides the following REST API endpoints:

#### Pipeline Management

- `GET /api/v1/pipelines` - Get Pipeline list
  - Query parameters: `search`, `status`, `start_time`, `end_time`, `limit`, `offset`
- `GET /api/v1/pipelines/:id` - Get Pipeline details
- `GET /api/v1/pipelines/:id/logs` - Get Pipeline logs
  - Query parameters: `search`, `type`, `start_time`, `end_time`, `limit`, `offset`
- `GET /api/v1/pipelines/:id/logs/export` - Export Pipeline logs
  - Query parameters: `format` (text|json), `search`, `type`, `start_time`, `end_time`
- `POST /api/v1/pipelines/:id/cancel` - Cancel Pipeline execution
- `DELETE /api/v1/pipelines/:id` - Delete Pipeline record
- `POST /api/v1/pipelines/batch/delete` - Batch delete Pipelines
- `POST /api/v1/pipelines/batch/cancel` - Batch cancel Pipelines

#### Queue Management

- `GET /api/v1/queue/stats` - Get queue statistics
- `GET /api/v1/queue` - Get queue list
- `DELETE /api/v1/queue/:id` - Cancel task in queue

### WebSocket Execution

Execute Pipeline via WebSocket connection:

- **Connection Path**: `ws://localhost:8080/ws` (default, configurable with `--path`; or `wss://` if using HTTPS)
- **Authentication**: If username and password are set, provide Basic Auth when connecting
- **Message Format**: JSON-formatted Action messages

## Usage Examples

### Example 1: Basic Startup

```bash
# Start server (default port 8080)
pipeline server

# Legacy /console now redirects to the web console (pipeline web)
```

### Q: Are tasks isolated from the server process?

- By default the server runs tasks **in-process** (goroutines). A panicking task
  is recovered and marked failed instead of crashing the server; however task
  resource usage still shares the server process.
- For stronger isolation use `--task-executor subprocess`: tasks run in separate
  subprocesses, so crashes and resource usage cannot affect the server process.
  Full sandboxing (containers) can be layered on top with containerized
  deployment.

## Deployment Recommendations

### Production Environment

1. **Use Reverse Proxy**: Use Nginx or Traefik as reverse proxy to provide HTTPS
2. **Enable Authentication**: Set username and password to protect the service
3. **Set Working Directory**: Use persistent storage as working directory
4. **Configure Concurrency**: Set reasonable concurrency based on server resources
5. **Monitoring and Logging**: Configure log collection and monitoring systems
