# web Command

The `pipeline web` command starts the full **management console**: it embeds the
built frontend (React + TypeScript, `web/ui/dist`) into the binary and serves it
together with the server backend (REST API + WebSocket + execution queue).

The frontend must be built before the binary is built:

```bash
cd web/ui && pnpm install && pnpm build
go build -o pipeline ./cmd/pipeline
```

## Basic Usage

```bash
pipeline web [options]
```

Then open <http://localhost:8080/> to access the console.

## Relationship with `pipeline server`

- `pipeline web` = embedded frontend + backend. It serves the SPA on `/`
  (with client-side routing fallback) and the full REST/WebSocket API.
- `pipeline server` = lightweight **API-only** service (REST + WebSocket +
  queue), without the embedded frontend. Use it when you only need the API
  (e.g. programmatic access or behind your own frontend).

Both commands share the same flags and backend behavior.

## Command Options

The options are identical to [`server`](./server.md):

| Flag | Env | Default | Description |
| --- | --- | --- | --- |
| `-p, --port` | `PORT` | `8080` | Listening port |
| `--path` | `ENDPOINT`, `SERVER_PATH` | `/ws` | WebSocket path |
| `-w, --workdir` | `WORKDIR` | `/tmp/go-idp/pipeline` | Work directory (stores `.pipeline_records/` and `.pipeline_configs/`) |
| `-u, --username` | `USERNAME` | — | Basic Auth username |
| `--password` | `PASSWORD` | — | Basic Auth password |
| `--allow-env` | `ALLOW_ENV` | — | Environment variables allowed to pass through (repeatable) |
| `--allow-all-env` | `ALLOW_ALL_ENV` | — | Allow all environment variables |
| `--max-concurrent` | `MAX_CONCURRENT` | `2` | Max concurrent executions |
| `--task-timeout` | `TASK_TIMEOUT` | `3600` | Default task timeout (seconds), `0` disables |
| `--task-executor` | `TASK_EXECUTOR` | `in-process` | `in-process` or `subprocess` |

**Example**:

```bash
# start the console on port 9090 with auth
pipeline web -p 9090 -u admin --password secret

# with 4 concurrent executions and a 2h default task timeout
pipeline web --max-concurrent 4 --task-timeout 7200
```

## Web Console

The management console provides (Linear-style monochrome UI, `⌘K` command
palette, zh/en i18n):

- **Overview**: today's runs, 7-day success rate, average duration, running
  count, 14-day run trend, status distribution, recent runs, quick pipelines
- **Pipelines**: pipeline definition (YAML-first) CRUD with a CodeMirror YAML
  editor and a visual structure preview, run history
- **Runs**: run list with status filters/search; run detail with a stage
  timeline, a stage/job/step execution tree and per-step logs (live while
  running), cancel / rerun / export logs
- **Queue**: active tasks and concurrency usage
- **Settings**: server info, concurrency, task timeout/executor, environment
  allowlist, storage, locale

### API / WebSocket

- REST API base: `/api/v1` (see [`server`](./server.md) for the endpoint list,
  plus the management extensions below)
- WebSocket: `ws://localhost:8080/ws`

#### Management API extensions

- `GET /api/v1/runs` - run list (same as `/pipelines`)
- `GET /api/v1/runs/:id` - run detail (definition tree + per-level states + logs)
- `POST /api/v1/runs` - create a run from YAML `{ "config": "...", "trigger": "manual" }`
- `POST /api/v1/runs/:id/rerun` / `POST /api/v1/pipelines/:id/rerun` - rerun a run
- `GET /api/v1/pipelines/:id/stages` - stage/job/step runtime states
- `GET|POST /api/v1/configs`, `GET|PUT|DELETE /api/v1/configs/:id` - pipeline
  definition templates (persisted under `.pipeline_configs/`)

### Templates & demo data

To help users get started quickly, the server injects demo data on startup
(idempotent: only when the corresponding storage is empty):

- **Built-in templates** (`GET /api/v1/templates`, read-only): 5 example
  pipelines (CI / release / deploy / docs / nightly). They are also copied into
  the config template library on first start. In the console, open **Pipelines →
  从模板创建 (From template)** to create a pipeline from a template or run it
  directly.
- **Demo runs**: 3 sample run records (a succeeded CI run, a failed deploy run
  and a succeeded docs run) with logs and per-level (stage/job/step) states, so
  the Overview / Runs / Run detail pages have content on first launch.

Real data (non-empty store) is never overwritten; once you create your own
records the seed stops.

### Runtime states

While a run executes (`in-process` executor), the server captures the
stage/job/step state machine (`pending → running → succeeded | failed`) through
an observer in the execution engine, exposed via `GET /api/v1/runs/:id`
(`definition` + `states`) and `GET /api/v1/pipelines/:id/stages`.

> Note: with `--task-executor subprocess`, per-level states are not collected
> (the task runs in a separate process); only run-level status is tracked.
