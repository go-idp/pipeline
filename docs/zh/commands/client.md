# client 命令

`pipeline client` 命令用于连接到 Pipeline Server 并通过 WebSocket 执行 Pipeline。客户端会将 Pipeline 配置发送到服务器，并实时接收执行日志和结果。

## 基本用法

```bash
pipeline client [选项]
```

## 命令选项

### `-c, --config`

指定 Pipeline 配置文件路径（必需）。

- **类型**: 字符串
- **环境变量**: `CONFIG`
- **必需**: 是
- **支持格式**:
  - 本地文件路径: `pipeline.yaml`
  - HTTP/HTTPS URL: `https://example.com/pipeline.yaml`

**示例**:

```bash
pipeline client -c pipeline.yaml -s ws://localhost:8080
```

### `-s, --server`

指定 Pipeline Server 地址（必需）。

- **类型**: 字符串
- **环境变量**: `SERVER`
- **必需**: 是
- **格式**: `ws://host:port` 或 `wss://host:port`

**示例**:

```bash
pipeline client -c pipeline.yaml -s ws://localhost:8080
pipeline client -c pipeline.yaml -s wss://pipeline.example.com
```

### `-u, --username`

设置 HTTP Basic Auth 用户名。

- **类型**: 字符串
- **环境变量**: `USERNAME`
- **说明**: 如果服务器启用了认证，需要提供用户名

**示例**:

```bash
pipeline client -c pipeline.yaml -s ws://localhost:8080 -u admin
```

### `-p, --password`

设置 HTTP Basic Auth 密码。

- **类型**: 字符串
- **环境变量**: `PASSWORD`
- **说明**: 如果服务器启用了认证，需要提供密码

**示例**:

```bash
pipeline client -c pipeline.yaml -s ws://localhost:8080 -u admin -p password123
```

### `--path`

指定服务器路径。

- **类型**: 字符串
- **环境变量**: `SERVER_PATH`
- **默认值**: `/`
- **说明**: 如果服务器使用了自定义路径，需要指定此选项

**示例**:

```bash
pipeline client -c pipeline.yaml -s ws://localhost:8080 --path /api/pipeline
```

## 工作流程

1. **加载配置**: 从本地文件或远程 URL 加载 Pipeline 配置
2. **连接服务器**: 通过 WebSocket 连接到 Pipeline Server
3. **发送配置**: 将 Pipeline 配置发送到服务器
4. **接收日志**: 实时接收 Pipeline 执行日志（stdout/stderr）
5. **等待完成**: 等待 Pipeline 执行完成或失败
6. **关闭连接**: 关闭 WebSocket 连接

## 使用示例

### 示例 1: 基本使用

```bash
# 启动服务器（在另一个终端）
pipeline server

# 使用客户端执行 Pipeline
pipeline client \
  -c pipeline.yaml \
  -s ws://localhost:8080
```

### 示例 2: 使用认证

```bash
pipeline client \
  -c pipeline.yaml \
  -s ws://localhost:8080 \
  -u admin \
  -p password123
```

### 示例 3: 使用远程配置

```bash
pipeline client \
  -c https://example.com/pipeline.yaml \
  -s ws://pipeline.example.com
```

### 示例 4: 使用自定义路径

```bash
pipeline client \
  -c pipeline.yaml \
  -s ws://localhost:8080 \
  --path /api/pipeline
```

### 示例 5: 使用环境变量

```bash
export CONFIG=pipeline.yaml
export SERVER=ws://localhost:8080
export USERNAME=admin
export PASSWORD=password123
pipeline client
```

## 输出格式

客户端会实时输出 Pipeline 的执行日志：

- **标准输出**: Pipeline 的 stdout 输出
- **标准错误**: Pipeline 的 stderr 输出
- **执行结果**: Pipeline 执行完成或失败的信息

**示例输出**:

```
[workflow] start
[workflow] version: 1.7.1
[workflow] name: My Pipeline
[stage(1/2): build] start
[job(1/1): build-job] start
[step(1/1): compile] start
Compiling...
[step(1/1): compile] done
[job(1/1): build-job] done
[stage(1/2): build] done
[workflow] done
```

## 错误处理

### 连接错误

如果无法连接到服务器，客户端会输出错误信息：

```
failed to connect to server: connection refused
```

### 认证错误

如果认证失败，客户端会输出错误信息：

```
authentication failed
```

### Pipeline 执行错误

如果 Pipeline 执行失败，客户端会输出错误信息并退出，退出码为非零。

## 常见问题

### Q: 如何连接到远程服务器？

A: 使用 `-s` 选项指定服务器地址，格式为 `ws://host:port` 或 `wss://host:port`。

### Q: 如何提供认证信息？

A: 使用 `-u` 和 `-p` 选项提供用户名和密码，或使用环境变量 `USERNAME` 和 `PASSWORD`。

### Q: 如何查看执行日志？

A: 客户端会实时输出执行日志到标准输出和标准错误。

### Q: 如何判断 Pipeline 是否成功？

A: 查看客户端退出码，0 表示成功，非零表示失败。
