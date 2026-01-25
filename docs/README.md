# Pipeline 文档索引

欢迎使用 Pipeline！本文档索引帮助您快速找到所需的文档。

## 📚 文档列表

### 入门文档

- **[使用文档](./USAGE.md)** - Pipeline 的完整使用指南
  - 安装方法
  - 快速开始
  - 配置文件格式
  - 使用场景示例

### 命令文档

- **[Run 命令文档](./COMMAND_RUN.md)** - `pipeline run` 命令详细说明
  - 命令选项
  - 配置文件查找
  - 使用示例
  - 错误处理

- **[Server 命令文档](./COMMAND_SERVER.md)** - `pipeline server` 命令详细说明
  - 服务器配置
  - Web Console 使用
  - REST API 文档
  - 队列管理
  - 部署建议

- **[Client 命令文档](./COMMAND_CLIENT.md)** - `pipeline client` 命令详细说明
  - 连接服务器
  - WebSocket 通信
  - 使用示例
  - 错误处理

### 技术文档

- **[架构文档](./ARCHITECTURE.md)** - Pipeline 的架构设计
  - 系统架构
  - 组件说明
  - 数据流
  - 扩展机制

- **[错误处理文档](./ERROR_HANDLING.md)** - 错误处理机制
  - 错误类型
  - Workdir 清理策略
  - 错误日志格式
  - 最佳实践

- **[优化文档](./OPTIMIZATION.md)** - 性能优化指南
  - 性能优化建议
  - 资源管理
  - 最佳实践

## 🚀 快速开始

### 1. 本地运行 Pipeline

```bash
# 创建配置文件
cat > .pipeline.yaml <<EOF
name: My Pipeline
stages:
  - name: build
    jobs:
      - name: build-job
        steps:
          - name: hello
            command: echo "Hello, Pipeline!"
EOF

# 运行 Pipeline
pipeline run
```

**参考**: [Run 命令文档](./COMMAND_RUN.md)

### 2. 启动 Pipeline Server

```bash
# 启动服务器
pipeline server

# 访问 Web Console
open http://localhost:8080/console
```

**参考**: [Server 命令文档](./COMMAND_SERVER.md)

### 3. 使用 Client 连接 Server

```bash
# 启动服务器（在另一个终端）
pipeline server

# 使用客户端执行 Pipeline
pipeline client -c pipeline.yaml -s ws://localhost:8080
```

**参考**: [Client 命令文档](./COMMAND_CLIENT.md)

## 📖 文档导航

### 按使用场景

- **本地开发**: [使用文档](./USAGE.md) → [Run 命令文档](./COMMAND_RUN.md)
- **服务部署**: [Server 命令文档](./COMMAND_SERVER.md) → [架构文档](./ARCHITECTURE.md)
- **远程执行**: [Client 命令文档](./COMMAND_CLIENT.md) → [Server 命令文档](./COMMAND_SERVER.md)
- **问题排查**: [错误处理文档](./ERROR_HANDLING.md) → [使用文档](./USAGE.md)

### 按用户角色

- **开发者**: [使用文档](./USAGE.md) → [Run 命令文档](./COMMAND_RUN.md) → [错误处理文档](./ERROR_HANDLING.md)
- **运维人员**: [Server 命令文档](./COMMAND_SERVER.md) → [架构文档](./ARCHITECTURE.md) → [优化文档](./OPTIMIZATION.md)
- **架构师**: [架构文档](./ARCHITECTURE.md) → [优化文档](./OPTIMIZATION.md)

## 🔍 常见问题

### 如何选择运行方式？

- **本地运行** (`pipeline run`): 适合本地开发和测试
- **Server 模式** (`pipeline server`): 适合生产环境，提供 Web Console 和 API
- **Client 模式** (`pipeline client`): 适合 CI/CD 集成，远程执行 Pipeline

### 如何配置 Pipeline？

参考 [使用文档 - 配置文件格式](./USAGE.md#3-配置文件格式)

### Pipeline 失败后如何调试？

参考 [错误处理文档](./ERROR_HANDLING.md)

### 如何优化 Pipeline 性能？

参考 [优化文档](./OPTIMIZATION.md)

## 📝 文档贡献

如果您发现文档有错误或需要改进，欢迎提交 Issue 或 Pull Request。

## 🔗 相关资源

- GitHub: https://github.com/go-idp/pipeline
- 示例配置: `examples/` 目录
- 单元测试: `*_test.go` 文件
