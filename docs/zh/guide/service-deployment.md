# 服务部署（Service）

Pipeline 支持直接以原生定义部署服务，无需生成 shell 脚本。Step 的 `service` 字段把
**docker-compose.yaml 或 Kubernetes manifest** 作为输入，由引擎使用 **Go SDK** 原生执行。

支持三种类型：

| 类型 | 实现 | 等价命令 |
|------|------|---------|
| `docker-compose` | Compose v2 SDK（`compose.Up`） | `docker compose --project-name <name> up -d` |
| `docker-swarm` | Docker CLI stack deploy SDK | `docker stack deploy --prune --with-registry-auth -c <config> <name>` |
| `kubernetes` | client-go server-side apply | `kubectl apply -f -` |

> 在 v1.8.0 之前，`service` 通过生成 shell 命令（heredoc 写临时文件 + `docker-compose` /
> `docker stack` / `kubectl` + shell 轮询就绪）实现。该方式已废弃移除，请使用新的原生配置。

## 配置结构

```yaml
steps:
  - name: deploy
    service:
      type: docker-compose          # 必填：docker-compose | docker-swarm | kubernetes
      name: my-services             # 必填：compose project 名 / swarm stack 名
      config: |                     # 必填：docker-compose.yaml 或 k8s manifest YAML（支持多文档）
        version: '3'
        services:
          web:
            image: nginx:alpine
      timeout: 120                  # 可选：启动就绪等待秒数，默认 120
      image_registry: registry.example.com   # 可选：私有镜像仓库地址
      image_registry_username: user          # 可选：镜像仓库用户名
      image_registry_password: pass          # 可选：镜像仓库密码
      namespace: default            # 可选（kubernetes）：命名空间
      kubeconfig: /path/to/kubeconfig        # 可选（kubernetes）：kubeconfig 路径
```

### 字段说明

- **type**: 服务类型，`docker-compose` / `docker-swarm` / `kubernetes`（`k8s` 是 `kubernetes` 的别名）。
- **name**: compose 的 project 名 / swarm 的 stack 名，用于隔离多个部署。例如
  `eunomia-task-10086`。kubernetes 不需要（由 manifest 中的 namespace 决定）。
- **config**: 服务定义内容。docker-compose / docker-swarm 传 docker-compose.yaml；
  kubernetes 传 manifest YAML，**支持多文档**（`---` 分隔）。
- **timeout**: 启动就绪等待的秒数，默认 `120`。部署本身（镜像拉取、创建资源）由 step 的
  `timeout` 限制（默认 86400 秒）。
- **image_registry / image_registry_username / image_registry_password**: 私有镜像仓库认证，
  等价于执行 `docker login`。拉取私有镜像时配置。
- **namespace / kubeconfig**: 仅 kubernetes 使用。`namespace` 缺省时取 kubeconfig 上下文或
  `default`；集群级资源（Namespace、ClusterRole、CRD 等）保持集群作用域。`kubeconfig` 缺省时
  依次尝试 `$KUBECONFIG`、`~/.kube/config`、集群内配置（in-cluster）。

## 环境变量插值

`config` 中的 `${VAR}` / `$VAR` 引用使用 **step 的环境变量**插值（与 `docker compose` 行为一致）：

```yaml
environment:
  EUNOMIA_TASK_ID: "10086"

steps:
  - name: deploy
    service:
      type: docker-compose
      name: task-${EUNOMIA_TASK_ID}
      config: |
        version: '3'
        services:
          web:
            image: registry.example.com/nginx:alpine
            environment:
              BUILD_ID: ${EUNOMIA_BUILD_ID}
```

支持 `${VAR}`、`$VAR`、`${VAR:-default}` 形式；`$$` 转义为字面量 `$`。

`name` / `image_registry*` / `namespace` / `kubeconfig` 等标量字段同样支持 `${VAR}` 插值
（在 Setup 阶段用 step 环境展开），例如 `name: task-${EUNOMIA_TASK_ID}`。

> 注意：compose 的 project 名 / swarm 的 stack 名需要**小写**（如 `eunomia-task-10086`），
> 大写名称会被校验拒绝。

## 就绪检查与失败诊断

引擎在部署完成后自动等待服务就绪，失败时采集诊断信息：

- **docker-compose**: 等待所有容器进入 running/healthy 状态（`compose.Up` + `Wait`）；
  失败时列出项目容器状态。
- **docker-swarm**: 轮询 stack 服务的副本收敛（running tasks >= desired replicas）；
  失败时采集服务与失败任务日志。
- **kubernetes**: 轮询 Deployment 的 `readyReplicas`；失败时采集 pods 与 events。

## 私有镜像仓库认证

通过 `image_registry` 字段配置，等价于 `docker login`，不再需要在流水线里写 shell：

```yaml
service:
  type: docker-compose
  name: my-services
  image_registry: registry.example.com
  image_registry_username: ${REGISTRY_USERNAME}
  image_registry_password: ${REGISTRY_PASSWORD}
  config: |
    version: '3'
    services:
      web:
        image: registry.example.com/myapp:latest
```

## 示例

- `examples/step-service-docker-compose.yaml`: Docker Compose 部署
- `examples/step-service-docker-swarm.yaml`: Docker Swarm stack 部署
- `examples/step-service-kubernetes.yaml`: Kubernetes manifest 部署
- `examples/service-deploy.yml`: 完整示例（构建 + 服务部署）

## 常见问题

### 执行位置

SDK 调用运行在 **pipeline 进程所在主机**（即执行 `pipeline run` 的 agent 主机），与之前
`docker compose` / `kubectl` CLI 所在主机一致。确保该主机有 docker / kubectl 访问权限。

### 与旧版本兼容

v1.8.0 移除了 `service.version: v1` 的 shell 生成方式。旧格式的配置需要按本文档迁移；
**agent 上的 pipeline CLI 需要升级到 v1.8.0+**，否则新配置会校验失败。
