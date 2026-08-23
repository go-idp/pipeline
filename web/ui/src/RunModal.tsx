import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { api, appendEnv } from './client';
import { I, Modal, useToast } from './components';
import { useI18n } from './i18n';

// 运行流水线确认弹窗：选择配置 + 追加环境变量 -> POST /api/v1/runs
export function RunPipelineModal({ configs, title, onClose }: {
  configs: { name: string; yaml: string; id?: string }[];
  title?: string;
  onClose: () => void;
}) {
  const { t } = useI18n();
  const toast = useToast();
  const nav = useNavigate();
  const [env, setEnv] = useState('');
  const [busy, setBusy] = useState(false);

  const names = configs.map((c) => c.name).join('、');

  const run = async () => {
    setBusy(true);
    try {
      let lastId = '';
      for (const c of configs) {
        const merged = appendEnv(c.yaml, env);
        const res = await api.createRun(merged, 'manual');
        lastId = res.id;
      }
      toast(`${names} ${t('pipelines.enqueued')}`);
      onClose();
      nav(`/runs/${lastId}`);
    } catch (e) {
      toast(String((e as Error).message), false);
    } finally {
      setBusy(false);
    }
  };

  return (
    <Modal
      title={
        <span style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          {I('play')}
          {title || t('pipelines.runConfirm')}
        </span>
      }
      onClose={onClose}
      footer={
        <>
          <button className="btn" onClick={onClose}>{t('common.cancel')}</button>
          <button className="btn btn-primary" disabled={busy} onClick={run}>
            {I('play')}
            {t('pipelines.enqueue')}
          </button>
        </>
      }
    >
      <div className="field" style={{ marginBottom: 10 }}>
        <label>{t('common.pipeline')}</label>
        <div style={{ fontSize: 14, fontWeight: 600, color: 'var(--text)' }}>{names}</div>
      </div>
      <div className="field">
        <label>{t('pipelines.envParams')}</label>
        <textarea
          className="inp"
          rows={3}
          placeholder={t('pipelines.envPlaceholder')}
          value={env}
          onChange={(e) => setEnv(e.target.value)}
        />
      </div>
      <div className="hint-line">{t('pipelines.enqueueHint')}</div>
    </Modal>
  );
}


