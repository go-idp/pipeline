# web 命令

`pipeline web` 命令启动完整的**管理后台**：将构建好的前端（React + TypeScript，`web/ui/dist`）
嵌入到二进制中，与后端服务（REST API + WebSocket + 执行队列）一起提供。

构建前端后再构建二进制：

```bash
cd web/ui && pnpm install && pnpm build
go build -o pipeline ./cmd/pipeline
```

## 基本用法

```bash
pipeline web [options]
```

然后访问 <http://localhost:8080/> 打开管理后台。

## 与 `pipeline server` 的关系

- `pipeline web` = 嵌入前端 + 后端：在 `/` 提供 SPA（含前端路由回退）与完整 REST/WebSocket API。
- `pipeline server` = 轻量 **API-only** 服务（REST + WebSocket + 队列），不含嵌入前端。

两个命令共享相同的参数与后端行为。

## 命令参数

与 [`server`](./server.md) 相同：

| 参数 | 环境变量 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `-p, --port` | `PORT` | `8080` | 监听端口 |
| `--path` | `ENDPOINT`, `SERVER_PATH` | `/ws` | WebSocket 路径 |
| `-w, --workdir` | `WORKDIR` | `/tmp/go-idp/pipeline` | 工作目录（存放 `.pipeline_records/` 与 `.pipeline_configs/`） |
| `-u, --username` | `USERNAME` | — | Basic Auth 用户名 |
| `--password` | `PASSWORD` | — | Basic Auth 密码 |
| `--allow-env` | `ALLOW_ENV` | — | 允许透传的环境变量（可重复） |
| `--allow-all-env` | `ALLOW_ALL_ENV` | — | 允许全部环境变量 |
| `--max-concurrent` | `MAX_CONCURRENT` | `2` | 最大并发执行数 |
| `--task-timeout` | `TASK_TIMEOUT` | `3600` | 默认任务超时（秒），`0` 不限制 |
| `--task-executor` | `TASK_EXECUTOR` | `in-process` | `in-process` 或 `subprocess` |

## Web 管理后台

管理后台提供（Linear 黑白灰极简风格、`⌘K` 命令面板、中英双语）：

- **总览**：今日运行、7 天成功率、平均耗时、运行中任务、14 天趋势、状态分布、最近运行、常用流水线
- **流水线**：YAML 优先的定义管理（CodeMirror 编辑器 + 可视化预览）、运行历史
- **运行**：运行列表（状态过滤/搜索）；运行详情（阶段时间条、stage/job/step 执行树、按步骤日志（运行中实时）、取消/重跑/导出）
- **队列**：活跃任务与并发占用
- **设置**：服务器信息、并发、任务超时/执行器、环境变量白名单、存储、语言

### API / WebSocket

- REST API 基础路径：`/api/v1`（端点列表见 [`server`](./server.md)，另有管理扩展）
- WebSocket：`ws://localhost:8080/ws`

#### 管理 API 扩展

- `GET /api/v1/runs` - 运行列表（同 `/pipelines`）
- `GET /api/v1/runs/:id` - 运行详情（执行结构树 + 三级状态 + 日志）
- `POST /api/v1/runs` - 从 YAML 创建运行 `{ "config": "...", "trigger": "manual" }`
- `POST /api/v1/runs/:id/rerun` / `POST /api/v1/pipelines/:id/rerun` - 重跑
- `GET /api/v1/pipelines/:id/stages` - stage/job/step 运行时状态
- `GET|POST /api/v1/configs`、`GET|PUT|DELETE /api/v1/configs/:id` - 流水线定义模板（持久化于 `.pipeline_configs/`）

### 模板与模拟数据

为帮助用户快速上手，服务启动时会注入演示数据（幂等：仅当对应存储为空时）：

- **内置模板**（`GET /api/v1/templates`，只读）：5 个示例流水线（CI / 发布 / 部署 /
  文档 / 夜间基准）。首次启动时也会复制到配置模板库。后台中通过 **流水线 → 从模板创建**
  即可基于模板创建流水线或直接运行。
- **演示运行**：3 条示例运行记录（成功 CI、失败部署、成功文档），含日志与
  stage/job/step 三级状态，让总览 / 运行 / 运行详情页首次打开即有内容。

已有数据（非空）不会被覆盖；一旦有了自己的记录，种子注入即停止。

### 运行时状态

`in-process` 执行器下，服务端通过执行引擎中的观察者采集 stage/job/step 状态机
（`pending → running → succeeded | failed`），经 `GET /api/v1/runs/:id`
（`definition` + `states`）与 `GET /api/v1/pipelines/:id/stages` 暴露。

> 注意：`--task-executor subprocess` 时任务在独立进程中运行，不采集三级状态，仅跟踪运行级状态。
