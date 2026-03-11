import { useState, useEffect, useMemo } from "react";
import { PageContainer } from "../components/layout/PageContainer";
import { PageHeader } from "../components/layout/PageHeader";
import { Card, CardContent } from "../components/ui/Card";
import { Button } from "../components/ui/Button";
import { Input } from "../components/ui/Input";
import { ConfirmDialog } from "../components/ui/ConfirmDialog";
import { useToast } from "../components/ui/ToastProvider";
import { RefreshCw, Trash2, Calendar, Search, Filter } from "lucide-react";
// @ts-ignore
import { GetLogsByDate, GetLogFiles, ClearLogs } from "../../wailsjs/go/app/App";

interface LogEntry {
    time: string;
    level: string;
    msg: string;
    [key: string]: any;
}

const MODULES = [
    { label: "所有模块", value: "all" },
    { label: "系统", keywords: ["app", "init", "start"] },
    { label: "数据库", keywords: ["db", "database", "sqlite", "sql"] },
    { label: "主机", keywords: ["host", "ssh", "connect"] },
    { label: "节点", keywords: ["node", "plugin", "deploy"] },
    { label: "监控", keywords: ["monitor", "status", "metric"] },
    { label: "终端", keywords: ["terminal", "pty", "shell"] },
];

export function LogsPage() {
    const toast = useToast();
    const [rawLogs, setRawLogs] = useState<string[]>([]);
    const [dates, setDates] = useState<string[]>([]);
    const [selectedDate, setSelectedDate] = useState<string>('');
    const [loading, setLoading] = useState(false);
    const [isClearDialogOpen, setIsClearDialogOpen] = useState(false);
    
    // Filters
    const [searchQuery, setSearchQuery] = useState("");
    const [selectedModule, setSelectedModule] = useState("all");

    useEffect(() => {
        loadDates();
    }, []);

    useEffect(() => {
        if (selectedDate) {
            loadLogs(selectedDate);
        }
    }, [selectedDate]);

    const loadDates = async () => {
        try {
            const result = await GetLogFiles();
            setDates(result || []);
            // Default to today or the latest available date
            const today = new Date().toISOString().split('T')[0];
            if (result && result.includes(today)) {
                setSelectedDate(today);
            } else if (result && result.length > 0) {
                setSelectedDate(result[result.length - 1]);
            }
        } catch (err) {
            console.error("Failed to load log dates:", err);
        }
    };

    const loadLogs = async (date: string) => {
        if (!date) return;
        setLoading(true);
        try {
            const data = await GetLogsByDate(date, 2000); // Increased limit
            setRawLogs(data || []);
        } catch (err) {
            console.error("Failed to load logs:", err);
        } finally {
            setLoading(false);
        }
    };

    const handleClearClick = () => {
        setIsClearDialogOpen(true);
    };

    const handleClearConfirm = async () => {
        setIsClearDialogOpen(false);
        try {
            await ClearLogs();
            loadLogs(selectedDate); // Reload current
            toast.success('日志已清空');
        } catch (err) {
            console.error("Failed to clear logs:", err);
            toast.error('日志清空失败', String(err));
        }
    };

    // Parse and Filter Logs
    const filteredLogs = useMemo(() => {
        return rawLogs
            .map(line => {
                try {
                    return JSON.parse(line) as LogEntry;
                } catch {
                    // Fallback for non-JSON lines
                    return { time: "", level: "UNKNOWN", msg: line } as LogEntry;
                }
            })
            .filter(log => {
                // 1. Module Filter
                if (selectedModule !== "all") {
                    const moduleConfig = MODULES.find(m => m.value === selectedModule || m.label === selectedModule); // simple matching
                    if (moduleConfig && moduleConfig.keywords) {
                        const content = JSON.stringify(log).toLowerCase();
                        const hasKeyword = moduleConfig.keywords.some(k => content.includes(k));
                        if (!hasKeyword) return false;
                    }
                }

                // 2. Search Query
                if (searchQuery) {
                    const query = searchQuery.toLowerCase();
                    const content = JSON.stringify(log).toLowerCase();
                    return content.includes(query);
                }

                return true;
            })
            .reverse(); // Show newest first
    }, [rawLogs, selectedModule, searchQuery]);

    const getLevelColor = (level: string) => {
        switch (level?.toUpperCase()) {
            case "INFO": return "text-blue-400";
            case "WARN": return "text-yellow-400";
            case "ERROR": return "text-red-400";
            case "DEBUG": return "text-gray-400";
            default: return "text-white";
        }
    };

    return (
        <PageContainer className="h-full flex flex-col space-y-4 overflow-hidden">
            <div className="flex flex-col gap-4 shrink-0">
                <div className="flex justify-between items-center">
                    <div className="flex flex-col gap-1">
                        <h1 className="text-2xl font-bold tracking-tight">日志管理</h1>
                        <p className="text-muted-foreground text-sm">查看和管理系统运行日志。</p>
                    </div>
                    <div className="flex gap-2">
                        <Button variant="outline" size="sm" onClick={() => loadLogs(selectedDate)} disabled={loading || !selectedDate}>
                            <RefreshCw className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
                            刷新
                        </Button>
                        <Button variant="destructive" size="sm" onClick={handleClearClick}>
                            <Trash2 className="mr-2 h-4 w-4" />
                            清空
                        </Button>
                    </div>
                </div>

                <ConfirmDialog 
                    open={isClearDialogOpen}
                    title="确认清空？"
                    content="此操作将永久删除当前的所有日志文件。此操作不可恢复。"
                    onConfirm={handleClearConfirm}
                    onCancel={() => setIsClearDialogOpen(false)}
                    confirmText="确认清空"
                    loadingText="清空中..."
                />

                {/* Toolbar */}
                <Card className="p-3 bg-card/50">
                    <div className="flex flex-wrap gap-4 items-center">
                        {/* Date Selector */}
                        <div className="flex items-center gap-2 min-w-[150px]">
                            <Calendar size={16} className="text-muted-foreground" />
                            <select 
                                className="bg-background border rounded px-2 py-1 text-sm focus:outline-none w-full"
                                value={selectedDate}
                                onChange={(e) => setSelectedDate(e.target.value)}
                            >
                                {dates.length === 0 && <option value="">无日志</option>}
                                {dates.map(date => (
                                    <option key={date} value={date}>{date}</option>
                                ))}
                            </select>
                        </div>

                        {/* Module Selector */}
                        <div className="flex items-center gap-2 min-w-[150px]">
                            <Filter size={16} className="text-muted-foreground" />
                            <select 
                                className="bg-background border rounded px-2 py-1 text-sm focus:outline-none w-full"
                                value={selectedModule}
                                onChange={(e) => setSelectedModule(e.target.value)}
                            >
                                {MODULES.map(m => (
                                    <option key={m.label} value={m.label}>{m.label}</option>
                                ))}
                            </select>
                        </div>

                        {/* Search Input */}
                        <div className="flex-1 flex items-center gap-2">
                            <Search size={16} className="text-muted-foreground" />
                            <Input 
                                placeholder="搜索关键字..." 
                                value={searchQuery}
                                onChange={(e) => setSearchQuery(e.target.value)}
                                className="h-8"
                            />
                        </div>
                    </div>
                </Card>
            </div>

            {/* Log List */}
            <Card className="flex-1 overflow-hidden border shadow-sm">
                <div className="h-full overflow-auto">
                    <table className="w-full text-sm text-left border-collapse">
                        <thead className="bg-muted/80 text-xs uppercase sticky top-0 z-10 backdrop-blur-md shadow-sm">
                            <tr>
                                <th className="px-4 py-3 font-semibold text-muted-foreground w-[180px]">时间</th>
                                <th className="px-4 py-3 font-semibold text-muted-foreground w-[80px]">级别</th>
                                <th className="px-4 py-3 font-semibold text-muted-foreground max-w-[600px]">消息内容</th>
                                <th className="px-4 py-3 font-semibold text-muted-foreground w-[300px]">其他信息</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-border/50">
                            {filteredLogs.length === 0 ? (
                                <tr>
                                    <td colSpan={4} className="px-4 py-12 text-center text-muted-foreground">
                                        <div className="flex flex-col items-center gap-2">
                                            <Search className="h-8 w-8 opacity-20" />
                                            <span>没有找到匹配的日志记录</span>
                                        </div>
                                    </td>
                                </tr>
                            ) : (
                                filteredLogs.map((log, index) => (
                                    <tr key={index} className="hover:bg-muted/30 transition-colors group border-b border-border/40 last:border-0">
                                        <td className="px-4 py-2 whitespace-nowrap text-muted-foreground font-mono text-xs w-[180px]">
                                            {log.time ? new Date(log.time).toLocaleString() : '-'}
                                        </td>
                                        <td className="px-4 py-2 w-[80px]">
                                            <span className={`px-2 py-0.5 rounded text-[10px] font-bold border ${
                                                log.level === 'INFO' ? 'bg-blue-500/10 text-blue-500 border-blue-500/20' :
                                                log.level === 'WARN' ? 'bg-yellow-500/10 text-yellow-500 border-yellow-500/20' :
                                                log.level === 'ERROR' ? 'bg-red-500/10 text-red-500 border-red-500/20' :
                                                'bg-gray-500/10 text-gray-500 border-gray-500/20'
                                            }`}>
                                                {log.level || 'UNKNOWN'}
                                            </span>
                                        </td>
                                        <td className="px-4 py-2 font-mono text-xs text-foreground/90 leading-relaxed max-w-[600px]">
                                            <div className="line-clamp-2" title={log.msg}>
                                                {log.msg}
                                            </div>
                                        </td>
                                        <td className="px-4 py-2 text-xs text-muted-foreground font-mono w-[300px]">
                                            <div className="flex flex-wrap gap-1">
                                                {Object.entries(log)
                                                    .filter(([k]) => !['time', 'level', 'msg'].includes(k))
                                                    .map(([k, v]) => (
                                                        <span key={k} className="inline-flex items-center px-1.5 py-0.5 rounded bg-muted/50 text-[10px] border border-border/50 whitespace-nowrap" title={String(v)}>
                                                            <span className="opacity-70 mr-1">{k}:</span>
                                                            <span className="max-w-full break-all">{String(v)}</span>
                                                        </span>
                                                    ))}
                                            </div>
                                        </td>
                                    </tr>
                                ))
                            )}
                        </tbody>
                    </table>
                </div>
            </Card>
        </PageContainer>
    );
}
