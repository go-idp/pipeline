# 介绍

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

## 🚀 快速开始

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

### 2. Server 模式

启动 Pipeline 服务，提供 Web Console 和 REST API：

```bash
# 启动服务器
pipeline server

# 访问 Web Console
open http://localhost:8080/console
```

### 3. Client 模式

连接到 Pipeline Server 并执行 Pipeline：

```bash
pipeline client -c pipeline.yaml -s ws://localhost:8080
```

## 💡 使用场景

- **CI/CD**: 作为 CI/CD 流水线执行引擎
- **自动化任务**: 执行各种自动化任务和脚本
- **构建系统**: 作为构建和部署系统
- **任务调度**: 作为任务调度和执行平台

---

**开始使用 Pipeline，让工作流执行更简单！** 🚀
