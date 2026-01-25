# Pipeline 错误处理文档

## 概述

Pipeline 在执行过程中可能会遇到各种错误情况。本文档详细说明了 Pipeline 的错误处理机制、workdir 清理策略以及错误信息的输出格式。

## 错误处理机制

### 1. 错误状态管理

当 Pipeline 执行失败时，会设置以下状态：

- **Status**: 设置为 `"failed"`
- **Error**: 包含详细的错误信息
- **FailedAt**: 记录失败时间戳

### 2. 错误类型

Pipeline 可能遇到的错误类型包括：

#### 2.1 Stage 执行错误
当某个 Stage 执行失败时，Pipeline 会立即终止并返回错误。

```yaml
stages:
  - name: build
    jobs:
      - name: build-job
        steps:
          - name: compile
            command: "make build"  # 如果失败，Pipeline 会终止
```

#### 2.2 超时错误
当 Pipeline 执行时间超过设定的 `timeout` 时，会触发超时错误。

```yaml
name: My Pipeline
timeout: 60  # 60 秒超时
```

#### 2.3 Context 取消错误
当外部 Context 被取消时，Pipeline 会立即终止。

```go
ctx, cancel := context.WithCancel(context.Background())
cancel()  // Pipeline 会检测到取消并终止
pipeline.Run(ctx)
```

## Workdir 清理策略

### 成功时清理

当 Pipeline 成功执行完成时，**workdir 会被自动清理**：

```bash
[workflow] done
[workflow] done to run (name: My Pipeline, workdir: /tmp/pipeline/abc123)
```

**注意**：
- 如果 `workdir` 是当前目录，则不会被清理（安全考虑）
- 如果清理失败，会记录警告日志，但不会影响 Pipeline 的成功状态

### 失败时保留

当 Pipeline 执行失败时，**workdir 会被保留**以便调试：

```bash
[workflow] error: stage "build" failed: job "build-job" failed: step "compile" failed: exit status 1
[workflow] workdir: /tmp/pipeline/abc123
[workflow] logs: check workdir for detailed logs and output files
[workflow] workdir preserved for debugging (not cleaned)
```

**保留 workdir 的好处**：
1. **调试方便**：可以查看失败时的文件状态
2. **日志保留**：可以查看详细的执行日志
3. **问题排查**：可以检查中间文件来定位问题

## 错误日志输出

### 成功时的日志

```
[workflow] start to run (name: My Pipeline)
[workflow] start
[workflow] version: 1.0.0
[workflow] name: My Pipeline
[workflow] workdir: /tmp/pipeline/abc123
[workflow] timeout: 3600 seconds
[workflow] done
[workflow] done to run (name: My Pipeline, workdir: /tmp/pipeline/abc123)
```

### 失败时的日志

```
[workflow] start to run (name: My Pipeline)
[workflow] start
[workflow] version: 1.0.0
[workflow] name: My Pipeline
[workflow] workdir: /tmp/pipeline/abc123
[workflow] timeout: 3600 seconds
[workflow] error: stage "build" failed: job "build-job" failed: step "compile" failed: exit status 1
[workflow] workdir: /tmp/pipeline/abc123
[workflow] logs: check workdir for detailed logs and output files
[workflow] workdir preserved for debugging (not cleaned)
[workflow] error: stage "build" failed: job "build-job" failed: step "compile" failed: exit status 1
[workflow] workdir: /tmp/pipeline/abc123
[workflow] logs: check workdir for detailed logs and output files
[workflow] workdir preserved for debugging (not cleaned)
```

### 超时错误的日志

```
[workflow] error: pipeline timeout after 60 seconds: context deadline exceeded
[workflow] workdir: /tmp/pipeline/abc123
[workflow] logs: check workdir for detailed logs and output files
[workflow] workdir preserved for debugging (not cleaned)
```

## 最佳实践

### 1. 设置合理的超时时间

```yaml
name: My Pipeline
timeout: 3600  # 根据实际需要设置，避免过长或过短
```

### 2. 检查失败后的 workdir

当 Pipeline 失败时，检查 workdir 中的文件：

```bash
# 查看 workdir 内容
ls -la /tmp/pipeline/abc123

# 查看日志文件
cat /tmp/pipeline/abc123/*.log

# 查看输出文件
cat /tmp/pipeline/abc123/output/*
```

### 3. 使用 Context 控制执行

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := pipeline.Run(ctx)
if err != nil {
    // 检查 workdir 进行调试
    log.Printf("Pipeline failed, workdir: %s", pipeline.Workdir)
}
```

### 4. 定期清理失败的 workdir

虽然失败的 workdir 会被保留，但建议定期清理以避免磁盘空间问题：

```bash
# 查找并清理旧的 workdir
find /tmp/pipeline -type d -mtime +7 -exec rm -rf {} \;
```

## 错误处理示例

### 示例 1: 处理 Stage 失败

```go
pipeline := &Pipeline{
    Name: "Build Pipeline",
    Stages: []*stage.Stage{
        {
            Name: "build",
            Jobs: []*job.Job{
                {
                    Name: "compile",
                    Steps: []*step.Step{
                        {
                            Name:    "build",
                            Command: "make build",
                        },
                    },
                },
            },
        },
    },
}

err := pipeline.Run(context.Background())
if err != nil {
    // Pipeline 失败，workdir 已保留
    log.Printf("Pipeline failed: %v", err)
    log.Printf("Workdir preserved at: %s", pipeline.Workdir)
    
    // 可以检查 workdir 中的文件
    // 例如：查看编译错误日志
}
```

### 示例 2: 处理超时

```go
pipeline := &Pipeline{
    Name:    "Long Running Pipeline",
    Timeout: 60, // 60 秒超时
    Stages: []*stage.Stage{
        // ...
    },
}

err := pipeline.Run(context.Background())
if err != nil {
    if strings.Contains(err.Error(), "timeout") {
        log.Printf("Pipeline timed out after 60 seconds")
        log.Printf("Check workdir for partial results: %s", pipeline.Workdir)
    }
}
```

## 常见问题

### Q: 为什么失败的 workdir 不自动清理？

A: 失败的 workdir 被保留是为了方便调试和问题排查。您可以手动清理，或者设置定期清理任务。

### Q: 如何禁用 workdir 清理？

A: 将 `workdir` 设置为当前目录，Pipeline 不会清理当前目录。

### Q: workdir 清理失败会影响 Pipeline 状态吗？

A: 不会。如果清理失败，只会记录警告日志，不会影响 Pipeline 的成功状态。

### Q: 如何查看详细的错误信息？

A: 检查 `pipeline.State.Error` 字段，或者查看 workdir 中的日志文件。

## 相关文档

- [使用文档](./USAGE.md)
- [架构文档](./ARCHITECTURE.md)
- [优化文档](./OPTIMIZATION.md)
