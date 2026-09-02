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
  失败时采集服务与失败任务日志。（在 docker engine ≥ 26.0 上 stack deploy SDK 使用
  `--detach=false` 由 CLI 自身等待收敛；较旧引擎回退为该轮询。）
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

## 远端引擎（SSH / IDP）

当 service 步骤带有 **远端引擎**（`engine: ssh://user:pass@host:22`、`idp://...`
等）时，Go SDK 无法驱动它：docker 的 Go client 不支持 `ssh://` 端点，且 docker
的 SSH connhelper 拒绝明文密码。此时 pipeline **生成一段 shell 命令**
（`docker compose up` / `docker stack deploy` / `kubectl apply`）并 **通过引擎在远端
主机执行**，与普通 command 步骤完全一致。本地（`host`）引擎仍使用 Go SDK，保留进程内
的就绪检查与诊断。

```yaml
steps:
  - name: deploy on remote
    engine: ssh://user:pass@10.0.0.2:22   # 远端引擎
    service:
      type: docker-compose
      name: my-services
      config: |
        version: '3'
        services:
          web:
            image: nginx:alpine
```

远端引擎下的行为差异：

- **远端必须有 CLI**：远端主机需安装 `docker compose`（v2）/ `docker` CLI / `kubectl`。
- **就绪检查为尽力而为**：`docker compose up -d --wait`、swarm 的 `--detach=false`
  部署（仅当远端 docker CLI 支持时，Docker ≥ 26.0）或 swarm 副本轮询、以及
  `kubectl wait --for=condition=Available deployment --all`；失败以步骤失败的形式浮现
  （无 SDK 级别的富诊断）。
- **私有镜像认证**：`image_registry*` 通过环境变量转发到远端，配合
  `docker login --password-stdin` 使用，凭据不会内嵌到命令或日志行。
- **`${VAR}` 插值** 发生在远端（步骤环境会转发到引擎会话）；相对路径 / build context
  在远端主机解析。

具体可运行的例子（Docker Compose / Swarm stack / Kubernetes / 构建后再部署）见
[通过 SSH / IDP 在远端部署服务](service-deployment-remote)。

## 示例

- `examples/step-service-docker-compose.yaml`: Docker Compose 部署
- `examples/step-service-docker-swarm.yaml`: Docker Swarm stack 部署
- `examples/step-service-kubernetes.yaml`: Kubernetes manifest 部署
- `examples/step-service-docker-compose-ssh.yaml`: 在远端 SSH 引擎上 Docker Compose 部署
- `examples/step-service-docker-compose-idp.yaml`: 在远端 idp 引擎上 Docker Compose 部署
- `examples/step-service-docker-swarm-ssh.yaml`: 在远端 SSH 引擎上 Docker Swarm 部署
- `examples/step-service-kubernetes-ssh.yaml`: 在远端 SSH 引擎上 Kubernetes 部署
- `examples/service-deploy.yml`: 完整示例（构建 + 服务部署）

## 常见问题

### 执行位置

SDK 调用运行在 **pipeline 进程所在主机**（即执行 `pipeline run` 的 agent 主机），与之前
`docker compose` / `kubectl` CLI 所在主机一致。确保该主机有 docker / kubectl 访问权限。

若步骤带有 **远端引擎**（`engine: ssh://...`），则不使用 Go SDK——pipeline 生成命令并在
远端主机执行（见 [远端引擎](#远端引擎ssh--idp)）。

### 与旧版本兼容

v1.8.0 移除了 `service.version: v1` 的 shell 生成方式。旧格式的配置需要按本文档迁移；
**agent 上的 pipeline CLI 需要升级到 v1.8.0+**，否则新配置会校验失败。

### docker-swarm / docker-compose 部署报 "no context store initialized"

**现象**：`docker-swarm` / `docker-compose` 部署步骤失败，日志如下：

```
[service] docker-swarm: deploy stack "eunomia-task-154" (timeout: 120s)
[service] registry auth configured for registry.ys.zcorky.com
Failed to initialize: unable to resolve docker endpoint: no context store initialized
failed to run command: exit status 1
```

**原因**：pipeline ≤ v1.8.4 创建 docker CLI SDK 时漏掉了 `Initialize()` 调用，CLI 的
context store 未初始化，`dockerCli.Client()` 解析 docker endpoint 失败后直接
`os.Exit(1)` 终止整个进程。该错误与 docker daemon、`DOCKER_HOST`、`DOCKER_CONTEXT`
均无关，在任何主机上都会发生。

**修复**：升级 agent 上的 pipeline CLI 到 **v1.8.5+**（`newDockerCli` 补上
`Initialize` 调用，将当前 context 解析为 `default`），无需修改配置。

> 升级后若仍报错且信息变为 `docker context "<name>" does not exist`，说明运行环境的
> `DOCKER_CONTEXT` 指向了不存在的 context——这是环境问题：用 `docker context create`
> 创建该 context，或取消 `DOCKER_CONTEXT` 环境变量。
