import { useEffect, useMemo, useRef, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { api } from './client';
import { I, usePolling, type IconName } from './components';
import { useI18n } from './i18n';
import type { QueueStats, ServerInfo } from './api';

const NAV = [
  ['overview', 'nav.overview', 'overview'],
  ['pipelines', 'nav.pipelines', 'layers'],
  ['runs', 'nav.runs', 'runs'],
  ['queue', 'nav.queue', 'queue'],
  ['settings', 'nav.settings', 'settings'],
] as const;

export function Layout({ children }: { children: React.ReactNode }) {
  const { t } = useI18n();
  const nav = useNavigate();
  const loc = useLocation();
  const path = loc.pathname.split('/')[1] || 'overview';
  const [paletteOpen, setPaletteOpen] = useState(false);

  const stats = usePolling<QueueStats>(() => api.queueStats(), []);
  const info = usePolling<ServerInfo>(() => api.serverInfo(), [], 30000);

  const running = stats?.running ?? 0;
  const pending = stats?.pending ?? 0;

  // Cmd+K
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault();
        setPaletteOpen((v) => !v);
      }
      if (e.key === 'Escape') setPaletteOpen(false);
      if (!paletteOpen && e.key === '/' && !(e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement)) {
        const input = document.querySelector<HTMLInputElement>('#tb-search');
        if (input) {
          e.preventDefault();
          input.focus();
        }
      }
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [paletteOpen]);

  const go = (p: string) => { nav(p); setPaletteOpen(false); };

  return (
    <div className="app">
      <aside className="sidebar">
        <div className="sb-brand">
          <div className="logo">
            <svg viewBox="0 0 24 24" fill="currentColor"><path d="M6 4h4v6h4V4h4v16h-4v-6h-4v6H6z" /></svg>
          </div>
          <span className="nm">Pipeline</span>
          <span className="ver">v{info?.version ?? '…'}</span>
        </div>
        <nav className="sb-nav">
          {NAV.map(([key, label, icon]) => (
            <div
              key={key}
              className={`nav-item ${path === key ? 'active' : ''}`}
              onClick={() => nav(`/${key}`)}
            >
              {I(icon)}
              {t(label)}
              {key === 'runs' && running > 0 && <span className="cnt">{running}</span>}
              {key === 'queue' && pending > 0 && <span className="cnt">{pending}</span>}
            </div>
          ))}
        </nav>
        <div className="sb-foot">
          <div className="srv-state">
            <span className="dot green" />
            {t('footer.serverRunning', { port: window.location.port || '80' })}
          </div>
        </div>
      </aside>

      <div className="main">
        <Topbar onPalette={() => setPaletteOpen(true)} />
        <div className="content">{children}</div>
      </div>

      {paletteOpen && (
        <CommandPalette onClose={() => setPaletteOpen(false)} onGo={go} />
      )}
    </div>
  );
}

function Topbar({ onPalette }: { onPalette: () => void }) {
  const { t } = useI18n();
  const loc = useLocation();
  const path = loc.pathname.split('/')[1] || 'overview';
  const sub = loc.pathname.split('/')[2];
  const nav = useNavigate();
  const searchRef = useRef<HTMLInputElement>(null);

  const title = path === 'overview' ? t('nav.overview')
    : path === 'pipelines' ? (sub ? t('pipelines.editPipeline') : t('nav.pipelines'))
    : path === 'runs' ? (sub ? t('runs.runDetail') : t('nav.runs'))
    : path === 'queue' ? t('nav.queue')
    : t('nav.settings');

  const crumb = sub
    ? path === 'pipelines' ? `${t('nav.pipelines')} / ${sub}` : `${t('nav.runs')} / ${sub}`
    : '';

  const onSearch = path === 'runs' || path === 'pipelines';

  return (
    <div className="topbar">
      <div className="tb-title">{title}</div>
      {crumb && <div className="tb-crumb">{crumb}</div>}
      <div className="tb-spacer" />
      {onSearch && (
        <div className="tb-search">
          {I('search')}
          <input
            id="tb-search"
            ref={searchRef}
            placeholder={t('common.search')}
            onChange={(e) => {
              const q = e.target.value;
              const el = document.querySelector<HTMLInputElement>('#page-search');
              if (el) el.value = q;
              el?.dispatchEvent(new Event('input', { bubbles: true }));
            }}
          />
          <span className="hint">/</span>
        </div>
      )}
      <button className="btn" onClick={onPalette} title="Cmd+K">{I('cmd')}</button>
      <button className="btn btn-primary" onClick={() => nav('/pipelines/new')}>
        {I('plus')}
        {t('common.new')}
      </button>
    </div>
  );
}

interface PaletteItem { label: string; icon: string; go: string | (() => void); group: string; hint?: string }

function CommandPalette({ onClose, onGo }: { onClose: () => void; onGo: (p: string) => void }) {
  const { t } = useI18n();
  const [q, setQ] = useState('');
  const [sel, setSel] = useState(0);

  const items = useMemo<PaletteItem[]>(() => {
    const pages: PaletteItem[] = [
      { label: t('nav.overview'), icon: 'overview', go: '/overview', group: t('palette.pages') },
      { label: t('nav.pipelines'), icon: 'layers', go: '/pipelines', group: t('palette.pages') },
      { label: t('nav.runs'), icon: 'runs', go: '/runs', group: t('palette.pages') },
      { label: t('nav.queue'), icon: 'queue', go: '/queue', group: t('palette.pages') },
      { label: t('nav.settings'), icon: 'settings', go: '/settings', group: t('palette.pages') },
    ];
    const actions: PaletteItem[] = [
      { label: t('palette.newPipeline'), icon: 'plus', go: '/pipelines/new', group: t('palette.actions') },
      { label: t('palette.runPipeline'), icon: 'play', go: '/pipelines', group: t('palette.actions') },
      { label: t('palette.viewRunning'), icon: 'runs', go: '/runs?status=running', group: t('palette.actions') },
    ];
    return [...pages, ...actions].filter(
      (it) => !q || it.label.toLowerCase().includes(q.toLowerCase()),
    );
  }, [q, t]);

  useEffect(() => setSel(0), [q]);

  const run = (it: PaletteItem) => {
    if (typeof it.go === 'string') onGo(it.go);
    else it.go();
    onClose();
  };

  return (
    <div className="palette" onKeyDown={(e) => {
      if (e.key === 'ArrowDown') { e.preventDefault(); setSel((s) => Math.min(s + 1, items.length - 1)); }
      if (e.key === 'ArrowUp') { e.preventDefault(); setSel((s) => Math.max(s - 1, 0)); }
      if (e.key === 'Enter' && items[sel]) { e.preventDefault(); run(items[sel]); }
      if (e.key === 'Escape') onClose();
    }}>
      <input
        autoFocus
        placeholder={t('palette.placeholder')}
        value={q}
        onChange={(e) => setQ(e.target.value)}
      />
      <div className="pl">
        {items.map((it, i) => (
          <div
            key={it.label + i}
            className={`pi ${i === sel ? 'sel' : ''}`}
            onMouseEnter={() => setSel(i)}
            onClick={() => run(it)}
          >
            {I(it.icon as IconName, 15)}
            {it.label}
            <span className="hk">{typeof it.go === 'string' && it.go.startsWith('/') ? 'Go' : '↵'}</span>
          </div>
        ))}
        {items.length === 0 && <div className="pg" style={{ padding: 12 }}>{t('common.empty')}</div>}
      </div>
    </div>
  );
}
