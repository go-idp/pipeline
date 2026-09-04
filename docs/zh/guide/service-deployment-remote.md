# 通过 SSH / IDP 在远端部署服务

默认情况下，`service` 步骤会**执行 `pipeline run` 的那台主机**上通过 **Go SDK** 部署。
当你需要部署到**另一台主机** —— 生产服务器、swarm manager，或只有通过 SSH 才能触达的集群 ——
请在步骤上设置一个**远端引擎**：

```yaml
engine: ssh://user:pass@10.0.0.2:22
```

由于 docker 的 Go client 无法连接 `ssh://` 端点（docker 的 SSH connhelper 也拒绝明文密码），
带远程 engine 的 service 步骤**不会走 SDK**。pipeline 会**生成对应的 CLI 命令**
（`docker compose up` / `docker stack deploy` / `kubectl apply`），并通过引擎在**远端主机**上执行 ——
与普通 command 步骤完全一致。本文展示具体、可直接运行的用例。

> **本地 vs 远端**：没有 `engine:`（或 `engine: host`）的步骤用 Go SDK，带进程内就绪检查
> 与富诊断；带远端 engine 的步骤走生成的命令。

## 支持的引擎

生成的命令是**引擎无关**的：只有 `engine:` 这一行会变，`service:` 配置（以及 pipeline 跑的
docker / kubectl 命令）保持不变。任何非 `host`/空的引擎都按命令执行；`host`（或无 `engine:`）
走 Go SDK。

| `engine:` | 部署方式 | 需要什么 |
|-----------|----------|----------|
| *(空)* / `host` | Go SDK（进程内） | 运行 `pipeline run` 的主机能访问 docker / kubectl |
| `ssh://user:pass@host:22` | 生成的命令走 SSH | 远端主机有 `docker compose` / `docker` / `kubectl` CLI |
| `idp://user:pass@host:8838` | 生成的命令在 idp agent 执行 | idp agent 主机上有 `docker compose` / `docker` / `kubectl` CLI |
| `docker://...` | 生成的命令在容器里执行 | 容器镜像含 docker CLI + 挂载 `/var/run/docker.sock`（边缘场景） |

同一个 service 在不同引擎下：

```yaml
steps:
  - name: deploy (本地 SDK)
    service:
      type: docker-compose
      name: web
      config: |
        services:
          web:
            image: nginx:alpine
  - name: deploy (远端 ssh)
    engine: ssh://user:pass@10.0.0.2:22
    service:
      type: docker-compose
      name: web
      config: |
        services:
          web:
            image: nginx:alpine
  - name: deploy (远端 idp)
    engine: idp://user:pass@10.0.0.5:8838
    service:
      type: docker-compose
      name: web
      config: |
        services:
          web:
            image: nginx:alpine
```

只有 `engine:`（以及每次部署需唯一的 `name`）变化 —— `service` 块完全一样。下面的用例以
`ssh:` 形式展示，并标注 idp 替代写法。

## 目标端需要什么

- **ssh**：`pipeline run` 可认证的 SSH 访问（`ssh://user:pass@host:22`，或基于 key），且远端主机
  有 docker / kubectl CLI。
- **idp**：运行中的 idp agent（`idp://user:pass@host:8838`），其主机有 docker / kubectl CLI。
- **compose**：安装 `docker compose`（v2）。
- **swarm**：安装 `docker` CLI，且该主机是 swarm manager。
- **kubernetes**：安装 `kubectl`，并能访问目标集群
  （`~/.kube/config`、`$KUBECONFIG`，或步骤里的 `kubeconfig` 路径）。

## 用例 1 —— 在远端服务器上部署 Docker Compose

把一个无状态 Web 应用部署到远端服务器，并等待其健康（`compose up -d --wait`）：

```yaml
name: myapp-deploy

environment:
  CI: "true"
  EUNOMIA_BUILD_ID: "10000"
  EUNOMIA_BUILD_TIMESTAMP: "1727070237"
  EUNOMIA_REGISTRY: registry.example.com

stages:
  - name: deploy
    jobs:
      - name: deploy web
        steps:
          - name: deploy
            engine: ssh://user:pass@10.0.0.2:22   # 远端引擎
            service:
              type: docker-compose
              name: myapp_task_${EUNOMIA_BUILD_ID} # Setup 时插值
              timeout: 120
              config: |
                version: '3'
                services:
                  web:
                    image: ${EUNOMIA_REGISTRY}/myapp:build_${EUNOMIA_BUILD_ID}
                    ports:
                      - "8080:80"
                    environment:
                      BUILD_TIMESTAMP: $EUNOMIA_BUILD_TIMESTAMP
                    restart: unless-stopped
```

远端主机实际执行：

```bash
set -e
export PIPELINE_SERVICE_FILE=$(mktemp)
trap 'rm -f "$PIPELINE_SERVICE_FILE"' EXIT
cat > "$PIPELINE_SERVICE_FILE" <<'PIPELINE_SERVICE_EOF'
version: '3'
services:
  web:
    ...
PIPELINE_SERVICE_EOF
docker compose -f "$PIPELINE_SERVICE_FILE" -p 'myapp_task_10000' up -d --wait --wait-timeout 120
```

- `config` 里的 `${VAR}` 由**远端**的 `docker compose` 用步骤环境插值（pipeline 会把步骤环境
  转发到 SSH 会话）。
- `name` 在 Setup 时展开（`${EUNOMIA_BUILD_ID}` → `10000`），因此 compose 项目名稳定、可覆盖。
- `--wait` 会阻塞到所有容器 running/healthy，或到 `--wait-timeout`（120 秒）。

> **引擎无关**：把 engine 一行改成 `engine: idp://user:pass@10.0.0.5:8838` 即可在 idp agent
> 主机上部署同一步骤；本地 SDK 部署只需去掉 `engine:`（或 `engine: host`）。

## 用例 2 —— 在远端 swarm manager 上部署 Swarm stack

把带副本的 stack 部署到 swarm manager，并轮询副本收敛：

```yaml
name: myapp-swarm-deploy

environment:
  CI: "true"
  EUNOMIA_BUILD_ID: "10000"
  EUNOMIA_REGISTRY: registry.example.com
  REGISTRY_USERNAME: deploy-user
  REGISTRY_PASSWORD: s3cret

stages:
  - name: deploy
    jobs:
      - name: deploy stack
        steps:
          - name: deploy
            engine: ssh://user:pass@10.0.0.3:22   # 远端 swarm manager
            service:
              type: docker-swarm
              name: myapp_task_${EUNOMIA_BUILD_ID}
              timeout: 180
              image_registry: ${EUNOMIA_REGISTRY}
              image_registry_username: ${REGISTRY_USERNAME}
              image_registry_password: ${REGISTRY_PASSWORD}
              config: |
                version: '3'
                services:
                  web:
                    image: ${EUNOMIA_REGISTRY}/myapp:build_${EUNOMIA_BUILD_ID}
                    deploy:
                      replicas: 2
                      update_config:
                        order: start-first
                      restart_policy:
                        condition: on-failure
                    ports:
                      - target: 80
                        published: 8080
                        mode: ingress
```

远端主机实际执行：

```bash
# ... 写入 $PIPELINE_SERVICE_FILE ...
echo "$PIPELINE_SERVICE_REGISTRY_PASS" | docker login -u "$PIPELINE_SERVICE_REGISTRY_USER" \
  --password-stdin "$PIPELINE_SERVICE_REGISTRY"
# 当远端 docker CLI 支持时才用 --detach=false（自身等待收敛）。
# `docker stack deploy --detach` 是 Docker Engine 26.0.0 才引入的，pipeline 用
# `docker stack deploy --help` 探测，而不用硬编码版本号。
if docker stack deploy --help 2>/dev/null | grep -q -- '--detach'; then
  docker stack deploy --detach=false --prune --with-registry-auth -c "$PIPELINE_SERVICE_FILE" 'myapp_task_10000'
else
  docker stack deploy --prune --with-registry-auth -c "$PIPELINE_SERVICE_FILE" 'myapp_task_10000'
  # 然后轮询 `docker service ls`，直到每个服务的 running >= desired，上限 180 秒
fi
```

- 仓库凭据以**环境变量**（`PIPELINE_SERVICE_REGISTRY*`）转发，配合 `docker login --password-stdin`
  使用，因此**不会出现在命令或日志行**。
- `--with-registry-auth` 让 swarm worker 能拉取私有镜像。
- `--detach=false` 仅在远端 docker CLI 支持时使用（Docker Engine ≥ 26.0，通过
  `docker stack deploy --help` 探测）；较旧引擎则回退为分离式部署，由 pipeline 轮询副本收敛。
- 就绪检查按 stack label `com.docker.stack.namespace=<name>` 轮询；超时则步骤失败并提示
  `docker-swarm: startup timeout`。

> **引擎无关**：把 engine 一行改成 `engine: idp://user:pass@10.0.0.5:8838` 即可在 idp agent
> 主机上部署同一 stack。

## 用例 3 —— 在远端集群应用 Kubernetes manifest

应用一份 `nginx` Deployment + Service 到远端集群，并等待就绪：

```yaml
name: myapp-k8s-deploy

stages:
  - name: deploy
    jobs:
      - name: apply manifest
        steps:
          - name: deploy
            engine: ssh://user:pass@10.0.0.4:22   # 远端 kubectl 主机
            service:
              type: kubernetes
              namespace: demo
              timeout: 120
              # kubeconfig: /root/.kube/config   # 可选；缺省用远端默认 kubectl 配置
              config: |
                apiVersion: v1
                kind: Namespace
                metadata:
                  name: demo
                ---
                apiVersion: apps/v1
                kind: Deployment
                metadata:
                  name: web
                  namespace: demo
                spec:
                  replicas: 2
                  selector:
                    matchLabels:
                      app: web
                  template:
                    metadata:
                      labels:
                        app: web
                    spec:
                      containers:
                        - name: web
                          image: nginx:alpine
                          ports:
                            - containerPort: 80
                ---
                apiVersion: v1
                kind: Service
                metadata:
                  name: web
                  namespace: demo
                spec:
                  selector:
                    app: web
                  ports:
                    - port: 80
                      targetPort: 80
```

远端主机实际执行：

```bash
# ... 写入 $PIPELINE_SERVICE_FILE ...
kubectl apply -f "$PIPELINE_SERVICE_FILE"
if [ "$(kubectl get deployment -n 'demo' -o name 2>/dev/null | wc -l)" -gt 0 ]; then
  kubectl wait --for=condition=Available deployment --all -n 'demo' --timeout=120s
fi
```

- `namespace` 缺省为 `default`。
- 若步骤设置了 `kubeconfig`，pipeline 会在 `kubectl apply` 之前于远端 `export KUBECONFIG`；
  否则用远端主机默认的 kubectl 配置。
- 就绪检查等待命名空间内所有 Deployment 变为 `Available`。命名空间内没有 Deployment 时跳过等待
  （与 SDK 行为一致）。

> **引擎无关**：把 engine 一行改成 `engine: idp://user:pass@10.0.0.5:8838` 即可在 idp agent
> 主机上应用同一 manifest。

## 用例 4 —— 构建后再部署到远端（端到端）

在一条 pipeline 里组合构建与远端部署：

```yaml
name: build-and-deploy

environment:
  CI: "true"
  EUNOMIA_BUILD_ID: "10000"
  EUNOMIA_REGISTRY: registry.example.com

stages:
  - name: build
    jobs:
      - name: build image
        steps:
          - name: build
            image: docker:latest
            command: |
              docker build -t ${EUNOMIA_REGISTRY}/myapp:build_${EUNOMIA_BUILD_ID} .
              docker push ${EUNOMIA_REGISTRY}/myapp:build_${EUNOMIA_BUILD_ID}

  - name: deploy
    jobs:
      - name: deploy on remote
        steps:
          - name: deploy
            engine: ssh://user:pass@10.0.0.2:22
            service:
              type: docker-compose
              name: myapp_task_${EUNOMIA_BUILD_ID}
              config: |
                version: '3'
                services:
                  web:
                    image: ${EUNOMIA_REGISTRY}/myapp:build_${EUNOMIA_BUILD_ID}
                    ports:
                      - "8080:80"
                    restart: unless-stopped
```

构建阶段用 **docker** 引擎（在容器里跑 CLI）；部署步骤用 **ssh** 远端引擎，因此刚推送的镜像
会在目标服务器上被拉取并启动。

## 每种类型生成的命令

下面的命令对**所有远端 engine**（`ssh` / `idp` / `docker`）都**完全相同**，仅传输方式不同。
本地（`host` / 无 `engine:`）步骤走 Go SDK，不执行这些命令。

| type | 远端主机执行的命令 | 就绪检查 |
|------|--------------------|----------|
| `docker-compose` | `docker compose -f <file> -p <name> up -d --wait --wait-timeout <timeout>` | `compose up --wait`（running/healthy） |
| `docker-swarm` | `docker stack deploy --prune --with-registry-auth -c <file> <name>`（仅当 docker CLI 支持时才加 `--detach=false`，Docker ≥ 26.0） | `--detach=false` 由 CLI 等待收敛；旧版回退为轮询 `docker service ls` 副本收敛 |
| `kubernetes` | `kubectl apply -f <file>`（可选 `export KUBECONFIG=<path>`） | `kubectl wait --for=condition=Available deployment --all` |

## 故障排查

- **`command not found: docker compose`** —— 远端缺 compose v2；安装 `docker compose` 或升级
  docker CLI 插件。
- **`no such host` / SSH 认证失败** —— 核对 `engine: ssh://user:pass@host:22` 端点与凭据；
  或改用 key 认证。
- **`docker stack deploy: This node is not a swarm manager`** —— 远端主机不是 swarm manager；
  先 `docker swarm init`（或加入已有集群）。
- **`unknown flag: --detach`** —— 远端主机 docker CLI 低于 26.0（`docker stack deploy --detach`
  是 Docker Engine 26.0.0 才引入的）。升级远端 docker，或保持 pipeline 现状（会自动回退为
  分离式部署 + 副本收敛轮询）。
- **`error saving credentials: mkdir /root: read-only file system`** —— 远端主机 `HOME`
  只读（如 macOS agent 的 `HOME` 是 `/root`），`docker login` 无法写 `~/.docker/config.json`。
  pipeline 会在登录前把 `DOCKER_CONFIG` 重定向到可写临时目录，并把 docker 用户
  `~/.docker/cli-plugins` 里的 CLI 插件软链进该目录，使 Compose 插件仍可被发现（否则
  `docker compose -f ...` 会报 `unknown shorthand flag: 'f' in -f`）。无需额外处理；保持最新的
  pipeline CLI。
- **`docker: command not found`** —— 远端主机 PATH 里没有 Homebrew 的 bin 目录。macOS 上
  docker 装在 `/opt/homebrew/bin`（Apple Silicon）或 `/usr/local/bin`（Intel），都不在非登录
  `/bin/sh` 的默认 PATH 里。pipeline 会在 docker/kubectl 不可用时把这些目录前置到 PATH，故无需
  额外处理。
- **`failed to connect to the docker API at unix:///var/run/docker.sock`** —— 远端主机是
  macOS，daemon socket 不在 `/var/run/docker.sock`；Docker Desktop / OrbStack / Colima 把
  socket 放在用户主目录下（`~/.docker/run/docker.sock`、`~/.orbstack/run/docker.sock`、
  `~/.colima/<profile>/docker.sock`）。pipeline 会探测这些路径（并兜底取 macOS 登录用户的家目录）
  并在跑 docker 前设置 `DOCKER_HOST`，故无需额外处理；请确保 agent 主机的 pipeline CLI 是最新的。
- **`kubectl: unable to load kubeconfig`** —— 远端主机没有有效 kubeconfig；把步骤的
  `kubeconfig` 指到远端正确路径。
- **`docker-swarm: startup timeout`** —— stack 未在 `timeout` 内收敛；检查 swarm 节点上的
  镜像 tag / 仓库认证。
- **相对路径** —— `config` 里的相对 bind mount 与 build context 会在**远端**主机解析，
  请确保路径在那里存在。
