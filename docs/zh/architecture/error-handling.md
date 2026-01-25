# 错误处理

Pipeline 在执行过程中可能会遇到各种错误情况。本文档详细说明了 Pipeline 的错误处理机制、workdir 清理策略以及错误信息的输出格式。

## 错误处理机制

### 错误状态管理

当 Pipeline 执行失败时，会设置以下状态：

- **Status**: 设置为 `"failed"`
- **Error**: 包含详细的错误信息
- **FailedAt**: 记录失败时间戳

### 错误类型

Pipeline 可能遇到的错误类型包括：

#### Stage 执行错误

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

#### 超时错误

当 Pipeline 执行时间超过设定的 `timeout` 时，会触发超时错误。

```yaml
name: My Pipeline
timeout: 60  # 60 秒超时
```

#### Context 取消错误

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

使用 Context 可以更好地控制 Pipeline 的执行：

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

if err := pipeline.Run(ctx); err != nil {
    log.Printf("Pipeline failed: %v", err)
}
```

### 4. 在 Post 钩子中添加清理逻辑

在 Post 钩子中添加清理逻辑，确保资源得到正确清理：

```yaml
post: |
  echo "Cleaning up..."
  rm -rf /tmp/build-artifacts
```

## 常见错误

### 命令执行失败

**原因**: 命令返回非零退出码

**解决**: 检查命令是否正确，查看 workdir 中的日志文件

### 超时错误

**原因**: Pipeline 执行时间超过设定的超时时间

**解决**: 增加超时时间或优化 Pipeline 执行效率

### 工作目录权限问题

**原因**: 无法创建工作目录或写入文件

**解决**: 检查目录权限，确保有足够的权限

### 环境变量未传递

**原因**: 环境变量未正确传递到 Pipeline

**解决**: 检查环境变量配置，使用 `--allow-env` 选项

## 调试技巧

### 启用调试模式

```bash
DEBUG=1 pipeline run -c pipeline.yaml
```

调试模式下会输出更详细的日志信息。

### 查看 workdir

失败后查看 workdir 中的文件：

```bash
ls -la /tmp/pipeline/abc123
cat /tmp/pipeline/abc123/*.log
```

### 检查环境变量

在 Pipeline 中添加步骤检查环境变量：

```yaml
steps:
  - name: check-env
    command: env | grep PIPELINE
```
