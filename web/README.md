# Pipeline Web 管理后台 · 产品设计文档（原型确认稿）

> 本文档用于确认 `pipeline web` 管理后台的产品设计。配套的可交互 HTML 原型见同目录 [`prototype.html`](./prototype.html)（直接用浏览器打开即可，无需构建）。
>
> 状态：**待确认** — 确认后再进入实现阶段（`pipeline web` 命令 + Go 后端 + 前端资源嵌入）。

---

## 1. 背景与目标

### 1.1 现状

项目已具备 `pipeline server` 命令，提供：

- WebSocket 实时执行通道（run / done / error / stdout / stderr 动作协议）
- REST API：`/api/v1/pipelines`（列表/详情/日志/导出/取消/删除/批量）、`/api/v1/queue`、`/api/v1/settings`、`/api/v1/configs/convert/*`
- 旧版 Web Console（`/console`，单文件 3789 行，紫色渐变风格，功能堆叠、无信息架构）

### 1.2 目标

将管理后台重新设计为 `pipeline web`，定位是 **CI/CD 风格的流水线管理 + 任务执行控制台**：

1. **流水线管理**：流水线定义（YAML 优先，可视化为辅）的增删改查、启停、运行历史。
2. **任务管理执行**：每次运行（Run）的全生命周期管理 — 排队 → 运行 → 成功/失败/取消，运行详情可视化（stage/job/step 树 + 实时日志）、取消、重跑、导出日志。
3. **视觉**：Linear 式黑白灰极简风格，克制、高效、键盘友好。

### 1.3 范围界定

| 属于本产品（v1） | 不属于本产品（v1） |
| --- | --- |
| 流水线定义 CRUD + YAML 编辑器 | 拖拽式图形化流水线编排器（DAG canvas） |
| 运行列表 / 详情 / 日志 / 取消 / 重跑 | 定时调度（cron）编排 |
| 队列监控与并发管理 | 多用户 RBAC 权限体系 |
| 系统设置（并发/超时/执行器/环境变量/认证） | 分布式 Runner 注册与管理 |
| 单机 server 进程托管 | 高可用 / 多节点 |

---

## 2. 业界标杆调研

| 产品 | 借鉴点 | 不借鉴点 |
| --- | --- | --- |
| **Linear** | 黑白灰极简视觉、键盘优先（Cmd+K）、信息密度与留白的平衡、状态色的克制使用 | — |
| **GitHub Actions** | 运行详情页的 **job → step 树 + 实时日志** 布局；Run 列表的过滤与徽标；Rerun（全部/失败） | 彩色渐变、装饰性元素 |
| **Buildkite** | 构建页顶部的 **stage 时间条可视化**（横向分段、悬停看耗时）、日志搜索与跟随 | 过于密集的仪表化 |
| **Jenkins（2025 重设计）** | 现代化重设计：减少层级、清晰的导航、列表优先 | 传统 Jenkins 的繁杂配置页 |
| **Argo Workflows UI** | 工作流运行的可视化表达（阶段/步骤流） | DAG canvas 的复杂度 |
| **Vercel / Railway** | 部署页的极简状态卡、一键重跑、部署时间线 | — |

> 参考链接：[聊聊 Linear 的设计变革](https://sspai.com/post/104449)、[Redesigning Jenkins (Part Two)](https://www.jenkins.io/blog/2025/07/24/redesigning-jenkins-part-two/)、[Buildkite: Introducing the new build page](https://buildkite.com/resources/changelog/266-introducing-the-new-build-page-engineered-for-scale-and-flexibility/)、[Buildkite: A simpler build page layout with a new list view](https://buildkite.com/resources/changelog/356-a-simpler-build-page-layout-with-a-new-list-view/)、[CI/CD Tools for Kubernetes](https://deepaksood619.github.io/devops/cicd/comparison/)

### 2.1 设计原则

1. **YAML 优先，可视化辅助**：流水线本质是 YAML 文件，编辑器以代码为主、可视化树为结构化预览，两者实时互转（后端已有 `/configs/convert/*` API）。
2. **运行详情三要素**：状态（一眼可见）、结构（stage/job/step 树）、日志（可搜索、可跟随、可导出）。三者同屏联动。
3. **黑白灰 + 克制的状态色**：界面骨架 100% 单色；仅状态点、徽标、运行指示使用极小面积的低饱和色（蓝=运行中、绿=成功、红=失败、灰=排队/取消）。
4. **键盘优先**：`Cmd+K` 全局命令面板、`R` 重跑、`Esc` 关闭、`/` 聚焦搜索。
5. **列表优先**：一切实体先给列表，详情通过行点击进入，避免大卡片堆叠。

---

## 3. 信息架构

```
pipeline web
├── Overview 总览            # 运行趋势 / 状态分布 / 最近运行 / 常用流水线
├── Pipelines 流水线         # 定义管理
│   ├── 列表（搜索/过滤/批量）
│   ├── 新建 / 编辑（YAML ⇄ Visual 双模式）
│   └── 详情（定义 / 运行历史 / 操作）
├── Runs 运行               # 任务执行管理
│   ├── 列表（状态过滤/搜索/时间范围）
│   └── 详情（stage 时间条 + step 树 + 实时日志 + 元信息）
├── Queue 队列              # 排队 / 运行中任务、并发监控
└── Settings 设置           # 服务器信息 / 并发 / 超时 / 执行器 / 环境变量 / 认证 / 存储
```

### 3.1 数据模型（与后端对应）

```
Pipeline (定义)                 Run (一次执行，即后端 PipelineRecord / QueueItem)
├── name / description          ├── id / name
├── stages[]                    ├── status: pending|running|succeeded|failed|cancelled
│   └── Stage                   ├── started_at / succeed_at / failed_at / cancelled_at
│       ├── run_mode(serial|parallel)  ├── duration
│       └── jobs[]              ├── error
│           └── Job             ├── config{workdir, timeout, image}
│               └── steps[]     ├── yaml (原始定义)
│                   └── Step    ├── logs[] {type: stdout|stderr, message, timestamp}
│                               └── stages_state[]  # 运行期各层状态（扩展点）
├── environment / image / timeout
└── state (最近一次运行状态)
```

> 说明：后端 Store 目前记录到 Run 粒度 + 扁平日志；`stages_state[]`（每层 stage/job/step 的运行时状态）与运行详情页的树/时间条可视化强相关，是 v1 实现期的核心扩展点（见 §7）。

---

## 4. 页面规格

### 4.1 Overview 总览

- **统计卡（4 枚，一行）**：今日运行数、成功率（近 7 天）、平均耗时、当前运行中。
- **近 14 天运行趋势**：紧凑柱状图（成功/失败双色堆叠），悬停显示明细。
- **状态分布**：排队 / 运行中 / 成功 / 失败 / 取消 计数。
- **最近运行**：最新 8 条运行，点击进入详情。
- **常用流水线**：最近运行的流水线，附「运行 ▸」快捷按钮。

### 4.2 Pipelines 流水线

**列表**（Table，Linear 风格）：
列 = 名称(+描述) · 状态(启用/禁用) · 最近运行状态徽标 · 最近运行耗时 · 最近运行时间 · 运行次数 · 操作（运行▸/更多⋮）。支持搜索、状态过滤、批量选择（批量运行/删除）。

**新建/编辑**（全屏抽屉）：
- 顶部：名称 + 描述 + 启用开关 + 保存。
- Tab 1 **YAML**：等宽编辑器（语法高亮、错误提示行内定位）。
- Tab 2 **Visual**：Stage 卡片列表（可折叠），每卡内 Job → Step 树，支持增删；字段与 YAML 双向同步。
- 底部：`pipeline run -c <yaml>` 的命令预览（复制按钮）。

**详情**：定义只读 YAML + Visual 预览；下方「运行历史」表格（复用 Runs 列表组件）；操作：运行、编辑、复制为模板、删除。

### 4.3 Runs 运行（核心页面）

**列表**：
- 过滤条：状态 Tabs（全部/排队/运行中/成功/失败/已取消）+ 搜索（名称/ID）+ 时间范围 + 只看我的（后续）。
- 表格列：流水线名 · ID 短码 · 状态徽标 · 耗时 · 触发方式(手动/API/重跑) · 开始时间。行点击进入详情。

**详情**（三区块联动布局）：

```
┌──────────────────────────────────────────────────────┐
│ Run 标题 + 状态徽标 + 耗时      [取消] [重跑▸] [⋮]      │
├──────────────────────────────────────────────────────┤
│ Stage 时间条  ▓▓▓▓▓▓▓▓▓▓▓▓▓  ▓▓▓▓▓▓▓  ▓▓▓▓▓          │
│  checkout 12s    build 45s   test 8s  deploy(等待)     │
├───────────────┬──────────────────────────────────────┤
│ Step 树        │  实时日志（等宽、stdout/stderr 分色、 │
│ ▸ checkout     │  行号、时间戳、搜索、跟随开关、导出）   │
│   · job        │                                      │
│     - step ⏱   │                                      │
│ ...           │                                      │
├───────────────┴──────────────────────────────────────┤
│ 元信息抽屉：ID / workdir / timeout / image / 环境变量   │
│ / 触发者 / 原始 YAML                                  │
└──────────────────────────────────────────────────────┘
```

- **Stage 时间条**：横向分段（Buildkite 式），每段=一个 stage，颜色=状态，宽度=耗时占比；悬停显示「名称 + 耗时」，点击定位到树中对应 stage。
- **Step 树**：Stage（可折叠）→ Job → Step 三级，左侧状态点 + 耗时；点击 step 切换日志视图；运行中的 step 有脉冲动画。
- **实时日志**：等宽字体，`stdout` 黑 / `stderr` 红灰；自动滚动（可关）；搜索高亮；行号 + 相对时间戳；复制/导出。
- **操作**：取消（运行中/排队）、重跑 ▸（重跑全部 / 仅失败步骤 / 复制 YAML 新跑）、导出日志（txt/json）、删除记录。

### 4.4 Queue 队列

- 统计：运行中 / 排队中 / 最大并发（进度条表达占用率）。
- 队列表格：任务名 · 状态 · 排队时长 · 操作（取消）。排队任务显示在运行中的任务之前，灰色处理。

### 4.5 Settings 设置

| 分组 | 项 |
| --- | --- |
| 服务器 | 版本、运行时长、HTTP 端口、Base Path |
| 执行 | 最大并发数、默认任务超时（秒）、任务执行方式（in-process / subprocess） |
| 环境 | 允许透传的环境变量（allow-env / allow-all-env 开关） |
| 认证 | 用户名 / 密码（Basic Auth） |
| 存储 | workdir 路径、最大记录数（当前 1000）、记录/配置持久化说明 |
| 危险区 | 清空运行历史（确认弹窗） |

---

## 5. 关键交互细节

| 交互 | 说明 |
| --- | --- |
| Cmd+K 命令面板 | 全局：跳转页面、新建流水线、运行流水线、查看队列；`↑↓` 选择、`Enter` 执行 |
| 状态流转 | pending→running→succeeded/failed；running/pending→cancelled；终态不可取消，仅可重跑/删除 |
| 重跑 | 「重跑全部」= 同一 YAML 新 Run ID；「仅失败步骤」= 复用定义 + 失败步骤上下文（v1 建议先做「重跑全部」） |
| 运行确认 | 点击「运行 ▸」展开确认层：显示流水线名 + 环境变量预览（可追加 KEY=VALUE）+ 执行按钮 |
| 空态 | 无流水线/无运行时有引导性空态（含示例 YAML 一键创建） |
| Toast | 所有变更操作给轻量 toast 反馈（成功/失败），不弹 alert |

---

## 6. 视觉规范（Linear 黑白灰）

| Token | 值 | 用途 |
| --- | --- | --- |
| 背景 | `#FAFAFA` | 页面底色 |
| 面板 | `#FFFFFF` | 卡片/侧栏/表头 |
| 边框 | `#E4E4E7` | 1px 分隔 |
| 主文字 | `#18181B` | 标题 |
| 次文字 | `#71717A` | 描述/次要信息 |
| 弱文字 | `#A1A1AA` | 占位/时间 |
| 强调 | `#18181B`（黑） | 主按钮/激活态 |
| 运行中 | `#3B82F6` | 状态点/时间条 |
| 成功 | `#16A34A` | 状态点/徽标 |
| 失败 | `#DC2626` | 状态点/徽标 |
| 排队/取消 | `#A1A1AA` | 状态点/徽标 |

- 字体：`-apple-system, "Inter", "SF Pro Text", "Segoe UI", sans-serif`；代码：`ui-monospace, "SF Mono", Menlo, monospace`。
- 圆角 6px、阴影极淡（`0 1px 2px rgba(0,0,0,.04)`）、聚焦态用 1px 黑边。
- 按钮：主按钮纯黑底白字；次按钮白底灰边；危险操作红字。图标 16px 线性。
- 间距 8px 基准网格；表格行高 40px；内容区最大宽度 1200px。

---

## 7. 实现阶段建议（确认后执行）

### 7.1 后端扩展（Go）

1. `cmd/pipeline/commands/web.go`：注册 `web` 命令（别名 `server`，参数向后兼容：`--port/--workdir/--username/--password/--max-concurrent/--task-timeout/--task-executor/--allow-env`）。
2. 运行详情结构化数据：在 `store.go` 的 `PipelineRecord` 增加 `StagesState`（每层 stage/job/step 的 `{id,name,status,started_at,succeed_at,failed_at,error}`），由 `pipeline.Run` 生命周期回调收集（在 `svc/server/run.go` 的 executor 层注入）。
3. REST 补齐：`GET /api/v1/runs`（Runs 列表，含 duration/trigger 计算）、`POST /api/v1/pipelines/:id/rerun`、`GET /api/v1/pipelines/:id/history`、`GET /api/v1/configs`（模板 CRUD 已存在于 configStore，需暴露 REST）。
4. 静态资源：新前端构建产物嵌入二进制（`go:embed`），`/console` 与 `/` 路由指向新 UI（保留旧路由一段时间或直接替换，待定）。

### 7.2 前端实现

- 单页应用：建议直接采用无框架或轻框架（vanilla/Preact/Vue3），随 Go 二进制 `go:embed` 分发，无独立构建链负担（与现状一致：单 HTML 文件）。
- 原型中的 mock 数据层替换为 REST 调用；WebSocket 动作协议（stdout/stderr/done/error）对接实时日志。
- 组件清单：Sidebar / CommandPalette / Table / StatusBadge / StageTimeline / StepTree / LogViewer / Drawer / Modal / Toast / StatCard。

---

## 8. 待确认问题

1. **旧 console 去留**：新 UI 直接替换 `/console`，还是新旧共存（`/console` 保留）？
2. **运行详情层级状态**：是否需要 v1 就展示 stage/job/step 三级实时状态（需要后端加 `StagesState` 采集），还是先只做 Run 粒度 + 扁平日志？
3. **流水线定义存放**：定义为「文件系统目录（如 `.pipeline/` 下的 yaml）+ 数据库/JSON 持久化」还是「仅配置模板库（现有 configStore）」？这决定 Pipelines 页的 CRUD 实现方式。
4. **认证**：v1 保持 Basic Auth（现状），还是引入登录页 + session/token？
5. **多语言**：界面文案中文 or English or 双语（i18n）？
