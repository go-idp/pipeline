# Best Practices

This document summarizes best practices and recommendations for using Pipeline.

## Configuration File Organization

### 1. Use Version Control

Include Pipeline configuration files in version control:

```bash
git add .pipeline.yaml
git commit -m "Add pipeline configuration"
```

### 2. Environment Separation

Create different configuration files for different environments:

```
.pipeline.dev.yaml
.pipeline.staging.yaml
.pipeline.prod.yaml
```

### 3. Configuration Templating

Use templated configurations to avoid duplication:

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

## Environment Variable Management

### 1. Use Environment Variables

Use environment variables instead of hardcoding:

```yaml
environment:
  BUILD_ID: "123"
  BRANCH: "main"
```

### 2. Protect Sensitive Information

Pass sensitive information via environment variables, don't hardcode in configuration files:

```bash
pipeline run -e GITHUB_TOKEN=xxx
```

### 3. Environment Variable Naming

Use clear naming conventions:

```yaml
environment:
  PIPELINE_BUILD_ID: "123"
  PIPELINE_BRANCH: "main"
```

## Error Handling

### 1. Set Reasonable Timeout

Set timeout based on actual needs:

```yaml
timeout: 3600  # 1 hour
```

### 2. Check Workdir After Failure

After Pipeline failure, check files in workdir:

```bash
ls -la /tmp/pipeline/abc123
cat /tmp/pipeline/abc123/*.log
```

### 3. Use Post Hook for Cleanup

Add cleanup logic in Post hook:

```yaml
post: |
  echo "Cleaning up..."
  rm -rf /tmp/build-artifacts
```

## Performance Optimization

### 1. Use Parallel Execution

For independent tasks, use parallel execution:

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

### 2. Optimize Docker Images

- Use smaller base images
- Reuse Docker image layers
- Use multi-stage builds

### 3. Set Reasonable Concurrency

Set reasonable concurrency based on server resources:

```bash
pipeline server --max-concurrent 5
```

## Security Practices

### 1. Enable Authentication

Enable authentication in production:

```bash
pipeline server -u admin --password secret123
```

### 2. Use HTTPS

Use reverse proxy to provide HTTPS:

```nginx
server {
    listen 443 ssl;
    server_name pipeline.example.com;
    
    location / {
        proxy_pass http://localhost:8080;
    }
}
```

### 3. Limit Environment Variables

Only pass necessary environment variables:

```bash
pipeline run --allow-env GITHUB_TOKEN
```

## Monitoring and Logging

### 1. Configure Log Collection

Configure log collection system:

```bash
pipeline server 2>&1 | tee pipeline.log
```

### 2. Monitor Pipeline Execution

Use Web Console to monitor Pipeline execution:

```bash
open http://localhost:8080/console
```

### 3. Set Up Alerts

Set up alert mechanisms to detect issues promptly.

## Deployment Recommendations

### 1. Use Docker

Deploy using Docker:

```bash
docker run -d \
  -p 8080:8080 \
  -v /var/lib/pipeline:/data \
  ghcr.io/go-idp/pipeline:latest server
```

### 2. Use Reverse Proxy

Use Nginx or Traefik as reverse proxy.

### 3. Persistent Storage

Use persistent storage as working directory:

```bash
pipeline server -w /var/lib/pipeline
```

## Development Recommendations

### 1. Local Testing

Test Pipeline locally:

```bash
pipeline run -c pipeline.yaml
```

### 2. Use Examples

Refer to example configurations:

```bash
pipeline run -c examples/basic.yml
```

### 3. Debug Mode

Use debug mode to view detailed information:

```bash
DEBUG=1 pipeline run -c pipeline.yaml
```

## Common Issues

### 1. Configuration File Not Found

Ensure configuration file exists, or use `-c` option:

```bash
pipeline run -c pipeline.yaml
```

### 2. Environment Variables Not Passed

Use `--allow-env` option:

```bash
pipeline run --allow-env GITHUB_TOKEN
```

### 3. Docker Command Failed

Ensure Docker is installed and running:

```bash
docker ps
```

## Summary

Following these best practices can help you use Pipeline better:

1. **Configuration File Organization**: Use version control, environment separation
2. **Environment Variable Management**: Use environment variables, protect sensitive information
3. **Error Handling**: Set reasonable timeout, check workdir after failure
4. **Performance Optimization**: Use parallel execution, optimize Docker images
5. **Security Practices**: Enable authentication, use HTTPS
6. **Monitoring and Logging**: Configure log collection, monitor Pipeline execution
7. **Deployment Recommendations**: Use Docker, use reverse proxy
