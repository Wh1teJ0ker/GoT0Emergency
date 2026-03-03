import { useState } from 'react';
import { PageContainer } from '../components/layout/PageContainer';
import { PageHeader } from '../components/layout/PageHeader';
import { XtermTerminal } from '../components/terminal/XtermTerminal';
import { FullHostMonitor } from '../components/monitor/FullHostMonitor';
import { Activity, Terminal } from 'lucide-react';

export function LocalPage() {
    const [activeTab, setActiveTab] = useState<'monitor' | 'terminal'>('monitor');

    return (
        <PageContainer className="h-full flex flex-col space-y-0">
            <div className="flex items-center justify-between mb-4">
                <PageHeader title="本地管理" description="管理您的本地系统资源和终端。" />
            </div>

            {/* Tabs */}
            <div className="flex border-b mb-4">
                <button
                    className={`px-4 py-2 font-medium text-sm transition-colors border-b-2 flex items-center gap-2 ${activeTab === 'monitor' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'}`}
                    onClick={() => setActiveTab('monitor')}
                >
                    <Activity size={16} />
                    主机监测
                </button>
                <button
                    className={`px-4 py-2 font-medium text-sm transition-colors border-b-2 flex items-center gap-2 ${activeTab === 'terminal' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'}`}
                    onClick={() => setActiveTab('terminal')}
                >
                    <Terminal size={16} />
                    本地终端
                </button>
            </div>

            <div className="flex-1 overflow-hidden relative">
                {/* Monitor Tab */}
                {activeTab === 'monitor' && (
                    <div className="absolute inset-0 overflow-auto">
                        <FullHostMonitor hostId={0} />
                    </div>
                )}

                {/* Terminal Tab */}
                <div className={`absolute inset-0 p-1 ${activeTab === 'terminal' ? 'block' : 'hidden'}`}>
                    <XtermTerminal hostId={0} />
                </div>
            </div>
        </PageContainer>
    );
}
