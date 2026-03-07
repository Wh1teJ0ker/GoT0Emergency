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
                        {hostName || `Host #${hostId}`}
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
                
                {/* Terminal with Mac/VSCode style container */}
                <div className={cn("absolute inset-0 p-4", activeTab === 'terminal' ? 'block' : 'hidden')}>
                    <div className="w-full h-full flex flex-col rounded-lg border border-zinc-800 bg-[#1e1e1e] shadow-xl overflow-hidden animate-in fade-in zoom-in-95 duration-200">
                        {/* Terminal Header */}
                        <div className="h-9 bg-[#1e1e1e] border-b border-zinc-800 flex items-center px-4 justify-between select-none shrink-0">
                            <div className="flex items-center gap-2">
                                <Terminal size={14} className="text-zinc-400" />
                                <span className="text-xs text-zinc-400 font-mono">
                                    {hostName ? `${hostName} (SSH)` : `Remote Terminal #${hostId}`}
                                </span>
                            </div>
                            <div className="flex items-center gap-3">
                                {/* Optional Actions */}
                                <div className="flex gap-1.5 opacity-50 hover:opacity-100 transition-opacity">
                                    <div className="w-2.5 h-2.5 rounded-full bg-red-500/20 border border-red-500/50 cursor-pointer hover:bg-red-500/40" title="Close" onClick={onClose}></div>
                                    <div className="w-2.5 h-2.5 rounded-full bg-yellow-500/20 border border-yellow-500/50 cursor-pointer hover:bg-yellow-500/40" title="Minimize"></div>
                                    <div className="w-2.5 h-2.5 rounded-full bg-green-500/20 border border-green-500/50 cursor-pointer hover:bg-green-500/40" title="Maximize"></div>
                                </div>
                            </div>
                        </div>
                        {/* Terminal Body with Padding */}
                        <div className="flex-1 p-3 relative bg-[#1e1e1e]">
                            <XtermTerminal 
                                hostId={hostId} 
                                onClose={onClose} 
                            />
                        </div>
                    </div>
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
