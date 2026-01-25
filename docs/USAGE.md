# Pipeline 使用文档

## 1. 安装

### 1.1 从源码编译

```bash
git clone https://github.com/go-idp/pipeline.git
cd pipeline
go build -o pipeline cmd/pipeline/main.go
```

### 1.2 使用 Go 安装

```bash
go install github.com/go-idp/pipeline/cmd/pipeline@latest
```

### 1.3 使用 Docker

```bash
docker pull ghcr.io/go-idp/pipeline:latest
```

## 2. 快速开始

### 2.1 创建配置文件

创建 `.pipeline.yaml` 或 `.go-idp/pipeline.yaml` 文件：

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

### 2.2 运行 Pipeline

```bash
pipeline run
```

或者指定配置文件：

```bash
pipeline run -c pipeline.yaml
```

## 3. 配置文件格式

### 3.1 基本结构

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

### 3.2 Stage 配置

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

**注意**: 字段名是 `run_mode`，不是 `mode`。某些示例文件中可能使用了旧的 `mode` 字段名，但实际应该使用 `run_mode`。

**配置继承规则**：
- Stage 会继承 Pipeline 的 `workdir`、`image`、`timeout`、`environment` 配置
- Stage 自己的配置优先级更高，会覆盖继承的配置
- 环境变量会合并，不会覆盖已存在的键

### 3.3 Job 配置

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

**配置继承规则**：
- Job 会继承 Stage 的 `workdir`、`image`、`timeout`、`environment` 配置
- Job 自己的配置优先级更高，会覆盖继承的配置
- 环境变量会合并，不会覆盖已存在的键

### 3.4 Step 配置

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

**配置继承规则**：
- Step 会继承 Job 的 `workdir`、`image`、`timeout`、`environment`、`image_registry` 相关配置
- Step 自己的配置优先级更高，会覆盖继承的配置
- 环境变量会合并，不会覆盖已存在的键
- 如果 `timeout` 未设置，默认值为 86400 秒（1 天）

**重要限制**：
- `language` 和 `plugin` 不能同时使用，会返回错误

## 4. 命令行使用

Pipeline 提供了三个主要命令：

- **`run`**: 在本地直接运行 Pipeline
- **`server`**: 启动 Pipeline 服务，提供 Web Console 和 REST API
- **`client`**: 连接到 Pipeline Server 并执行 Pipeline

### 4.1 run 命令

运行一个 Pipeline：

```bash
pipeline run [选项]
```

**快速示例**:

```bash
# 使用默认配置文件
pipeline run

# 指定配置文件
pipeline run -c my-pipeline.yaml

# 使用远程配置文件
pipeline run -c https://example.com/pipeline.yaml

# 设置工作目录和环境变量
pipeline run -w /tmp/pipeline -e BUILD_ID=123
```

**详细文档**: 请参考 [Run 命令文档](./COMMAND_RUN.md)

### 4.2 server 命令

启动 Pipeline 服务端：

```bash
pipeline server [选项]
```

**快速示例**:

```bash
# 启动服务器（默认端口 8080）
pipeline server

# 自定义端口和工作目录
pipeline server -p 9090 -w /var/lib/pipeline

# 启用认证
pipeline server -u admin --password secret123

# 设置并发数
pipeline server --max-concurrent 5
```

**详细文档**: 请参考 [Server 命令文档](./COMMAND_SERVER.md)

### 4.3 client 命令

连接到 Pipeline Server 并执行 Pipeline：

```bash
pipeline client [选项]
```

**快速示例**:

```bash
# 基本使用
pipeline client -c pipeline.yaml -s ws://localhost:8080

# 使用认证
pipeline client -c pipeline.yaml -s ws://localhost:8080 -u admin -p password123

# 使用远程配置
pipeline client -c https://example.com/pipeline.yaml -s ws://pipeline.example.com
```

**详细文档**: 请参考 [Client 命令文档](./COMMAND_CLIENT.md)

## 5. 使用场景示例

### 5.1 基本构建流程

```yaml
name: Build Application

stages:
  - name: checkout
    jobs:
      - name: checkout
        steps:
          - name: git-clone
            command: |
              git clone https://github.com/user/repo.git .
              git checkout $BRANCH

  - name: build
    jobs:
      - name: build
        steps:
          - name: build-app
            image: golang:1.20
            command: |
              go build -o app ./cmd/app

  - name: test
    jobs:
      - name: test
        steps:
          - name: run-tests
            image: golang:1.20
            command: |
              go test ./...
```

### 5.2 使用 Docker 构建

```yaml
name: Docker Build

image: docker:dind

stages:
  - name: build
    jobs:
      - name: build-image
        steps:
          - name: docker-build
            command: |
              docker build -t myapp:latest .
              docker push myapp:latest
```

### 5.2.1 使用私有镜像仓库

```yaml
name: Use Private Registry

stages:
  - name: build
    jobs:
      - name: build
        # Job 级别配置镜像仓库（会传递给所有 Step）
        image_registry: registry.example.com
        image_registry_username: myuser
        image_registry_password: mypassword
        steps:
          - name: build-step
            image: my-private-image:latest
            command: echo "building..."
          
          - name: another-step
            # 也可以在每个 Step 单独配置
            image: another-private-image:latest
            image_registry: another-registry.com
            image_registry_username: anotheruser
            image_registry_password: anotherpass
            command: echo "another step"
```

**镜像仓库配置说明**：
- 镜像仓库配置从 Job 继承到 Step
- Step 可以覆盖 Job 的镜像仓库配置
- 用于拉取私有 Docker 镜像时进行认证

### 5.3 使用插件

```yaml
name: Use Plugin

stages:
  - name: checkout
    jobs:
      - name: checkout
        steps:
          - name: git-checkout
            plugin:
              image: ghcr.io/go-idp/pipeline-checkout:latest
              settings:
                repository: https://github.com/user/repo.git
                branch: main
                token: ${GITHUB_TOKEN}  # 支持环境变量替换
              entrypoint: /custom/entrypoint  # 可选，默认 /pipeline/plugin/run
```

**插件配置说明**：

1. **Settings 环境变量**：
   - Plugin 的 `settings` 会通过环境变量传递给插件
   - 格式：`PIPELINE_PLUGIN_SETTINGS_<UPPERCASE_KEY>=value`
   - 例如：`settings: { "api_key": "secret" }` → `PIPELINE_PLUGIN_SETTINGS_API_KEY=secret`

2. **环境变量替换**：
   - Settings 的值支持 `${ENV}` 格式的环境变量替换
   - 例如：`token: ${GITHUB_TOKEN}` 会从当前环境变量中获取 `GITHUB_TOKEN` 的值

3. **命令传递**：
   - Step 的 `command` 会通过 `PIPELINE_PLUGIN_COMMAND` 环境变量传递给插件（base64 编码）

4. **环境变量继承**：
   - 使用 `language` 时，插件会自动继承 Step 的环境变量
   - 使用自定义 `plugin` 时，默认不继承环境变量（除非插件内部处理）

### 5.4 使用语言运行时

```yaml
name: Node.js Build

stages:
  - name: build
    jobs:
      - name: build
        steps:
          - name: install-deps
            language:
              name: node
              version: 16
            command: |
              npm install
              
          - name: build-app
            language:
              name: node
              version: 16
            command: |
              npm run build
```

**语言运行时说明**：

1. **自动插件转换**：
   - 使用 `language` 时，会自动转换为对应的插件
   - 镜像格式：`ghcr.io/go-idp/pipeline-language-<name>:<version>`
   - 例如：`language: { name: "node", version: "16" }` → `ghcr.io/go-idp/pipeline-language-node:16`

2. **环境变量继承**：
   - 语言运行时插件会自动继承 Step 的环境变量
   - 你可以在 Step 中设置环境变量，插件中可以使用

3. **限制**：
   - `language` 和 `plugin` **不能同时使用**，会返回错误
   - 如果同时设置，Pipeline 会返回错误：`you can not use language and plugin at the same time`

### 5.5 并行执行任务

```yaml
name: Parallel Jobs

stages:
  - name: build
    run_mode: parallel  # 并行执行所有 job
    jobs:
      - name: build-frontend
        steps:
          - name: build
            command: npm run build
            
      - name: build-backend
        steps:
          - name: build
            command: go build ./cmd/server
```

### 5.6 串行执行阶段

```yaml
name: Serial Stages

stages:
  - name: build
    run_mode: serial  # 串行执行所有 job
    jobs:
      - name: build-step1
        steps:
          - name: step1
            command: echo "step 1"
            
      - name: build-step2
        steps:
          - name: step2
            command: echo "step 2"
```

### 5.7 使用服务编排

```yaml
name: Test with Services

stages:
  - name: setup-services
    jobs:
      - name: start-db
        steps:
          - name: start-postgres
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

  - name: test
    jobs:
      - name: run-tests
        steps:
          - name: test-with-db
            command: |
              # 测试代码，可以连接到 postgres:5432
              pytest tests/
```

### 5.8 使用 SSH 远程执行

```yaml
name: Remote Deploy

stages:
  - name: deploy
    jobs:
      - name: deploy-app
        steps:
          - name: deploy
            engine: ssh://user:password@server.example.com:22
            command: |
              cd /opt/app
              git pull
              docker-compose up -d
```

### 5.9 使用 IDP Agent 执行

```yaml
name: Remote Execution

stages:
  - name: build
    jobs:
      - name: build
        steps:
          - name: build-on-remote
            engine: idp://client_id:client_secret@agent.example.com:8080
            command: |
              go build -o app ./cmd/app
```

### 5.10 使用 Pre/Post 钩子

```yaml
name: With Hooks

pre: |
  echo "Pipeline started at $(date)"
  echo "Setting up environment..."

post: |
  echo "Pipeline finished at $(date)"
  echo "Cleaning up..."

stages:
  - name: build
    jobs:
      - name: build
        steps:
          - name: build
            command: echo "building..."
```

## 6. 环境变量

### 6.1 自动注入的环境变量

Pipeline 会自动注入以下环境变量，可以在任何步骤中使用：

- `PIPELINE_RUNNER`: "pipeline"
- `PIPELINE_RUNNER_OS`: 操作系统（如 "linux", "darwin"）
- `PIPELINE_RUNNER_ARCH`: 架构（如 "amd64", "arm64"）
- `PIPELINE_RUNNER_VERSION`: Pipeline 版本
- `PIPELINE_RUNNER_USER`: 运行用户
- `PIPELINE_RUNNER_WORKDIR`: 运行工作目录
- `PIPELINE_NAME`: 流水线名称
- `PIPELINE_WORKDIR`: 流水线工作目录

### 6.2 使用环境变量

在配置文件中使用环境变量：

```yaml
environment:
  BUILD_ID: "123"
  BRANCH: "main"

stages:
  - name: build
    jobs:
      - name: build
        steps:
          - name: build
            command: |
              echo "Building for branch: $BRANCH"
              echo "Build ID: $BUILD_ID"
```

### 6.3 传递系统环境变量

默认情况下，Pipeline 不会传递系统环境变量。需要显式允许：

```bash
# 允许特定环境变量
pipeline run --allow-env GITHUB_TOKEN --allow-env CI

# 允许所有环境变量
pipeline run --allow-all-env
```

## 7. 配置文件查找顺序

`pipeline run` 命令会按以下顺序查找配置文件：

1. 命令行参数 `-c` 指定的文件
2. 当前目录下的 `.pipeline.yaml`
3. 当前目录下的 `.go-idp/pipeline.yaml`

## 8. 远程配置文件

支持从 HTTP/HTTPS URL 加载配置文件：

```bash
pipeline run -c https://example.com/pipeline.yaml
```

配置文件会被下载到临时文件，执行完成后自动删除（调试模式下保留）。

## 9. 配置继承与合并

### 9.1 配置继承链

配置按照以下层级继承：**Pipeline → Stage → Job → Step**

每一级都会继承父级的配置，子级配置优先级更高。

### 9.2 可继承的配置项

以下配置项支持继承：

- **工作目录** (`workdir`): Pipeline → Stage → Job → Step
- **Docker 镜像** (`image`): Pipeline → Stage → Job → Step
- **超时时间** (`timeout`): Pipeline → Stage → Job → Step
- **环境变量** (`environment`): Pipeline → Stage → Job → Step
- **镜像仓库配置** (`image_registry`, `image_registry_username`, `image_registry_password`): Job → Step

### 9.3 配置合并规则

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

### 9.4 工作目录

- 如果不指定 `workdir`，Pipeline 使用当前目录
- 如果指定了 `workdir`，会在该目录下创建工作目录
- Pipeline 执行完成后会自动清理工作目录（如果不是当前目录）
- 工作目录从 Pipeline → Stage → Job → Step 逐级继承，子级可以覆盖

## 10. 超时控制

### 10.1 设置超时

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

### 10.2 默认超时

- Pipeline: 86400 秒（1 天）
- Step: 86400 秒（1 天）

## 11. 错误处理

### 11.1 失败行为

- 任何 Step 失败，整个 Job 失败
- 任何 Job 失败，整个 Stage 失败
- 任何 Stage 失败，整个 Pipeline 失败
- 失败后立即停止，不执行后续步骤

### 11.2 查看错误

错误信息会输出到 stderr，并记录在 Pipeline 状态中。

## 12. 调试技巧

### 12.1 启用调试模式

```bash
DEBUG=1 pipeline run
```

调试模式下会：
- 保留临时文件
- 输出详细的调试信息
- 显示 Pipeline 配置的 JSON 格式

### 12.2 查看日志

Pipeline 会输出详细的执行日志，包括：
- 每个层级的开始和结束时间
- 执行模式（串行/并行）
- 错误信息

### 12.3 测试单个步骤

可以创建一个简单的 Pipeline 来测试单个步骤：

```yaml
name: Test Step

stages:
  - name: test
    jobs:
      - name: test
        steps:
          - name: my-step
            command: echo "test"
```

## 13. 最佳实践

### 13.1 配置文件组织

- 将配置文件放在项目根目录
- 使用 `.pipeline.yaml` 或 `.go-idp/pipeline.yaml` 作为默认配置
- 为不同环境创建不同的配置文件

### 13.2 环境变量管理

- 使用配置文件中的 `environment` 字段管理环境变量
- 敏感信息通过命令行参数或环境变量传递，不要硬编码在配置文件中
- 使用 `--allow-env` 而不是 `--allow-all-env` 以提高安全性

### 13.3 工作目录

- 为每个 Pipeline 指定独立的工作目录
- 避免使用系统关键目录作为工作目录
- 确保工作目录有足够的磁盘空间

### 13.4 超时设置

- 根据实际需要设置合理的超时时间
- 为长时间运行的任务设置足够的超时时间
- 避免设置过长的超时时间，以免资源浪费

### 13.5 错误处理

- 在关键步骤中添加错误检查
- 使用 `set -e` 在 Shell 脚本中启用错误即退出
- 在 Post 钩子中添加清理逻辑

### 13.6 并行执行

- 对于独立的任务，使用并行执行提高效率
- 对于有依赖关系的任务，使用串行执行
- 注意并行执行时的资源竞争问题

## 14. 常见问题

### 14.1 配置文件找不到

**问题**: `config is required`

**解决**: 
- 确保配置文件存在
- 使用 `-c` 参数指定配置文件路径
- 检查配置文件名称是否正确（`.pipeline.yaml` 或 `.go-idp/pipeline.yaml`）

### 14.2 环境变量未传递

**问题**: 在 Pipeline 中无法访问系统环境变量

**解决**: 
- 使用 `--allow-env` 允许特定环境变量
- 使用 `--allow-all-env` 允许所有环境变量
- 在配置文件的 `environment` 字段中显式设置

### 14.3 Docker 命令失败

**问题**: Docker 相关命令执行失败

**解决**: 
- 确保 Docker 已安装并运行
- 使用 `docker:dind` 镜像以支持 Docker-in-Docker
- 检查 Docker 权限

### 14.4 SSH 连接失败

**问题**: SSH 远程执行失败

**解决**: 
- 检查 SSH 服务器地址和端口
- 验证用户名和密码
- 如果使用私钥，确保 base64 编码正确
- 检查网络连接

### 14.5 工作目录权限问题

**问题**: 无法创建工作目录或写入文件

**解决**: 
- 检查目录权限
- 确保有足够的磁盘空间
- 使用有写权限的目录作为工作目录

### 14.6 配置继承问题

**问题**: 环境变量或配置没有按预期继承

**解决**: 
- 检查配置继承链：Pipeline → Stage → Job → Step
- 记住环境变量合并规则：子级已存在的键不会被父级覆盖
- 确保子级配置为空（空字符串、0、nil）才会继承父级配置
- 查看日志确认最终使用的配置值

### 14.7 Language 和 Plugin 冲突

**问题**: 同时使用 `language` 和 `plugin` 时报错

**解决**: 
- `language` 和 `plugin` 不能同时使用
- 如果使用语言运行时，移除 `plugin` 配置
- 如果需要自定义插件，移除 `language` 配置

## 15. 更多示例

查看 `examples/` 目录下的示例文件：

- `basic.yml`: 基本示例
- `docker.yaml`: Docker 构建示例
- `github.yaml`: GitHub Actions 风格示例
- `plugin.yml`: 插件使用示例
- `language.yml`: 语言运行时示例
- `step-engine-ssh.yaml`: SSH 引擎示例
- `step-service-docker-compose.yaml`: 服务编排示例

