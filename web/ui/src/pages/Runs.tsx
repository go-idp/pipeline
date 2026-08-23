import { useEffect, useMemo, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { api, fmtDur, relTime, runDuration, runTrigger, shortId } from '../client';
import { ConfirmModal, Empty, I, StatusPill, useToast } from '../components';
import { useI18n } from '../i18n';
import type { PipelineRecord, RunStatus } from '../api';

const FILTERS: (RunStatus | 'all')[] = ['all', 'running', 'pending', 'succeeded', 'failed', 'cancelled'];

export default function Runs() {
  const { t } = useI18n();
  const nav = useNavigate();
  const toast = useToast();
  const [params, setParams] = useSearchParams();

  const [runs, setRuns] = useState<PipelineRecord[]>([]);
  const [q, setQ] = useState('');
  const [filter, setFilter] = useState<RunStatus | 'all'>(
    (params.get('status') as RunStatus | 'all') || 'all',
  );
  const [confirmCancel, setConfirmCancel] = useState<PipelineRecord | null>(null);

  useEffect(() => {
    let alive = true;
    const load = () =>
      api.listRuns({ limit: '200' })
        .then((r) => alive && setRuns(r.data))
        .catch(() => {});
    load();
    const timer = setInterval(load, 4000);
    return () => {
      alive = false;
      clearInterval(timer);
    };
  }, []);

  const counts = useMemo(() => {
    const c: Record<string, number> = { all: runs.length };
    for (const f of FILTERS) if (f !== 'all') c[f] = runs.filter((r) => r.status === f).length;
    return c;
  }, [runs]);

  const list = runs.filter((r) => {
    if (filter !== 'all' && r.status !== filter) return false;
    if (q && !r.name.toLowerCase().includes(q.toLowerCase()) && !r.id.toLowerCase().includes(q.toLowerCase())) return false;
    return true;
  });

  const setFilterAndParam = (f: RunStatus | 'all') => {
    setFilter(f);
    setParams(f === 'all' ? {} : { status: f }, { replace: true });
  };

  const cancel = async (r: PipelineRecord) => {
    try {
      await api.cancelRun(r.id);
      toast(`${t('runs.cancelled')} #${shortId(r.id)}`);
      setRuns((rs) => rs.map((x) => (x.id === r.id ? { ...x, status: 'cancelled' } : x)));
    } catch (e) {
      toast(String((e as Error).message), false);
    }
  };

  const rerun = async (r: PipelineRecord) => {
    try {
      const res = await api.rerun(r.id);
      toast(t('runs.rerunCreated'));
      nav(`/runs/${res.id}`);
    } catch (e) {
      toast(String((e as Error).message), false);
    }
  };

  return (
    <>
      <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 16, flexWrap: 'wrap' }}>
        <div className="seg">
          {FILTERS.map((f) => (
            <button key={f} className={filter === f ? 'on' : ''} onClick={() => setFilterAndParam(f)}>
              {f === 'all' ? t('common.all') : t(`status.${f}`)} {counts[f] ?? 0}
            </button>
          ))}
        </div>
        <div style={{ flex: 1 }} />
        <input
          className="inp"
          id="page-search"
          style={{ width: 220 }}
          placeholder={t('common.search')}
          onChange={(e) => setQ(e.target.value)}
        />
      </div>

      <div className="card">
        <table className="tbl">
          <thead>
            <tr>
              <th>{t('common.pipeline')}</th>
              <th>{t('common.runId')}</th>
              <th>{t('common.status')}</th>
              <th>{t('common.duration')}</th>
              <th>{t('common.trigger')}</th>
              <th className="num">{t('common.startedAt')}</th>
              <th style={{ textAlign: 'right' }}>{t('common.actions')}</th>
            </tr>
          </thead>
          <tbody>
            {list.length === 0 && (
              <tr>
                <td colSpan={7}>
                  <Empty
                    icon={I('runs')}
                    title={q ? t('runs.noMatch') : t('runs.noRuns')}
                    sub={q ? t('runs.noMatchSub') : t('runs.noRunsSub')}
                  />
                </td>
              </tr>
            )}
            {list.map((r) => (
              <tr key={r.id} onClick={() => nav(`/runs/${r.id}`)}>
                <td>
                  <div style={{ fontWeight: 500 }}>{r.name}</div>
                  {r.error && (
                    <div style={{ fontSize: 11.5, color: 'var(--red)', maxWidth: 260, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                      {r.error}
                    </div>
                  )}
                </td>
                <td style={{ fontFamily: 'var(--mono)', color: 'var(--text-2)' }}>#{shortId(r.id)}</td>
                <td><StatusPill status={r.status} t={t} /></td>
                <td style={{ fontVariantNumeric: 'tabular-nums' }}>
                  {runDuration(r) !== null ? fmtDur(runDuration(r)) : r.status === 'running' ? '…' : '—'}
                </td>
                <td style={{ color: 'var(--text-2)' }}>{t(`trigger.${runTrigger(r)}`)}</td>
                <td className="num" style={{ color: 'var(--text-3)' }}>
                  {new Date(r.started_at).toLocaleTimeString()} · {relTime(r.started_at)}
                </td>
                <td>
                  <div className="row-actions" onClick={(e) => e.stopPropagation()}>
                    {(r.status === 'running' || r.status === 'pending') && (
                      <button className="btn btn-sm" onClick={() => setConfirmCancel(r)}>
                        {I('stop')}
                        {t('common.cancel')}
                      </button>
                    )}
                    {(r.status === 'succeeded' || r.status === 'failed' || r.status === 'cancelled') && (
                      <button className="btn btn-sm" onClick={() => rerun(r)}>
                        {I('rr')}
                        {t('common.rerun')}
                      </button>
                    )}
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {confirmCancel && (
        <ConfirmModal
          title={t('runs.cancelRun', { id: `#${shortId(confirmCancel.id)}` })}
          message={t('runs.cancelWarn')}
          onConfirm={() => cancel(confirmCancel)}
          onClose={() => setConfirmCancel(null)}
        />
      )}
    </>
  );
}
