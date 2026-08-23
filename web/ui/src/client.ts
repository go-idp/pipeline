import type {
  ConfigListResp, ConfigTemplate, PipelineRecord, QueueItem, QueueStats,
  Template,
  RunDetail, RunListResp, RunStatus, ServerInfo, Settings,
} from './api';

const API = '/api/v1';

async function request<T>(path: string, opts?: RequestInit): Promise<T> {
  const res = await fetch(API + path, {
    headers: { 'Content-Type': 'application/json' },
    ...opts,
  });
  if (!res.ok) {
    let msg = `${res.status} ${res.statusText}`;
    try {
      const body = await res.json();
      if (body?.error) msg = body.error;
    } catch {
      /* ignore */
    }
    throw new Error(msg);
  }
  return res.json() as Promise<T>;
}

export const api = {
  // runs
  listRuns: (params: Record<string, string> = {}) => {
    const q = new URLSearchParams(params).toString();
    return request<RunListResp>(`/runs${q ? `?${q}` : ''}`);
  },
  getRun: (id: string) => request<RunDetail>(`/runs/${id}`),
  createRun: (config: string, trigger = 'manual') =>
    request<{ id: string; name: string; status: string }>(`/runs`, {
      method: 'POST',
      body: JSON.stringify({ config, trigger }),
    }),
  rerun: (id: string) => request<{ id: string }>(`/runs/${id}/rerun`, { method: 'POST' }),
  cancelRun: (id: string) =>
    request<{ message: string }>(`/pipelines/${id}/cancel`, { method: 'POST' }),
  deleteRun: (id: string) =>
    request<{ message: string }>(`/pipelines/${id}`, { method: 'DELETE' }),

  // templates (built-in examples)
  listTemplates: () => request<{ data: Template[]; total: number }>(`/templates`),

  // config templates (pipelines)
  listConfigs: () => request<ConfigListResp>(`/configs`),
  getConfig: (id: string) => request<ConfigTemplate>(`/configs/${id}`),
  createConfig: (body: { name: string; description?: string; yaml: string }) =>
    request<ConfigTemplate>(`/configs`, { method: 'POST', body: JSON.stringify(body) }),
  updateConfig: (id: string, body: { name: string; description?: string; yaml: string }) =>
    request<ConfigTemplate>(`/configs/${id}`, { method: 'PUT', body: JSON.stringify(body) }),
  deleteConfig: (id: string) =>
    request<{ message: string }>(`/configs/${id}`, { method: 'DELETE' }),

  // queue
  queueStats: () => request<QueueStats>(`/queue/stats`),
  listQueue: () => request<{ data: QueueItem[]; total: number }>(`/queue`),

  // settings
  getSettings: () => request<Settings>(`/settings`),
  saveSettings: (settings: Record<string, unknown>) =>
    request<{ message: string }>(`/settings`, { method: 'POST', body: JSON.stringify(settings) }),
  // 根路径版本信息（不在 /api/v1 前缀下）
  serverInfo: () => fetch(`/`).then((r) => r.json() as Promise<ServerInfo>),
};

// 状态 -> 展示
export const STATUS_ORDER: RunStatus[] = ['running', 'pending', 'succeeded', 'failed', 'cancelled'];

export function fmtDur(sec?: number | null): string {
  if (sec === undefined || sec === null || Number.isNaN(sec)) return '—';
  if (sec < 60) return `${sec}s`;
  if (sec < 3600) return `${Math.floor(sec / 60)}m ${Math.round(sec % 60)}s`;
  return `${Math.floor(sec / 3600)}h ${Math.round((sec % 3600) / 60)}m`;
}

export function fmtClock(t?: string): string {
  if (!t) return '—';
  const d = new Date(t);
  if (Number.isNaN(d.getTime())) return t;
  const p = (n: number) => String(n).padStart(2, '0');
  return `${p(d.getHours())}:${p(d.getMinutes())}:${p(d.getSeconds())}`;
}

export function relTime(t?: string): string {
  if (!t) return '—';
  const s = (Date.now() - new Date(t).getTime()) / 1000;
  if (s < 60) return 'just now';
  if (s < 3600) return `${Math.floor(s / 60)}m ago`;
  if (s < 86400) return `${Math.floor(s / 3600)}h ago`;
  return `${Math.floor(s / 86400)}d ago`;
}

export function shortId(id: string): string {
  return id.length > 9 ? id.slice(0, 9) : id;
}

export function runDuration(r: PipelineRecord): number | null {
  const start = new Date(r.started_at).getTime();
  if (Number.isNaN(start)) return null;
  if (r.status === 'running') return Math.round((Date.now() - start) / 1000);
  const end = r.succeed_at || r.failed_at || r.cancelled_at;
  if (!end) return null;
  const e = new Date(end).getTime();
  if (Number.isNaN(e)) return null;
  return Math.max(0, Math.round((e - start) / 1000));
}

export function runTrigger(r: PipelineRecord | { config?: Record<string, unknown> }): string {
  const v = r.config?.trigger;
  return typeof v === 'string' ? v : 'manual';
}

// 轻量 YAML 结构解析（用于可视化预览；完整编辑在 YAML 页）
export interface LiteStage { name: string; mode: string; jobs: { name: string; steps: { name: string; cmd: string }[] }[] }
export function parseYamlLite(yaml: string): LiteStage[] {
  const out: LiteStage[] = [];
  let curS: LiteStage | null = null;
  let curJ: { name: string; steps: { name: string; cmd: string }[] } | null = null;
  const lines = yaml.split('\n');
  for (const raw of lines) {
    const line = raw.replace(/\s+$/, '');
    const t = line.trim();
    if (!t || t.startsWith('#')) continue;
    const indent = line.length - line.trimStart().length;
    let m: RegExpMatchArray | null;
    if ((m = t.match(/^-\s*name:\s*(.*)$/))) {
      const nm = m[1].replace(/^['"]|['"]$/g, '').trim();
      if (indent === 2) {
        curS = { name: nm, mode: 'parallel', jobs: [] };
        out.push(curS);
        curJ = null;
      } else if (indent === 6 && curS) {
        curJ = { name: nm, steps: [] };
        curS.jobs.push(curJ);
      } else if (indent === 10 && curJ) {
        curJ.steps.push({ name: nm, cmd: '' });
      }
    } else if ((m = t.match(/^(name|run_mode|mode|command|type|description):\s*(.*)$/))) {
      const k = m[1];
      const v = m[2].replace(/^['"]|['"]$/g, '').trim();
      if (k === 'name' && indent === 0) {
        /* pipeline name */
      } else if ((k === 'run_mode' || k === 'mode') && indent === 4 && curS) curS.mode = v;
      else if (k === 'command' && indent === 12 && curJ && curJ.steps.length) curJ.steps[curJ.steps.length - 1].cmd = v;
    }
  }
  return out;
}

// 把 KEY=VALUE 行追加为 pipeline environment
export function appendEnv(yaml: string, envText: string): string {
  const lines = envText.split('\n').map((l) => l.trim()).filter((l) => l && !l.startsWith('#'));
  if (lines.length === 0) return yaml;

  const envLines = lines.map((l) => {
    const [k, ...rest] = l.split('=');
    const v = rest.join('=');
    return `  ${k}: "${v}"`;
  });

  // 在 name: 行之后插入 environment 块
  const idx = yaml.indexOf('\n');
  const head = idx >= 0 ? yaml.slice(0, idx) : yaml;
  const rest = idx >= 0 ? yaml.slice(idx) : '';
  return `${head}\nenvironment:\n${envLines.join('\n')}${rest}`;
}

export function escapeHtml(s: string): string {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}
