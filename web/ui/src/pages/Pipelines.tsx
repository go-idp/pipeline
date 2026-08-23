import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { api, fmtDur, relTime, runDuration } from '../client';
import { ConfirmModal, Empty, I, StatusPill, useToast } from '../components';
import { useI18n } from '../i18n';
import type { ConfigTemplate, PipelineRecord } from '../api';
import { useRunModal } from '../App';
import { TemplatePicker } from '../TemplatePicker';

export default function Pipelines() {
  const { t } = useI18n();
  const nav = useNavigate();
  const toast = useToast();
  const { openRunModal } = useRunModal();

  const [configs, setConfigs] = useState<ConfigTemplate[] | null>(null);
  const [runs, setRuns] = useState<PipelineRecord[]>([]);
  const [q, setQ] = useState('');
  const [confirmDel, setConfirmDel] = useState<ConfigTemplate | null>(null);
  const [templatePicker, setTemplatePicker] = useState(false);

  const reload = () => {
    api.listConfigs().then((r) => setConfigs(r.data)).catch(() => setConfigs([]));
    api.listRuns({ limit: '200' }).then((r) => setRuns(r.data)).catch(() => setRuns([]));
  };

  useEffect(reload, []);

  const lastRunOf = (name: string): PipelineRecord | undefined =>
    runs.find((r) => r.name === name);

  const list = (configs ?? []).filter((c) => {
    if (q && !c.name.toLowerCase().includes(q.toLowerCase()) && !(c.description ?? '').toLowerCase().includes(q.toLowerCase())) return false;
    return true;
  });

  const del = async (c: ConfigTemplate) => {
    try {
      await api.deleteConfig(c.id);
      setConfigs((cs) => (cs ?? []).filter((x) => x.id !== c.id));
      toast(t('pipelines.deleted'));
    } catch (e) {
      toast(String((e as Error).message), false);
    }
  };

  return (
    <>
      <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 16, flexWrap: 'wrap' }}>
        <div className="seg">
          <button className="on">{t('common.all')} {(configs ?? []).length}</button>
        </div>
        <div style={{ flex: 1 }} />
        <input
          className="inp"
          id="page-search"
          style={{ width: 220 }}
          placeholder={t('common.search')}
          onChange={(e) => setQ(e.target.value)}
        />
        <button className="btn" onClick={() => setTemplatePicker(true)}>
          {I('layers')}
          {t('pipelines.fromTemplate')}
        </button>
        <button className="btn btn-primary" onClick={() => nav('/pipelines/new')}>
          {I('plus')}
          {t('pipelines.newPipeline')}
        </button>
      </div>

      <div className="card">
        <table className="tbl">
          <thead>
            <tr>
              <th>{t('common.pipeline')}</th>
              <th>{t('pipelines.lastRun')}</th>
              <th>{t('common.duration')}</th>
              <th>{t('pipelines.runCount')}</th>
              <th className="num">{t('pipelines.lastRunAt')}</th>
              <th style={{ textAlign: 'right' }}>{t('common.actions')}</th>
            </tr>
          </thead>
          <tbody>
            {list.length === 0 && (
              <tr>
                <td colSpan={6}>
                  <Empty
                    icon={I('layers')}
                    title={q ? t('pipelines.noMatch') : t('pipelines.templates')}
                    sub={q ? t('pipelines.noMatchSub') : t('pipelines.templatesSub')}
                    action={
                      <div style={{ display: 'flex', gap: 8, justifyContent: 'center' }}>
                        <button className="btn" onClick={() => setTemplatePicker(true)}>
                          {I('layers')}
                          {t('pipelines.fromTemplate')}
                        </button>
                        <button className="btn btn-primary" onClick={() => nav('/pipelines/new')}>
                          {I('plus')}
                          {t('pipelines.newPipeline')}
                        </button>
                      </div>
                    }
                  />
                </td>
              </tr>
            )}
            {list.map((c) => {
              const lr = lastRunOf(c.name);
              return (
                <tr key={c.id} onClick={() => nav(`/pipelines/${c.id}`)}>
                  <td>
                    <div style={{ fontWeight: 500 }}>{c.name}</div>
                    <div style={{ fontSize: 12, color: 'var(--text-3)' }}>{c.description}</div>
                  </td>
                  <td>{lr ? <StatusPill status={lr.status} t={t} /> : <span style={{ color: 'var(--text-3)' }}>—</span>}</td>
                  <td>{lr && runDuration(lr) !== null ? fmtDur(runDuration(lr)) : '—'}</td>
                  <td>{runs.filter((r) => r.name === c.name).length}</td>
                  <td className="num" style={{ color: 'var(--text-3)' }}>{lr ? relTime(lr.started_at) : '—'}</td>
                  <td>
                    <div className="row-actions">
                      <button
                        className="btn btn-sm btn-primary"
                        onClick={(e) => {
                          e.stopPropagation();
                          openRunModal([{ name: c.name, yaml: c.yaml, id: c.id }]);
                        }}
                      >
                        {I('play')}
                        {t('common.run')}
                      </button>
                      <button
                        className="btn btn-sm"
                        onClick={(e) => {
                          e.stopPropagation();
                          nav(`/pipelines/${c.id}/edit`);
                        }}
                      >
                        {t('common.edit')}
                      </button>
                      <button
                        className="btn btn-icon btn-sm"
                        onClick={(e) => {
                          e.stopPropagation();
                          setConfirmDel(c);
                        }}
                      >
                        {I('trash')}
                      </button>
                    </div>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      {confirmDel && (
        <ConfirmModal
          title={t('pipelines.deleteConfirm', { name: confirmDel.name })}
          message={t('pipelines.deleteWarn')}
          onConfirm={() => del(confirmDel)}
          onClose={() => setConfirmDel(null)}
        />
      )}

      {templatePicker && (
        <TemplatePicker
          onClose={() => setTemplatePicker(false)}
          onUsed={reload}
        />
      )}
    </>
  );
}
