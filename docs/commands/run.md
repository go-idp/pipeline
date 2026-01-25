# run Command

The `pipeline run` command is used to run Pipeline directly on your local machine. This is the most commonly used command in Pipeline, supporting loading configuration from local files or remote URLs and executing Pipeline workflows.

## Basic Usage

```bash
pipeline run [options]
```

## Command Options

### `-c, --config`

Specify the Pipeline configuration file path.

- **Type**: String
- **Environment Variable**: `PIPELINE_CONFIG`
- **Default**: Auto-detect (`.pipeline.yaml` or `.go-idp/pipeline.yaml`)
- **Supported Formats**:
  - Local file path: `pipeline.yaml`
  - HTTP/HTTPS URL: `https://example.com/pipeline.yaml`

**Examples**:

```bash
# Use local configuration file
pipeline run -c pipeline.yaml

# Use remote configuration file
pipeline run -c https://example.com/pipeline.yaml

# Use environment variable
export PIPELINE_CONFIG=pipeline.yaml
pipeline run
```

### `-w, --workdir`

Specify the Pipeline working directory.

- **Type**: String
- **Environment Variable**: `PIPELINE_WORKDIR`
- **Default**: Current directory

**Example**:

```bash
pipeline run -w /tmp/my-pipeline
```

### `-e, --env`

Set environment variables (can be used multiple times).

- **Type**: String array
- **Environment Variable**: `ENV`
- **Format**: `KEY=VALUE`

**Example**:

```bash
pipeline run -e GITHUB_TOKEN=xxx -e BUILD_NUMBER=123
```

## Configuration File Search

If the `-c` option is not specified, `pipeline run` will automatically search for configuration files in the following order:

1. `.pipeline.yaml` (current directory)
2. `.go-idp/pipeline.yaml` (current directory)

If a configuration file is found, it will be used automatically; otherwise, an error will be reported.

## Usage Examples

### Example 1: Basic Usage

```bash
# Create configuration file
cat > .pipeline.yaml <<EOF
name: Hello Pipeline
stages:
  - name: greet
    jobs:
      - name: say-hello
        steps:
          - name: hello
            command: echo "Hello, World!"
EOF

# Run Pipeline
pipeline run
```

## Execution Flow

1. **Load Configuration**: Load Pipeline configuration from local file or remote URL
2. **Parse Configuration**: Parse YAML configuration and validate
3. **Apply Options**: Apply command-line options (workdir, image, environment variables, etc.)
4. **Execute Pipeline**: Execute each Stage in order
5. **Cleanup**: Clean workdir on success, preserve workdir on failure for debugging

## Error Handling

When Pipeline execution fails:

- **workdir Preserved**: Failed workdir will be preserved for debugging
- **Error Logs**: Output detailed error information, including workdir location
- **Status Information**: Record error information in Pipeline State

For detailed error handling, see [Error Handling Documentation](/architecture/error-handling.md).

## Environment Variables

You can set command options via environment variables:

```bash
export PIPELINE_CONFIG=pipeline.yaml
export PIPELINE_WORKDIR=/tmp/pipeline
export PIPELINE_IMAGE=alpine:latest
pipeline run
```

## Debug Mode

Enable debug mode to view detailed execution information:

```bash
DEBUG=1 pipeline run -c pipeline.yaml
```

Debug mode will:
- Display Pipeline configuration in JSON format
- Preserve temporarily downloaded remote configuration files
- Output more detailed log information
