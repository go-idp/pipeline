# 第一个 Pipeline

让我们创建一个简单的 Pipeline 来快速上手。

## 创建配置文件

创建 `.pipeline.yaml` 文件：

```yaml
name: My First Pipeline

stages:
  - name: build
    jobs:
      - name: build-job
        steps:
          - name: hello
            command: echo "Hello, Pipeline!"
```

## 运行 Pipeline

```bash
pipeline run
```

Pipeline 会自动查找当前目录下的 `.pipeline.yaml` 文件并执行。

## 查看输出

执行成功后，您应该会看到类似以下的输出：

```
[Pipeline] My First Pipeline started
[Stage] build started
[Job] build-job started
[Step] hello started
Hello, Pipeline!
[Step] hello completed
[Job] build-job completed
[Stage] build completed
[Pipeline] My First Pipeline completed
```

## 下一步

- 了解更多配置选项，请查看 [配置文件](./configuration.md)
- 了解核心概念，请查看 [核心概念](./concepts.md)
- 查看更多示例，请查看 [命令文档](/commands/)
