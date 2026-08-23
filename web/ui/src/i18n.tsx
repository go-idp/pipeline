import { createContext, useContext, useEffect, useState, type ReactNode } from 'react';

export type Locale = 'zh' | 'en';

const zh = {
  nav: { overview: '总览', pipelines: '流水线', runs: '运行', queue: '队列', settings: '设置' },
  common: {
    search: '搜索', new: '新建', run: '运行', cancel: '取消', save: '保存', delete: '删除',
    edit: '编辑', copy: '复制', close: '关闭', confirm: '确认', export: '导出', rerun: '重跑',
    loading: '加载中…', empty: '暂无数据', all: '全部', enabled: '已启用', disabled: '已停用',
    status: '状态', duration: '耗时', trigger: '触发', startedAt: '开始时间', actions: '操作',
    pipeline: '流水线', runId: '运行 ID', name: '名称', description: '描述', yes: '是', no: '否',
  },
  status: { pending: '排队中', running: '运行中', succeeded: '成功', failed: '失败', cancelled: '已取消' },
  trigger: { manual: '手动', api: 'API', rerun: '重跑', ws: 'WS' },
  overview: {
    title: '总览', todayRuns: '今日运行', rate7: '近 7 天成功率', avgDur: '平均耗时', runningNow: '当前运行中',
    trend14: '近 14 天运行', dist: '状态分布', recentRuns: '最近运行', quickPipelines: '常用流水线',
    times: '次', okRate: '成功 {n} 次', avg7: '近 7 天 {n} 次运行', concurrency: '并发上限 {n}',
  },
  pipelines: {
    title: '流水线', newPipeline: '新建流水线', editPipeline: '编辑流水线', name: '名称', desc: '描述',
    stagesCount: '{n} 个阶段', lastRun: '最近运行', runCount: '运行次数', lastRunAt: '最近运行时间',
    history: '运行历史', definition: '定义', config: '配置', yamlTab: 'YAML', visualTab: '可视化',
    yamlEditor: '流水线定义', addStage: '添加 Stage', saveOk: '流水线已保存', created: '已创建',
    runConfirm: '运行流水线', envParams: '运行参数（可选）', envPlaceholder: '追加环境变量，每行一个，如：\nBUILD_ID=10086',
    enqueueHint: '任务将进入队列，按序执行', enqueue: '加入队列', enqueued: '已加入队列',
    copyCmd: '复制命令', deleted: '流水线已删除', deleteConfirm: '删除流水线 {name}',
    deleteWarn: '流水线定义将被删除，历史运行记录保留。', noMatch: '没有匹配的流水线',
    noMatchSub: '调整搜索条件，或创建一个新的流水线',
    fromTemplate: '从模板创建', useTemplate: '使用模板', templates: '模板与案例',
    templatesSub: '内置流水线模板，点击「使用模板」创建为流水线，或直接运行',
    templateUsed: '已从模板创建流水线', stages: '{n} 个阶段', directRun: '直接运行',
  },
  runs: {
    title: '运行', noMatch: '没有匹配的运行', noMatchSub: '调整搜索条件', noRuns: '还没有运行记录',
    noRunsSub: '从流水线页点击「运行」开始', cancelRun: '取消运行 {id}', cancelWarn: '运行将被终止，正在执行的步骤会立即停止。此操作不可撤销。',
    cancelled: '已取消', rerunAll: '重跑全部步骤', rerunFailed: '仅重跑失败的步骤', rerunHint: '以相同 YAML 创建一次全新的运行',
    rerunFailedHint: '跳过已成功的步骤，从失败点继续（v1 规划中）', startRerun: '开始重跑', rerunCreated: '已创建重跑',
    exportLogs: '导出日志', logsExported: '日志已导出', viewYaml: '查看原始定义', runDetail: '运行详情',
    timing: '耗时 {d}', meta: '元信息', execStructure: '执行结构', realtime: '实时输出中', autoScroll: '自动滚动',
    lines: '{n} 行', waiting: '（该步骤尚未开始，等待上游完成）', noLogs: '（无日志输出）',
    logsCopied: '日志已复制', config: '配置', env: 'environment', workdir: 'workdir', timeout: 'timeout',
    image: 'image', trigger: '触发', stepNotStarted: '该步骤尚未开始',
  },
  queue: {
    title: '队列', running: '运行中', pending: '排队中', maxConcurrent: '最大并发', executor: '执行器',
    usage: '占用 {cur}/{max} 并发', queueActive: '{n} 个活跃任务', enqueueAt: '入队时间', waitTime: '排队时长',
    idle: '队列空闲', idleSub: '当前没有运行中或排队的任务', cancelTask: '取消任务 {id}', cancelWarn: '任务将从队列移除或立即终止。',
  },
  settings: {
    title: '设置', server: '服务器', version: '版本', uptime: '运行时长', addr: '地址', basePath: 'Base Path',
    auth: '认证', username: '用户名', password: '密码', set: '已设置', execution: '执行',
    maxConcurrent: '最大并发数（max-concurrent）', taskTimeout: '默认任务超时（秒，0 不限制）',
    executor: '任务执行方式（task-executor）', restartHint: '变更需重启服务生效',
    envAllow: '环境变量透传', allowAll: '允许全部环境变量（allow-all-env）', add: '添加',
    storage: '存储', workdir: 'workdir', maxRecords: '记录上限', configTpl: '配置模板',
    save: '保存设置', saved: '设置已保存（重启服务后生效）', clearHistory: '清空运行历史',
    clearConfirm: '清空运行历史', clearWarn: '将删除全部运行记录（{n} 条），此操作不可恢复。', cleared: '运行历史已清空',
    locale: '界面语言',
  },
  palette: {
    placeholder: '输入命令或搜索…（↑↓ 选择，Enter 执行）', pages: '页面', actions: '操作',
    newPipeline: '新建流水线', runPipeline: '运行流水线…', viewRunning: '查看运行中的任务',
  },
  toasts: { welcome: '欢迎使用 Pipeline Web' },
  footer: { serverRunning: '服务运行中 · :{port}' },
};
export type Dict = typeof zh;

const en: Dict = {
  nav: { overview: 'Overview', pipelines: 'Pipelines', runs: 'Runs', queue: 'Queue', settings: 'Settings' },
  common: {
    search: 'Search', new: 'New', run: 'Run', cancel: 'Cancel', save: 'Save', delete: 'Delete',
    edit: 'Edit', copy: 'Copy', close: 'Close', confirm: 'Confirm', export: 'Export', rerun: 'Rerun',
    loading: 'Loading…', empty: 'No data', all: 'All', enabled: 'Enabled', disabled: 'Disabled',
    status: 'Status', duration: 'Duration', trigger: 'Trigger', startedAt: 'Started', actions: 'Actions',
    pipeline: 'Pipeline', runId: 'Run ID', name: 'Name', description: 'Description', yes: 'Yes', no: 'No',
  },
  status: { pending: 'Queued', running: 'Running', succeeded: 'Succeeded', failed: 'Failed', cancelled: 'Cancelled' },
  trigger: { manual: 'Manual', api: 'API', rerun: 'Rerun', ws: 'WS' },
  overview: {
    title: 'Overview', todayRuns: 'Runs today', rate7: '7d success rate', avgDur: 'Avg duration', runningNow: 'Running now',
    trend14: 'Runs (14 days)', dist: 'Status distribution', recentRuns: 'Recent runs', quickPipelines: 'Frequent pipelines',
    times: '', okRate: '{n} succeeded', avg7: '{n} runs in 7 days', concurrency: 'Max concurrency {n}',
  },
  pipelines: {
    title: 'Pipelines', newPipeline: 'New pipeline', editPipeline: 'Edit pipeline', name: 'Name', desc: 'Description',
    stagesCount: '{n} stages', lastRun: 'Last run', runCount: 'Runs', lastRunAt: 'Last run at',
    history: 'Run history', definition: 'Definition', config: 'Config', yamlTab: 'YAML', visualTab: 'Visual',
    yamlEditor: 'Pipeline definition', addStage: 'Add stage', saveOk: 'Pipeline saved', created: 'Created',
    runConfirm: 'Run pipeline', envParams: 'Run parameters (optional)', envPlaceholder: 'Extra environment variables, one per line, e.g.\nBUILD_ID=10086',
    enqueueHint: 'The run will be queued and executed in order', enqueue: 'Enqueue', enqueued: 'enqueued',
    copyCmd: 'Copy command', deleted: 'Pipeline deleted', deleteConfirm: 'Delete pipeline {name}',
    deleteWarn: 'The pipeline definition will be deleted; run history is kept.', noMatch: 'No matching pipelines',
    noMatchSub: 'Adjust filters or create a new pipeline',
    fromTemplate: 'From template', useTemplate: 'Use template', templates: 'Templates & Examples',
    templatesSub: 'Built-in pipeline templates. Click "Use template" to create a pipeline, or run directly.',
    templateUsed: 'Pipeline created from template', stages: '{n} stages', directRun: 'Run directly',
  },
  runs: {
    title: 'Runs', noMatch: 'No matching runs', noMatchSub: 'Adjust filters', noRuns: 'No runs yet',
    noRunsSub: 'Click "Run" on a pipeline to start', cancelRun: 'Cancel run {id}', cancelWarn: 'The run will be terminated immediately. This cannot be undone.',
    cancelled: 'Cancelled', rerunAll: 'Rerun all steps', rerunFailed: 'Rerun failed steps only', rerunHint: 'Create a brand-new run with the same YAML',
    rerunFailedHint: 'Skip succeeded steps and continue from the failure (planned in v1)', startRerun: 'Start rerun', rerunCreated: 'Rerun created',
    exportLogs: 'Export logs', logsExported: 'Logs exported', viewYaml: 'View raw definition', runDetail: 'Run detail',
    timing: 'Duration {d}', meta: 'Metadata', execStructure: 'Execution structure', realtime: 'Streaming', autoScroll: 'Auto-scroll',
    lines: '{n} lines', waiting: '(step has not started yet)', noLogs: '(no log output)',
    logsCopied: 'Logs copied', config: 'Config', env: 'environment', workdir: 'workdir', timeout: 'timeout',
    image: 'image', trigger: 'Trigger', stepNotStarted: 'This step has not started yet',
  },
  queue: {
    title: 'Queue', running: 'Running', pending: 'Queued', maxConcurrent: 'Max concurrency', executor: 'Executor',
    usage: '{cur}/{max} used', queueActive: '{n} active tasks', enqueueAt: 'Enqueued', waitTime: 'Wait time',
    idle: 'Queue is idle', idleSub: 'No running or queued tasks', cancelTask: 'Cancel task {id}', cancelWarn: 'The task will be removed from the queue or terminated.',
  },
  settings: {
    title: 'Settings', server: 'Server', version: 'Version', uptime: 'Uptime', addr: 'Address', basePath: 'Base Path',
    auth: 'Authentication', username: 'Username', password: 'Password', set: 'set', execution: 'Execution',
    maxConcurrent: 'Max concurrency (max-concurrent)', taskTimeout: 'Default task timeout (seconds, 0 = disabled)',
    executor: 'Task executor', restartHint: 'Changes require a server restart',
    envAllow: 'Environment allowlist', allowAll: 'Allow all environment variables (allow-all-env)', add: 'Add',
    storage: 'Storage', workdir: 'workdir', maxRecords: 'Max records', configTpl: 'Config templates',
    save: 'Save settings', saved: 'Settings saved (restart required)', clearHistory: 'Clear run history',
    clearConfirm: 'Clear run history', clearWarn: 'All {n} run records will be deleted. This cannot be undone.', cleared: 'Run history cleared',
    locale: 'Language',
  },
  palette: {
    placeholder: 'Type a command or search… (↑↓ select, Enter run)', pages: 'Pages', actions: 'Actions',
    newPipeline: 'New pipeline', runPipeline: 'Run pipeline…', viewRunning: 'View running tasks',
  },
  toasts: { welcome: 'Welcome to Pipeline Web' },
  footer: { serverRunning: 'Server running · :{port}' },
};

const dicts: Record<Locale, Dict> = { zh, en };

interface I18nCtx {
  locale: Locale;
  setLocale: (l: Locale) => void;
  t: (key: string, vars?: Record<string, string | number>) => string;
}

const Ctx = createContext<I18nCtx>({
  locale: 'zh',
  setLocale: () => {},
  t: (k: string) => k,
});

export function I18nProvider({ children }: { children: ReactNode }) {
  const [locale, setLocale] = useState<Locale>(() => (localStorage.getItem('pipeline.locale') as Locale) || 'zh');

  useEffect(() => {
    localStorage.setItem('pipeline.locale', locale);
    document.documentElement.lang = locale;
  }, [locale]);

  const t = (key: string, vars?: Record<string, string | number>) => {
    const found: unknown = key.split('.').reduce(
      (o: unknown, k) => (o as Record<string, unknown> | undefined)?.[k],
      dicts[locale] as unknown,
    );
    let s: string = typeof found === 'string' ? found : key;
    if (vars) {
      for (const [k, v] of Object.entries(vars)) s = s.replaceAll(`{${k}}`, String(v));
    }
    return s;
  };

  return <Ctx.Provider value={{ locale, setLocale, t }}>{children}</Ctx.Provider>;
}

export const useI18n = () => useContext(Ctx);
