# Service Deployment

Pipeline can deploy services natively from raw definitions — no shell generation.
The step `service` field takes a **docker-compose.yaml or Kubernetes manifest** as
input and the engine executes it with **Go SDKs**.

Three types are supported:

| Type | Implementation | Equivalent command |
|------|----------------|--------------------|
| `docker-compose` | Compose v2 SDK (`compose.Up`) | `docker compose --project-name <name> up -d` |
| `docker-swarm` | Docker CLI stack deploy SDK | `docker stack deploy --prune --with-registry-auth -c <config> <name>` |
| `kubernetes` | client-go server-side apply | `kubectl apply -f -` |

> Before v1.8.0, `service` was implemented by generating shell commands (heredoc
> temp files + `docker-compose` / `docker stack` / `kubectl` + shell readiness
> polling). That approach is removed; use the native configuration below.

## Configuration

```yaml
steps:
  - name: deploy
    service:
      type: docker-compose          # required: docker-compose | docker-swarm | kubernetes
      name: my-services             # required: compose project / swarm stack name
      config: |                     # required: docker-compose.yaml or k8s manifest YAML (multi-doc)
        version: '3'
        services:
          web:
            image: nginx:alpine
      timeout: 120                  # optional: startup readiness wait in seconds (default 120)
      image_registry: registry.example.com   # optional: private registry address
      image_registry_username: user          # optional: registry username
      image_registry_password: pass          # optional: registry password
      namespace: default            # optional (kubernetes): namespace
      kubeconfig: /path/to/kubeconfig        # optional (kubernetes): kubeconfig path
```

### Fields

- **type**: service type. `docker-compose` / `docker-swarm` / `kubernetes`
  (`k8s` is an alias of `kubernetes`).
- **name**: compose project name / swarm stack name, used to isolate deployments,
  e.g. `eunomia-task-10086`. Not used by kubernetes.
- **config**: the service definition. docker-compose.yaml for docker-compose /
  docker-swarm; manifest YAML (multi-document, `---` separated) for kubernetes.
- **timeout**: startup readiness wait in seconds, default `120`. The deploy itself
  (image pulls, resource creation) is bounded by the step `timeout` (default 86400s).
- **image_registry / image_registry_username / image_registry_password**: private
  registry authentication, equivalent to `docker login`.
- **namespace / kubeconfig**: kubernetes only. `namespace` falls back to the
  kubeconfig context or `default`; cluster-scoped resources (Namespace, ClusterRole,
  CRD, ...) stay cluster-scoped. `kubeconfig` falls back to `$KUBECONFIG`,
  `~/.kube/config`, then in-cluster config.

## Environment Interpolation

`${VAR}` / `$VAR` references in `config` are interpolated with the **step
environment** (same behavior as `docker compose`):

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

Supported forms: `${VAR}`, `$VAR`, `${VAR:-default}`; `$$` escapes to a literal `$`.

Scalar fields like `name` / `image_registry*` / `namespace` / `kubeconfig` also
support `${VAR}` interpolation (expanded at Setup with the step environment),
e.g. `name: task-${EUNOMIA_TASK_ID}`.

> Note: compose project / swarm stack names must be **lowercase**
> (e.g. `eunomia-task-10086`); uppercase names are rejected by validation.

## Readiness Check & Failure Diagnostics

The engine waits for the service to become ready and collects diagnostics on failure:

- **docker-compose**: waits until all containers are running/healthy
  (`compose.Up` + `Wait`); on failure lists project container states.
- **docker-swarm**: polls stack service replica convergence
  (running tasks >= desired replicas); on failure collects service and failed-task logs.
  (On a docker engine ≥ 26.0 the stack deploy SDK uses `--detach=false` so the CLI
  itself waits for convergence; older engines fall back to this poll.)
- **kubernetes**: polls Deployment `readyReplicas`; on failure collects pods and events.

## Private Registry Authentication

Configured via `image_registry`, equivalent to `docker login` — no shell needed:

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

## Remote Engine (SSH / IDP)

When a service step carries a **remote engine** (`engine: ssh://user:pass@host:22`,
`idp://...`, etc.), the Go SDK cannot drive it: the docker Go client does not
support an `ssh://` endpoint, and docker's SSH connhelper rejects inline
passwords. Instead, pipeline **generates a shell command**
(`docker compose up` / `docker stack deploy` / `kubectl apply`) and runs it
**through the engine on the remote host**, exactly like a command step. A local
(`host`) engine keeps using the Go SDK with in-process readiness checks and
diagnostics.

```yaml
steps:
  - name: deploy on remote
    engine: ssh://user:pass@10.0.0.2:22   # remote engine
    service:
      type: docker-compose
      name: my-services
      config: |
        version: '3'
        services:
          web:
            image: nginx:alpine
```

Behavior differences on a remote engine:

- **CLI must exist on the remote**: the remote host needs `docker compose` (v2)
  / `docker` CLI / `kubectl` installed.
- **Readiness is best-effort**: `docker compose up -d --wait`, a swarm `--detach=false`
  deploy (only when the remote docker CLI supports it, Docker ≥ 26.0) or a swarm
  replica poll, and `kubectl wait --for=condition=Available deployment --all`.
  Failures surface as the step failing (no rich SDK diagnostics).
- **Registry auth**: `image_registry*` are forwarded to the remote via env vars
  and used with `docker login --password-stdin`, so credentials are never
  embedded in the command or log line.
- **`${VAR}` interpolation** happens on the remote (the step environment is
  forwarded to the engine session); relative paths and build contexts resolve on
  the remote host.

For concrete, runnable examples (Docker Compose / Swarm stack / Kubernetes /
build-and-deploy), see [Remote Service Deployment over SSH / IDP]
(service-deployment-remote).

## Examples

- `examples/step-service-docker-compose.yaml`: Docker Compose deployment
- `examples/step-service-docker-swarm.yaml`: Docker Swarm stack deployment
- `examples/step-service-kubernetes.yaml`: Kubernetes manifest deployment
- `examples/step-service-docker-compose-ssh.yaml`: Docker Compose on a remote SSH engine
- `examples/step-service-docker-compose-idp.yaml`: Docker Compose on a remote idp engine
- `examples/step-service-docker-swarm-ssh.yaml`: Docker Swarm on a remote SSH engine
- `examples/step-service-kubernetes-ssh.yaml`: Kubernetes on a remote SSH engine
- `examples/service-deploy.yml`: Full example (build + service deployment)

## FAQ

### Where does the SDK run?

The SDK calls run in the **pipeline process host** (the agent host that executes
`pipeline run`), same host that had `docker compose` / `kubectl` access before.
Make sure that host can reach docker / kubectl.

If the step has a **remote engine** (`engine: ssh://...`), the Go SDK is not used
— pipeline generates a command and runs it on the remote host (see
[Remote Engine](#remote-engine-ssh--idp)).

### Compatibility with old versions

v1.8.0 removed the `service.version: v1` shell generation. Migrate old configs to
the format above; **the pipeline CLI on agents must be upgraded to v1.8.0+**,
otherwise the new config fails validation.

### docker-swarm / docker-compose deploy fails with "no context store initialized"

**Symptom**: a `docker-swarm` / `docker-compose` deploy step fails with:

```
[service] docker-swarm: deploy stack "eunomia-task-154" (timeout: 120s)
[service] registry auth configured for registry.ys.zcorky.com
Failed to initialize: unable to resolve docker endpoint: no context store initialized
failed to run command: exit status 1
```

**Cause**: pipeline ≤ v1.8.4 created the docker CLI SDK without calling
`Initialize()`, so the CLI context store was never set up; `dockerCli.Client()`
failed to resolve the docker endpoint and called `os.Exit(1)`, killing the whole
process. The error is unrelated to the docker daemon, `DOCKER_HOST` or
`DOCKER_CONTEXT` — it happens on any host.

**Fix**: upgrade the pipeline CLI on agents to **v1.8.5+** (`newDockerCli` now
calls `Initialize`, resolving the current context as `default`). No config change
is needed.

> After upgrading, if the error changes to `docker context "<name>" does not
> exist`, the runner's `DOCKER_CONTEXT` points to a nonexistent context — that is
> an environment issue: create the context with `docker context create`, or unset
> the `DOCKER_CONTEXT` environment variable.
