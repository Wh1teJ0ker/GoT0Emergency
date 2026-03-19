import { useState, useRef } from 'react';
import { PageContainer } from '../components/layout/PageContainer';
import { PageHeader } from '../components/layout/PageHeader';
import { XtermTerminal } from '../components/terminal/XtermTerminal';
import { FullHostMonitor } from '../components/monitor/FullHostMonitor';
import { Activity, Terminal, X, Plus } from 'lucide-react';
import { Button } from '../components/ui/Button';
import { cn } from '../lib/utils';

interface TerminalTab {
    id: string;
    name: string;
    hostId: number;
}

export function LocalPage() {
    const [activeTab, setActiveTab] = useState<'monitor' | 'terminal'>('monitor');
    const [terminals, setTerminals] = useState<TerminalTab[]>([]);
    const [activeTerminalId, setActiveTerminalId] = useState<string | null>(null);
    const terminalCounter = useRef(0);

    const addTerminal = () => {
        terminalCounter.current += 1;
        const newTerminal: TerminalTab = {
            id: `local-term-${terminalCounter.current}`,
            name: `终端 ${terminalCounter.current}`,
            hostId: 0,
        };
        setTerminals(prev => [...prev, newTerminal]);
        setActiveTerminalId(newTerminal.id);
        setActiveTab('terminal');
    };

    const closeTerminal = (e: React.MouseEvent, id: string) => {
        e.stopPropagation();
        setTerminals(prev => {
            const newTerminals = prev.filter(t => t.id !== id);
            if (activeTerminalId === id && newTerminals.length > 0) {
                setActiveTerminalId(newTerminals[newTerminals.length - 1].id);
            } else if (newTerminals.length === 0) {
                setActiveTerminalId(null);
                setActiveTab('monitor');
            }
            return newTerminals;
        });
    };

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
                        <FullHostMonitor hostId={0} showHistory={false} />
                    </div>
                )}

                {/* Terminal Tab */}
                {activeTab === 'terminal' && (
                    <div className="absolute inset-0 flex flex-col">
                        {/* Terminal Tabs Bar */}
                        <div className="flex items-center gap-1 px-2 py-1 border-b bg-muted/30 overflow-x-auto">
                            {terminals.map(term => (
                                <div
                                    key={term.id}
                                    className={cn(
                                        "flex items-center gap-2 px-3 py-1.5 rounded-t text-sm cursor-pointer border border-b-0 transition-colors min-w-[150px]",
                                        activeTerminalId === term.id
                                            ? "bg-background border-primary text-foreground"
                                            : "bg-muted/50 border-transparent text-muted-foreground hover:bg-muted"
                                    )}
                                    onClick={() => setActiveTerminalId(term.id)}
                                >
                                    <Terminal size={14} />
                                    <span className="flex-1 truncate">{term.name}</span>
                                    <button
                                        onClick={(e) => closeTerminal(e, term.id)}
                                        className="hover:bg-destructive/20 hover:text-destructive rounded p-0.5 transition-colors"
                                    >
                                        <X size={12} />
                                    </button>
                                </div>
                            ))}
                            <button
                                className="p-1 text-muted-foreground hover:text-foreground hover:bg-muted rounded transition-colors"
                                onClick={addTerminal}
                                title="新建终端"
                            >
                                <Plus size={16} />
                            </button>
                        </div>

                        {/* Terminal Content */}
                        <div className="flex-1 p-0 overflow-hidden">
                            {terminals.length === 0 ? (
                                <div className="h-full flex flex-col items-center justify-center text-muted-foreground">
                                    <Terminal size={48} className="mb-4 opacity-50" />
                                    <p className="text-sm mb-4">暂无终端</p>
                                    <Button onClick={addTerminal} size="sm" className="gap-2">
                                        <Plus size={16} />
                                        新建终端
                                    </Button>
                                </div>
                            ) : (
                                terminals.map(term => (
                                    <div
                                        key={term.id}
                                        className={cn(
                                            "w-full h-full rounded-lg border border-zinc-800 bg-[#1e1e1e] shadow-xl overflow-hidden",
                                            activeTerminalId === term.id ? "block" : "hidden"
                                        )}
                                    >
                                        {/* Terminal Body */}
                                        <div className="w-full h-full relative p-2">
                                            <XtermTerminal key={term.id} hostId={term.hostId} />
                                        </div>
                                    </div>
                                ))
                            )}
                        </div>
                    </div>
                )}
            </div>
        </PageContainer>
    );
}
