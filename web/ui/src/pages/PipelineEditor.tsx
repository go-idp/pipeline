import { useEffect, useMemo, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import CodeMirror from '@uiw/react-codemirror';
import { yaml } from '@codemirror/lang-yaml';
import { api, parseYamlLite } from '../client';
import { Drawer, I, useToast } from '../components';
import { useI18n } from '../i18n';

const DEFAULT_YAML = `name: My Pipeline

stages:
  - name: build
    jobs:
      - name: build
        steps:
          - name: hello
            command: echo "Hello, Pipeline!"
`;

export default function PipelineEditor() {
  const { id } = useParams();
  const isEdit = !!id;
  const nav = useNavigate();
  const toast = useToast();
  const { t } = useI18n();

  const [name, setName] = useState('');
  const [desc, setDesc] = useState('');
  const [yamlText, setYamlText] = useState(DEFAULT_YAML);
  const [tab, setTab] = useState<'yaml' | 'visual'>('yaml');
  const [loaded, setLoaded] = useState(!isEdit);

  useEffect(() => {
    if (!isEdit) return;
    api.getConfig(id!)
      .then((c) => {
        setName(c.name);
        setDesc(c.description ?? '');
        setYamlText(c.yaml);
        setLoaded(true);
      })
      .catch((e) => {
        toast(String((e as Error).message), false);
        nav('/pipelines');
      });
  }, [id, isEdit, nav, toast]);

  const visual = useMemo(() => parseYamlLite(yamlText), [yamlText]);

  const save = async () => {
    if (!name.trim()) {
      toast(t('pipelines.name') + ' required', false);
      return;
    }
    try {
      if (isEdit) {
        await api.updateConfig(id!, { name: name.trim(), description: desc.trim(), yaml: yamlText });
      } else {
        await api.createConfig({ name: name.trim(), description: desc.trim(), yaml: yamlText });
      }
      toast(t('pipelines.saveOk'));
      nav('/pipelines');
    } catch (e) {
      toast(String((e as Error).message), false);
    }
  };

  const addStage = () => {
    setYamlText(
      yamlText.replace(/\s*$/, '') +
        '\n  - name: new-stage\n    jobs:\n      - name: job\n        steps:\n          - name: step\n            command: echo "hello"\n',
    );
  };

  return (
    <Drawer
      title={isEdit ? t('pipelines.editPipeline') : t('pipelines.newPipeline')}
      onClose={() => nav('/pipelines')}
      footer={
        <>
          <div className="hint-line" style={{ marginRight: 'auto' }}>
            <span className="kbd">⌘</span><span className="kbd">S</span> save
          </div>
          <button className="btn" onClick={() => nav('/pipelines')}>{t('common.cancel')}</button>
          <button className="btn btn-primary" onClick={save}>{t('common.save')}</button>
        </>
      }
    >
      {!loaded ? (
        <div className="hint-line">{t('common.loading')}</div>
      ) : (
        <>
          <div style={{ display: 'flex', gap: 12 }}>
            <div className="field" style={{ flex: 1 }}>
              <label>{t('common.name')}</label>
              <input className="inp" value={name} onChange={(e) => setName(e.target.value)} placeholder="CI / release / deploy" />
            </div>
            <div className="field" style={{ flex: 1 }}>
              <label>{t('common.description')}</label>
              <input className="inp" value={desc} onChange={(e) => setDesc(e.target.value)} placeholder="one line" />
            </div>
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: 14, marginBottom: 16 }}>
            <div style={{ flex: 1 }} />
            <div className="tabs2">
              <button className={tab === 'yaml' ? 'on' : ''} onClick={() => setTab('yaml')}>{t('pipelines.yamlTab')}</button>
              <button className={tab === 'visual' ? 'on' : ''} onClick={() => setTab('visual')}>{t('pipelines.visualTab')}</button>
            </div>
          </div>

          {tab === 'yaml' ? (
            <>
              <div className="qlabel">{t('pipelines.yamlEditor')}</div>
              <div style={{ border: '1px solid var(--border)', borderRadius: 8, overflow: 'hidden' }}>
                <CodeMirror
                  value={yamlText}
                  height="420px"
                  extensions={[yaml()]}
                  onChange={(v) => setYamlText(v)}
                  basicSetup={{ lineNumbers: true, foldGutter: true, highlightActiveLine: false }}
                />
              </div>
              <div className="hint-line" style={{ marginTop: 8 }}>
                YAML first · switch to Visual to check the structure
              </div>
            </>
          ) : (
            <>
              <div className="qlabel">{t('pipelines.visualTab')}（read-only preview）</div>
              <div style={{ marginTop: 10 }}>
                {visual.length === 0 && <div className="hint-line">{t('common.empty')}</div>}
                {visual.map((s, si) => (
                  <div className="vstage" key={si}>
                    <div className="vstage-h">
                      <span className="dot" style={{ background: 'var(--text)' }} />
                      {s.name}
                      <span className="chip">{s.mode}</span>
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
                <button className="btn" style={{ width: '100%', borderStyle: 'dashed' }} onClick={addStage}>
                  {I('plus')}
                  {t('pipelines.addStage')}
                </button>
              </div>
              <div className="hint-line" style={{ marginTop: 10 }}>
                structural changes append YAML blocks; edit fields in the YAML tab
              </div>
            </>
          )}
        </>
      )}
    </Drawer>
  );
}
