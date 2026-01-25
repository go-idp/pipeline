# 核心概念

了解 Pipeline 的核心概念有助于更好地使用和理解 Pipeline。

## Pipeline

Pipeline 是最高级别的执行单元，包含多个 Stage。一个 Pipeline 代表一个完整的工作流。

```yaml
name: My Pipeline
stages:
  - name: stage1
    jobs: [...]
```

### Pipeline 生命周期

1. **准备阶段** (`prepare`): 创建工作目录、设置环境变量、添加 pre/post 钩子
2. **执行阶段** (`Run`): 遍历所有 Stage 并执行
3. **清理阶段** (`clean`): 清理工作目录

## Stage

Stage 是 Pipeline 的一个执行阶段，可以包含多个 Job，支持并行或串行执行。

```yaml
stages:
  - name: build
    run_mode: parallel  # parallel 或 serial
    jobs: [...]
```

### Stage 执行模式

- **parallel**（并行）: 所有 Job 同时执行，适合独立任务
- **serial**（串行）: Job 按顺序执行，适合有依赖关系的任务

### Stage 执行流程

1. 根据 `run_mode` 决定执行方式
2. 执行所有 Job
3. 如果任何 Job 失败，Stage 失败
4. 如果 Stage 失败，Pipeline 终止

## Job

Job 是 Stage 中的任务单元，包含多个 Step。Step 按顺序执行。

```yaml
jobs:
  - name: build-job
    steps: [...]
```

### Job 执行流程

1. 顺序执行所有 Step
2. 如果任何 Step 失败，Job 失败
3. 如果 Job 失败，所属 Stage 失败

## Step

Step 是最小的执行单元，执行具体的命令或操作。

```yaml
steps:
  - name: compile
    command: make build
    image: golang:1.20
```

### Step 执行引擎

Step 支持多种执行引擎：

- **host**: 在宿主机执行（默认）
- **docker**: 在 Docker 容器中执行
- **ssh**: SSH 远程执行
- **idp**: IDP Agent 执行

### Step 类型

Step 可以执行不同类型的操作：

1. **命令执行**: 使用 `command` 字段执行 Shell 命令
2. **插件执行**: 使用 `plugin` 字段执行自定义插件
3. **语言运行时**: 使用 `language` 字段使用语言运行时
4. **服务编排**: 使用 `service` 字段启动依赖服务

## 执行流程

```
Pipeline.prepare() - 准备阶段
  ├── 创建工作目录
  ├── 设置环境变量
  ├── 添加 pre/post 钩子
  └── 初始化所有 Stage

Pipeline.Run() - 执行阶段
  └── 遍历所有 Stage
      └── Stage.Run()
          └── 根据 RunMode 执行 Job
              ├── serial: 串行执行
              └── parallel: 并行执行
                  └── Job.Run()
                      └── 顺序执行所有 Step
                          └── Step.Run()
                              └── 根据 Engine 执行命令

Pipeline.clean() - 清理阶段
  └── 清理工作目录
```

## 错误处理

### 失败行为

- 任何 Step 失败，整个 Job 失败
- 任何 Job 失败，整个 Stage 失败
- 任何 Stage 失败，整个 Pipeline 失败
- 失败后立即停止，不执行后续步骤

### 错误状态

当 Pipeline 执行失败时，会设置以下状态：

- **Status**: 设置为 `"failed"`
- **Error**: 包含详细的错误信息
- **FailedAt**: 记录失败时间戳

## 配置继承

配置按照以下层级继承：**Pipeline → Stage → Job → Step**

每一级都会继承父级的配置，子级配置优先级更高。

### 可继承的配置项

- **工作目录** (`workdir`)
- **Docker 镜像** (`image`)
- **超时时间** (`timeout`)
- **环境变量** (`environment`)
- **镜像仓库配置** (`image_registry`, `image_registry_username`, `image_registry_password`)

### 环境变量合并规则

- 如果子级 `environment` 为 `nil`，则直接使用父级的 `environment`
- 如果子级 `environment` 已设置，则合并父级和子级的环境变量
- **重要**：子级已存在的键不会被父级覆盖，只会添加新的键

## 自动注入的环境变量

Pipeline 会自动注入以下环境变量，可以在任何步骤中使用：

- `PIPELINE_RUNNER`: "pipeline"
- `PIPELINE_RUNNER_OS`: 操作系统（如 "linux", "darwin"）
- `PIPELINE_RUNNER_ARCH`: 架构（如 "amd64", "arm64"）
- `PIPELINE_RUNNER_VERSION`: Pipeline 版本
- `PIPELINE_RUNNER_USER`: 运行用户
- `PIPELINE_RUNNER_WORKDIR`: 运行工作目录
- `PIPELINE_NAME`: 流水线名称
- `PIPELINE_WORKDIR`: 流水线工作目录

## 超时控制

### 超时设置

可以在 Pipeline、Stage、Job、Step 级别设置超时时间：

```yaml
name: Timeout Example

timeout: 3600  # Pipeline 级别超时（秒）

stages:
  - name: build
    jobs:
      - name: build
        steps:
          - name: build
            timeout: 1800  # Step 级别超时（秒）
            command: |
              # 长时间运行的命令
```

### 默认超时

- Pipeline: 86400 秒（1 天）
- Step: 86400 秒（1 天）

## 工作目录

### 工作目录规则

- 如果不指定 `workdir`，Pipeline 使用当前目录
- 如果指定了 `workdir`，会在该目录下创建工作目录
- Pipeline 执行完成后会自动清理工作目录（如果不是当前目录）
- 工作目录从 Pipeline → Stage → Job → Step 逐级继承，子级可以覆盖

### 工作目录结构

```
workdir/
  ├── pipeline-<id>/          # Pipeline 工作目录
  │   ├── stage-<id>/         # Stage 工作目录
  │   │   ├── job-<id>/       # Job 工作目录
  │   │   │   └── step-<id>/  # Step 工作目录
  │   │   └── ...
  │   └── ...
  └── ...
```

## 钩子

### Pre 钩子

在 Pipeline 执行前运行，用于准备工作。

```yaml
pre: |
  echo "Pipeline started at $(date)"
  echo "Setting up environment..."
```

### Post 钩子

在 Pipeline 执行后运行（无论成功或失败），用于清理工作。

```yaml
post: |
  echo "Pipeline finished at $(date)"
  echo "Cleaning up..."
```

## 最佳实践

### 1. 使用并行执行提高效率

对于独立的任务，使用并行执行：

```yaml
stages:
  - name: build
    run_mode: parallel
    jobs:
      - name: build-frontend
        steps: [...]
      - name: build-backend
        steps: [...]
```

### 2. 使用串行执行处理依赖

对于有依赖关系的任务，使用串行执行：

```yaml
stages:
  - name: build
    run_mode: serial
    jobs:
      - name: build-step1
        steps: [...]
      - name: build-step2
        steps: [...]  # 依赖 step1
```

### 3. 合理设置超时时间

根据实际需要设置合理的超时时间：

```yaml
timeout: 3600  # 1 小时
```

### 4. 使用环境变量管理配置

使用环境变量而不是硬编码：

```yaml
environment:
  BUILD_ID: "123"
  BRANCH: "main"
```

### 5. 在 Post 钩子中添加清理逻辑

确保资源得到正确清理：

```yaml
post: |
  echo "Cleaning up..."
  rm -rf /tmp/build-artifacts
```
