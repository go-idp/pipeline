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

## 功能特性

### Web Console

访问 `http://localhost:8080/console` 可以打开 Web Console，提供：

- **Pipeline 管理**: 创建、查看、删除 Pipeline
- **队列管理**: 查看队列状态、取消任务
- **Pipeline 取消**: 支持取消正在执行或等待中的 Pipeline
- **历史记录**: 查看 Pipeline 执行历史
- **实时日志**: 查看 Pipeline 执行日志
- **Pipeline 定义查看**: 查看完整的 Pipeline YAML 配置，支持一键复制
- **系统设置**: 配置队列并发数等设置

### REST API

服务器提供以下 REST API 端点：

#### Pipeline 管理

- `GET /api/v1/pipelines` - 获取 Pipeline 列表
- `GET /api/v1/pipelines/:id` - 获取 Pipeline 详情
- `GET /api/v1/pipelines/:id/logs` - 获取 Pipeline 日志
- `POST /api/v1/pipelines/:id/cancel` - 取消 Pipeline 执行
- `DELETE /api/v1/pipelines/:id` - 删除 Pipeline 记录

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

### Q: 如何取消正在执行的 Pipeline？

A: 在 Web Console 中点击"取消"按钮，或通过 REST API `POST /api/v1/pipelines/:id/cancel` 取消。
