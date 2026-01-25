# client Command

The `pipeline client` command is used to connect to a Pipeline Server and execute Pipeline via WebSocket. The client sends Pipeline configuration to the server and receives execution logs and results in real-time.

## Basic Usage

```bash
pipeline client [options]
```

## Command Options

### `-c, --config`

Specify the Pipeline configuration file path (required).

- **Type**: String
- **Environment Variable**: `CONFIG`
- **Required**: Yes
- **Supported Formats**:
  - Local file path: `pipeline.yaml`
  - HTTP/HTTPS URL: `https://example.com/pipeline.yaml`

**Example**:

```bash
pipeline client -c pipeline.yaml -s ws://localhost:8080
```

### `-s, --server`

Specify the Pipeline Server address (required).

- **Type**: String
- **Environment Variable**: `SERVER`
- **Required**: Yes
- **Format**: `ws://host:port` or `wss://host:port`

**Example**:

```bash
pipeline client -c pipeline.yaml -s ws://localhost:8080
pipeline client -c pipeline.yaml -s wss://pipeline.example.com
```

## Workflow

1. **Load Configuration**: Load Pipeline configuration from local file or remote URL
2. **Connect to Server**: Connect to Pipeline Server via WebSocket
3. **Send Configuration**: Send Pipeline configuration to server
4. **Receive Logs**: Receive Pipeline execution logs (stdout/stderr) in real-time
5. **Wait for Completion**: Wait for Pipeline execution to complete or fail
6. **Close Connection**: Close WebSocket connection

## Usage Examples

### Example 1: Basic Usage

```bash
# Start server (in another terminal)
pipeline server

# Execute Pipeline with client
pipeline client \
  -c pipeline.yaml \
  -s ws://localhost:8080
```

## Output Format

The client will output Pipeline execution logs in real-time:

- **Standard Output**: Pipeline stdout output
- **Standard Error**: Pipeline stderr output
- **Execution Result**: Pipeline execution completion or failure information

**Example Output**:

```
[workflow] start
[workflow] version: 1.7.1
[workflow] name: My Pipeline
[stage(1/2): build] start
[job(1/1): build-job] start
[step(1/1): compile] start
Compiling...
[step(1/1): compile] done
[job(1/1): build-job] done
[stage(1/2): build] done
[workflow] done
```

## Error Handling

### Connection Error

If unable to connect to server, client will output error:

```
failed to connect to server: connection refused
```

### Authentication Error

If authentication fails, client will output error:

```
authentication failed
```

### Pipeline Execution Error

If Pipeline execution fails, client will output error and exit with non-zero exit code.
