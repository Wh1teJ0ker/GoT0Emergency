import { useState } from 'react';
import { XtermTerminal } from '../terminal/XtermTerminal';
import { FileManager } from './FileManager';
import { FullHostMonitor } from '../monitor/FullHostMonitor';
import { Button } from '../ui/Button';
import { Terminal, Folder, Activity, ArrowLeft } from 'lucide-react';
import { cn } from '../../lib/utils';

interface SessionViewProps {
    hostId: number;
    hostName?: string;
    initialTab?: 'monitor' | 'terminal' | 'files';
    onClose: () => void;
}

export function SessionView({ hostId, hostName, initialTab = 'monitor', onClose }: SessionViewProps) {
    const [activeTab, setActiveTab] = useState<'monitor' | 'terminal' | 'files'>(initialTab);

    return (
        <div className="h-full flex flex-col bg-background">
            <div className="flex items-center justify-between p-2 border-b bg-muted/40">
                <div className="flex gap-2">
                    <Button 
                        variant={activeTab === 'monitor' ? 'secondary' : 'ghost'} 
                        size="sm" 
                        onClick={() => setActiveTab('monitor')}
                        className="gap-2"
                    >
                        <Activity size={14} />
                        主机监控
                    </Button>
                    <Button 
                        variant={activeTab === 'terminal' ? 'secondary' : 'ghost'} 
                        size="sm" 
                        onClick={() => setActiveTab('terminal')}
                        className="gap-2"
                    >
                        <Terminal size={14} />
                        终端
                    </Button>
                    <Button 
                        variant={activeTab === 'files' ? 'secondary' : 'ghost'} 
                        size="sm" 
                        onClick={() => setActiveTab('files')}
                        className="gap-2"
                    >
                        <Folder size={14} />
                        文件管理
                    </Button>
                </div>
                <div className="flex items-center gap-2">
                    <span className="text-sm font-medium mr-4">
                        {hostName ? `${hostName} (#${hostId})` : `会话 #${hostId}`}
                    </span>
                    <Button variant="ghost" size="sm" onClick={onClose} className="gap-2">
                        <ArrowLeft size={14} />
                        返回列表
                    </Button>
                </div>
            </div>
            <div className="flex-1 overflow-hidden relative">
                {activeTab === 'monitor' && (
                    <div className="absolute inset-0 overflow-auto bg-background">
                        <FullHostMonitor hostId={hostId} />
                    </div>
                )}
                
                {/* Terminal should be kept mounted if possible to avoid reconnection, but for now simple conditional is fine */}
                <div className={cn("absolute inset-0 p-2", activeTab === 'terminal' ? 'block' : 'hidden')}>
                    <XtermTerminal 
                        hostId={hostId} 
                        onClose={onClose} 
                    />
                </div>
                
                {activeTab === 'files' && (
                    <div className="absolute inset-0 overflow-auto bg-background p-4">
                        <FileManager hostId={hostId} />
                    </div>
                )}
            </div>
        </div>
    );
}
