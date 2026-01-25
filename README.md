# Pipeline

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

Pipeline 是一个强大的工作流执行引擎，支持本地执行和服务化部署。它提供了灵活的配置方式、丰富的执行引擎、以及完整的 Web Console 和 REST API。

## ✨ 特性

- 🚀 **多种执行模式**: 支持本地运行、Server 模式和 Client 模式
- 🐳 **多执行引擎**: 支持 host、docker、ssh、idp 等多种执行引擎
- 📊 **Web Console**: 提供完整的 Web 界面，支持 Pipeline 管理和监控
- 🔄 **队列系统**: 内置队列系统，支持并发控制和任务管理
- 📝 **完整日志**: 详细的执行日志和错误信息
- 🔧 **灵活配置**: 支持 YAML 配置文件和命令行参数
- 🔌 **插件系统**: 支持自定义插件扩展功能
- 🌐 **服务化**: 支持通过 WebSocket 和 REST API 远程执行

## 🚀 快速开始

### 安装

#### 从源码编译

```bash
git clone https://github.com/go-idp/pipeline.git
cd pipeline
go build -o pipeline cmd/pipeline/main.go
```

#### 使用 Go 安装

```bash
go install github.com/go-idp/pipeline/cmd/pipeline@latest
```

#### 使用 Docker

```bash
docker pull ghcr.io/go-idp/pipeline:latest
```

### 第一个 Pipeline

1. **创建配置文件** `.pipeline.yaml`:

```yaml
name: My First Pipeline

stages:
  - name: build
    jobs:
      - name: build-job
        steps:
          - name: hello
            command: echo "Hello, Pipeline!"
```

2. **运行 Pipeline**:

```bash
pipeline run
```

## 📖 使用方式

### 1. 本地运行

直接在本地执行 Pipeline：

```bash
pipeline run -c pipeline.yaml
```

**详细文档**: [Run 命令文档](./docs/COMMAND_RUN.md)

### 2. Server 模式

启动 Pipeline 服务，提供 Web Console 和 REST API：

```bash
# 启动服务器
pipeline server

# 访问 Web Console
open http://localhost:8080/console
```

**详细文档**: [Server 命令文档](./docs/COMMAND_SERVER.md)

### 3. Client 模式

连接到 Pipeline Server 并执行 Pipeline：

```bash
pipeline client -c pipeline.yaml -s ws://localhost:8080
```

**详细文档**: [Client 命令文档](./docs/COMMAND_CLIENT.md)

## 📚 文档

- **[使用文档](./docs/USAGE.md)** - 完整的使用指南和配置说明
- **[Run 命令文档](./docs/COMMAND_RUN.md)** - 本地运行 Pipeline 的详细说明
- **[Server 命令文档](./docs/COMMAND_SERVER.md)** - Server 模式的配置和使用
- **[Client 命令文档](./docs/COMMAND_CLIENT.md)** - Client 模式的配置和使用
- **[错误处理文档](./docs/ERROR_HANDLING.md)** - 错误处理机制和调试指南
- **[架构文档](./docs/ARCHITECTURE.md)** - 系统架构和设计说明
- **[优化文档](./docs/OPTIMIZATION.md)** - 性能优化建议
- **[文档索引](./docs/README.md)** - 所有文档的索引

## 🎯 核心概念

### Pipeline

Pipeline 是最高级别的执行单元，包含多个 Stage。

```yaml
name: My Pipeline
stages:
  - name: stage1
    jobs: [...]
```

### Stage

Stage 是 Pipeline 的一个执行阶段，可以包含多个 Job，支持并行或串行执行。

```yaml
stages:
  - name: build
    run_mode: parallel  # parallel 或 serial
    jobs: [...]
```

### Job

Job 是 Stage 中的任务单元，包含多个 Step。

```yaml
jobs:
  - name: build-job
    steps: [...]
```

### Step

Step 是最小的执行单元，执行具体的命令或操作。

```yaml
steps:
  - name: compile
    command: make build
    image: golang:1.20
```

## 🔧 配置示例

### 基本配置

```yaml
name: Build Application

stages:
  - name: checkout
    jobs:
      - name: checkout
        steps:
          - name: git-clone
            command: git clone https://github.com/user/repo.git .

  - name: build
    jobs:
      - name: build
        steps:
          - name: build-app
            image: golang:1.20
            command: go build -o app ./cmd/app
```

### 使用 Docker

```yaml
name: Docker Build

stages:
  - name: build
    jobs:
      - name: build-image
        steps:
          - name: build
            image: docker:latest
            command: docker build -t myapp:latest .
```

### 使用插件

```yaml
name: Plugin Example

stages:
  - name: deploy
    jobs:
      - name: deploy
        steps:
          - name: deploy-step
            plugin:
              image: my-plugin:latest
              settings:
                token: ${GITHUB_TOKEN}
```

更多示例请查看 [examples](./examples/) 目录。

## 🌟 主要功能

### Web Console

Pipeline Server 提供完整的 Web Console，支持：

- 📊 Pipeline 管理：创建、查看、删除 Pipeline
- 📈 队列监控：实时查看队列状态和统计信息
- 📝 日志查看：查看 Pipeline 执行日志和 Pipeline 定义
- ⚙️ 系统设置：配置队列并发数等系统参数
- 🔄 自动刷新：自动刷新 Pipeline 状态和队列信息

### 队列系统

- **并发控制**: 可配置最大并发执行数
- **自动执行**: 队列自动检测并执行待执行的 Pipeline
- **状态管理**: 完整的 Pipeline 状态跟踪（pending、running、succeeded、failed）
- **任务取消**: 支持取消队列中的任务

### 错误处理

- **Workdir 保留**: 失败时保留 workdir 以便调试
- **详细日志**: 输出详细的错误信息和调试提示
- **状态跟踪**: 完整的执行状态和错误信息记录

## 🛠️ 开发

### 运行测试

```bash
go test ./...
```

### 构建

```bash
go build -o pipeline cmd/pipeline/main.go
```

### 运行示例

```bash
# 运行基本示例
pipeline run -c examples/basic.yml

# 运行 Docker 示例（需要 Docker）
pipeline run -c examples/docker.yaml
```

## 📦 项目结构

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
├── examples/              # 示例配置
├── docs/                  # 文档
└── *.go                   # 核心代码
```

## 🤝 贡献

欢迎贡献代码、报告问题或提出建议！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 📄 许可证

本项目采用 MIT 许可证。详情请查看 [LICENSE](LICENSE) 文件。

## 🔗 相关链接

- **GitHub**: https://github.com/go-idp/pipeline
- **文档**: [docs/](./docs/)
- **示例**: [examples/](./examples/)

## 💡 使用场景

- **CI/CD**: 作为 CI/CD 流水线执行引擎
- **自动化任务**: 执行各种自动化任务和脚本
- **构建系统**: 作为构建和部署系统
- **任务调度**: 作为任务调度和执行平台

---

**开始使用 Pipeline，让工作流执行更简单！** 🚀
