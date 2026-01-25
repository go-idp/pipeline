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

## More Examples

See example files in the `examples/` directory:
- `basic.yml`: Basic example
- `docker.yaml`: Docker build example
- `github.yaml`: GitHub Actions style example
- `plugin.yml`: Plugin usage example
- `language.yml`: Language runtime example
- `step-engine-ssh.yaml`: SSH engine example
- `step-service-docker-compose.yaml`: Service orchestration example
