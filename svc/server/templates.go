package server

// Template 是内置的流水线模板/案例，帮助用户快速上手。
type Template struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	YAML        string `json:"yaml"`
}

// BuiltinTemplates 内置模板列表。
// 启动时注入到配置模板库（见 seed.go），并可通过 GET /api/v1/templates 获取。
var BuiltinTemplates = []*Template{
	{
		ID:          "ci",
		Name:        "CI",
		Description: "检出 → 构建 → 测试 → 部署（Docker 镜像）",
		YAML: `name: CI

environment:
  CI: "true"

stages:
  - name: checkout
    jobs:
      - name: frontend
        steps:
          - name: checkout
            command: |
              git clone --progress --depth 1 -b main https://github.com/go-idp/pipeline $PWD
  - name: build
    jobs:
      - name: build
        steps:
          - name: go build
            command: go build -o pipeline ./cmd/pipeline
  - name: test
    jobs:
      - name: test
        steps:
          - name: go test
            command: go test ./...
  - name: deploy
    type: deploy
    run_mode: serial
    jobs:
      - name: deploy
        steps:
          - name: docker build
            command: |
              docker buildx build --push -t registry.example.com/idp/backend:test-pipeline .`,
	},
	{
		ID:          "release",
		Name:        "发布",
		Description: "lint → 测试 → 构建产物 → 发布到 registry（goreleaser）",
		YAML: `name: release

stages:
  - name: lint
    jobs:
      - name: lint
        steps:
          - name: golangci-lint
            command: golangci-lint run ./...
  - name: test
    jobs:
      - name: test
        steps:
          - name: go test
            command: go test -race ./...
  - name: build
    jobs:
      - name: build
        steps:
          - name: goreleaser
            command: goreleaser release --skip-publish
  - name: publish
    type: deploy
    run_mode: serial
    jobs:
      - name: publish
        steps:
          - name: push artifacts
            command: goreleaser release`,
	},
	{
		ID:          "deploy",
		Name:        "部署",
		Description: "构建镜像 → 推送 → 更新 Kubernetes 部署",
		YAML: `name: deploy

environment:
  REGISTRY: registry.example.com

stages:
  - name: build-image
    jobs:
      - name: build
        steps:
          - name: docker build
            command: docker build -t $REGISTRY/idp/backend:$BUILD_ID .
  - name: push
    jobs:
      - name: push
        steps:
          - name: docker push
            command: docker push $REGISTRY/idp/backend:$BUILD_ID
  - name: deploy
    type: deploy
    run_mode: serial
    jobs:
      - name: k8s
        steps:
          - name: kubectl apply
            command: kubectl set image deployment/idp-backend backend=$REGISTRY/idp/backend:$BUILD_ID`,
	},
	{
		ID:          "docs",
		Name:        "文档",
		Description: "构建并发布 VitePress 文档站点",
		YAML: `name: docs

stages:
  - name: build-docs
    jobs:
      - name: docs
        steps:
          - name: vitepress build
            command: pnpm run docs:build
  - name: publish-docs
    type: deploy
    run_mode: serial
    jobs:
      - name: publish
        steps:
          - name: gh-pages
            command: pnpm run docs:publish`,
	},
	{
		ID:          "nightly",
		Name:        "夜间基准",
		Description: "每日自动运行的性能基准测试",
		YAML: `name: nightly-benchmark

environment:
  CI: "true"

stages:
  - name: bench
    jobs:
      - name: benchmark
        steps:
          - name: go bench
            command: go test -bench=. -benchmem ./...`,
	},
}
