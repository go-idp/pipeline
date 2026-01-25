# 架构概述

Pipeline 是一个用 Go 语言编写的轻量级 CI/CD 流水线执行引擎，类似于 Jenkins、GitLab CI 等工具，但更加轻量和灵活。它支持多种执行引擎、插件系统、服务编排等功能。

## 核心架构

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

## 执行流程

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

## 核心模块

### Pipeline 模块

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

### Stage 模块

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

### Job 模块

**主要文件**:
- `job.go`: Job 结构定义
- `run.go`: Job 执行逻辑
- `setup.go`: Job 初始化
- `state.go`: Job 状态管理

**核心功能**:
- 任务级别的步骤管理
- 步骤按顺序执行
- 环境变量和工作目录继承

### Step 模块

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

Pipeline 支持多种执行引擎：

- **Host**: 在本地主机上直接执行命令（默认）
- **Docker**: 在 Docker 容器中执行命令
- **SSH**: 通过 SSH 在远程服务器上执行命令
- **IDP**: 通过 WebSocket 连接到远程 IDP Agent 执行命令

## 插件系统

Pipeline 支持插件机制，通过 Docker 镜像封装可复用的功能。插件可以：
- 封装复杂的操作逻辑
- 提供语言运行时支持
- 集成第三方服务

## 服务编排

Pipeline 支持服务编排，可以启动依赖服务（如数据库、缓存等）供 Pipeline 使用。

## 配置继承

配置按照以下层级继承：**Pipeline → Stage → Job → Step**

每一级都会继承父级的配置，子级配置优先级更高。

## 更多信息

- [系统架构](./system.md) - 详细的系统架构说明
- [错误处理](./error-handling.md) - 错误处理机制
- [性能优化](./optimization.md) - 性能优化建议
