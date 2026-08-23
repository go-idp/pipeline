import { useEffect, useMemo, useRef, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { api, fmtDur, shortId } from '../client';
import { ConfirmModal, I, StatusDot, useToast } from '../components';
import { useI18n } from '../i18n';
import type { RunDetail, RunStageDef, RunStatus } from '../api';

export default function RunDetail() {
  const { id } = useParams();
  const nav = useNavigate();
  const toast = useToast();
  const { t } = useI18n();

  const [detail, setDetail] = useState<RunDetail | null>(null);
  const [err, setErr] = useState('');
  const [selKey, setSelKey] = useState<string>('');
  const [confirmCancel, setConfirmCancel] = useState(false);
  const [showYaml, setShowYaml] = useState(false);
  const [follow, setFollow] = useState(true);
  const [search, setSearch] = useState('');

  useEffect(() => {
    if (!id) return;
    let alive = true;
    const load = async () => {
      try {
        const d = await api.getRun(id);
        if (!alive) return;
        setDetail(d);
        setErr('');
        // 默认选中第一个 step
        const first = d.definition?.stages?.[0]?.jobs?.[0]?.steps?.[0];
        if (first) setSelKey((k) => k || first.id);
      } catch (e) {
        if (alive) setErr(String((e as Error).message));
      }
    };
    load();
    const timer = setInterval(load, 1500);
    return () => {
      alive = false;
      clearInterval(timer);
    };
  }, [id]);

  const isActive = detail?.status === 'running' || detail?.status === 'pending';
  const isTerminal = detail && !isActive;

  const steps = useMemo(() => {
    if (!detail?.definition) return [];
    const out: { id: string; name: string; command: string; status: RunStatus; started_at?: string; ended_at?: string }[] = [];
    for (const s of detail.definition.stages) {
      for (const j of s.jobs) {
        for (const k of j.steps) {
          const st = detail.states?.[k.id];
          out.push({
            id: k.id,
            name: k.name,
            command: k.command,
            status: k.status,
            started_at: st?.started_at,
            ended_at: st?.ended_at,
          });
        }
      }
    }
    return out;
  }, [detail]);

  // 选中 step 的日志（按时间窗口过滤）
  const selected = steps.find((s) => s.id === selKey) ?? null;
  const logs = useMemo(() => {
    if (!detail?.logs) return [];
    if (!selected) return detail.logs;
    const start = selected.started_at ? new Date(selected.started_at).getTime() : 0;
    const end = selected.ended_at ? new Date(selected.ended_at).getTime() : Infinity;
    return detail.logs.filter((l) => {
      const ts = new Date(l.timestamp).getTime();
      return ts >= start && ts <= end;
    });
  }, [detail, selected]);

  const bodyRef = useRef<HTMLDivElement>(null);
  useEffect(() => {
    if (follow && bodyRef.current) bodyRef.current.scrollTop = bodyRef.current.scrollHeight;
  }, [logs.length, follow]);

  const highlight = (msg: string) => {
    if (!search) return msg;
    const idx = msg.toLowerCase().indexOf(search.toLowerCase());
    if (idx < 0) return msg;
    return (
      <>
        {msg.slice(0, idx)}
        <mark>{msg.slice(idx, idx + search.length)}</mark>
        {msg.slice(idx + search.length)}
      </>
    );
  };

  const cancel = async () => {
    if (!detail) return;
    try {
      await api.cancelRun(detail.id);
      toast(`${t('runs.cancelled')} #${shortId(detail.id)}`);
      const d = await api.getRun(detail.id);
      setDetail(d);
    } catch (e) {
      toast(String((e as Error).message), false);
    }
  };

  const rerun = async () => {
    if (!detail) return;
    try {
      const res = await api.rerun(detail.id);
      toast(t('runs.rerunCreated'));
      nav(`/runs/${res.id}`);
    } catch (e) {
      toast(String((e as Error).message), false);
    }
  };

  const exportLogs = () => {
    if (!detail) return;
    const lines: string[] = [
      `Pipeline: ${detail.name} (#${detail.id})`,
      `Status: ${detail.status}  Duration: ${fmtDur(detail.duration)}`,
      `Exported at: ${new Date().toISOString()}`,
      '='.repeat(70),
    ];
    for (const l of detail.logs ?? []) {
      lines.push(`[${new Date(l.timestamp).toISOString()}] [${l.type}] ${l.message}`);
    }
    const blob = new Blob([lines.join('\n')], { type: 'text/plain;charset=utf-8' });
    const a = document.createElement('a');
    a.href = URL.createObjectURL(blob);
    a.download = `pipeline-${detail.id}-logs.txt`;
    a.click();
    URL.revokeObjectURL(a.href);
    toast(t('runs.logsExported'));
  };

  if (err && !detail) {
    return (
      <div className="empty">
        <div className="et">{err}</div>
        <button className="btn" onClick={() => nav('/runs')}>{t('common.close')}</button>
      </div>
    );
  }
  if (!detail) return <div className="hint-line">{t('common.loading')}</div>;

  const duration = detail.duration ?? (detail.status === 'running' && detail.started_at
    ? Math.round((Date.now() - new Date(detail.started_at).getTime()) / 1000)
    : 0);

  return (
    <>
      <div style={{ display: 'flex', alignItems: 'flex-start', gap: 12, marginBottom: 14, flexWrap: 'wrap' }}>
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            <span className="badge" style={{ fontSize: 14, fontWeight: 600 }}>
              <StatusDot status={detail.status} />
              {t(`status.${detail.status}`)}
            </span>
            <h1 style={{ fontSize: 19, fontWeight: 650, letterSpacing: '-.01em' }}>
              {detail.name}
              <span style={{ color: 'var(--text-3)', fontWeight: 500, fontFamily: 'var(--mono)', fontSize: 13 }}>
                {' '}#{shortId(detail.id)}
              </span>
            </h1>
          </div>
          <div style={{ display: 'flex', gap: 14, marginTop: 6, fontSize: 12, color: 'var(--text-3)' }}>
            <span>{t(`trigger.${detail.trigger}`)}</span>
            <span>{new Date(detail.started_at).toLocaleString()}</span>
            <span>{t('runs.timing', { d: fmtDur(duration) })}</span>
          </div>
          {detail.error && (
            <div
              style={{
                display: 'flex', alignItems: 'center', gap: 8, marginTop: 8,
                color: 'var(--red)', fontSize: 12.5, background: '#fef2f2', border: '1px solid #fecaca',
                borderRadius: 6, padding: '6px 10px', fontFamily: 'var(--mono)',
              }}
            >
              {I('warn', 13)}
              {detail.error}
            </div>
          )}
        </div>
        <div style={{ flex: 1 }} />
        {isActive && (
          <button className="btn" onClick={() => setConfirmCancel(true)}>
            {I('stop')}
            {t('common.cancel')}
          </button>
        )}
        {isTerminal && (
          <button className="btn" onClick={rerun}>
            {I('rr')}
            {t('common.rerun')}
          </button>
        )}
        <button className="btn" onClick={exportLogs}>
          {I('dl')}
          {t('runs.exportLogs')}
        </button>
        <button className="btn" onClick={() => setShowYaml(true)}>{I('copy')}YAML</button>
      </div>

      {detail.definition && <StageTimeline stages={detail.definition.stages} onSelect={setSelKey} />}

      <div style={{ display: 'grid', gridTemplateColumns: '320px 1fr', gap: 14, alignItems: 'start', marginTop: 14 }}>
        <StepTree stages={detail.definition?.stages ?? []} selKey={selKey} onSelect={setSelKey} />
        <div className="log">
          <div className="log-h">
            <span className="lt">{selected ? selected.name : t('common.loading')}</span>
            <span className="lsub">{selected?.command ?? ''}</span>
            <span className="sp" />
            <div className="srch">
              {I('search')}
              <input
                placeholder="search"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
              />
            </div>
            <button
              className="btn btn-icon btn-sm"
              title="copy"
              onClick={() => {
                navigator.clipboard?.writeText(logs.map((l) => l.message).join('\n'));
                toast(t('runs.logsCopied'));
              }}
            >
              {I('copy')}
            </button>
          </div>
          <div className="log-b" ref={bodyRef}>
            {logs.length === 0 && (
              <div className="ll">
                <span className="txt" style={{ color: 'var(--text-3)' }}>
                  {selected?.status === 'pending' ? t('runs.waiting') : t('runs.noLogs')}
                </span>
              </div>
            )}
            {logs.map((l, i) => (
              <div className="ll" key={i}>
                <span className="ln">{i + 1}</span>
                <span className="txt" style={{ color: l.type === 'stderr' ? '#b91c1c' : undefined }}>
                  {highlight(l.message)}
                </span>
              </div>
            ))}
          </div>
          <div className="log-foot">
            <span>{t('runs.lines', { n: logs.length })}</span>
            {detail.status === 'running' && (
              <span style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                <span className="dot blue" />
                {t('runs.realtime')}
              </span>
            )}
            <span style={{ flex: 1 }} />
            <span className={`switch ${follow ? 'on' : ''}`} onClick={() => setFollow((v) => !v)}>
              <i />
              {t('runs.autoScroll')}
            </span>
          </div>
        </div>
      </div>

      <div className="card" style={{ marginTop: 14 }}>
        <div className="card-h">
          {t('runs.meta')}
          <span className="sub">config / environment / yaml</span>
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0 30px', padding: '12px 16px' }}>
          <div className="kv"><div className="k">run id</div><div className="v">{detail.id}</div></div>
          <div className="kv"><div className="k">{t('common.pipeline')}</div><div className="v">{detail.name}</div></div>
          <div className="kv"><div className="k">{t('runs.workdir')}</div><div className="v">/tmp/go-idp/pipeline/{detail.id}</div></div>
          <div className="kv"><div className="k">{t('runs.timeout')}</div><div className="v">{(detail.config?.timeout as number) ?? 0}s</div></div>
          <div className="kv"><div className="k">{t('runs.image')}</div><div className="v">{(detail.config?.image as string) || '—'}</div></div>
          <div className="kv"><div className="k">{t('runs.trigger')}</div><div className="v plain">{t(`trigger.${detail.trigger}`)}</div></div>
          <div className="kv" style={{ borderBottom: 0 }}><div className="k">yaml</div><div className="v plain"><span style={{ cursor: 'pointer', textDecoration: 'underline' }} onClick={() => setShowYaml(true)}>{t('runs.viewYaml')}</span></div></div>
        </div>
      </div>

      {confirmCancel && detail && (
        <ConfirmModal
          title={t('runs.cancelRun', { id: `#${shortId(detail.id)}` })}
          message={t('runs.cancelWarn')}
          onConfirm={cancel}
          onClose={() => setConfirmCancel(false)}
        />
      )}

      {showYaml && detail?.yaml && (
        <div className="overlay" onClick={(e) => e.target === e.currentTarget && setShowYaml(false)}>
          <div className="modal m-lg">
            <div className="m-h">
              <div className="mt">YAML</div>
              <button className="btn btn-icon" style={{ marginLeft: 'auto' }} onClick={() => setShowYaml(false)}>{I('x')}</button>
            </div>
            <div className="m-b" style={{ paddingTop: 10 }}>
              <div className="codeblock" style={{ maxHeight: '60vh' }}>{detail.yaml}</div>
            </div>
            <div className="m-f">
              <button className="btn btn-primary" onClick={() => setShowYaml(false)}>{t('common.close')}</button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}

/* ---------- stage timeline ---------- */
function StageTimeline({ stages, onSelect }: { stages: RunStageDef[]; onSelect: (stepId: string) => void }) {
  const total = stages.length || 1;
  return (
    <>
      <div className="tl">
        {stages.map((s) => (
          <div
            key={s.id}
            className={`tl-seg s-${s.status}`}
            style={{ flex: 1 }}
            title={s.name}
            onClick={() => {
              const first = s.jobs[0]?.steps[0];
              if (first) onSelect(first.id);
            }}
          >
            {s.status === 'running' && <span className="dot" style={{ background: '#fff', width: 5, height: 5 }} />}
            {s.name}
          </div>
        ))}
        {stages.length === 0 && <div className="tl-seg s-pending" style={{ flex: 1 }}>—</div>}
      </div>
      <div className="tl-cap">
        {stages.map((s) => (
          <div key={s.id}>
            <StatusDot status={s.status} />
            <span className="nm">{s.name}</span>
            <span className="dm">{s.status === 'pending' ? 'wait' : s.status}</span>
          </div>
        ))}
        <span style={{ marginLeft: 'auto', color: 'var(--text-3)' }}>{total} stages</span>
      </div>
    </>
  );
}

/* ---------- step tree ---------- */
function StepTree({ stages, selKey, onSelect }: {
  stages: RunStageDef[]; selKey: string; onSelect: (stepId: string) => void;
}) {
  const [open, setOpen] = useState<Record<string, boolean>>({});

  const toggle = (id: string) => setOpen((o) => ({ ...o, [id]: !o[id] }));

  return (
    <div className="tree">
      <div className="tree-h">
        {I('layers')}
        structure
        <span className="sub">stage / job / step</span>
      </div>
      <div className="tree-b">
        {stages.length === 0 && <div className="hint-line" style={{ padding: 8 }}>—</div>}
        {stages.map((s, si) => {
          const isOpen = open[s.id] !== false;
          return (
            <div key={s.id}>
              <div className={`tr l1 ${isOpen ? 'open' : ''}`} onClick={() => toggle(s.id)}>
                <span className="chev">{I('chev')}</span>
                <StatusDot status={s.status} />
                <span className="tname">{s.name}</span>
                <span className="chip">{s.mode}</span>
              </div>
              <div className={`kids ${isOpen ? 'open' : ''}`}>
                {s.jobs.map((j) => (
                  <div key={j.id}>
                    <div className="tr l2 indent-1">
                      <StatusDot status={j.status} />
                      <span className="tname">{j.name}</span>
                    </div>
                    {j.steps.map((k) => (
                      <div
                        key={k.id}
                        className={`tr l3 indent-2 ${k.id === selKey ? 'sel' : ''}`}
                        onClick={() => onSelect(k.id)}
                      >
                        <StatusDot status={k.status} />
                        {I('term', 13)}
                        <span className="tname">{k.name}</span>
                      </div>
                    ))}
                  </div>
                ))}
                {si < stages.length - 1 && <div style={{ height: 4 }} />}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
