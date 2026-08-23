// API 类型定义（与 svc/server 的 JSON 响应对齐）

export type RunStatus = 'pending' | 'running' | 'succeeded' | 'failed' | 'cancelled';
export type LogType = 'stdout' | 'stderr';

export interface LogEntry {
  type: LogType;
  message: string;
  timestamp: string;
}

export interface StageState {
  id: string;
  level: 'pipeline' | 'stage' | 'job' | 'step';
  name: string;
  status: RunStatus;
  error?: string;
  started_at: string;
  ended_at?: string;
}

export interface PipelineRecord {
  id: string;
  name: string;
  status: RunStatus;
  started_at: string;
  succeed_at?: string;
  failed_at?: string;
  cancelled_at?: string;
  error?: string;
  config?: Record<string, unknown>;
  yaml?: string;
  logs?: LogEntry[];
  stages_state?: Record<string, StageState>;
}

export interface RunListResp {
  data: PipelineRecord[];
  total: number;
  limit: number;
  offset: number;
}

export interface RunStepDef {
  id: string;
  name: string;
  command: string;
  status: RunStatus;
}

export interface RunJobDef {
  id: string;
  name: string;
  status: RunStatus;
  steps: RunStepDef[];
}

export interface RunStageDef {
  id: string;
  name: string;
  mode: string;
  status: RunStatus;
  jobs: RunJobDef[];
}

export interface RunDefinition {
  stages: RunStageDef[];
}

export interface RunDetail {
  id: string;
  name: string;
  status: RunStatus;
  trigger: string;
  started_at: string;
  ended_at?: string;
  duration: number;
  error?: string;
  yaml?: string;
  config?: Record<string, unknown>;
  definition?: RunDefinition;
  states?: Record<string, StageState>;
  logs?: LogEntry[];
}

export interface ConfigTemplate {
  id: string;
  name: string;
  description?: string;
  yaml: string;
  created_at: string;
  updated_at: string;
}

export interface Template {
  id: string;
  name: string;
  description: string;
  yaml: string;
}

export interface ConfigListResp {
  data: ConfigTemplate[];
  total: number;
}

export interface QueueItem {
  id: string;
  name: string;
  status: RunStatus;
  trigger?: string;
  created_at: string;
  started_at?: string;
  ended_at?: string;
  error?: string;
}

export interface QueueStats {
  total: number;
  pending: number;
  running: number;
  succeeded: number;
  failed: number;
  cancelled: number;
  max_concurrent: number;
  current_concurrent: number;
}

export interface QueueListResp {
  data: QueueItem[];
  total: number;
}

export interface Settings {
  max_concurrent: number;
  max_records: number;
  refresh_interval: number;
}

export interface ServerInfo {
  version: string;
  running_at: string;
}
