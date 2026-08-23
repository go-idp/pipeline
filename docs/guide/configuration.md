# Configuration

Pipeline uses YAML format configuration files to define workflows. This document details the configuration file format and options.

## Basic Structure

```yaml
name: Pipeline Name                    # Required: Pipeline name

workdir: /tmp/my-pipeline             # Optional: Working directory, default current directory

image: alpine:latest                   # Optional: Default Docker image

timeout: 3600                         # Optional: Timeout in seconds, default 86400

environment:                           # Optional: Environment variables
  KEY1: value1
  KEY2: value2

pre: echo "pre hook"                  # Optional: Pre hook

post: echo "post hook"                # Optional: Post hook

stages:                                # Required: Stage list
  - name: stage1
    jobs:
      - name: job1
        steps:
          - name: step1
            command: echo "step1"
```

## Configuration Inheritance

Configuration is inherited in the following hierarchy: **Pipeline → Stage → Job → Step**

Each level inherits configuration from its parent, with child-level configuration taking higher priority.

### Inheritable Configuration Items

- **Working Directory** (`workdir`): Pipeline → Stage → Job → Step
- **Docker Image** (`image`): Pipeline → Stage → Job → Step
- **Timeout** (`timeout`): Pipeline → Stage → Job → Step
- **Environment Variables** (`environment`): Pipeline → Stage → Job → Step
- **Image Registry Configuration** (`image_registry`, `image_registry_username`, `image_registry_password`): Job → Step

## Service (native deploy)

A step can deploy a service natively with Go SDKs (no shell generation):

```yaml
steps:
  - name: deploy
    service:
      type: docker-compose          # docker-compose | docker-swarm | kubernetes
      name: my-services             # compose project / swarm stack name
      config: |                     # docker-compose.yaml or k8s manifest YAML (multi-doc)
        version: '3'
        services:
          db:
            image: postgres:13
      timeout: 120                  # startup readiness wait in seconds (default 120)
      image_registry: registry.example.com   # optional: private registry auth (docker login equivalent)
      image_registry_username: user
      image_registry_password: pass
      namespace: default            # kubernetes only
      kubeconfig: /path/to/kubeconfig        # kubernetes only
```

- `docker-compose`: Compose v2 SDK (`docker compose --project-name <name> up -d` equivalent)
- `docker-swarm`: Docker CLI stack deploy SDK (`docker stack deploy` equivalent)
- `kubernetes`: client-go server-side apply (`kubectl apply -f -` equivalent)

`${VAR}` references in `config` are interpolated with the step environment. The SDK
runs in the pipeline process (the agent host that executes `pipeline run`), which is
the same host that has docker / kubectl access. The legacy `version: v1` shell-based
service mode is removed.

## More Examples

See example files in the `examples/` directory:
- `basic.yml`: Basic example
- `docker.yaml`: Docker build example
- `github.yaml`: GitHub Actions style example
- `plugin.yml`: Plugin usage example
- `language.yml`: Language runtime example
- `step-engine-ssh.yaml`: SSH engine example
- `step-service-docker-compose.yaml`: Docker Compose service deployment example
- `step-service-docker-swarm.yaml`: Docker Swarm service deployment example
- `step-service-kubernetes.yaml`: Kubernetes service deployment example
- `service-deploy.yml`: Full example (build + service deployment)
