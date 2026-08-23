import { cloneElement, createContext, useCallback, useContext, useEffect, useRef, useState, type ReactNode } from 'react';
import type { RunStatus } from './api';

/* ---------- icons (inline, stroke 1.6) ---------- */
const S = {
  overview: <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round"><rect x="3" y="3" width="7" height="9" rx="1.5" /><rect x="14" y="3" width="7" height="5" rx="1.5" /><rect x="14" y="12" width="7" height="9" rx="1.5" /><rect x="3" y="16" width="7" height="5" rx="1.5" /></svg>,
  layers: <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round"><path d="M12 2 2 7l10 5 10-5-10-5z" /><path d="m2 12 10 5 10-5" /><path d="m2 17 10 5 10-5" /></svg>,
  runs: <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round"><path d="M4 12a8 8 0 1 1 2.3 5.6" /><path d="M4 12V6" /><path d="M4 12h6" /></svg>,
  queue: <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round"><path d="M8 6h13M8 12h13M8 18h13" /><circle cx="3.5" cy="6" r="1" fill="currentColor" stroke="none" /><circle cx="3.5" cy="12" r="1" fill="currentColor" stroke="none" /><circle cx="3.5" cy="18" r="1" fill="currentColor" stroke="none" /></svg>,
  settings: <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round"><path d="M4 21v-7M4 10V3M12 21v-9M12 8V3M20 21v-5M20 12V3" /><circle cx="4" cy="12" r="2" /><circle cx="12" cy="10" r="2" /><circle cx="20" cy="14" r="2" /></svg>,
  play: <svg viewBox="0 0 24 24" fill="currentColor"><path d="M7 4.5v15l13-7.5-13-7.5z" /></svg>,
  stop: <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8"><rect x="6" y="6" width="12" height="12" rx="1.5" fill="currentColor" stroke="none" /></svg>,
  x: <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round"><path d="M6 6l12 12M18 6 6 18" /></svg>,
  chev: <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="m9 6 6 6-6 6" /></svg>,
  search: <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round"><circle cx="11" cy="11" r="7" /><path d="m21 21-4.3-4.3" /></svg>,
  check: <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="m4.5 12.5 5 5 10-11" /></svg>,
  warn: <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round"><path d="M12 3 2 20h20L12 3z" /><path d="M12 10v4M12 17.5v.5" /></svg>,
  copy: <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round"><rect x="9" y="9" width="12" height="12" rx="2" /><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" /></svg>,
  dl: <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round"><path d="M12 3v12m0 0 4-4m-4 4-4-4" /><path d="M4 17v2a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-2" /></svg>,
  trash: <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round"><path d="M4 7h16M9 7V5a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2M6 7l1 13a1 1 0 0 0 1 1h8a1 1 0 0 0 1-1l1-13" /></svg>,
  rr: <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round"><path d="M20 11a8 8 0 1 0-2.3 5.6" /><path d="M20 4v7h-7" /></svg>,
  more: <svg viewBox="0 0 24 24" fill="currentColor"><circle cx="5" cy="12" r="1.7" /><circle cx="12" cy="12" r="1.7" /><circle cx="19" cy="12" r="1.7" /></svg>,
  plus: <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round"><path d="M12 5v14M5 12h14" /></svg>,
  term: <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round"><rect x="3" y="4" width="18" height="16" rx="2" /><path d="m7 9 3 3-3 3M13 15h4" /></svg>,
  cmd: <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round"><path d="M9 6a3 3 0 1 0 0 6h6a3 3 0 1 0 0-6 3 3 0 0 0 0 6H9a3 3 0 1 0 0-6z" /><path d="M15 18H9a3 3 0 1 0 0 6 3 3 0 0 0 0-6h6a3 3 0 1 0 0 6 3 3 0 0 0 0-6z" transform="scale(1,-1) translate(0,-24)" /></svg>,
};

// 简单做法：直接返回已构造的 svg 元素（通过 cloneElement 注入尺寸）
// 图标名称（S 的键）
export type IconName = keyof typeof S;

// 渲染图标：cloneElement 保留原始 svg 的 fill/stroke/strokeWidth 等绘制属性，
// 只覆盖 width/height 完成缩放（直接重建 svg 会丢失绘制属性导致图标坏掉）。
export function I(name: IconName, size = 16) {
  const el = S[name];
  return (
    <span style={{ display: 'inline-flex', width: size, height: size, flex: 'none' }}>
      {cloneElement(el, { width: size, height: size })}
    </span>
  );
}

/* ---------- status ---------- */
const STATUS_DOT: Record<RunStatus, string> = {
  pending: '', running: 'blue', succeeded: 'green', failed: 'red', cancelled: '',
};

export function StatusPill({ status, t }: { status: RunStatus; t: (k: string) => string }) {
  return (
    <span className="pill">
      <span className={`dot ${STATUS_DOT[status]}`} />
      {t(`status.${status}`)}
    </span>
  );
}

export function StatusDot({ status }: { status: RunStatus }) {
  return <span className={`dot ${STATUS_DOT[status]}`} />;
}

/* ---------- polling hook ---------- */
export function usePolling<T>(fn: () => Promise<T>, deps: unknown[], interval = 1500, active = true): T | null {
  const [data, setData] = useState<T | null>(null);
  const fnRef = useRef(fn);
  fnRef.current = fn;
  const activeRef = useRef(active);
  activeRef.current = active;

  useEffect(() => {
    let alive = true;
    let timer: ReturnType<typeof setInterval> | null = null;

    const tick = async () => {
      if (!activeRef.current) return;
      try {
        const d = await fnRef.current();
        if (alive) setData(d);
      } catch {
        /* 静默失败，下次重试 */
      }
    };

    tick();
    timer = setInterval(tick, interval);
    return () => {
      alive = false;
      if (timer) clearInterval(timer);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [...deps, interval, active]);

  return data;
}

/* ---------- modal / drawer ---------- */
export function Modal({ title, children, footer, onClose, wide }: {
  title: ReactNode; children: ReactNode; footer?: ReactNode; onClose: () => void; wide?: boolean;
}) {
  return (
    <div className="overlay" onClick={(e) => e.target === e.currentTarget && onClose()}>
      <div className={`modal ${wide ? 'm-lg' : ''}`}>
        <div className="m-h">
          <div className="mt">{title}</div>
        </div>
        <div className="m-b">{children}</div>
        {footer && <div className="m-f">{footer}</div>}
      </div>
    </div>
  );
}

export function Drawer({ title, onClose, footer, children, width = 760 }: {
  title: ReactNode; onClose: () => void; footer?: ReactNode; children: ReactNode; width?: number;
}) {
  return (
    <>
      <div className="overlay" style={{ zIndex: 55, background: 'rgba(9,9,11,.2)' }} onClick={onClose} />
      <div className="drawer" style={{ width }}>
        <div className="d-h">
          <div className="dt">{title}</div>
          <span style={{ flex: 1 }} />
          <button className="btn btn-icon" onClick={onClose}>{I('x')}</button>
        </div>
        <div className="d-b">{children}</div>
        {footer && <div className="d-f">{footer}</div>}
      </div>
    </>
  );
}

/* ---------- toast ---------- */
interface Toast { id: number; msg: string; ok: boolean }
const ToastCtx = createContext<(msg: string, ok?: boolean) => void>(() => {});
export const useToast = () => useContext(ToastCtx);

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);
  const idRef = useRef(0);

  const push = useCallback((msg: string, ok = true) => {
    const id = ++idRef.current;
    setToasts((ts) => [...ts, { id, msg, ok }]);
    setTimeout(() => setToasts((ts) => ts.filter((t) => t.id !== id)), 2600);
  }, []);

  return (
    <ToastCtx.Provider value={push}>
      {children}
      <div className="toasts">
        {toasts.map((t) => (
          <div key={t.id} className="toast">
            <span style={{ color: t.ok ? '#4ade80' : '#f87171', display: 'flex' }}>{t.ok ? I('check', 14) : I('warn', 14)}</span>
            {t.msg}
          </div>
        ))}
      </div>
    </ToastCtx.Provider>
  );
}

/* ---------- confirm helper ---------- */
export function ConfirmModal({ title, message, onConfirm, onClose }: {
  title: string; message: string; onConfirm: () => void; onClose: () => void;
}) {
  return (
    <Modal
      title={title}
      onClose={onClose}
      footer={
        <>
          <button className="btn" onClick={onClose}>Cancel</button>
          <button
            className="btn btn-danger"
            onClick={() => { onConfirm(); onClose(); }}
          >
            Confirm
          </button>
        </>
      }
    >
      {message}
    </Modal>
  );
}

/* ---------- empty ---------- */
export function Empty({ icon, title, sub, action }: { icon?: ReactNode; title: string; sub?: string; action?: ReactNode }) {
  return (
    <div className="empty">
      <div className="ei">{icon}</div>
      <div className="et">{title}</div>
      {sub && <div className="es">{sub}</div>}
      {action}
    </div>
  );
}
