# 命令概述

Pipeline 提供了三个主要命令来满足不同的使用场景。

## 命令列表

### run

在本地直接运行 Pipeline。这是 Pipeline 最常用的命令。

```bash
pipeline run [选项]
```

**适用场景**:
- 本地开发和测试
- CI/CD 流水线执行
- 自动化脚本执行

**详细文档**: [run 命令](./run.md)

### server

启动 Pipeline 服务，提供 Web Console 和 REST API。

```bash
pipeline server [选项]
```

**适用场景**:
- 生产环境部署
- 需要 Web 界面管理
- 需要 REST API 集成
- 需要队列管理

**详细文档**: [server 命令](./server.md)

### client

连接到 Pipeline Server 并通过 WebSocket 执行 Pipeline。

```bash
pipeline client [选项]
```

**适用场景**:
- 远程执行 Pipeline
- CI/CD 集成
- 分布式执行

**详细文档**: [client 命令](./client.md)

## 命令选择指南

### 本地开发

使用 `run` 命令：

```bash
pipeline run -c pipeline.yaml
```

### 服务部署

使用 `server` 命令：

```bash
pipeline server -p 8080
```

### 远程执行

使用 `client` 命令：

```bash
pipeline client -c pipeline.yaml -s ws://server:8080
```

## 命令组合使用

### Server + Client 模式

1. 启动 Server：

```bash
pipeline server -p 8080
```

2. 使用 Client 连接：

```bash
pipeline client -c pipeline.yaml -s ws://localhost:8080
```

这种模式适合：
- 集中式 Pipeline 管理
- 多客户端执行
- 队列和并发控制

## 环境变量

所有命令都支持通过环境变量设置选项：

```bash
export PIPELINE_CONFIG=pipeline.yaml
export PIPELINE_WORKDIR=/tmp/pipeline
pipeline run
```

## 配置文件

所有命令都支持配置文件，查找顺序：

1. 命令行参数 `-c` 指定的文件
2. `.pipeline.yaml`（当前目录）
3. `.go-idp/pipeline.yaml`（当前目录）

## 调试模式

所有命令都支持调试模式：

```bash
DEBUG=1 pipeline run -c pipeline.yaml
DEBUG=1 pipeline server
DEBUG=1 pipeline client -c pipeline.yaml -s ws://localhost:8080
```

调试模式下会输出详细的执行信息。
