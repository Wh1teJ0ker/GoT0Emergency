import { HashRouter, Routes, Route, Navigate } from 'react-router-dom';
import { MainLayout } from './components/layout/MainLayout';
import { SettingsPage } from './pages/SettingsPage';
import { PageContainer } from './components/layout/PageContainer';
import { PageHeader } from './components/layout/PageHeader';
import { Card } from './components/ui/Card';
import { RemotePage } from './pages/RemotePage';
import { LocalPage } from './pages/LocalPage';
import { PluginManager } from './pages/node/PluginManager';
import { NodeManager } from './pages/node/NodeManager';
import { LogsPage } from './pages/LogsPage';

// Placeholder for future implementation
const PlaceholderPage = ({ title }: { title: string }) => (
    <PageContainer>
        <PageHeader title={title} description="此功能正在开发中。" />
        <Card className="flex items-center justify-center p-12 min-h-[300px] border-dashed">
            <p className="text-muted-foreground">开发中</p>
        </Card>
    </PageContainer>
);

function App() {
    return (
        <HashRouter basename="/">
            <Routes>
                <Route path="/" element={<MainLayout />}>
                    <Route index element={<Navigate to="/local" replace />} />
                    <Route path="local" element={<LocalPage />} />
                    <Route path="remote" element={<RemotePage />} />
                    <Route path="remote/:id" element={<RemotePage />} />
                    <Route path="node/plugins" element={<PluginManager />} />
                    <Route path="node/manager" element={<NodeManager />} />
                    <Route path="logs" element={<LogsPage />} />
                    <Route path="settings" element={<SettingsPage />} />
                    
                    {/* Fallback to local */}
                    <Route path="*" element={<Navigate to="/local" replace />} />
                </Route>
            </Routes>
        </HashRouter>
    );
}

export default App;
