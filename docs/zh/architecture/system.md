# 系统架构

## 项目概述

Pipeline 是一个用 Go 语言编写的轻量级 CI/CD 流水线执行引擎，类似于 Jenkins、GitLab CI 等工具，但更加轻量和灵活。它支持多种执行引擎、插件系统、服务编排等功能。

**版本**: 1.6.0  
**语言**: Go 1.22.1

## 层次结构

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

## 核心模块

### Pipeline 模块 (`pipeline/`)

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

### Stage 模块 (`stage/`)

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

### Job 模块 (`job/`)

**主要文件**:
- `job.go`: Job 结构定义
- `run.go`: Job 执行逻辑
- `setup.go`: Job 初始化
- `state.go`: Job 状态管理

**核心功能**:
- 任务级别的步骤管理
- 步骤按顺序执行
- 环境变量和工作目录继承

### Step 模块 (`step/`)

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

## 执行引擎

Pipeline 支持多种执行引擎，通过 `Step.Engine` 字段指定：

### Host 引擎（默认）

在本地主机上直接执行命令。

```yaml
steps:
  - name: example
    command: echo "hello"
    # engine: host (默认)
```

### Docker 引擎

在 Docker 容器中执行命令。如果指定了 `image`，自动使用 Docker 引擎。

```yaml
steps:
  - name: example
    image: alpine:latest
    command: echo "hello"
    # engine: docker (自动)
```

### SSH 引擎

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

### IDP 引擎

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

## 插件系统

Pipeline 支持插件机制，通过 Docker 镜像封装可复用的功能。

### 插件配置

```yaml
steps:
  - name: checkout
    plugin:
      image: ghcr.io/go-idp/pipeline-checkout:latest
      settings:
        repository: https://github.com/user/repo.git
        branch: main
        token: ${GITHUB_TOKEN}
```

### 插件工作原理

1. Pipeline 将插件配置转换为环境变量
2. 启动 Docker 容器运行插件镜像
3. 插件通过环境变量接收配置
4. 插件执行完成后返回结果

## 服务编排

Pipeline 支持服务编排，可以启动依赖服务（如数据库、缓存等）供 Pipeline 使用。

### Docker Compose 服务

```yaml
steps:
  - name: test-with-db
    service:
      type: docker-compose
      name: test-db
      version: v1
      config: |
        version: '3'
        services:
          postgres:
            image: postgres:13
            environment:
              POSTGRES_PASSWORD: test
            ports:
              - "5432:5432"
    command: |
      # 测试代码，可以连接到 postgres:5432
      pytest tests/
```

## 配置继承

配置按照以下层级继承：**Pipeline → Stage → Job → Step**

每一级都会继承父级的配置，子级配置优先级更高。

### 可继承的配置项

- **工作目录** (`workdir`)
- **Docker 镜像** (`image`)
- **超时时间** (`timeout`)
- **环境变量** (`environment`)
- **镜像仓库配置** (`image_registry`, `image_registry_username`, `image_registry_password`)

## 数据流

### Pipeline 执行数据流

```
配置文件 (YAML)
  ↓
Pipeline 对象
  ↓
Stage 对象（继承配置）
  ↓
Job 对象（继承配置）
  ↓
Step 对象（继承配置）
  ↓
执行引擎（Host/Docker/SSH/IDP）
  ↓
命令执行
  ↓
结果返回
```

## 扩展机制

### 自定义执行引擎

可以通过实现 `step.Engine` 接口来添加自定义执行引擎。

### 自定义插件

可以通过创建 Docker 镜像来创建自定义插件。

## 项目结构

```
pipeline/
├── cmd/pipeline/          # 命令行入口
│   └── commands/          # 命令实现
│       ├── run.go         # run 命令
│       ├── server.go       # server 命令
│       └── client.go       # client 命令
├── svc/                   # 服务层
│   ├── server/            # Server 实现
│   │   ├── server.go      # Server 主逻辑
│   │   ├── queue.go       # 队列系统
│   │   ├── store.go       # 存储系统
│   │   └── console.html   # Web Console
│   └── client/            # Client 实现
├── pipeline/              # Pipeline 核心
├── stage/                  # Stage 核心
├── job/                    # Job 核心
├── step/                   # Step 核心
├── examples/               # 示例配置
└── docs/                   # 文档
```
