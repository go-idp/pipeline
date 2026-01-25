# server Command

The `pipeline server` command starts a Pipeline service that provides Web Console and REST API, supports executing Pipeline via WebSocket, and manages Pipeline execution queues.

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

### Web Console

Access `http://localhost:8080/console` to open the Web Console, providing:

- **Pipeline Management**: Create, view, delete Pipelines
- **Search and Filter**: Search Pipelines by name or ID, filter by status and time range
- **Queue Management**: View queue status, cancel tasks
- **Pipeline Cancellation**: Support canceling executing or pending Pipelines
- **History**: View Pipeline execution history
- **Enhanced Logs**: Search, filter, and export Pipeline execution logs
- **Real-time Logs**: View Pipeline execution logs with real-time streaming support
- **Pipeline Definition View**: View complete Pipeline YAML configuration with one-click copy
- **Dark Mode**: Toggle between light and dark themes
- **System Settings**: Configure queue concurrency and other settings

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

- **Connection Path**: `ws://localhost:8080/` (or `wss://` if using HTTPS)
- **Authentication**: If username and password are set, provide Basic Auth when connecting
- **Message Format**: JSON-formatted Action messages

## Usage Examples

### Example 1: Basic Startup

```bash
# Start server (default port 8080)
pipeline server

# Access Web Console
open http://localhost:8080/console
```

## Deployment Recommendations

### Production Environment

1. **Use Reverse Proxy**: Use Nginx or Traefik as reverse proxy to provide HTTPS
2. **Enable Authentication**: Set username and password to protect the service
3. **Set Working Directory**: Use persistent storage as working directory
4. **Configure Concurrency**: Set reasonable concurrency based on server resources
5. **Monitoring and Logging**: Configure log collection and monitoring systems
