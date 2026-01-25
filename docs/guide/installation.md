# Installation

Pipeline supports multiple installation methods. Choose the one that suits your needs.

## Build from Source

```bash
git clone https://github.com/go-idp/pipeline.git
cd pipeline
go build -o pipeline cmd/pipeline/main.go
```

## Install with Go

```bash
go install github.com/go-idp/pipeline/cmd/pipeline@latest
```

After installation, make sure `$GOPATH/bin` or `$HOME/go/bin` is in your `PATH` environment variable.

## Use Docker

```bash
docker pull ghcr.io/go-idp/pipeline:latest
```

Run with Docker:

```bash
docker run --rm -v $(pwd):/workspace ghcr.io/go-idp/pipeline:latest run -c /workspace/pipeline.yaml
```

## Verify Installation

After installation, verify with:

```bash
pipeline --version
```

This should output the Pipeline version information.
