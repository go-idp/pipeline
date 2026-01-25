# Pipeline Run 命令文档

## 概述

`pipeline run` 命令用于在本地直接运行 Pipeline。这是 Pipeline 最常用的命令，支持从本地文件或远程 URL 加载配置，并执行 Pipeline 工作流。

## 基本用法

```bash
pipeline run [选项]
```

## 命令选项

### `-c, --config`

指定 Pipeline 配置文件路径。

- **类型**: 字符串
- **环境变量**: `PIPELINE_CONFIG`
- **默认值**: 自动查找（`.pipeline.yaml` 或 `.go-idp/pipeline.yaml`）
- **支持格式**:
  - 本地文件路径: `pipeline.yaml`
  - HTTP/HTTPS URL: `https://example.com/pipeline.yaml`

**示例**:

```bash
# 使用本地配置文件
pipeline run -c pipeline.yaml

# 使用远程配置文件
pipeline run -c https://example.com/pipeline.yaml

# 使用环境变量
export PIPELINE_CONFIG=pipeline.yaml
pipeline run
```

### `-w, --workdir`

指定 Pipeline 的工作目录。

- **类型**: 字符串
- **环境变量**: `PIPELINE_WORKDIR`
- **默认值**: 当前目录

**示例**:

```bash
pipeline run -w /tmp/my-pipeline
```

### `-i, --image`

指定默认的 Docker 镜像。

- **类型**: 字符串
- **环境变量**: `PIPELINE_IMAGE`
- **说明**: 如果 Pipeline 配置中没有指定 `image`，将使用此值

**示例**:

```bash
pipeline run -i alpine:latest
```

### `-e, --env`

设置环境变量（可多次使用）。

- **类型**: 字符串切片
- **环境变量**: `ENV`
- **格式**: `KEY=VALUE`

**示例**:

```bash
pipeline run -e GITHUB_TOKEN=xxx -e BUILD_NUMBER=123
```

### `--allow-env`

允许传递指定的环境变量到 Pipeline（可多次使用）。

- **类型**: 字符串切片
- **环境变量**: `ALLOW_ENV`
- **说明**: 从当前 shell 环境中选择性地传递环境变量

**示例**:

```bash
# 允许传递 GITHUB_TOKEN 和 CI 相关的环境变量
pipeline run --allow-env GITHUB_TOKEN --allow-env CI_BUILD_NUMBER
```

### `--allow-all-env`

允许传递所有环境变量到 Pipeline。

- **类型**: 布尔值
- **环境变量**: `ALLOW_ALL_ENV`
- **说明**: 将当前 shell 的所有环境变量传递给 Pipeline

**示例**:

```bash
pipeline run --allow-all-env
```

## 配置文件查找

如果不指定 `-c` 选项，`pipeline run` 会自动查找配置文件，按以下顺序：

1. `.pipeline.yaml`（当前目录）
2. `.go-idp/pipeline.yaml`（当前目录）

如果找到配置文件，将自动使用；否则会报错。

## 使用示例

### 示例 1: 基本使用

```bash
# 创建配置文件
cat > .pipeline.yaml <<EOF
name: Hello Pipeline
stages:
  - name: greet
    jobs:
      - name: say-hello
        steps:
          - name: hello
            command: echo "Hello, World!"
EOF

# 运行 Pipeline
pipeline run
```

### 示例 2: 使用远程配置

```bash
# 从远程 URL 加载配置
pipeline run -c https://raw.githubusercontent.com/example/pipeline/main/pipeline.yaml
```

### 示例 3: 设置工作目录和环境变量

```bash
pipeline run \
  -c pipeline.yaml \
  -w /tmp/my-pipeline \
  -e BUILD_NUMBER=123 \
  -e GITHUB_TOKEN=xxx
```

### 示例 4: 传递环境变量

```bash
# 只传递特定的环境变量
export GITHUB_TOKEN=xxx
export CI_BUILD_NUMBER=123
pipeline run --allow-env GITHUB_TOKEN --allow-env CI_BUILD_NUMBER

# 传递所有环境变量
pipeline run --allow-all-env
```

### 示例 5: 使用自定义 Docker 镜像

```bash
pipeline run -c pipeline.yaml -i node:16-alpine
```

## 配置文件格式

配置文件必须是有效的 YAML 格式，包含 Pipeline 定义。详细格式请参考 [使用文档](./USAGE.md#3-配置文件格式)。

**最小配置示例**:

```yaml
name: My Pipeline
stages:
  - name: build
    jobs:
      - name: build-job
        steps:
          - name: build
            command: echo "Building..."
```

## 执行流程

1. **加载配置**: 从本地文件或远程 URL 加载 Pipeline 配置
2. **解析配置**: 解析 YAML 配置并验证
3. **应用选项**: 应用命令行选项（workdir、image、环境变量等）
4. **执行 Pipeline**: 按顺序执行各个 Stage
5. **清理**: 成功时清理 workdir，失败时保留 workdir 以便调试

## 错误处理

当 Pipeline 执行失败时：

- **workdir 保留**: 失败的 workdir 会被保留，方便调试
- **错误日志**: 输出详细的错误信息，包括 workdir 位置
- **状态信息**: 在 Pipeline State 中记录错误信息

详细错误处理说明请参考 [错误处理文档](./ERROR_HANDLING.md)。

## 环境变量

可以通过环境变量设置命令选项：

```bash
export PIPELINE_CONFIG=pipeline.yaml
export PIPELINE_WORKDIR=/tmp/pipeline
export PIPELINE_IMAGE=alpine:latest
pipeline run
```

## 调试模式

启用调试模式可以查看详细的执行信息：

```bash
DEBUG=1 pipeline run -c pipeline.yaml
```

调试模式下会：
- 显示 Pipeline 配置的 JSON 格式
- 保留临时下载的远程配置文件
- 输出更详细的日志信息

## 常见问题

### Q: 如何指定配置文件？

A: 使用 `-c` 选项指定配置文件路径，或将其放在当前目录并命名为 `.pipeline.yaml`。

### Q: 如何传递敏感信息（如密码）？

A: 使用 `-e` 选项传递环境变量，避免在配置文件中硬编码敏感信息。

### Q: 远程配置文件会被缓存吗？

A: 不会。每次运行都会重新下载远程配置文件。调试模式下会保留临时文件。

### Q: 如何查看 Pipeline 的执行日志？

A: Pipeline 的执行日志会输出到标准输出。如果失败，可以查看 workdir 中的详细日志文件。

### Q: 如何设置超时时间？

A: 在 Pipeline 配置文件中设置 `timeout` 字段（单位：秒）。

## 相关文档

- [使用文档](./USAGE.md)
- [错误处理文档](./ERROR_HANDLING.md)
- [架构文档](./ARCHITECTURE.md)
