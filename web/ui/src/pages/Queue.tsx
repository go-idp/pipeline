import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { api, fmtDur, relTime, shortId } from '../client';
import { ConfirmModal, Empty, I, StatusPill, useToast } from '../components';
import { useI18n } from '../i18n';
import type { QueueItem, QueueStats } from '../api';

export default function Queue() {
  const { t } = useI18n();
  const nav = useNavigate();
  const toast = useToast();

  const [stats, setStats] = useState<QueueStats | null>(null);
  const [items, setItems] = useState<QueueItem[]>([]);
  const [confirmCancel, setConfirmCancel] = useState<QueueItem | null>(null);

  useEffect(() => {
    let alive = true;
    const load = async () => {
      try {
        const [s, q] = await Promise.all([api.queueStats(), api.listQueue()]);
        if (!alive) return;
        setStats(s);
        setItems(q.data);
      } catch {
        /* ignore */
      }
    };
    load();
    const timer = setInterval(load, 2000);
    return () => {
      alive = false;
      clearInterval(timer);
    };
  }, []);

  const cancel = async (item: QueueItem) => {
    try {
      await api.cancelRun(item.id);
      toast(`${t('runs.cancelled')} #${shortId(item.id)}`);
    } catch (e) {
      toast(String((e as Error).message), false);
    }
  };

  const running = items.filter((i) => i.status === 'running');
  const pending = items.filter((i) => i.status === 'pending');
  const active = [...running, ...pending];

  return (
    <>
      <div className="stat-grid">
        <div className="stat">
          <div className="lb">{t('queue.running')}</div>
          <div className="vl" style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            {stats?.running ?? 0}
            {(stats?.running ?? 0) > 0 && <span className="dot blue" style={{ width: 8, height: 8 }} />}
          </div>
          <div className="sb">{t('queue.usage', { cur: stats?.current_concurrent ?? 0, max: stats?.max_concurrent ?? 2 })}</div>
        </div>
        <div className="stat">
          <div className="lb">{t('queue.pending')}</div>
          <div className="vl">{stats?.pending ?? 0}</div>
          <div className="sb">{pending[0] ? relTime(pending[pending.length - 1].created_at) : '—'}</div>
        </div>
        <div className="stat">
          <div className="lb">{t('queue.maxConcurrent')}</div>
          <div className="vl">{stats?.max_concurrent ?? 2}</div>
        </div>
        <div className="stat">
          <div className="lb">{t('queue.executor')}</div>
          <div className="vl" style={{ fontSize: 18, lineHeight: 1.4 }}>in-process</div>
        </div>
      </div>

      <div className="card" style={{ marginBottom: 14 }}>
        <div className="card-h">
          concurrency
          <span className="sub">max-concurrent = {stats?.max_concurrent ?? 2}</span>
          <span style={{ flex: 1 }} />
          <span style={{ fontSize: 12, color: 'var(--text-2)' }}>
            {stats?.current_concurrent ?? 0}/{stats?.max_concurrent ?? 2}
          </span>
        </div>
        <div className="card-b" style={{ padding: '14px 16px' }}>
          <div className="prog">
            <i style={{ width: `${((stats?.current_concurrent ?? 0) / Math.max(1, stats?.max_concurrent ?? 2)) * 100}%` }} />
          </div>
        </div>
      </div>

      <div className="card">
        <div className="card-h">
          {t('nav.queue')}
          <span className="sub">{t('queue.queueActive', { n: active.length })}</span>
        </div>
        <table className="tbl">
          <thead>
            <tr>
              <th>{t('common.pipeline')}</th>
              <th>{t('common.runId')}</th>
              <th>{t('common.status')}</th>
              <th>{t('queue.enqueueAt')}</th>
              <th className="num">{t('queue.waitTime')}</th>
              <th style={{ textAlign: 'right' }}>{t('common.actions')}</th>
            </tr>
          </thead>
          <tbody>
            {active.length === 0 && (
              <tr>
                <td colSpan={6}>
                  <Empty icon={I('queue')} title={t('queue.idle')} sub={t('queue.idleSub')} />
                </td>
              </tr>
            )}
            {active.map((it) => (
              <tr key={it.id} onClick={() => nav(`/runs/${it.id}`)}>
                <td><div style={{ fontWeight: 500 }}>{it.name}</div></td>
                <td style={{ fontFamily: 'var(--mono)', color: 'var(--text-2)' }}>#{shortId(it.id)}</td>
                <td><StatusPill status={it.status} t={t} /></td>
                <td style={{ color: 'var(--text-3)' }}>{relTime(it.created_at)}</td>
                <td className="num" style={{ fontVariantNumeric: 'tabular-nums' }}>
                  {fmtDur(Math.round((Date.now() - new Date(it.created_at).getTime()) / 1000))}
                </td>
                <td>
                  <div className="row-actions" onClick={(e) => e.stopPropagation()}>
                    <button className="btn btn-sm" onClick={() => setConfirmCancel(it)}>
                      {I('stop')}
                      {t('common.cancel')}
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {confirmCancel && (
        <ConfirmModal
          title={t('queue.cancelTask', { id: `#${shortId(confirmCancel.id)}` })}
          message={t('queue.cancelWarn')}
          onConfirm={() => cancel(confirmCancel)}
          onClose={() => setConfirmCancel(null)}
        />
      )}
    </>
  );
}
