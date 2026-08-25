# Remote Service Deployment over SSH / IDP

By default a `service` step runs the **Go SDK** on the host that executes
`pipeline run`. When you need to deploy to a **different host** — a production
server, a swarm manager, or a cluster reachable only through SSH — set a **remote
engine** on the step:

```yaml
engine: ssh://user:pass@10.0.0.2:22
```

Because the docker Go client cannot talk to an `ssh://` endpoint (and docker's
SSH connhelper rejects inline passwords), a remote-engine service step is **not**
driven by the SDK. Pipeline **generates the equivalent CLI command**
(`docker compose up` / `docker stack deploy` / `kubectl apply`) and runs it on the
**remote host** through the engine — exactly like a command step. This page shows
concrete, runnable use cases.

> **Local vs remote**: a step without an `engine:` (or with `engine: host`) uses
> the Go SDK with in-process readiness checks and rich diagnostics. A step with a
> remote engine uses the generated command instead.

## Supported engines

The generated command is **engine-agnostic**: only the `engine:` line changes,
the `service:` config (and the docker / kubectl command that pipeline runs) stays
the same. Any engine other than `host`/empty is executed as a command; `host`
(or no `engine:`) runs the Go SDK.

| `engine:` | how it deploys | what it needs |
|-----------|----------------|---------------|
| *(empty)* / `host` | Go SDK (in-process) | docker / kubectl access on the `pipeline run` host |
| `ssh://user:pass@host:22` | generated command over SSH | remote host: `docker compose` / `docker` / `kubectl` CLI |
| `idp://user:pass@host:8838` | generated command on the idp agent | the idp agent host: `docker compose` / `docker` / `kubectl` CLI |
| `docker://...` | generated command inside a container | container image with docker CLI + `/var/run/docker.sock` mounted (edge case) |

The same service under different engines:

```yaml
steps:
  - name: deploy (local SDK)
    service:
      type: docker-compose
      name: web
      config: |
        services:
          web:
            image: nginx:alpine
  - name: deploy (remote ssh)
    engine: ssh://user:pass@10.0.0.2:22
    service:
      type: docker-compose
      name: web
      config: |
        services:
          web:
            image: nginx:alpine
  - name: deploy (remote idp)
    engine: idp://user:pass@10.0.0.5:8838
    service:
      type: docker-compose
      name: web
      config: |
        services:
          web:
            image: nginx:alpine
```

Only `engine:` (and the `name`, which must be unique per deployment) changes —
the `service` block is identical. Use cases below show the `ssh:` form and call
out the idp alternative.

## What you need on the target

- For **ssh**: SSH access that `pipeline run` can authenticate to
  (`ssh://user:pass@host:22`, or key-based auth), and the docker / kubectl CLI on
  the remote host.
- For **idp**: a running idp agent (`idp://user:pass@host:8838`) whose host has
  the docker / kubectl CLI.
- For **compose**: `docker compose` (v2) installed.
- For **swarm**: `docker` CLI installed and the host is a swarm manager.
- For **kubernetes**: `kubectl` installed and configured to reach the target
  cluster (`~/.kube/config`, `$KUBECONFIG`, or the step's `kubeconfig` path).

## Example 1 — Docker Compose on a remote server

Deploy a stateless web app to a remote server and wait until it is healthy
(`compose up -d --wait`):

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
            engine: ssh://user:pass@10.0.0.2:22   # remote engine
            service:
              type: docker-compose
              name: myapp_task_${EUNOMIA_BUILD_ID} # interpolated at Setup
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

What happens on the remote host:

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

- `${VAR}` in `config` is interpolated by `docker compose` **on the remote**, using
  the step environment (which pipeline forwards to the SSH session).
- `name` is expanded at Setup (`${EUNOMIA_BUILD_ID}` → `10000`), so the compose
  project name is stable and overridable.
- `--wait` blocks until every container is running/healthy or `--wait-timeout`
  (120s) elapses.

> **Engine-agnostic**: the same step deploys over **idp** by only changing the
> engine line — `engine: idp://user:pass@10.0.0.5:8838`. A local SDK deploy is
> just the step with no `engine:` (or `engine: host`).

## Example 2 — Docker Swarm stack on a remote swarm manager

Deploy a replicated stack to a swarm manager and poll replica convergence:

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
            engine: ssh://user:pass@10.0.0.3:22   # remote swarm manager
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

What happens on the remote host:

```bash
# ... write $PIPELINE_SERVICE_FILE ...
echo "$PIPELINE_SERVICE_REGISTRY_PASS" | docker login -u "$PIPELINE_SERVICE_REGISTRY_USER" \
  --password-stdin "$PIPELINE_SERVICE_REGISTRY"
# prefer --detach=false (waits for convergence) when the engine is >= 17.05
PIPELINE_SWARM_VERSION=$(docker version --format '{{.Server.Version}}' 2>/dev/null || echo "0.0.0")
if ... version >= 17.05 ...; then
  docker stack deploy --detach=false --prune --with-registry-auth -c "$PIPELINE_SERVICE_FILE" 'myapp_task_10000'
else
  docker stack deploy --prune --with-registry-auth -c "$PIPELINE_SERVICE_FILE" 'myapp_task_10000'
  # then poll `docker service ls` until every service's running >= desired, up to 180s
fi
```

- Registry credentials are forwarded as **env vars** (`PIPELINE_SERVICE_REGISTRY*`)
  and consumed with `docker login --password-stdin`, so they never appear in the
  command or log line.
- `--with-registry-auth` lets swarm workers pull the private image.
- Readiness polls the stack label `com.docker.stack.namespace=<name>`; on timeout
  the step fails with `docker-swarm: startup timeout`.

> **Engine-agnostic**: swap the engine line for
> `engine: idp://user:pass@10.0.0.5:8838` to deploy the same stack on an idp
> agent host.

## Example 3 — Kubernetes manifest to a remote cluster

Apply an `nginx` Deployment + Service to a remote cluster and wait for readiness:

```yaml
name: myapp-k8s-deploy

stages:
  - name: deploy
    jobs:
      - name: apply manifest
        steps:
          - name: deploy
            engine: ssh://user:pass@10.0.0.4:22   # remote kubectl host
            service:
              type: kubernetes
              namespace: demo
              timeout: 120
              # kubeconfig: /root/.kube/config   # optional; else remote default
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

What happens on the remote host:

```bash
# ... write $PIPELINE_SERVICE_FILE ...
kubectl apply -f "$PIPELINE_SERVICE_FILE"
if [ "$(kubectl get deployment -n 'demo' -o name 2>/dev/null | wc -l)" -gt 0 ]; then
  kubectl wait --for=condition=Available deployment --all -n 'demo' --timeout=120s
fi
```

- `namespace` defaults to `default` when unset.
- If the step sets `kubeconfig`, pipeline exports `KUBECONFIG` on the remote before
  `kubectl apply`; otherwise the remote host's default kubectl config is used.
- Readiness waits for every Deployment in the namespace to become `Available`.
  If the namespace has no Deployment, the wait is skipped (like the SDK).

> **Engine-agnostic**: swap the engine line for
> `engine: idp://user:pass@10.0.0.5:8838` to apply the same manifest on an idp
> agent host.

## Example 4 — Build then deploy to a remote server (end to end)

Combine a build stage and a remote deploy stage in one pipeline:

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

The build stage uses the **docker** engine (runs the CLIs in a container); the
deploy step uses the **ssh** remote engine, so the previously pushed image is
pulled and started on the target server.

## Reference — what the engine generates per type

The command below is **identical for every remote engine** (`ssh` / `idp` / `docker`);
only the transport differs. A local (`host` / no `engine:`) step runs the Go SDK
instead of any of these commands.

| type | command run on the remote host | readiness |
|------|--------------------------------|-----------|
| `docker-compose` | `docker compose -f <file> -p <name> up -d --wait --wait-timeout <timeout>` | `compose up --wait` (running/healthy) |
| `docker-swarm` | `docker stack deploy --prune --with-registry-auth -c <file> <name>` (`--detach=false` when the engine is ≥ 17.05) | `--detach=false` waits for convergence; older engines poll `docker service ls` replica convergence |
| `kubernetes` | `kubectl apply -f <file>` (+ optional `export KUBECONFIG=<path>`) | `kubectl wait --for=condition=Available deployment --all` |

## Troubleshooting

- **`command not found: docker compose`** — the remote host is missing compose v2;
  install `docker compose` or upgrade the docker CLI plugin.
- **`no such host` / SSH auth failure** — verify the `engine: ssh://user:pass@host:22`
  endpoint and credentials; or use key-based auth.
- **`docker stack deploy: This node is not a swarm manager`** — the remote host is not
  a swarm manager; run `docker swarm init` (or join an existing cluster) first.
- **`kubectl: unable to load kubeconfig`** — the remote host has no valid kubeconfig;
  set the step's `kubeconfig` to the correct path on the remote.
- **`docker-swarm: startup timeout`** — the stack did not converge within `timeout`;
  check the image tag / registry auth on the swarm nodes.
- **Relative paths** — relative bind mounts and build contexts in `config` are resolved
  on the **remote** host, so make sure the referenced paths exist there.
