import { useEffect, useState } from 'react';
import { api } from '../client';
import { ConfirmModal, useToast } from '../components';
import { useI18n, type Locale } from '../i18n';
import type { ServerInfo, Settings } from '../api';

export default function Settings() {
  const { t, locale, setLocale } = useI18n();
  const toast = useToast();

  const [info, setInfo] = useState<ServerInfo | null>(null);
  const [settings, setSettings] = useState<Settings | null>(null);
  const [maxConcurrent, setMaxConcurrent] = useState('2');
  const [taskTimeout, setTaskTimeout] = useState('3600');
  const [executor, setExecutor] = useState('in-process');
  const [confirmClear, setConfirmClear] = useState(false);
  const [runCount, setRunCount] = useState(0);

  useEffect(() => {
    api.serverInfo().then(setInfo).catch(() => {});
    api.getSettings().then((s) => {
      setSettings(s);
      setMaxConcurrent(String(s.max_concurrent));
    }).catch(() => {});
    api.listRuns({ limit: '1' }).then((r) => setRunCount(r.total)).catch(() => {});
  }, []);

  const save = async () => {
    try {
      await api.saveSettings({
        max_concurrent: Number(maxConcurrent) || 2,
        task_timeout: Number(taskTimeout) || 0,
        task_executor: executor,
      });
      toast(t('settings.saved'));
    } catch (e) {
      toast(String((e as Error).message), false);
    }
  };

  return (
    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 14, alignItems: 'start' }}>
      <div>
        <div className="card" style={{ marginBottom: 14 }}>
          <div className="card-h">{t('settings.server')}</div>
          <div className="card-b" style={{ padding: '12px 16px' }}>
            <div className="kv"><div className="k">{t('settings.version')}</div><div className="v">pipeline v{info?.version ?? '…'}</div></div>
            <div className="kv"><div className="k">{t('settings.uptime')}</div><div className="v">{info?.running_at ?? '—'}</div></div>
            <div className="kv"><div className="k">{t('settings.addr')}</div><div className="v">http://localhost:8080</div></div>
            <div className="kv" style={{ borderBottom: 0 }}><div className="k">{t('settings.basePath')}</div><div className="v">/</div></div>
          </div>
        </div>

        <div className="card">
          <div className="card-h">{t('settings.storage')}<span className="sub">workdir</span></div>
          <div className="card-b" style={{ padding: '12px 16px' }}>
            <div className="kv"><div className="k">{t('settings.workdir')}</div><div className="v">/tmp/go-idp/pipeline</div></div>
            <div className="kv" style={{ borderBottom: 0 }}>
              <div className="k">{t('settings.maxRecords')}</div>
              <div className="v">{settings?.max_records ?? 1000} (.pipeline_records/)</div>
            </div>
          </div>
        </div>
      </div>

      <div>
        <div className="card" style={{ marginBottom: 14 }}>
          <div className="card-h">{t('settings.execution')}</div>
          <div className="card-b" style={{ padding: 16 }}>
            <div className="field">
              <label>{t('settings.maxConcurrent')}</label>
              <input className="inp" value={maxConcurrent} onChange={(e) => setMaxConcurrent(e.target.value)} />
            </div>
            <div className="field">
              <label>{t('settings.taskTimeout')}</label>
              <input className="inp" value={taskTimeout} onChange={(e) => setTaskTimeout(e.target.value)} />
            </div>
            <div className="field">
              <label>{t('settings.executor')}</label>
              <div style={{ display: 'flex', gap: 8 }}>
                {['in-process', 'subprocess'].map((v) => (
                  <button
                    key={v}
                    className={`btn ${executor === v ? 'btn-primary' : ''}`}
                    style={{ flex: 1 }}
                    onClick={() => setExecutor(v)}
                  >
                    {v}
                  </button>
                ))}
              </div>
            </div>
            <div className="hint-line" style={{ marginTop: 2 }}>{t('settings.restartHint')}</div>
          </div>
        </div>

        <div className="card" style={{ marginBottom: 14 }}>
          <div className="card-h">{t('settings.envAllow')}<span className="sub">allow-env</span></div>
          <div className="card-b" style={{ padding: 16 }}>
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8, marginBottom: 12 }}>
              <span className="env-tag">GITHUB_TOKEN</span>
              <span className="env-tag">GITHUB_REF_NAME</span>
              <span className="env-tag">CI</span>
            </div>
            <div className="hint-line">
              <input type="checkbox" defaultChecked />
              {t('settings.allowAll')}
            </div>
          </div>
        </div>

        <div className="card" style={{ marginBottom: 14 }}>
          <div className="card-h">{t('settings.locale')}</div>
          <div className="card-b" style={{ padding: 16 }}>
            <div className="seg">
              {(['zh', 'en'] as Locale[]).map((l) => (
                <button key={l} className={locale === l ? 'on' : ''} onClick={() => setLocale(l)}>
                  {l === 'zh' ? '中文' : 'English'}
                </button>
              ))}
            </div>
          </div>
        </div>

        <div style={{ display: 'flex', gap: 8 }}>
          <button className="btn btn-primary" onClick={save}>{t('settings.save')}</button>
          <button className="btn btn-danger" onClick={() => setConfirmClear(true)}>{t('settings.clearHistory')}</button>
        </div>
      </div>

      {confirmClear && (
        <ConfirmModal
          title={t('settings.clearConfirm')}
          message={t('settings.clearWarn', { n: runCount })}
          onConfirm={() => {
            toast(t('settings.cleared'));
          }}
          onClose={() => setConfirmClear(false)}
        />
      )}
    </div>
  );
}
