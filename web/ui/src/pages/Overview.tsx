import { useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { api, fmtDur, relTime, runDuration, shortId } from '../client';
import { I, StatusDot, usePolling } from '../components';
import { useI18n } from '../i18n';
import type { PipelineRecord } from '../api';
import { useRunModal } from '../App';

export default function Overview() {
  const { t } = useI18n();
  const nav = useNavigate();
  const runs = usePolling<PipelineRecord[]>(async () => {
    const res = await api.listRuns({ limit: '200' });
    return res.data;
  }, [], 3000);

  const stats = useMemo(() => {
    const list = runs ?? [];
    const day = 86400000;
    const now = Date.now();
    const todayStart = new Date();
    todayStart.setHours(0, 0, 0, 0);
    const today = list.filter((r) => new Date(r.started_at).getTime() >= todayStart.getTime()).length;
    const last7 = list.filter((r) => new Date(r.started_at).getTime() >= now - 7 * day);
    const ok7 = last7.filter((r) => r.status === 'succeeded').length;
    const rate = last7.length ? Math.round((ok7 / last7.length) * 100) : 0;
    const durs = last7.map(runDuration).filter((d): d is number => d !== null);
    const avg = durs.length ? Math.round(durs.reduce((a, b) => a + b, 0) / durs.length) : 0;
    const running = list.filter((r) => r.status === 'running').length;

    const days: { label: string; ok: number; fail: number }[] = [];
    for (let i = 13; i >= 0; i--) {
      const d = new Date();
      d.setHours(0, 0, 0, 0);
      d.setDate(d.getDate() - i);
      const s = d.getTime();
      const rr = list.filter((r) => {
        const ts = new Date(r.started_at).getTime();
        return ts >= s && ts < s + day;
      });
      days.push({
        label: `${d.getMonth() + 1}/${d.getDate()}`,
        ok: rr.filter((r) => r.status === 'succeeded').length,
        fail: rr.filter((r) => r.status === 'failed').length,
      });
    }

    const dist = (['pending', 'running', 'succeeded', 'failed', 'cancelled'] as const).map((s) => ({
      status: s,
      n: list.filter((r) => r.status === s).length,
    }));

    return { today, rate, ok7, last7: last7.length, avg, running, days, dist, list };
  }, [runs]);

  const maxN = Math.max(1, ...stats.days.map((d) => d.ok + d.fail));
  const recent = stats.list.slice(0, 8);

  return (
    <>
      <div className="stat-grid">
        <div className="stat">
          <div className="lb">{t('overview.todayRuns')}</div>
          <div className="vl">{stats.today}<small>{t('overview.times')}</small></div>
        </div>
        <div className="stat">
          <div className="lb">{t('overview.rate7')}</div>
          <div className="vl">{stats.rate}<small>%</small></div>
          <div className="sb">{t('overview.okRate', { n: stats.ok7 })}</div>
        </div>
        <div className="stat">
          <div className="lb">{t('overview.avgDur')}</div>
          <div className="vl">{fmtDur(stats.avg)}</div>
          <div className="sb">{t('overview.avg7', { n: stats.last7 })}</div>
        </div>
        <div className="stat">
          <div className="lb">{t('overview.runningNow')}</div>
          <div className="vl" style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            {stats.running}
            {stats.running > 0 && <span className="dot blue" style={{ width: 8, height: 8 }} />}
          </div>
          <div className="sb">{t('overview.concurrency', { n: 2 })}</div>
        </div>
      </div>

      <div className="ov-grid">
        <div className="card">
          <div className="card-h">{t('overview.trend14')}</div>
          <div className="card-b">
            <div className="bars">
              {stats.days.map((d) => (
                <div
                  className="bar-wrap"
                  key={d.label}
                  title={`${d.label}: ok ${d.ok} / fail ${d.fail}`}
                >
                  <div className="bar">
                    <i className="ok" style={{ height: `${(d.ok / maxN) * 100}%` }} />
                    <i className="fail" style={{ height: `${(d.fail / maxN) * 100}%` }} />
                  </div>
                  <div className="dl">{d.label}</div>
                </div>
              ))}
            </div>
          </div>
        </div>
        <div className="card">
          <div className="card-h">{t('overview.dist')}</div>
          <div className="card-b">
            {stats.dist.map((d) => (
              <div className="dist-row" key={d.status}>
                <div className="dk">
                  <StatusDot status={d.status} />
                  {t(`status.${d.status}`)}
                </div>
                <div className="dist-bar">
                  <i
                    style={{
                      width: stats.list.length ? `${(d.n / stats.list.length) * 100}%` : 0,
                      background: d.status === 'running' ? '#3b82f6' : d.status === 'succeeded' ? '#16a34a' : d.status === 'failed' ? '#dc2626' : '#a1a1aa',
                    }}
                  />
                </div>
                <div className="dv">{d.n}</div>
              </div>
            ))}
          </div>
        </div>
      </div>

      <div className="ov-grid">
        <div className="card">
          <div className="card-h">
            {t('overview.recentRuns')}
            <span className="sub">view details</span>
          </div>
          <div className="card-b" style={{ padding: '6px 16px' }}>
            <div className="mini-list">
              {recent.length === 0 && <div className="hint-line" style={{ padding: '16px 0' }}>{t('common.empty')}</div>}
              {recent.map((r) => (
                <div className="mi" key={r.id} onClick={() => nav(`/runs/${r.id}`)}>
                  <StatusDot status={r.status} />
                  <div className="mn">
                    <div className="t">
                      {r.name}
                      <span style={{ color: 'var(--text-3)', fontWeight: 400, fontFamily: 'var(--mono)', fontSize: 11.5 }}>
                        {' '}#{shortId(r.id)}
                      </span>
                    </div>
                    <div className="s">{relTime(r.started_at)}</div>
                  </div>
                  <div className="mt">{runDuration(r) !== null ? fmtDur(runDuration(r)) : '—'}</div>
                </div>
              ))}
            </div>
          </div>
        </div>
        <div className="card">
          <div className="card-h">{t('overview.quickPipelines')}</div>
          <div className="card-b" style={{ padding: '6px 16px' }}>
            <QuickPipelines />
          </div>
        </div>
      </div>
    </>
  );
}

function QuickPipelines() {
  const { t } = useI18n();
  const nav = useNavigate();
  const { openRunModal } = useRunModal();
  const configs = usePolling(async () => (await api.listConfigs()).data, [], 10000);

  if (!configs) return <div className="hint-line" style={{ padding: '16px 0' }}>{t('common.loading')}</div>;

  return (
    <div className="mini-list">
      {configs.length === 0 && <div className="hint-line" style={{ padding: '16px 0' }}>{t('common.empty')}</div>}
      {configs.slice(0, 5).map((c) => (
        <div className="mi" key={c.id} onClick={() => nav(`/pipelines/${c.id}`)}>
          {I('layers')}
          <div className="mn">
            <div className="t">{c.name}</div>
            <div className="s">{t('pipelines.stagesCount', { n: stagesOf(c.yaml).length })}</div>
          </div>
          <button
            className="btn btn-sm"
            style={{ flex: 'none' }}
            onClick={(e) => {
              e.stopPropagation();
              openRunModal([{ name: c.name, yaml: c.yaml, id: c.id }]);
            }}
          >
            {I('play')}
            {t('common.run')}
          </button>
        </div>
      ))}
    </div>
  );
}

function stagesOf(yaml: string): string[] {
  return yaml
    .split('\n')
    .filter((l) => /^\s{2}- name:/.test(l))
    .map((l) => l.replace(/^\s*-\s*name:\s*/, '').trim());
}
