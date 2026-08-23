# server 命令

`pipeline server` 命令启动一个 Pipeline 服务，提供 Web Console 和 REST API，支持通过 WebSocket 执行 Pipeline，并管理 Pipeline 执行队列。

## 基本用法

```bash
pipeline server [选项]
```

## 命令选项

### `-p, --port`

指定服务器监听端口。

- **类型**: 整数
- **环境变量**: `PORT`
- **默认值**: `8080`

**示例**:

```bash
pipeline server -p 9090
```

### `--path`

指定服务器路径（WebSocket 和 API 的基础路径）。

- **类型**: 字符串
- **环境变量**: `ENDPOINT`, `SERVER_PATH`
- **默认值**: `/`

**示例**:

```bash
pipeline server --path /api/pipeline
```

### `-w, --workdir`

指定 Pipeline 执行的工作目录。

- **类型**: 字符串
- **环境变量**: `WORKDIR`
- **默认值**: `/tmp/go-idp/pipeline`

**示例**:

```bash
pipeline server -w /var/lib/pipeline
```

### `-u, --username`

设置 HTTP Basic Auth 用户名。

- **类型**: 字符串
- **环境变量**: `USERNAME`
- **说明**: 如果设置了用户名或密码，将启用 HTTP Basic Auth

**示例**:

```bash
pipeline server -u admin -p password123
```

### `--password`

设置 HTTP Basic Auth 密码。

- **类型**: 字符串
- **环境变量**: `PASSWORD`
- **说明**: 如果设置了用户名或密码，将启用 HTTP Basic Auth

**示例**:

```bash
pipeline server -u admin --password password123
```

### `--allow-env`

允许传递指定的环境变量到 Pipeline（可多次使用）。

- **类型**: 字符串切片
- **环境变量**: `ALLOW_ENV`
- **说明**: 从服务器环境中选择性地传递环境变量到 Pipeline

**示例**:

```bash
pipeline server --allow-env GITHUB_TOKEN --allow-env CI_BUILD_NUMBER
```

### `--allow-all-env`

允许传递所有环境变量到 Pipeline。

- **类型**: 布尔值
- **环境变量**: `ALLOW_ALL_ENV`
- **说明**: 将服务器的所有环境变量传递给 Pipeline

**示例**:

```bash
pipeline server --allow-all-env
```

### `--max-concurrent`

设置最大并发执行的 Pipeline 数量。

- **类型**: 整数
- **环境变量**: `MAX_CONCURRENT`
- **默认值**: `2`
- **说明**: 控制同时执行的 Pipeline 数量，超过此数量的 Pipeline 将进入队列等待

**示例**:

```bash
pipeline server --max-concurrent 5
```

### `--task-timeout`

未显式设置 `timeout` 的 Pipeline 的默认执行超时（秒），防止任务无限挂起占用并发槽。

- **类型**: 整数
- **环境变量**: `TASK_TIMEOUT`
- **默认值**: `3600`
- **说明**: 仅对 YAML 中未显式设置 `timeout` 的任务生效，显式设置优先；`0` 表示不限制。该配置只在 server 模式生效，不影响 `run` 模式（run 默认仍为 86400 秒）。

**示例**:

```bash
# 未设超时的任务最长执行 1 小时
pipeline server --task-timeout 3600

# 关闭默认超时
pipeline server --task-timeout 0
```

### `--task-executor`

任务执行方式，控制任务与 server 进程的隔离程度。

- **类型**: 字符串（`in-process` | `subprocess`）
- **环境变量**: `TASK_EXECUTOR`
- **默认值**: `in-process`
- **说明**:
  - `in-process`（默认）：任务在 server 进程内的 goroutine 中执行（与旧行为一致）。
  - `subprocess`：每个任务通过独立 `pipeline run` 子进程执行（复用当前 pipeline 二进制），
    任务的崩溃、panic（含原生服务 SDK）与资源消耗都发生在子进程内，不影响 server 进程本体。
    取消时按进程组终止（Linux/macOS），避免遗留 step 子进程。

**示例**:

```bash
# 任务级进程隔离
pipeline server --task-executor subprocess
```

## 功能特性

### Web Console

访问 `http://localhost:8080/console` 可以打开 Web Console，提供：

- **Pipeline 管理**: 创建、查看、删除 Pipeline
- **搜索和过滤**: 按名称或ID搜索 Pipeline，按状态和时间范围过滤
- **队列管理**: 查看队列状态、取消任务
- **Pipeline 取消**: 支持取消正在执行或等待中的 Pipeline
- **历史记录**: 查看 Pipeline 执行历史
- **日志增强**: 搜索、过滤和导出 Pipeline 执行日志
- **实时日志**: 查看 Pipeline 执行日志，支持实时流式传输
- **Pipeline 定义查看**: 查看完整的 Pipeline YAML 配置，支持一键复制
- **深色模式**: 支持浅色和深色主题切换
- **系统设置**: 配置队列并发数等设置

### REST API

服务器提供以下 REST API 端点：

#### Pipeline 管理

- `GET /api/v1/pipelines` - 获取 Pipeline 列表
  - 查询参数: `search`, `status`, `start_time`, `end_time`, `limit`, `offset`
- `GET /api/v1/pipelines/:id` - 获取 Pipeline 详情
- `GET /api/v1/pipelines/:id/logs` - 获取 Pipeline 日志
  - 查询参数: `search`, `type`, `start_time`, `end_time`, `limit`, `offset`
- `GET /api/v1/pipelines/:id/logs/export` - 导出 Pipeline 日志
  - 查询参数: `format` (text|json), `search`, `type`, `start_time`, `end_time`
- `POST /api/v1/pipelines/:id/cancel` - 取消 Pipeline 执行
- `DELETE /api/v1/pipelines/:id` - 删除 Pipeline 记录
- `POST /api/v1/pipelines/batch/delete` - 批量删除 Pipeline
- `POST /api/v1/pipelines/batch/cancel` - 批量取消 Pipeline

#### 队列管理

- `GET /api/v1/queue/stats` - 获取队列统计信息
- `GET /api/v1/queue` - 获取队列列表
- `DELETE /api/v1/queue/:id` - 取消队列中的任务

#### 系统设置

- `GET /api/v1/settings` - 获取系统设置
- `POST /api/v1/settings` - 保存系统设置（需要重启生效）

### WebSocket 执行

通过 WebSocket 连接可以执行 Pipeline：

- **连接路径**: `ws://localhost:8080/`（或 `wss://` 如果使用 HTTPS）
- **认证**: 如果设置了用户名和密码，需要在连接时提供 Basic Auth
- **消息格式**: JSON 格式的 Action 消息

## 使用示例

### 示例 1: 基本启动

```bash
# 启动服务器（默认端口 8080）
pipeline server

# 访问 Web Console
open http://localhost:8080/console
```

### 示例 2: 自定义端口和工作目录

```bash
pipeline server -p 9090 -w /var/lib/pipeline
```

### 示例 3: 启用认证

```bash
pipeline server -u admin --password secret123
```

### 示例 4: 设置并发数

```bash
pipeline server --max-concurrent 5
```

### 示例 5: 传递环境变量

```bash
pipeline server --allow-env GITHUB_TOKEN --allow-env CI_BUILD_NUMBER
```

## 部署建议

### 生产环境

1. **使用反向代理**: 使用 Nginx 或 Traefik 作为反向代理，提供 HTTPS
2. **启用认证**: 设置用户名和密码保护服务
3. **设置工作目录**: 使用持久化存储作为工作目录
4. **配置并发数**: 根据服务器资源设置合理的并发数
5. **监控和日志**: 配置日志收集和监控系统

### Docker 部署

```bash
docker run -d \
  -p 8080:8080 \
  -v /var/lib/pipeline:/data \
  -e USERNAME=admin \
  -e PASSWORD=secret123 \
  ghcr.io/go-idp/pipeline:latest server
```

## 常见问题

### Q: 如何访问 Web Console？

A: 启动服务器后，访问 `http://localhost:8080/console`（或您配置的端口）。

### Q: 如何启用 HTTPS？

A: 使用反向代理（如 Nginx）提供 HTTPS，或使用 `wss://` 协议连接 WebSocket。

### Q: 如何设置认证？

A: 使用 `-u` 和 `--password` 选项设置用户名和密码。

### Q: 如何查看 Pipeline 执行日志？

A: 在 Web Console 中查看，或通过 REST API `GET /api/v1/pipelines/:id/logs` 获取。

### Q: 任务与 server 进程的隔离？

- server 默认在**进程内**执行任务（goroutine），单个任务 panic 不会拖垮 server（已加 panic 恢复，
  任务标记为 failed）；但任务资源消耗仍在 server 进程内。
- 需要更强的隔离时使用 `--task-executor subprocess`：任务在独立子进程中执行，崩溃与资源
  消耗均不影响 server 进程。彻底隔离（容器沙盒）可在此基础上配合容器化部署实现。

### Q: 如何取消正在执行的 Pipeline？

A: 在 Web Console 中点击"取消"按钮，或通过 REST API `POST /api/v1/pipelines/:id/cancel` 取消。
