# Your First Pipeline

Let's create a simple Pipeline to get started quickly.

## Create Configuration File

Create a `.pipeline.yaml` file:

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

## Run Pipeline

```bash
pipeline run
```

Pipeline will automatically find the `.pipeline.yaml` file in the current directory and execute it.

## View Output

After successful execution, you should see output similar to:

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

## Next Steps

- Learn more configuration options, see [Configuration](./configuration.md)
- Understand core concepts, see [Core Concepts](./concepts.md)
- View more examples, see [Commands Documentation](/commands/)
