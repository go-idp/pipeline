import { createContext, useCallback, useContext, useState, type ReactNode } from 'react';
import { Navigate, Route, Routes } from 'react-router-dom';
import { Layout } from './Layout';
import Overview from './pages/Overview';
import Pipelines from './pages/Pipelines';
import PipelineDetail from './pages/PipelineDetail';
import PipelineEditor from './pages/PipelineEditor';
import Runs from './pages/Runs';
import RunDetail from './pages/RunDetail';
import Queue from './pages/Queue';
import Settings from './pages/Settings';
import { RunPipelineModal } from './RunModal';

// 全局「运行流水线」弹窗（Overview / Pipelines / RunDetail 共用）
interface RunModalCtx {
  openRunModal: (configs: { name: string; yaml: string; id?: string }[], title?: string) => void;
}
const Ctx = createContext<RunModalCtx>({ openRunModal: () => {} });
export const useRunModal = () => useContext(Ctx);

export default function App() {
  const [runModal, setRunModal] = useState<{ configs: { name: string; yaml: string; id?: string }[]; title?: string } | null>(null);

  const openRunModal = useCallback((configs: { name: string; yaml: string; id?: string }[], title?: string) => {
    setRunModal({ configs, title });
  }, []);

  return (
    <Ctx.Provider value={{ openRunModal }}>
      <Layout>
        <Routes>
          <Route path="/" element={<Overview />} />
          <Route path="/overview" element={<Overview />} />
          <Route path="/pipelines" element={<Pipelines />} />
          <Route path="/pipelines/new" element={<PipelineEditor />} />
          <Route path="/pipelines/:id" element={<PipelineDetail />} />
          <Route path="/pipelines/:id/edit" element={<PipelineEditor />} />
          <Route path="/runs" element={<Runs />} />
          <Route path="/runs/:id" element={<RunDetail />} />
          <Route path="/queue" element={<Queue />} />
          <Route path="/settings" element={<Settings />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Layout>
      {runModal && (
        <RunPipelineModal
          configs={runModal.configs}
          title={runModal.title}
          onClose={() => setRunModal(null)}
        />
      )}
    </Ctx.Provider>
  );
}

export type { ReactNode };
