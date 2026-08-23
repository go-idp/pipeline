import { useEffect, useMemo, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { api, fmtDur, parseYamlLite, relTime, runDuration, runTrigger, shortId } from '../client';
import { ConfirmModal, Empty, I, StatusPill, useToast } from '../components';
import { useI18n } from '../i18n';
import type { ConfigTemplate, PipelineRecord } from '../api';
import { useRunModal } from '../App';

export default function PipelineDetail() {
  const { id } = useParams();
  const nav = useNavigate();
  const toast = useToast();
  const { t } = useI18n();
  const { openRunModal } = useRunModal();

  const [cfg, setCfg] = useState<ConfigTemplate | null>(null);
  const [runs, setRuns] = useState<PipelineRecord[]>([]);
  const [showYaml, setShowYaml] = useState(false);
  const [confirmDel, setConfirmDel] = useState(false);

  useEffect(() => {
    if (!id) return;
    api.getConfig(id).then(setCfg).catch((e) => toast(String((e as Error).message), false));
  }, [id, toast]);

  const history = useMemo(
    () => runs.filter((r) => cfg && r.name === cfg.name),
    [runs, cfg],
  );

  useEffect(() => {
    if (!cfg) return;
    let alive = true;
    const load = () =>
      api.listRuns({ limit: '200' }).then((r) => {
        if (alive) setRuns(r.data);
      });
    load();
    const timer = setInterval(load, 5000);
    return () => {
      alive = false;
      clearInterval(timer);
    };
  }, [cfg]);

  if (!cfg) return <div className="hint-line">{t('common.loading')}</div>;

  const visual = parseYamlLite(cfg.yaml);

  const del = async () => {
    try {
      await api.deleteConfig(cfg.id);
      toast(t('pipelines.deleted'));
      nav('/pipelines');
    } catch (e) {
      toast(String((e as Error).message), false);
    }
  };

  return (
    <>
      <div style={{ display: 'flex', alignItems: 'flex-start', gap: 12, marginBottom: 18, flexWrap: 'wrap' }}>
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            <h1 style={{ fontSize: 19, fontWeight: 650, letterSpacing: '-.01em' }}>{cfg.name}</h1>
          </div>
          {cfg.description && <div style={{ color: 'var(--text-2)', marginTop: 4, fontSize: 13 }}>{cfg.description}</div>}
          <div style={{ display: 'flex', gap: 14, marginTop: 8, fontSize: 12, color: 'var(--text-3)' }}>
            <span>{t('pipelines.stagesCount', { n: visual.length })}</span>
            <span>{t('pipelines.runCount')} {history.length}</span>
            <span>
              ID <span style={{ fontFamily: 'var(--mono)' }}>{cfg.id}</span>
            </span>
          </div>
        </div>
        <div style={{ flex: 1 }} />
        <button
          className="btn btn-primary"
          onClick={() => openRunModal([{ name: cfg.name, yaml: cfg.yaml, id: cfg.id }])}
        >
          {I('play')}
          {t('common.run')}
        </button>
        <button className="btn" onClick={() => nav(`/pipelines/${cfg.id}/edit`)}>{t('common.edit')}</button>
        <button className="btn" onClick={() => setShowYaml(true)}>{I('copy')}YAML</button>
        <button className="btn btn-danger" onClick={() => setConfirmDel(true)}>{I('trash')}</button>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1.5fr 1fr', gap: 14, marginBottom: 18 }}>
        <div className="card">
          <div className="card-h">{t('pipelines.definition')}</div>
          <div className="card-b" style={{ padding: 14 }}>
            {visual.length === 0 && <div className="hint-line">{t('common.empty')}</div>}
            {visual.map((s, si) => (
              <div className="vstage" key={si}>
                <div className="vstage-h">
                  <span className="dot" style={{ background: 'var(--text)' }} />
                  {s.name}
                  <span className="chip" style={{ marginLeft: 'auto' }}>{s.mode}</span>
                </div>
                {s.jobs.map((j, ji) => (
                  <div key={ji}>
                    <div className="vjob">
                      {I('layers')}
                      <span style={{ fontWeight: 500 }}>{j.name}</span>
                    </div>
                    {j.steps.map((k, ki) => (
                      <div className="vstep" key={ki}>
                        {I('term')}
                        <span>{k.name}</span>
                        <span className="cmd">{k.cmd}</span>
                      </div>
                    ))}
                  </div>
                ))}
              </div>
            ))}
          </div>
        </div>
        <div className="card">
          <div className="card-h">{t('pipelines.config')}</div>
          <div className="card-b" style={{ padding: '10px 16px' }}>
            <div className="kv"><div className="k">yaml</div><div className="v" style={{ cursor: 'pointer', textDecoration: 'underline' }} onClick={() => setShowYaml(true)}>{cfg.yaml.length} chars</div></div>
            <div className="kv"><div className="k">created</div><div className="v plain">{relTime(cfg.created_at)}</div></div>
            <div className="kv" style={{ borderBottom: 0 }}><div className="k">updated</div><div className="v plain">{relTime(cfg.updated_at)}</div></div>
          </div>
        </div>
      </div>

      <div className="card">
        <div className="card-h">{t('pipelines.history')}<span className="sub">{history.length}</span></div>
        <table className="tbl">
          <thead>
            <tr>
              <th>{t('common.runId')}</th>
              <th>{t('common.status')}</th>
              <th>{t('common.trigger')}</th>
              <th>{t('common.duration')}</th>
              <th className="num">{t('common.startedAt')}</th>
            </tr>
          </thead>
          <tbody>
            {history.length === 0 && (
              <tr>
                <td colSpan={5}>
                  <Empty icon={I('runs')} title={t('runs.noRuns')} sub={t('runs.noRunsSub')} />
                </td>
              </tr>
            )}
            {history.slice(0, 30).map((r) => (
              <tr key={r.id} onClick={() => nav(`/runs/${r.id}`)}>
                <td style={{ fontFamily: 'var(--mono)', color: 'var(--text-2)' }}>#{shortId(r.id)}</td>
                <td><StatusPill status={r.status} t={t} /></td>
                <td style={{ color: 'var(--text-2)' }}>{t(`trigger.${runTrigger(r)}`)}</td>
                <td>{runDuration(r) !== null ? fmtDur(runDuration(r)) : '—'}</td>
                <td className="num" style={{ color: 'var(--text-3)' }}>{fmtClock(r.started_at)} · {relTime(r.started_at)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {showYaml && (
        <div className="overlay" onClick={(e) => e.target === e.currentTarget && setShowYaml(false)}>
          <div className="modal m-lg">
            <div className="m-h">
              <div className="mt">YAML</div>
              <button className="btn btn-icon" style={{ marginLeft: 'auto' }} onClick={() => setShowYaml(false)}>{I('x')}</button>
            </div>
            <div className="m-b" style={{ paddingTop: 10 }}>
              <div className="codeblock" style={{ maxHeight: '60vh' }}>{cfg.yaml}</div>
            </div>
            <div className="m-f">
              <button
                className="btn"
                onClick={() => {
                  navigator.clipboard?.writeText(cfg.yaml);
                  toast(t('runs.logsCopied'));
                }}
              >
                {I('copy')}Copy
              </button>
              <button className="btn btn-primary" onClick={() => setShowYaml(false)}>{t('common.close')}</button>
            </div>
          </div>
        </div>
      )}

      {confirmDel && (
        <ConfirmModal
          title={t('pipelines.deleteConfirm', { name: cfg.name })}
          message={t('pipelines.deleteWarn')}
          onConfirm={del}
          onClose={() => setConfirmDel(false)}
        />
      )}
    </>
  );
}

function fmtClock(t?: string): string {
  if (!t) return '—';
  const d = new Date(t);
  if (Number.isNaN(d.getTime())) return t;
  const p = (n: number) => String(n).padStart(2, '0');
  return `${p(d.getHours())}:${p(d.getMinutes())}:${p(d.getSeconds())}`;
}
