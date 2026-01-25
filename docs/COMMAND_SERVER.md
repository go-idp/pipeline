# Pipeline Server 命令文档

## 概述

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

### 1. Web Console

访问 `http://localhost:8080/console` 可以打开 Web Console，提供：

- **Pipeline 管理**: 创建、查看、删除 Pipeline
- **队列管理**: 查看队列状态、取消任务
- **Pipeline 取消**: 支持取消正在执行或等待中的 Pipeline
- **历史记录**: 查看 Pipeline 执行历史
- **实时日志**: 查看 Pipeline 执行日志
- **Pipeline 定义查看**: 查看完整的 Pipeline YAML 配置，支持一键复制
- **系统设置**: 配置队列并发数等设置

#### Web Console 功能说明

**任务卡片布局**：
- 状态标签显示在左侧，便于快速识别 Pipeline 状态
- Pipeline 名称和状态在同一行显示
- Pipeline ID 单独显示在第二行
- 对于 `pending` 或 `running` 状态的 Pipeline，右侧显示"取消"按钮

**Pipeline 详情查看**：
- 支持查看 Pipeline 的完整信息（状态、时间、错误等）
- 支持查看 Pipeline 执行日志
- 支持查看 Pipeline 定义（YAML 格式）
- Pipeline 定义右上角提供"复制"按钮，可一键复制完整的 YAML 配置到剪贴板

**取消 Pipeline**：
- 在 Pipeline 列表中，`pending` 或 `running` 状态的 Pipeline 会显示红色的"取消"按钮
- 点击取消按钮后，会弹出确认对话框
- 取消成功后，Pipeline 状态会变为 `cancelled`，并显示成功通知

### 2. REST API

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

### 3. WebSocket 执行

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
pipeline server \
  -p 9090 \
  -w /var/lib/pipeline
```

### 示例 3: 启用认证

```bash
pipeline server \
  -u admin \
  --password secret123
```

### 示例 4: 设置并发数和环境变量

```bash
pipeline server \
  --max-concurrent 5 \
  --allow-env GITHUB_TOKEN \
  --allow-env CI_BUILD_NUMBER
```

### 示例 5: 使用环境变量配置

```bash
export PORT=9090
export WORKDIR=/var/lib/pipeline
export USERNAME=admin
export PASSWORD=secret123
export MAX_CONCURRENT=5
pipeline server
```

### 示例 6: 使用自定义路径

```bash
pipeline server \
  --path /api/pipeline

# WebSocket 连接地址变为: ws://localhost:8080/api/pipeline
# Web Console 地址变为: http://localhost:8080/api/pipeline/console
```

## 队列系统

服务器内置队列系统，支持：

- **并发控制**: 通过 `--max-concurrent` 控制同时执行的 Pipeline 数量
- **自动执行**: 队列会自动检测并执行待执行的 Pipeline
- **状态管理**: Pipeline 状态包括 `pending`、`running`、`succeeded`、`failed`

### 队列状态

- **pending**: 等待执行
- **running**: 正在执行
- **succeeded**: 执行成功
- **failed**: 执行失败
- **cancelled**: 已取消（由用户主动取消）

### 取消 Pipeline

Pipeline 支持取消操作，可以取消以下状态的 Pipeline：

- **pending**: 等待中的 Pipeline 可以直接取消，状态会变为 `cancelled`
- **running**: 正在执行的 Pipeline 会通过 context 取消机制停止执行，状态会变为 `cancelled`

**注意**：
- 已完成的 Pipeline（`succeeded`、`failed`、`cancelled`）不能取消
- 取消操作会立即停止 Pipeline 执行，但不会清理工作目录（workdir），以便调试
- 取消的 Pipeline 会保留在历史记录中，状态为 `cancelled`

**通过 API 取消**：

```bash
# 取消 Pipeline
curl -X POST http://localhost:8080/api/v1/pipelines/{id}/cancel

# 取消队列中的任务（等同于上面的操作）
curl -X DELETE http://localhost:8080/api/v1/queue/{id}
```

**通过 Web Console 取消**：

在 Web Console 的 Pipeline 列表中，对于 `pending` 或 `running` 状态的 Pipeline，会显示红色的"取消"按钮（位于任务卡片右侧），点击即可取消。

**通过 Web Console 复制 Pipeline 定义**：

1. 在 Pipeline 列表中点击任务卡片，打开详情 Drawer
2. 切换到"Pipeline 定义"标签页
3. 点击右上角的"📋 复制"按钮
4. Pipeline YAML 配置会自动复制到剪贴板，并显示成功通知

## 数据存储

服务器使用内存存储 Pipeline 执行记录：

- **存储位置**: 工作目录下的 `.pipeline_records/` 目录
- **记录格式**: JSON 格式
- **持久化**: 记录会保存到文件，重启后可以恢复

## 安全考虑

### 认证

如果设置了用户名和密码，所有 API 和 WebSocket 连接都需要 Basic Auth 认证。

### 环境变量

默认情况下，服务器的环境变量不会传递给 Pipeline。只有通过 `--allow-env` 或 `--allow-all-env` 明确允许的环境变量才会传递。

### 工作目录隔离

每个 Pipeline 执行都有独立的工作目录，确保执行隔离。

## 部署建议

### 生产环境

```bash
# 使用 systemd 管理服务
cat > /etc/systemd/system/pipeline-server.service <<EOF
[Unit]
Description=Pipeline Server
After=network.target

[Service]
Type=simple
User=pipeline
WorkingDirectory=/var/lib/pipeline
ExecStart=/usr/local/bin/pipeline server \
  -p 8080 \
  -w /var/lib/pipeline \
  -u admin \
  --password $(openssl rand -base64 32) \
  --max-concurrent 5
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

systemctl enable pipeline-server
systemctl start pipeline-server
```

### Docker 部署

```bash
docker run -d \
  --name pipeline-server \
  -p 8080:8080 \
  -v /var/lib/pipeline:/tmp/go-idp/pipeline \
  -e PORT=8080 \
  -e MAX_CONCURRENT=5 \
  ghcr.io/go-idp/pipeline:latest \
  server
```

## API 使用示例

### 获取 Pipeline 列表

```bash
curl http://localhost:8080/api/v1/pipelines
```

### 获取 Pipeline 详情

```bash
curl http://localhost:8080/api/v1/pipelines/{id}
```

### 获取队列统计

```bash
curl http://localhost:8080/api/v1/queue/stats
```

### 取消 Pipeline

```bash
# 取消 Pipeline（推荐）
curl -X POST http://localhost:8080/api/v1/pipelines/{id}/cancel

# 或取消队列中的任务
curl -X DELETE http://localhost:8080/api/v1/queue/{id}
```

### 使用认证

```bash
curl -u admin:password123 http://localhost:8080/api/v1/pipelines
```

## 常见问题

### Q: 如何查看服务器日志？

A: 服务器日志输出到标准输出，可以通过重定向保存到文件：

```bash
pipeline server > server.log 2>&1
```

### Q: 如何重启服务器？

A: 停止服务器（Ctrl+C）后重新启动。Pipeline 执行记录会保留在工作目录中。

### Q: 如何清理旧的 Pipeline 记录？

A: 删除工作目录下的 `.pipeline_records/` 目录中的旧记录文件。

### Q: 队列中的任务会丢失吗？

A: 不会。队列中的任务会持久化到工作目录，重启服务器后会恢复。

### Q: 如何限制并发数？

A: 使用 `--max-concurrent` 选项设置最大并发数。

### Q: Web Console 支持哪些浏览器？

A: 支持所有现代浏览器（Chrome、Firefox、Safari、Edge）。

## 相关文档

- [Client 命令文档](./COMMAND_CLIENT.md)
- [使用文档](./USAGE.md)
- [架构文档](./ARCHITECTURE.md)
