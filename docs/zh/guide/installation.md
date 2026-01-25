# 安装

Pipeline 支持多种安装方式，您可以根据需要选择合适的方式。

## 从源码编译

```bash
git clone https://github.com/go-idp/pipeline.git
cd pipeline
go build -o pipeline cmd/pipeline/main.go
```

## 使用 Go 安装

```bash
go install github.com/go-idp/pipeline/cmd/pipeline@latest
```

安装后，确保 `$GOPATH/bin` 或 `$HOME/go/bin` 在您的 `PATH` 环境变量中。

## 使用 Docker

```bash
docker pull ghcr.io/go-idp/pipeline:latest
```

使用 Docker 运行：

```bash
docker run --rm -v $(pwd):/workspace ghcr.io/go-idp/pipeline:latest run -c /workspace/pipeline.yaml
```

## 验证安装

安装完成后，可以通过以下命令验证：

```bash
pipeline --version
```

应该会输出 Pipeline 的版本信息。
