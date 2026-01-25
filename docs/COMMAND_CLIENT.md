# Pipeline Client 命令文档

## 概述

`pipeline client` 命令用于连接到 Pipeline Server 并通过 WebSocket 执行 Pipeline。客户端会将 Pipeline 配置发送到服务器，并实时接收执行日志和结果。

## 基本用法

```bash
pipeline client [选项]
```

## 命令选项

### `-c, --config`

指定 Pipeline 配置文件路径（必需）。

- **类型**: 字符串
- **环境变量**: `CONFIG`
- **必需**: 是
- **支持格式**:
  - 本地文件路径: `pipeline.yaml`
  - HTTP/HTTPS URL: `https://example.com/pipeline.yaml`

**示例**:

```bash
pipeline client -c pipeline.yaml -s ws://localhost:8080
```

### `-s, --server`

指定 Pipeline Server 地址（必需）。

- **类型**: 字符串
- **环境变量**: `SERVER`
- **必需**: 是
- **格式**: `ws://host:port` 或 `wss://host:port`

**示例**:

```bash
pipeline client -c pipeline.yaml -s ws://localhost:8080
pipeline client -c pipeline.yaml -s wss://pipeline.example.com
```

### `-u, --username`

设置 HTTP Basic Auth 用户名。

- **类型**: 字符串
- **环境变量**: `USERNAME`
- **说明**: 如果服务器启用了认证，需要提供用户名

**示例**:

```bash
pipeline client -c pipeline.yaml -s ws://localhost:8080 -u admin
```

### `-p, --password`

设置 HTTP Basic Auth 密码。

- **类型**: 字符串
- **环境变量**: `PASSWORD`
- **说明**: 如果服务器启用了认证，需要提供密码

**示例**:

```bash
pipeline client -c pipeline.yaml -s ws://localhost:8080 -u admin -p password123
```

### `--path`

指定服务器路径。

- **类型**: 字符串
- **环境变量**: `SERVER_PATH`
- **默认值**: `/`
- **说明**: 如果服务器使用了自定义路径，需要指定此选项

**示例**:

```bash
pipeline client -c pipeline.yaml -s ws://localhost:8080 --path /api/pipeline
```

## 工作流程

1. **加载配置**: 从本地文件或远程 URL 加载 Pipeline 配置
2. **连接服务器**: 通过 WebSocket 连接到 Pipeline Server
3. **发送配置**: 将 Pipeline 配置发送到服务器
4. **接收日志**: 实时接收 Pipeline 执行日志（stdout/stderr）
5. **等待完成**: 等待 Pipeline 执行完成或失败
6. **关闭连接**: 关闭 WebSocket 连接

## 使用示例

### 示例 1: 基本使用

```bash
# 启动服务器（在另一个终端）
pipeline server

# 使用客户端执行 Pipeline
pipeline client \
  -c pipeline.yaml \
  -s ws://localhost:8080
```

### 示例 2: 使用认证

```bash
pipeline client \
  -c pipeline.yaml \
  -s ws://localhost:8080 \
  -u admin \
  -p password123
```

### 示例 3: 使用远程配置

```bash
pipeline client \
  -c https://example.com/pipeline.yaml \
  -s ws://pipeline.example.com
```

### 示例 4: 使用自定义路径

```bash
pipeline client \
  -c pipeline.yaml \
  -s ws://localhost:8080 \
  --path /api/pipeline
```

### 示例 5: 使用环境变量

```bash
export CONFIG=pipeline.yaml
export SERVER=ws://localhost:8080
export USERNAME=admin
export PASSWORD=password123
pipeline client
```

### 示例 6: 在 URL 中包含认证信息

```bash
# 用户名和密码可以包含在 URL 中
pipeline client \
  -c pipeline.yaml \
  -s ws://admin:password123@localhost:8080
```

## 输出格式

客户端会实时输出 Pipeline 的执行日志：

- **标准输出**: Pipeline 的 stdout 输出
- **标准错误**: Pipeline 的 stderr 输出
- **执行结果**: Pipeline 执行完成或失败的信息

**示例输出**:

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

## 错误处理

### 连接错误

如果无法连接到服务器，客户端会输出错误信息：

```
failed to connect to server: connection refused
```

### 认证错误

如果认证失败，客户端会输出错误信息：

```
authentication failed
```

### Pipeline 执行错误

如果 Pipeline 执行失败，客户端会输出错误信息并返回非零退出码：

```
[workflow] error: stage "build" failed: job "build-job" failed: step "compile" failed: exit status 1
```

## 退出码

- `0`: Pipeline 执行成功
- `非零`: Pipeline 执行失败或连接错误

## 与 Server 的交互

### WebSocket 消息格式

客户端发送的消息格式：

```json
{
  "type": "run",
  "payload": "<YAML 格式的 Pipeline 配置>"
}
```

服务器返回的消息类型：

- `stdout`: 标准输出日志
- `stderr`: 标准错误日志
- `done`: 执行完成
- `error`: 执行错误

### 连接流程

1. 客户端建立 WebSocket 连接
2. 客户端发送 `run` 消息（包含 Pipeline 配置）
3. 服务器将 Pipeline 加入队列
4. 服务器返回 `done` 消息（表示已加入队列）
5. 服务器执行 Pipeline 并发送日志
6. 服务器发送最终结果（`done` 或 `error`）

## 最佳实践

### 1. 使用环境变量管理敏感信息

```bash
export SERVER=ws://pipeline.example.com
export USERNAME=admin
export PASSWORD=$(cat ~/.pipeline/password)
pipeline client -c pipeline.yaml
```

### 2. 在 CI/CD 中使用

```yaml
# GitHub Actions 示例
- name: Run Pipeline
  run: |
    pipeline client \
      -c pipeline.yaml \
      -s ws://${{ secrets.PIPELINE_SERVER }} \
      -u ${{ secrets.PIPELINE_USERNAME }} \
      -p ${{ secrets.PIPELINE_PASSWORD }}
```

### 3. 检查执行结果

```bash
if pipeline client -c pipeline.yaml -s ws://localhost:8080; then
  echo "Pipeline executed successfully"
else
  echo "Pipeline failed"
  exit 1
fi
```

## 常见问题

### Q: 如何知道 Pipeline 是否执行成功？

A: 检查命令的退出码。退出码为 0 表示成功，非零表示失败。

### Q: 客户端会等待 Pipeline 执行完成吗？

A: 是的。客户端会等待 Pipeline 执行完成或失败后才退出。

### Q: 如何查看详细的执行日志？

A: 客户端会实时输出所有日志。如果需要保存日志，可以重定向输出：

```bash
pipeline client -c pipeline.yaml -s ws://localhost:8080 > pipeline.log 2>&1
```

### Q: 支持多个 Pipeline 同时执行吗？

A: 支持。可以在多个终端同时运行客户端，服务器会管理执行队列。

### Q: 如何取消正在执行的 Pipeline？

A: 在客户端端按 Ctrl+C 可以取消连接，但 Pipeline 可能仍在服务器端执行。需要通过服务器的 Web Console 或 API 取消。

### Q: 客户端和直接运行有什么区别？

A: 
- **直接运行** (`pipeline run`): 在本地执行 Pipeline
- **客户端运行** (`pipeline client`): 将 Pipeline 发送到服务器执行，适合远程执行和集中管理

## 相关文档

- [Server 命令文档](./COMMAND_SERVER.md)
- [使用文档](./USAGE.md)
- [架构文档](./ARCHITECTURE.md)
