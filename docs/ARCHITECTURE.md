# Pipeline 技术架构文档

## 1. 项目概述

Pipeline 是一个用 Go 语言编写的轻量级 CI/CD 流水线执行引擎，类似于 Jenkins、GitLab CI 等工具，但更加轻量和灵活。它支持多种执行引擎、插件系统、服务编排等功能。

**版本**: 1.6.0  
**语言**: Go 1.22.1

## 2. 核心架构

### 2.1 层次结构

Pipeline 采用四层架构模型：

```
Pipeline (管道)
  └── Stage (阶段)
      └── Job (任务)
          └── Step (步骤)
```

- **Pipeline**: 整个流水线的容器，包含多个阶段
- **Stage**: 流水线的阶段，可以串行或并行执行多个任务
- **Job**: 任务单元，包含多个步骤，步骤按顺序执行
- **Step**: 最小执行单元，执行具体的命令或操作

### 2.2 执行流程

```
1. Pipeline.prepare() - 准备阶段
   ├── 创建工作目录
   ├── 设置环境变量
   ├── 添加 pre/post 钩子
   └── 初始化所有 Stage

2. Pipeline.Run() - 执行阶段
   └── 遍历所有 Stage
       └── Stage.Run()
           └── 根据 RunMode 执行 Job
               ├── serial: 串行执行
               └── parallel: 并行执行
                   └── Job.Run()
                       └── 顺序执行所有 Step
                           └── Step.Run()
                               └── 根据 Engine 执行命令

3. Pipeline.clean() - 清理阶段
   └── 清理工作目录
```

## 3. 核心模块

### 3.1 Pipeline 模块 (`pipeline/`)

**主要文件**:
- `pipeline.go`: Pipeline 结构定义和基础方法
- `run.go`: Pipeline 执行逻辑
- `state.go`: Pipeline 状态管理

**核心功能**:
- 流水线生命周期管理
- 环境变量管理
- 工作目录管理
- Pre/Post 钩子支持
- 超时控制（默认 86400 秒）

**关键方法**:
- `Run(ctx, opts...)`: 执行流水线
- `prepare(id)`: 准备阶段
- `clean()`: 清理阶段
- `SetWorkdir(workdir)`: 设置工作目录
- `SetEnvironment(env)`: 设置环境变量
- `SetImage(image)`: 设置默认镜像
- `SetTimeout(timeout)`: 设置超时时间

### 3.2 Stage 模块 (`stage/`)

**主要文件**:
- `stage.go`: Stage 结构定义
- `run.go`: Stage 执行逻辑
- `setup.go`: Stage 初始化
- `state.go`: Stage 状态管理
- `constants.go`: 运行模式常量

**核心功能**:
- 阶段级别的任务编排
- 支持串行（serial）和并行（parallel）两种执行模式
- 环境变量和工作目录继承

**运行模式**:
- `serial`: 串行执行，任务按顺序执行
- `parallel`: 并行执行，所有任务同时执行（默认）

### 3.3 Job 模块 (`job/`)

**主要文件**:
- `job.go`: Job 结构定义
- `run.go`: Job 执行逻辑
- `setup.go`: Job 初始化
- `state.go`: Job 状态管理

**核心功能**:
- 任务级别的步骤管理
- 步骤按顺序执行
- 环境变量和工作目录继承

### 3.4 Step 模块 (`step/`)

**主要文件**:
- `step.go`: Step 结构定义
- `run.go`: Step 执行逻辑
- `setup.go`: Step 初始化
- `state.go`: Step 状态管理
- `plugin.go`: 插件支持
- `service.go`: 服务编排支持
- `init.go`: 执行引擎注册

**核心功能**:
- 命令执行
- 多种执行引擎支持
- 插件系统
- 语言运行时支持
- 服务编排支持

## 4. 执行引擎

Pipeline 支持多种执行引擎，通过 `Step.Engine` 字段指定：

### 4.1 Host 引擎（默认）

在本地主机上直接执行命令。

```yaml
steps:
  - name: example
    command: echo "hello"
    # engine: host (默认)
```

### 4.2 Docker 引擎

在 Docker 容器中执行命令。如果指定了 `image`，自动使用 Docker 引擎。

```yaml
steps:
  - name: example
    image: alpine:latest
    command: echo "hello"
    # engine: docker (自动)
```

### 4.3 SSH 引擎

通过 SSH 在远程服务器上执行命令。

```yaml
steps:
  - name: example
    engine: ssh://username:password@host:port
    command: echo "hello"
```

支持私钥认证：
```yaml
steps:
  - name: example
    engine: ssh://private_key:base64_encoded_key@host:port
    command: echo "hello"
```

### 4.4 IDP 引擎

通过 WebSocket 连接到远程 IDP Agent 执行命令。

```yaml
steps:
  - name: example
    engine: idp://client_id:client_secret@host:port
    command: echo "hello"
```

支持 TLS：
```yaml
steps:
  - name: example
    engine: idps://client_id:client_secret@host:port
    command: echo "hello"
```

## 5. 插件系统

Pipeline 支持插件机制，通过 Docker 镜像封装可复用的功能。

### 5.1 插件配置

```yaml
steps:
  - name: checkout
    plugin:
      image: docker.io/library/busybox:latest
      settings:
        username: "test"
        token: "test"
        repository: https://github.com/go-idp/pipeline
      entrypoint: /bin/env  # 可选，默认 /pipeline/plugin/run
```

### 5.2 插件环境变量

插件的 `settings` 会自动转换为环境变量：
- 格式: `PIPELINE_PLUGIN_SETTINGS_<KEY>` (KEY 转为大写)
- 例如: `{"key": "value"}` => `PIPELINE_PLUGIN_SETTINGS_KEY=value`

## 6. 语言运行时支持

Pipeline 支持通过语言运行时执行脚本，自动选择合适的 Docker 镜像。

```yaml
steps:
  - name: build
    command: |
      console.log('Hello from Node.js')
    language:
      name: node
      version: 16
```

支持的语言：
- `node`: Node.js
- `go`: Go
- `python`: Python
- 更多语言通过插件镜像提供

## 7. 服务编排

Pipeline 支持在步骤中启动服务（如数据库、缓存等），支持以下类型：

### 7.1 Docker Compose

```yaml
steps:
  - name: start-services
    service:
      type: docker-compose
      name: my-services
      version: v1
      config: |
        version: '3'
        services:
          db:
            image: postgres:13
```

### 7.2 Docker Swarm

```yaml
steps:
  - name: deploy-stack
    service:
      type: docker-swarm
      name: my-stack
      version: v1
      config: |
        version: '3'
        services:
          app:
            image: myapp:latest
```

### 7.3 Kubernetes

```yaml
steps:
  - name: deploy-k8s
    service:
      type: kubernetes
      name: my-deployment
      version: v1
      config: |
        apiVersion: v1
        kind: Pod
        ...
```

## 8. 服务端/客户端模式

### 8.1 服务端模式 (`server`)

启动一个 WebSocket 服务器，接收客户端提交的流水线任务。

**架构**:
```
Client (WebSocket)
  └── Server (WebSocket Server)
      └── Pipeline Execution
```

**功能**:
- WebSocket 连接管理
- 流水线任务接收和执行
- 实时输出流式传输（stdout/stderr）
- 错误处理和状态报告

**关键文件**:
- `svc/server/server.go`: 服务器接口
- `svc/server/core.go`: WebSocket 处理逻辑
- `svc/server/config.go`: 服务器配置
- `svc/server/run.go`: 服务器启动

### 8.2 客户端模式 (`client`)

连接到服务端，提交流水线任务并接收执行结果。

**架构**:
```
Client
  ├── Connect to Server (WebSocket)
  ├── Send Pipeline Config
  ├── Receive stdout/stderr
  └── Receive Done/Error
```

**关键文件**:
- `svc/client/client.go`: 客户端接口
- `svc/client/connect.go`: 连接逻辑
- `svc/client/run.go`: 任务提交和执行
- `svc/client/config.go`: 客户端配置

### 8.3 通信协议

使用 JSON 格式的 WebSocket 消息：

**Action 类型**:
- `run`: 执行流水线
- `stdout`: 标准输出
- `stderr`: 标准错误
- `done`: 执行完成
- `error`: 执行错误

**消息格式**:
```json
{
  "type": "run|stdout|stderr|done|error",
  "payload": "..."
}
```

## 9. 状态管理

每个层级（Pipeline、Stage、Job、Step）都有独立的状态管理：

**状态类型**:
- `pending`: 等待执行
- `running`: 正在执行
- `succeeded`: 执行成功
- `failed`: 执行失败

**状态字段**:
- `ID`: 唯一标识
- `Status`: 当前状态
- `StartedAt`: 开始时间
- `SucceedAt`: 成功时间
- `FailedAt`: 失败时间
- `Error`: 错误信息

## 10. 环境变量

### 10.1 自动注入的环境变量

Pipeline 会自动注入以下环境变量：

- `PIPELINE_RUNNER`: 固定值 "pipeline"
- `PIPELINE_RUNNER_OS`: 运行系统（如 "linux", "darwin"）
- `PIPELINE_RUNNER_ARCH`: 运行架构（如 "amd64", "arm64"）
- `PIPELINE_RUNNER_VERSION`: Pipeline 版本
- `PIPELINE_RUNNER_USER`: 运行用户
- `PIPELINE_RUNNER_WORKDIR`: 运行工作目录
- `PIPELINE_NAME`: 流水线名称
- `PIPELINE_WORKDIR`: 流水线工作目录

### 10.2 环境变量继承

环境变量从 Pipeline -> Stage -> Job -> Step 逐级继承，子级可以覆盖父级的值。

## 11. 工作目录管理

### 11.1 工作目录继承

工作目录从 Pipeline -> Stage -> Job -> Step 逐级继承。

### 11.2 工作目录清理

Pipeline 执行完成后会自动清理工作目录（如果工作目录不是当前目录）。

## 12. 超时控制

Pipeline 支持在 Pipeline、Stage、Job、Step 四个层级设置超时时间，通过 Context 超时机制实现。

### 12.1 超时继承机制

- **Pipeline 级别**：默认 86400 秒（1 天），可通过 `SetTimeout()` 方法或配置文件的 `timeout` 字段设置
- **Stage 级别**：继承 Pipeline 的超时设置，可通过配置文件的 `timeout` 字段覆盖
- **Job 级别**：继承 Stage 的超时设置，可通过配置文件的 `timeout` 字段覆盖
- **Step 级别**：继承 Job 的超时设置，默认 86400 秒，可通过配置文件的 `timeout` 字段覆盖

### 12.2 Context 超时实现

每个层级的 `Run()` 方法都会根据该层级的 `Timeout` 字段创建带超时的 Context：

```go
// Pipeline.Run()
if p.Timeout > 0 {
    ctx, cancel = context.WithTimeout(ctx, time.Duration(p.Timeout)*time.Second)
    defer cancel()
}

// Stage.Run()
if s.Timeout > 0 {
    ctx, cancel = context.WithTimeout(ctx, time.Duration(s.Timeout)*time.Second)
    defer cancel()
}

// Job.Run()
if j.Timeout > 0 {
    ctx, cancel = context.WithTimeout(ctx, time.Duration(j.Timeout)*time.Second)
    defer cancel()
}

// Step.Run()
if s.Timeout > 0 {
    ctx, cancel = context.WithTimeout(ctx, time.Duration(s.Timeout)*time.Second)
    defer cancel()
}
```

### 12.3 超时行为

1. **超时传播**：超时 Context 会向下传递给子层级，子层级可以设置更短的超时时间
2. **超时检测**：当超时发生时，会检测 `context.DeadlineExceeded` 或 `context.Canceled` 错误
3. **错误信息**：超时错误会包含层级信息和超时时间，例如：`"pipeline timeout after 60 seconds: context deadline exceeded"`
4. **状态更新**：超时发生时，对应层级的 `State.Status` 会被设置为 `"failed"`，`State.Error` 包含超时信息

### 12.4 并行执行中的超时

在 Stage 的并行执行模式中，`errgroup.WithContext()` 会从带超时的 Context 创建新的 Context，确保：
- 任何一个 Job 超时，会取消所有并行执行的 Job
- 超时错误会正确传播到 Stage 层级

### 12.5 配置示例

```yaml
name: "test pipeline"
timeout: 3600  # Pipeline 级别：1 小时

stages:
  - name: "build"
    timeout: 1800  # Stage 级别：30 分钟（覆盖 Pipeline 的 1 小时）
    jobs:
      - name: "build job"
        timeout: 900  # Job 级别：15 分钟（覆盖 Stage 的 30 分钟）
        steps:
          - name: "build step"
            command: "make build"
            timeout: 600  # Step 级别：10 分钟（覆盖 Job 的 15 分钟）
```

### 12.6 超时日志

每个层级在设置超时时会输出日志，便于调试和监控：

```
[workflow] timeout: 3600 seconds
[stage(1/2): build] timeout: 1800 seconds
[job(1/1): build job] timeout: 900 seconds
[step(1/1): build step] timeout: 600 seconds
```

## 13. 依赖关系

### 13.1 核心依赖

- `github.com/go-idp/agent`: IDP Agent 支持
- `github.com/go-zoox/command`: 命令执行引擎
- `github.com/go-zoox/cli`: CLI 框架
- `github.com/go-zoox/websocket`: WebSocket 支持
- `github.com/go-zoox/zoox`: Web 框架

### 13.2 其他依赖

- `golang.org/x/sync`: 并发控制
- Docker SDK: Docker 容器支持
- 其他 go-zoox 系列工具库

## 14. 扩展性

### 14.1 自定义执行引擎

通过实现 `github.com/go-zoox/command/engine.Engine` 接口，可以添加自定义执行引擎。

### 14.2 插件开发

开发插件只需：
1. 创建 Docker 镜像
2. 实现 `/pipeline/plugin/run` 入口点（或自定义）
3. 通过环境变量接收配置

### 14.3 服务编排扩展

通过修改 `step/service.go` 可以添加新的服务编排类型支持。

## 15. 安全考虑

1. **嵌套 Pipeline 防护**: 防止在 Pipeline 内部再次执行 Pipeline
2. **环境变量隔离**: 通过 `allow-env` 和 `allow-all-env` 控制环境变量传递
3. **工作目录隔离**: 每个 Pipeline 执行都有独立的工作目录
4. **超时保护**: 防止长时间运行的命令占用资源
5. **SSH 密钥管理**: 支持私钥认证，密钥通过 base64 编码传递

## 16. 性能优化

1. **并行执行**: Stage 和 Job 支持并行执行，提高效率
2. **资源复用**: 工作目录和容器可以复用
3. **流式输出**: 通过 WebSocket 实现实时输出，无需等待完成
4. **错误快速失败**: 任何步骤失败立即停止，避免浪费资源

## 17. 日志系统

Pipeline 使用 `github.com/go-zoox/logger` 进行日志记录：

- 支持结构化日志
- 可配置输出目标（stdout/stderr）
- 日志包含层级信息（Pipeline -> Stage -> Job -> Step）

## 18. 错误处理

- 每个层级都有独立的错误处理
- 错误信息会向上传播
- 失败状态会记录在 State 中
- 支持错误恢复（通过重试机制，需要外部实现）


