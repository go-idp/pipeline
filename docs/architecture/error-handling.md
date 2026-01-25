# Error Handling

Pipeline may encounter various error situations during execution. This document details Pipeline's error handling mechanisms, workdir cleanup strategies, and error message output formats.

## Error Handling Mechanism

### Error Status Management

When Pipeline execution fails, the following status is set:

- **Status**: Set to `"failed"`
- **Error**: Contains detailed error information
- **FailedAt**: Records failure timestamp

### Error Types

Pipeline may encounter the following error types:

#### Stage Execution Error

When a Stage execution fails, Pipeline immediately terminates and returns an error.

```yaml
stages:
  - name: build
    jobs:
      - name: build-job
        steps:
          - name: compile
            command: "make build"  # If fails, Pipeline terminates
```

#### Timeout Error

When Pipeline execution time exceeds the set `timeout`, a timeout error is triggered.

```yaml
name: My Pipeline
timeout: 60  # 60 seconds timeout
```

## Workdir Cleanup Strategy

### Cleanup on Success

When Pipeline executes successfully, **workdir is automatically cleaned**:

```bash
[workflow] done
[workflow] done to run (name: My Pipeline, workdir: /tmp/pipeline/abc123)
```

**Note**:
- If `workdir` is the current directory, it will not be cleaned (safety consideration)
- If cleanup fails, a warning log is recorded, but it does not affect Pipeline's success status

### Preserved on Failure

When Pipeline execution fails, **workdir is preserved** for debugging:

```bash
[workflow] error: stage "build" failed: job "build-job" failed: step "compile" failed: exit status 1
[workflow] workdir: /tmp/pipeline/abc123
[workflow] logs: check workdir for detailed logs and output files
[workflow] workdir preserved for debugging (not cleaned)
```

**Benefits of preserving workdir**:
1. **Easy Debugging**: Can view file state at failure
2. **Log Preservation**: Can view detailed execution logs
3. **Problem Troubleshooting**: Can check intermediate files to locate issues

## Best Practices

### 1. Set Reasonable Timeout

```yaml
name: My Pipeline
timeout: 3600  # Set based on actual needs, avoid too long or too short
```

### 2. Check Workdir After Failure

When Pipeline fails, check files in workdir:

```bash
# View workdir contents
ls -la /tmp/pipeline/abc123

# View log files
cat /tmp/pipeline/abc123/*.log

# View output files
cat /tmp/pipeline/abc123/output/*
```

### 3. Use Context to Control Execution

Using Context allows better control over Pipeline execution:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

if err := pipeline.Run(ctx); err != nil {
    log.Printf("Pipeline failed: %v", err)
}
```
