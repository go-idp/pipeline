import { useEffect, useState } from 'react';
import { api } from './client';
import { I, Modal, useToast } from './components';
import { useI18n } from './i18n';
import type { Template } from './api';
import { useRunModal } from './App';

// 模板/案例选择器：内置模板 -> 使用模板（创建为流水线）或直接运行
export function TemplatePicker({ onClose, onUsed }: { onClose: () => void; onUsed: () => void }) {
  const { t } = useI18n();
  const toast = useToast();
  const { openRunModal } = useRunModal();
  const [templates, setTemplates] = useState<Template[] | null>(null);
  const [busy, setBusy] = useState<string | null>(null);

  useEffect(() => {
    api.listTemplates()
      .then((r) => setTemplates(r.data))
      .catch(() => setTemplates([]));
  }, []);

  const useTemplate = async (tpl: Template) => {
    setBusy(tpl.id);
    try {
      await api.createConfig({ name: tpl.name, description: tpl.description, yaml: tpl.yaml });
      toast(t('pipelines.templateUsed'));
      onUsed();
      onClose();
    } catch (e) {
      toast(String((e as Error).message), false);
    } finally {
      setBusy(null);
    }
  };

  const stagesCount = (yaml: string) => yaml.split('\n').filter((l) => /^\s{2}- name:/.test(l)).length;

  return (
    <Modal
      title={
        <span style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          {I('layers')}
          {t('pipelines.templates')}
        </span>
      }
      onClose={onClose}
      wide
      footer={
        <button className="btn" onClick={onClose}>
          {t('common.close')}
        </button>
      }
    >
      <div className="hint-line" style={{ marginBottom: 12 }}>
        {t('pipelines.templatesSub')}
      </div>
      {!templates && <div className="hint-line">{t('common.loading')}</div>}
      {templates && templates.length === 0 && <div className="hint-line">{t('common.empty')}</div>}
      <div style={{ display: 'flex', flexDirection: 'column', gap: 8, maxHeight: '55vh', overflowY: 'auto' }}>
        {templates?.map((tpl) => (
          <div
            key={tpl.id}
            style={{
              display: 'flex', alignItems: 'center', gap: 12, padding: '10px 12px',
              border: '1px solid var(--border)', borderRadius: 8, background: '#fcfcfd',
            }}
          >
            {I('layers')}
            <div style={{ flex: 1, minWidth: 0 }}>
              <div style={{ fontWeight: 600 }}>{tpl.name}</div>
              <div style={{ fontSize: 12, color: 'var(--text-3)' }}>{tpl.description}</div>
            </div>
            <span className="chip">{t('pipelines.stages', { n: stagesCount(tpl.yaml) })}</span>
            <button
              className="btn btn-sm"
              onClick={() => openRunModal([{ name: tpl.name, yaml: tpl.yaml, id: tpl.id }])}
            >
              {I('play')}
              {t('pipelines.directRun')}
            </button>
            <button
              className="btn btn-sm btn-primary"
              disabled={busy === tpl.id}
              onClick={() => useTemplate(tpl)}
            >
              {I('plus')}
              {t('pipelines.useTemplate')}
            </button>
          </div>
        ))}
      </div>
    </Modal>
  );
}
