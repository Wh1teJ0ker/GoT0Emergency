import { useState, useRef, useEffect } from 'react';
import { XtermTerminal } from '../terminal/XtermTerminal';
import { FileManager } from './FileManager';
import { FullHostMonitor } from '../monitor/FullHostMonitor';
import { Button } from '../ui/Button';
import { Terminal, Folder, Activity, ArrowLeft, Plus, X, Zap, CheckCircle, PowerOff } from 'lucide-react';
import { cn } from '../../lib/utils';
// @ts-ignore
import { ConnectSSH, IsConnected } from '../../../wailsjs/go/app/App';
import { useToast } from '../ui/ToastProvider';

interface TerminalTab {
    id: string;
    name: string;
    hostId: number;
}

interface SessionViewProps {
    hostId: number;
    hostName?: string;
    initialTab?: 'monitor' | 'terminal' | 'files';
    onClose: () => void;
}

export function SessionView({ hostId, hostName, initialTab = 'monitor', onClose }: SessionViewProps) {
    const toast = useToast();
    const [activeTab, setActiveTab] = useState<'monitor' | 'terminal' | 'files'>(initialTab);
    const [terminals, setTerminals] = useState<TerminalTab[]>([]);
    const [activeTerminalId, setActiveTerminalId] = useState<string | null>(null);
    const terminalCounter = useRef(0);
    const [isConnecting, setIsConnecting] = useState(false);
    const [isConnected, setIsConnected] = useState(false);

    // Auto-connect when entering the session view
    useEffect(() => {
        const checkAndConnect = async () => {
            try {
                const connected = await IsConnected(hostId);
                if (connected) {
                    setIsConnected(true);
                    return;
                }

                // Not connected, try to connect
                setIsConnecting(true);
                await ConnectSSH(hostId);
                setIsConnected(true);
                toast.success('SSH 连接成功');
            } catch (err) {
                console.error("Connection failed:", err);
                toast.error('SSH 连接失败', String(err));
                setIsConnected(false);
            } finally {
                setIsConnecting(false);
            }
        };

        checkAndConnect();
    }, [hostId]);

    // Reset terminals when hostId changes
    useEffect(() => {
        setTerminals([]);
        setActiveTerminalId(null);
        terminalCounter.current = 0;
    }, [hostId]);

    const addTerminal = () => {
        terminalCounter.current += 1;
        const newTerminal: TerminalTab = {
            id: `ssh-term-${terminalCounter.current}`,
            name: `终端 ${terminalCounter.current}`,
            hostId: hostId,
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
            }
            return newTerminals;
        });
    };

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
                    <span className="text-sm font-medium mr-2">
                        {hostName || `Host #${hostId}`}
                    </span>
                    {isConnecting ? (
                        <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-yellow-100 text-yellow-800">
                            <Zap size={10} className="mr-1 animate-pulse" /> 连接中...
                        </span>
                    ) : isConnected ? (
                        <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-green-100 text-green-800">
                            <CheckCircle size={10} className="mr-1" /> 在线
                        </span>
                    ) : (
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={async () => {
                                setIsConnecting(true);
                                try {
                                    await ConnectSSH(hostId);
                                    setIsConnected(true);
                                    toast.success('SSH 连接成功');
                                } catch (err) {
                                    toast.error('SSH 连接失败', String(err));
                                    setIsConnected(false);
                                } finally {
                                    setIsConnecting(false);
                                }
                            }}
                            className="gap-1 h-7 text-xs"
                        >
                            <PowerOff size={12} /> 点击连接
                        </Button>
                    )}
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

                {/* Terminal with multi-tab support */}
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
                        <div className="flex-1 p-4 overflow-hidden">
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
                                            "w-full h-full rounded-lg border border-zinc-800 bg-[#1e1e1e] shadow-xl overflow-hidden animate-in fade-in zoom-in-95 duration-200",
                                            activeTerminalId === term.id ? "block" : "hidden"
                                        )}
                                    >
                                        {/* Terminal Header */}
                                        <div className="h-9 bg-[#1e1e1e] border-b border-zinc-800 flex items-center px-4 justify-between select-none shrink-0">
                                            <div className="flex items-center gap-2">
                                                <Terminal size={14} className="text-zinc-400" />
                                                <span className="text-xs text-zinc-400 font-mono">
                                                    {term.name} - {hostName ? hostName : `Host #${hostId}`}
                                                </span>
                                            </div>
                                            <div className="flex items-center gap-3">
                                                <div className="flex gap-1.5 opacity-50 hover:opacity-100 transition-opacity">
                                                    <div className="w-2.5 h-2.5 rounded-full bg-red-500/20 border border-red-500/50 cursor-pointer hover:bg-red-500/40" title="Close" onClick={onClose}></div>
                                                    <div className="w-2.5 h-2.5 rounded-full bg-yellow-500/20 border border-yellow-500/50 cursor-pointer hover:bg-yellow-500/40" title="Minimize"></div>
                                                    <div className="w-2.5 h-2.5 rounded-full bg-green-500/20 border border-green-500/50 cursor-pointer hover:bg-green-500/40" title="Maximize"></div>
                                                </div>
                                            </div>
                                        </div>
                                        {/* Terminal Body */}
                                        <div className="flex-1 p-3 relative bg-[#1e1e1e]">
                                            <XtermTerminal
                                                key={term.id}
                                                hostId={term.hostId}
                                                onClose={onClose}
                                            />
                                        </div>
                                    </div>
                                ))
                            )}
                        </div>
                    </div>
                )}

                {activeTab === 'files' && (
                    <div className="absolute inset-0 overflow-auto bg-background p-4">
                        <FileManager hostId={hostId} />
                    </div>
                )}
            </div>
        </div>
    );
}
