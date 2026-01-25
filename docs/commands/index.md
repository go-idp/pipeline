# Commands Overview

Pipeline provides three main commands to meet different usage scenarios.

## Command List

### run

Run Pipeline directly on your local machine. This is the most commonly used command.

```bash
pipeline run [options]
```

**Use Cases**:
- Local development and testing
- CI/CD pipeline execution
- Automation script execution

**Documentation**: [run command](./run.md)

### server

Start a Pipeline service that provides Web Console and REST API.

```bash
pipeline server [options]
```

**Use Cases**:
- Production deployment
- Web interface management
- REST API integration
- Queue management

**Documentation**: [server command](./server.md)

### client

Connect to a Pipeline Server and execute Pipeline via WebSocket.

```bash
pipeline client [options]
```

**Use Cases**:
- Remote Pipeline execution
- CI/CD integration
- Distributed execution

**Documentation**: [client command](./client.md)

## Command Selection Guide

### Local Development

Use the `run` command:

```bash
pipeline run -c pipeline.yaml
```

### Service Deployment

Use the `server` command:

```bash
pipeline server -p 8080
```

### Remote Execution

Use the `client` command:

```bash
pipeline client -c pipeline.yaml -s ws://server:8080
```

## Command Combinations

### Server + Client Mode

1. Start Server:

```bash
pipeline server -p 8080
```

2. Connect with Client:

```bash
pipeline client -c pipeline.yaml -s ws://localhost:8080
```

This mode is suitable for:
- Centralized Pipeline management
- Multi-client execution
- Queue and concurrency control

## Environment Variables

All commands support setting options via environment variables:

```bash
export PIPELINE_CONFIG=pipeline.yaml
export PIPELINE_WORKDIR=/tmp/pipeline
pipeline run
```

## Configuration Files

All commands support configuration files, searched in the following order:

1. File specified by `-c` command-line argument
2. `.pipeline.yaml` (current directory)
3. `.go-idp/pipeline.yaml` (current directory)

## Debug Mode

All commands support debug mode:

```bash
DEBUG=1 pipeline run -c pipeline.yaml
DEBUG=1 pipeline server
DEBUG=1 pipeline client -c pipeline.yaml -s ws://localhost:8080
```

Debug mode will output detailed execution information.
