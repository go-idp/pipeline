# 配置文件

Pipeline 使用 YAML 格式的配置文件来定义工作流。本文档详细介绍配置文件的格式和选项。

## 基本结构

```yaml
name: Pipeline Name                    # 必需：流水线名称

workdir: /tmp/my-pipeline             # 可选：工作目录，默认当前目录

image: alpine:latest                   # 可选：默认 Docker 镜像

timeout: 3600                         # 可选：超时时间（秒），默认 86400

environment:                           # 可选：环境变量
  KEY1: value1
  KEY2: value2

pre: echo "pre hook"                  # 可选：前置钩子

post: echo "post hook"                # 可选：后置钩子

stages:                                # 必需：阶段列表
  - name: stage1
    jobs:
      - name: job1
        steps:
          - name: step1
            command: echo "step1"
```

## Pipeline 级别配置

### name

流水线名称，必需字段。

```yaml
name: My Pipeline
```

### workdir

工作目录，可选。默认为当前目录。

```yaml
workdir: /tmp/my-pipeline
```

### image

默认 Docker 镜像，可选。会被 Stage、Job、Step 继承。

```yaml
image: alpine:latest
```

### timeout

超时时间（秒），可选。默认 86400（1 天）。

```yaml
timeout: 3600  # 1 小时
```

### environment

环境变量，可选。会被 Stage、Job、Step 继承并合并。

```yaml
environment:
  KEY1: value1
  KEY2: value2
```

### pre

前置钩子，可选。在 Pipeline 执行前运行。

```yaml
pre: |
  echo "Pipeline started at $(date)"
  echo "Setting up environment..."
```

### post

后置钩子，可选。在 Pipeline 执行后运行（无论成功或失败）。

```yaml
post: |
  echo "Pipeline finished at $(date)"
  echo "Cleaning up..."
```

## Stage 配置

```yaml
stages:
  - name: stage-name                   # 必需：阶段名称
    run_mode: parallel                 # 可选：执行模式，parallel（并行）或 serial（串行），默认 parallel
    workdir: /tmp/stage                # 可选：工作目录（继承自 Pipeline）
    image: alpine:latest                # 可选：Docker 镜像（继承自 Pipeline）
    timeout: 3600                       # 可选：超时时间（秒，继承自 Pipeline）
    environment:                        # 可选：环境变量（合并自 Pipeline）
      KEY: value
    jobs:                              # 必需：任务列表
      - name: job-name
        steps:
          - name: step-name
            command: echo "hello"
```

### run_mode

执行模式，可选。支持 `parallel`（并行）和 `serial`（串行），默认为 `parallel`。

**注意**: 字段名是 `run_mode`，不是 `mode`。

```yaml
run_mode: parallel  # 并行执行所有 job
run_mode: serial    # 串行执行所有 job
```

## Job 配置

```yaml
jobs:
  - name: job-name                     # 必需：任务名称
    workdir: /tmp/job                  # 可选：工作目录（继承自 Stage）
    image: alpine:latest                # 可选：Docker 镜像（继承自 Stage）
    timeout: 3600                       # 可选：超时时间（秒，继承自 Stage）
    environment:                        # 可选：环境变量（合并自 Stage）
      KEY: value
    image_registry: docker.io           # 可选：镜像仓库地址
    image_registry_username: user       # 可选：镜像仓库用户名
    image_registry_password: pass       # 可选：镜像仓库密码
    steps:                              # 必需：步骤列表
      - name: step-name
        command: echo "hello"
```

### image_registry

镜像仓库配置，用于拉取私有 Docker 镜像。

```yaml
image_registry: registry.example.com
image_registry_username: myuser
image_registry_password: mypassword
```

## Step 配置

```yaml
steps:
  - name: step-name                    # 必需：步骤名称
    command: echo "hello"              # 可选：执行的命令
    workdir: /tmp                      # 可选：工作目录（继承自 Job）
    image: alpine:latest               # 可选：Docker 镜像（继承自 Job）
    engine: host                        # 可选：执行引擎（host/docker/ssh/idp）
    shell: /bin/sh                     # 可选：Shell 类型
    timeout: 3600                      # 可选：超时时间（秒，继承自 Job，默认 86400）
    environment:                        # 可选：环境变量（合并自 Job）
      KEY: value
    image_registry: docker.io           # 可选：镜像仓库地址（继承自 Job）
    image_registry_username: user       # 可选：镜像仓库用户名（继承自 Job）
    image_registry_password: pass       # 可选：镜像仓库密码（继承自 Job）
    data_dir_inner: /data              # 可选：容器内数据目录
    data_dir_outer: /host/data         # 可选：宿主机数据目录
    plugin:                            # 可选：插件配置
      image: my-plugin:latest
      settings:
        key: value
        token: ${GITHUB_TOKEN}         # 支持环境变量替换
      entrypoint: /custom/entrypoint  # 可选，默认 /pipeline/plugin/run
    language:                          # 可选：语言运行时
      name: node
      version: 16
    service:                           # 可选：服务编排
      type: docker-compose
      name: my-services
      version: v1
      config: |
        version: '3'
        services:
          db:
            image: postgres:13
```

### engine

执行引擎，可选。支持：
- `host`: 在宿主机执行（默认）
- `docker`: 在 Docker 容器中执行
- `ssh://user:password@host:port`: SSH 远程执行
- `idp://client_id:client_secret@host:port`: IDP Agent 执行

```yaml
engine: host
engine: docker
engine: ssh://user:password@server.example.com:22
engine: idp://client_id:client_secret@agent.example.com:8080
```

### plugin

插件配置，可选。用于使用自定义插件。

```yaml
plugin:
  image: my-plugin:latest
  settings:
    key: value
    token: ${GITHUB_TOKEN}  # 支持环境变量替换
  entrypoint: /custom/entrypoint  # 可选，默认 /pipeline/plugin/run
```

**重要限制**：`language` 和 `plugin` 不能同时使用。

### language

语言运行时，可选。自动转换为对应的插件。

```yaml
language:
  name: node
  version: 16
```

支持的运行时：
- `node`: Node.js
- `python`: Python
- `go`: Go
- `java`: Java
- `rust`: Rust

**重要限制**：`language` 和 `plugin` 不能同时使用。

### service

服务编排，可选。用于启动依赖服务（如数据库）。

```yaml
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

## 配置继承

配置按照以下层级继承：**Pipeline → Stage → Job → Step**

每一级都会继承父级的配置，子级配置优先级更高。

### 可继承的配置项

- **工作目录** (`workdir`): Pipeline → Stage → Job → Step
- **Docker 镜像** (`image`): Pipeline → Stage → Job → Step
- **超时时间** (`timeout`): Pipeline → Stage → Job → Step
- **环境变量** (`environment`): Pipeline → Stage → Job → Step
- **镜像仓库配置** (`image_registry`, `image_registry_username`, `image_registry_password`): Job → Step

### 配置合并规则

1. **基本配置合并**：
   - 如果子级配置为空（空字符串、0、nil），则使用父级配置
   - 如果子级配置已设置，则使用子级配置（覆盖父级）

2. **环境变量合并**：
   - 如果子级 `environment` 为 `nil`，则直接使用父级的 `environment`
   - 如果子级 `environment` 已设置，则合并父级和子级的环境变量
   - **重要**：子级已存在的键不会被父级覆盖，只会添加新的键

**示例**：

```yaml
# Pipeline 级别
environment:
  GLOBAL_VAR: global
  SHARED_VAR: pipeline

stages:
  - name: build
    # Stage 级别
    environment:
      SHARED_VAR: stage  # 这个值会被使用，不会覆盖
      STAGE_VAR: stage
    
    jobs:
      - name: build-job
        # Job 级别
        environment:
          SHARED_VAR: job  # 这个值会被使用
          JOB_VAR: job
        
        steps:
          - name: build-step
            # Step 级别
            environment:
              SHARED_VAR: step  # 这个值会被使用
              STEP_VAR: step
            command: |
              # 最终环境变量：
              # GLOBAL_VAR=global (来自 Pipeline)
              # SHARED_VAR=step (来自 Step，覆盖了所有父级)
              # STAGE_VAR=stage (来自 Stage)
              # JOB_VAR=job (来自 Job)
              # STEP_VAR=step (来自 Step)
              env
```

## 配置文件查找顺序

`pipeline run` 命令会按以下顺序查找配置文件：

1. 命令行参数 `-c` 指定的文件
2. 当前目录下的 `.pipeline.yaml`
3. 当前目录下的 `.go-idp/pipeline.yaml`

## 远程配置文件

支持从 HTTP/HTTPS URL 加载配置文件：

```bash
pipeline run -c https://example.com/pipeline.yaml
```

配置文件会被下载到临时文件，执行完成后自动删除（调试模式下保留）。

## 更多示例

查看 `examples/` 目录下的示例文件：
- `basic.yml`: 基本示例
- `docker.yaml`: Docker 构建示例
- `github.yaml`: GitHub Actions 风格示例
- `plugin.yml`: 插件使用示例
- `language.yml`: 语言运行时示例
- `step-engine-ssh.yaml`: SSH 引擎示例
- `step-service-docker-compose.yaml`: 服务编排示例
