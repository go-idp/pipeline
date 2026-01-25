# 最佳实践

本文档总结了使用 Pipeline 的最佳实践和建议。

## 配置文件组织

### 1. 使用版本控制

将 Pipeline 配置文件纳入版本控制：

```bash
git add .pipeline.yaml
git commit -m "Add pipeline configuration"
```

### 2. 环境分离

为不同环境创建不同的配置文件：

```
.pipeline.dev.yaml
.pipeline.staging.yaml
.pipeline.prod.yaml
```

### 3. 配置模板化

使用模板化配置，避免重复：

```yaml
# base.yaml
name: ${PIPELINE_NAME}
stages:
  - name: build
    jobs:
      - name: build-job
        steps:
          - name: build
            command: ${BUILD_COMMAND}
```

## 环境变量管理

### 1. 使用环境变量

使用环境变量而不是硬编码：

```yaml
environment:
  BUILD_ID: "123"
  BRANCH: "main"
```

### 2. 敏感信息保护

敏感信息通过环境变量传递，不要硬编码在配置文件中：

```bash
pipeline run -e GITHUB_TOKEN=xxx
```

### 3. 环境变量命名

使用清晰的命名约定：

```yaml
environment:
  PIPELINE_BUILD_ID: "123"
  PIPELINE_BRANCH: "main"
```

## 错误处理

### 1. 设置合理的超时时间

根据实际需要设置超时时间：

```yaml
timeout: 3600  # 1 小时
```

### 2. 检查失败后的 workdir

失败后检查 workdir 中的文件：

```bash
ls -la /tmp/pipeline/abc123
cat /tmp/pipeline/abc123/*.log
```

### 3. 使用 Post 钩子清理

在 Post 钩子中添加清理逻辑：

```yaml
post: |
  echo "Cleaning up..."
  rm -rf /tmp/build-artifacts
```

## 性能优化

### 1. 使用并行执行

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

### 2. 优化 Docker 镜像

- 使用较小的基础镜像
- 复用 Docker 镜像层
- 使用多阶段构建

### 3. 合理设置并发数

根据服务器资源设置合理的并发数：

```bash
pipeline server --max-concurrent 5
```

## 安全实践

### 1. 启用认证

在生产环境中启用认证：

```bash
pipeline server -u admin --password secret123
```

### 2. 使用 HTTPS

使用反向代理提供 HTTPS：

```nginx
server {
    listen 443 ssl;
    server_name pipeline.example.com;
    
    location / {
        proxy_pass http://localhost:8080;
    }
}
```

### 3. 限制环境变量

只传递必要的环境变量：

```bash
pipeline run --allow-env GITHUB_TOKEN
```

## 监控和日志

### 1. 配置日志收集

配置日志收集系统：

```bash
pipeline server 2>&1 | tee pipeline.log
```

### 2. 监控 Pipeline 执行

使用 Web Console 监控 Pipeline 执行：

```bash
open http://localhost:8080/console
```

### 3. 设置告警

设置告警机制，及时发现问题。

## 部署建议

### 1. 使用 Docker

使用 Docker 部署：

```bash
docker run -d \
  -p 8080:8080 \
  -v /var/lib/pipeline:/data \
  ghcr.io/go-idp/pipeline:latest server
```

### 2. 使用反向代理

使用 Nginx 或 Traefik 作为反向代理。

### 3. 持久化存储

使用持久化存储作为工作目录：

```bash
pipeline server -w /var/lib/pipeline
```

## 开发建议

### 1. 本地测试

在本地测试 Pipeline：

```bash
pipeline run -c pipeline.yaml
```

### 2. 使用示例

参考示例配置：

```bash
pipeline run -c examples/basic.yml
```

### 3. 调试模式

使用调试模式查看详细信息：

```bash
DEBUG=1 pipeline run -c pipeline.yaml
```

## 常见问题

### 1. 配置文件找不到

确保配置文件存在，或使用 `-c` 选项指定：

```bash
pipeline run -c pipeline.yaml
```

### 2. 环境变量未传递

使用 `--allow-env` 选项：

```bash
pipeline run --allow-env GITHUB_TOKEN
```

### 3. Docker 命令失败

确保 Docker 已安装并运行：

```bash
docker ps
```

## 总结

遵循这些最佳实践可以帮助您更好地使用 Pipeline：

1. **配置文件组织**: 使用版本控制，环境分离
2. **环境变量管理**: 使用环境变量，保护敏感信息
3. **错误处理**: 设置合理的超时时间，检查失败后的 workdir
4. **性能优化**: 使用并行执行，优化 Docker 镜像
5. **安全实践**: 启用认证，使用 HTTPS
6. **监控和日志**: 配置日志收集，监控 Pipeline 执行
7. **部署建议**: 使用 Docker，使用反向代理
