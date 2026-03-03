import { useState, useEffect, useRef } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '../ui/Card';
import { Button } from '../ui/Button';
import { 
    Activity, Cpu, HardDrive, Network, Zap, Server, 
    Monitor, Clock, Disc, Layers, List, Box, Info 
} from 'lucide-react';
import { 
    LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer 
} from 'recharts';
// @ts-ignore
import { CheckHostStatus, GetHostMetrics } from '../../../wailsjs/go/app/App';
// @ts-ignore
import { monitor } from '../../../wailsjs/go/models';

import { Loading } from '../ui/Loading';

interface FullHostMonitorProps {
    hostId: number;
}

export function FullHostMonitor({ hostId }: FullHostMonitorProps) {
    const [status, setStatus] = useState<monitor.HostStatus | null>(null);
    const [metrics, setMetrics] = useState<monitor.MetricPoint[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [activeTab, setActiveTab] = useState('overview');

    useEffect(() => {
        loadStatus();
        loadHistory();
        const interval = setInterval(() => {
            loadStatus();
            loadHistory();
        }, 60000); // 1 minute interval
        return () => clearInterval(interval);
    }, [hostId]);

    const loadStatus = async () => {
        // Don't set loading on refresh to avoid flickering
        if (!status) setLoading(true);
        try {
            const result = await CheckHostStatus(hostId);
            setStatus(result);
            setError(null);
        } catch (err) {
            console.error("Failed to check host status:", err);
            setError(String(err));
        } finally {
            setLoading(false);
        }
    };

    const loadHistory = async () => {
        if (hostId === 0) return; // Skip history for local host for now if not persisted
        try {
            const data = await GetHostMetrics(hostId, "24h");
            if (data) {
                // Format timestamp
                const formattedData = data.map((d: any) => ({
                    ...d,
                    time: new Date(d.timestamp).toLocaleTimeString(),
                }));
                setMetrics(formattedData);
            }
        } catch (err) {
            console.error("Failed to load metrics history:", err);
        }
    };

    const formatBytes = (bytes: number | undefined) => {
        if (bytes === undefined || bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    };

    if (loading && !status) {
        return <Loading text="Loading monitor data..." className="py-20" />;
    }

    if (error && !status) {
        return (
            <div className="p-8 text-center text-destructive">
                <p>Failed to load monitor data</p>
                <p className="text-sm mt-2">{error}</p>
                <Button variant="outline" size="sm" onClick={loadStatus} className="mt-4">
                    Retry
                </Button>
            </div>
        );
    }

    if (!status) return null;

    const tabs = [
        { id: 'overview', label: '总览', icon: Activity },
        { id: 'history', label: '历史趋势', icon: Clock },
        { id: 'cpu', label: 'CPU', icon: Cpu },
        { id: 'memory', label: '内存', icon: Zap },
        { id: 'disk', label: '磁盘', icon: HardDrive },
        { id: 'network', label: '网络', icon: Network },
        { id: 'process', label: '进程', icon: List },
        { id: 'hardware', label: '硬件', icon: Server },
    ];

    return (
        <div className="h-full flex flex-col space-y-4 p-4">
            {/* Header Info */}
            <div className="flex items-center justify-between bg-card p-4 rounded-lg border shadow-sm">
                <div className="flex items-center gap-4">
                    <div className="p-2 bg-primary/10 rounded-full text-primary">
                        <Monitor size={24} />
                    </div>
                    <div>
                        <h2 className="text-lg font-bold">{status.system?.hostname || 'Unknown Host'}</h2>
                        <div className="text-sm text-muted-foreground flex items-center gap-2">
                            <span>{status.system?.os}</span>
                            <span>•</span>
                            <span>{status.system?.platform}</span>
                            <span>•</span>
                            <span className="font-mono">{status.system?.kernel_arch}</span>
                        </div>
                    </div>
                </div>
                <div className="flex gap-6 text-sm">
                    <div className="text-right">
                        <div className="text-muted-foreground">Uptime</div>
                        <div className="font-medium font-mono">{status.system?.uptime_str || '-'}</div>
                    </div>
                    <div className="text-right">
                        <div className="text-muted-foreground">Boot Time</div>
                        <div className="font-medium font-mono">
                            {status.system?.boot_time ? new Date(status.system.boot_time * 1000).toLocaleString() : '-'}
                        </div>
                    </div>
                    <div className="text-right">
                        <div className="text-muted-foreground">User</div>
                        <div className="font-medium">{status.system?.current_user || '-'}</div>
                    </div>
                </div>
            </div>

            {/* Tabs */}
            <div className="flex gap-2 overflow-x-auto pb-2">
                {tabs.map(tab => (
                    <button
                        key={tab.id}
                        onClick={() => setActiveTab(tab.id)}
                        className={`flex items-center gap-2 px-4 py-2 rounded-md text-sm font-medium transition-colors whitespace-nowrap
                            ${activeTab === tab.id 
                                ? 'bg-primary text-primary-foreground shadow-sm' 
                                : 'bg-muted/50 text-muted-foreground hover:bg-muted hover:text-foreground'
                            }`}
                    >
                        <tab.icon size={16} />
                        {tab.label}
                    </button>
                ))}
            </div>

            {/* Content Area */}
            <div className="flex-1 overflow-y-auto">
                <div className="space-y-4">
                    {/* OVERVIEW TAB */}
                    {activeTab === 'overview' && (
                        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                            <Card>
                                <CardHeader className="pb-2">
                                    <CardTitle className="text-sm font-medium text-muted-foreground">CPU Usage</CardTitle>
                                </CardHeader>
                                <CardContent>
                                    <div className="text-2xl font-bold">{status.cpu?.usage_total?.toFixed(1) || 0}%</div>
                                    <p className="text-xs text-muted-foreground mt-1">
                                        {status.cpu?.cores_logical} Cores / {status.cpu?.model}
                                    </p>
                                    <div className="h-1.5 w-full bg-muted mt-3 rounded-full overflow-hidden">
                                        <div 
                                            className="h-full bg-primary transition-all duration-500" 
                                            style={{ width: `${status.cpu?.usage_total || 0}%` }} 
                                        />
                                    </div>
                                </CardContent>
                            </Card>
                            <Card>
                                <CardHeader className="pb-2">
                                    <CardTitle className="text-sm font-medium text-muted-foreground">Memory Usage</CardTitle>
                                </CardHeader>
                                <CardContent>
                                    <div className="text-2xl font-bold">{status.memory?.usage?.toFixed(1) || 0}%</div>
                                    <p className="text-xs text-muted-foreground mt-1">
                                        {formatBytes(status.memory?.used)} / {formatBytes(status.memory?.total)}
                                    </p>
                                    <div className="h-1.5 w-full bg-muted mt-3 rounded-full overflow-hidden">
                                        <div 
                                            className="h-full bg-primary transition-all duration-500" 
                                            style={{ width: `${status.memory?.usage || 0}%` }} 
                                        />
                                    </div>
                                </CardContent>
                            </Card>
                            <Card>
                                <CardHeader className="pb-2">
                                    <CardTitle className="text-sm font-medium text-muted-foreground">Disk Usage</CardTitle>
                                </CardHeader>
                                <CardContent>
                                    <div className="text-2xl font-bold">{status.disk?.usage?.toFixed(1) || 0}%</div>
                                    <p className="text-xs text-muted-foreground mt-1">
                                        Used: {formatBytes(status.disk?.used)} / Total: {formatBytes(status.disk?.total)}
                                    </p>
                                    <div className="h-1.5 w-full bg-muted mt-3 rounded-full overflow-hidden">
                                        <div 
                                            className="h-full bg-primary transition-all duration-500" 
                                            style={{ width: `${status.disk?.usage || 0}%` }} 
                                        />
                                    </div>
                                </CardContent>
                            </Card>
                            <Card>
                                <CardHeader className="pb-2">
                                    <CardTitle className="text-sm font-medium text-muted-foreground">Network I/O</CardTitle>
                                </CardHeader>
                                <CardContent>
                                    <div className="flex justify-between items-baseline">
                                        <div className="text-sm font-medium">Rx: {formatBytes(status.network?.total_rx)}</div>
                                    </div>
                                    <div className="flex justify-between items-baseline mt-1">
                                        <div className="text-sm font-medium">Tx: {formatBytes(status.network?.total_tx)}</div>
                                    </div>
                                    <p className="text-xs text-muted-foreground mt-2">
                                        {status.network?.tcp_connections} TCP / {status.network?.udp_connections} UDP
                                    </p>
                                </CardContent>
                            </Card>
                            
                            {/* System Availability / Load */}
                            <Card className="col-span-full">
                                <CardHeader>
                                    <CardTitle className="text-base flex items-center gap-2">
                                        <Activity size={18} /> 系统负载
                                    </CardTitle>
                                </CardHeader>
                                <CardContent>
                                    <div className="flex items-center gap-8">
                                        <div>
                                            <div className="text-sm text-muted-foreground mb-1">Load Average (1m, 5m, 15m)</div>
                                            <div className="text-xl font-mono">{status.cpu?.load_avg || 'N/A'}</div>
                                        </div>
                                        <div>
                                            <div className="text-sm text-muted-foreground mb-1">Process Count</div>
                                            <div className="text-xl font-mono">{status.process?.total || 0}</div>
                                        </div>
                                        <div>
                                            <div className="text-sm text-muted-foreground mb-1">Service Status</div>
                                            <div className="text-sm font-medium text-green-500 flex items-center gap-1">
                                                <div className="w-2 h-2 rounded-full bg-green-500" />
                                                Active
                                            </div>
                                        </div>
                                    </div>
                                </CardContent>
                            </Card>
                        </div>
                    )}

                    {activeTab === 'history' && (
                        <div className="space-y-6">
                            <Card>
                                <CardHeader>
                                    <CardTitle>CPU & Memory History (24h)</CardTitle>
                                    <CardDescription>Historical resource usage trends.</CardDescription>
                                </CardHeader>
                                <CardContent className="h-[400px]">
                                    <ResponsiveContainer width="100%" height="100%">
                                        <LineChart data={metrics}>
                                            <CartesianGrid strokeDasharray="3 3" stroke="#333" />
                                            <XAxis dataKey="time" stroke="#888" fontSize={12} tickFormatter={(val) => val.split(':').slice(0,2).join(':')} />
                                            <YAxis stroke="#888" fontSize={12} domain={[0, 100]} />
                                            <Tooltip 
                                                contentStyle={{ backgroundColor: '#1f1f1f', border: '1px solid #333' }}
                                                itemStyle={{ fontSize: '12px' }}
                                            />
                                            <Legend />
                                            <Line type="monotone" dataKey="cpu_usage" name="CPU %" stroke="#3b82f6" strokeWidth={2} dot={false} />
                                            <Line type="monotone" dataKey="memory_usage" name="Memory %" stroke="#eab308" strokeWidth={2} dot={false} />
                                        </LineChart>
                                    </ResponsiveContainer>
                                </CardContent>
                            </Card>
                            
                            <Card>
                                <CardHeader>
                                    <CardTitle>Disk Usage History (24h)</CardTitle>
                                </CardHeader>
                                <CardContent className="h-[300px]">
                                    <ResponsiveContainer width="100%" height="100%">
                                        <LineChart data={metrics}>
                                            <CartesianGrid strokeDasharray="3 3" stroke="#333" />
                                            <XAxis dataKey="time" stroke="#888" fontSize={12} tickFormatter={(val) => val.split(':').slice(0,2).join(':')} />
                                            <YAxis stroke="#888" fontSize={12} domain={[0, 100]} />
                                            <Tooltip 
                                                contentStyle={{ backgroundColor: '#1f1f1f', border: '1px solid #333' }}
                                                itemStyle={{ fontSize: '12px' }}
                                            />
                                            <Legend />
                                            <Line type="monotone" dataKey="disk_usage" name="Disk %" stroke="#22c55e" strokeWidth={2} dot={false} />
                                        </LineChart>
                                    </ResponsiveContainer>
                                </CardContent>
                            </Card>
                        </div>
                    )}

                    {/* CPU TAB */}
                    {activeTab === 'cpu' && (
                        <div className="grid gap-4">
                            <Card>
                                <CardHeader><CardTitle>CPU 详情</CardTitle></CardHeader>
                                <CardContent className="space-y-4">
                                    <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                                        <div className="p-3 bg-muted/30 rounded-lg">
                                            <div className="text-xs text-muted-foreground">Model</div>
                                            <div className="font-medium text-sm">{status.cpu?.model}</div>
                                        </div>
                                        <div className="p-3 bg-muted/30 rounded-lg">
                                            <div className="text-xs text-muted-foreground">Frequency</div>
                                            <div className="font-medium text-sm">{status.cpu?.frequency?.toFixed(0)} MHz</div>
                                        </div>
                                        <div className="p-3 bg-muted/30 rounded-lg">
                                            <div className="text-xs text-muted-foreground">Logical Cores</div>
                                            <div className="font-medium text-sm">{status.cpu?.cores_logical}</div>
                                        </div>
                                        <div className="p-3 bg-muted/30 rounded-lg">
                                            <div className="text-xs text-muted-foreground">Physical Cores</div>
                                            <div className="font-medium text-sm">{status.cpu?.cores_physical}</div>
                                        </div>
                                    </div>
                                    
                                    <div>
                                        <h3 className="text-sm font-medium mb-2">Load Average</h3>
                                        <div className="p-3 bg-muted/30 rounded-lg font-mono text-sm">
                                            {status.cpu?.load_avg}
                                        </div>
                                    </div>
                                    
                                    {/* Per Core Usage */}
                                    {status.cpu?.usage_per_core && status.cpu.usage_per_core.length > 0 && (
                                        <div>
                                            <h3 className="text-sm font-medium mb-2">Per Core Usage</h3>
                                            <div className="grid grid-cols-2 sm:grid-cols-4 md:grid-cols-6 gap-2">
                                                {status.cpu.usage_per_core.map((usage, i) => (
                                                    <div key={i} className="bg-muted/30 p-2 rounded text-center">
                                                        <div className="text-xs text-muted-foreground mb-1">Core {i}</div>
                                                        <div className="h-16 w-4 mx-auto bg-muted rounded-full relative overflow-hidden">
                                                            <div 
                                                                className="absolute bottom-0 left-0 w-full bg-primary transition-all duration-500"
                                                                style={{ height: `${usage}%` }}
                                                            />
                                                        </div>
                                                        <div className="text-xs font-medium mt-1">{usage.toFixed(0)}%</div>
                                                    </div>
                                                ))}
                                            </div>
                                        </div>
                                    )}
                                </CardContent>
                            </Card>
                        </div>
                    )}

                    {/* MEMORY TAB */}
                    {activeTab === 'memory' && (
                        <div className="grid gap-4">
                            <Card>
                                <CardHeader><CardTitle>内存详情</CardTitle></CardHeader>
                                <CardContent className="space-y-6">
                                    {/* Physical Memory */}
                                    <div>
                                        <h3 className="text-sm font-medium mb-3">Physical Memory</h3>
                                        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-3">
                                            <div className="p-3 bg-muted/30 rounded-lg">
                                                <div className="text-xs text-muted-foreground">Total</div>
                                                <div className="font-medium text-sm">{formatBytes(status.memory?.total)}</div>
                                            </div>
                                            <div className="p-3 bg-muted/30 rounded-lg">
                                                <div className="text-xs text-muted-foreground">Used</div>
                                                <div className="font-medium text-sm">{formatBytes(status.memory?.used)}</div>
                                            </div>
                                            <div className="p-3 bg-muted/30 rounded-lg">
                                                <div className="text-xs text-muted-foreground">Free</div>
                                                <div className="font-medium text-sm">{formatBytes(status.memory?.free)}</div>
                                            </div>
                                            <div className="p-3 bg-muted/30 rounded-lg">
                                                <div className="text-xs text-muted-foreground">Usage</div>
                                                <div className="font-medium text-sm">{status.memory?.usage?.toFixed(1)}%</div>
                                            </div>
                                        </div>
                                        <div className="h-4 bg-muted rounded-full overflow-hidden relative">
                                            <div 
                                                className="h-full bg-primary absolute left-0 top-0" 
                                                style={{ width: `${status.memory?.usage || 0}%` }} 
                                            />
                                            <div className="absolute inset-0 flex items-center justify-center text-[10px] font-medium text-white mix-blend-difference">
                                                {status.memory?.usage?.toFixed(1)}%
                                            </div>
                                        </div>
                                    </div>

                                    {/* Swap Memory */}
                                    <div>
                                        <h3 className="text-sm font-medium mb-3">Swap Memory</h3>
                                        <div className="grid grid-cols-2 md:grid-cols-3 gap-4 mb-3">
                                            <div className="p-3 bg-muted/30 rounded-lg">
                                                <div className="text-xs text-muted-foreground">Total Swap</div>
                                                <div className="font-medium text-sm">{formatBytes(status.memory?.swap_total)}</div>
                                            </div>
                                            <div className="p-3 bg-muted/30 rounded-lg">
                                                <div className="text-xs text-muted-foreground">Used Swap</div>
                                                <div className="font-medium text-sm">{formatBytes(status.memory?.swap_used)}</div>
                                            </div>
                                            <div className="p-3 bg-muted/30 rounded-lg">
                                                <div className="text-xs text-muted-foreground">Usage</div>
                                                <div className="font-medium text-sm">
                                                    {(status.memory?.swap_total ? (status.memory.swap_used / status.memory.swap_total * 100) : 0).toFixed(1)}%
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                </CardContent>
                            </Card>
                        </div>
                    )}

                    {/* DISK TAB */}
                    {activeTab === 'disk' && (
                        <div className="grid gap-4">
                            <Card>
                                <CardHeader><CardTitle>磁盘 I/O 统计</CardTitle></CardHeader>
                                <CardContent>
                                    <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                                        <div className="p-3 bg-muted/30 rounded-lg">
                                            <div className="text-xs text-muted-foreground">Read Bytes</div>
                                            <div className="font-medium text-sm">{formatBytes(status.disk?.read_bytes)}</div>
                                        </div>
                                        <div className="p-3 bg-muted/30 rounded-lg">
                                            <div className="text-xs text-muted-foreground">Write Bytes</div>
                                            <div className="font-medium text-sm">{formatBytes(status.disk?.write_bytes)}</div>
                                        </div>
                                        <div className="p-3 bg-muted/30 rounded-lg">
                                            <div className="text-xs text-muted-foreground">Read Ops</div>
                                            <div className="font-medium text-sm">{status.disk?.read_ops}</div>
                                        </div>
                                        <div className="p-3 bg-muted/30 rounded-lg">
                                            <div className="text-xs text-muted-foreground">Write Ops</div>
                                            <div className="font-medium text-sm">{status.disk?.write_ops}</div>
                                        </div>
                                    </div>
                                </CardContent>
                            </Card>

                            <Card>
                                <CardHeader><CardTitle>分区列表</CardTitle></CardHeader>
                                <CardContent>
                                    <div className="overflow-x-auto">
                                        <table className="w-full text-sm">
                                            <thead>
                                                <tr className="border-b">
                                                    <th className="text-left py-2 font-medium text-muted-foreground">Mount Point</th>
                                                    <th className="text-left py-2 font-medium text-muted-foreground">FS Type</th>
                                                    <th className="text-right py-2 font-medium text-muted-foreground">Total</th>
                                                    <th className="text-right py-2 font-medium text-muted-foreground">Used</th>
                                                    <th className="text-right py-2 font-medium text-muted-foreground">Usage</th>
                                                </tr>
                                            </thead>
                                            <tbody>
                                                {status.disk?.partitions?.map((p: any, i: number) => (
                                                    <tr key={i} className="border-b last:border-0 hover:bg-muted/50">
                                                        <td className="py-2 font-mono">{p.path}</td>
                                                        <td className="py-2 text-muted-foreground">{p.fstype}</td>
                                                        <td className="py-2 text-right">{formatBytes(p.total)}</td>
                                                        <td className="py-2 text-right">{formatBytes(p.used)}</td>
                                                        <td className="py-2 text-right">
                                                            <div className="flex items-center justify-end gap-2">
                                                                <span className="w-12 text-right">{p.usage?.toFixed(1)}%</span>
                                                                <div className="w-16 h-2 bg-muted rounded-full overflow-hidden">
                                                                    <div 
                                                                        className={`h-full ${p.usage > 90 ? 'bg-destructive' : p.usage > 70 ? 'bg-yellow-500' : 'bg-green-500'}`} 
                                                                        style={{ width: `${p.usage}%` }} 
                                                                    />
                                                                </div>
                                                            </div>
                                                        </td>
                                                    </tr>
                                                ))}
                                            </tbody>
                                        </table>
                                    </div>
                                </CardContent>
                            </Card>
                        </div>
                    )}

                    {/* NETWORK TAB */}
                    {activeTab === 'network' && (
                        <div className="grid gap-4">
                            <Card>
                                <CardHeader><CardTitle>网络概览</CardTitle></CardHeader>
                                <CardContent>
                                    <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                                        <div className="p-3 bg-muted/30 rounded-lg">
                                            <div className="text-xs text-muted-foreground">Total Rx</div>
                                            <div className="font-medium text-sm">{formatBytes(status.network?.total_rx)}</div>
                                        </div>
                                        <div className="p-3 bg-muted/30 rounded-lg">
                                            <div className="text-xs text-muted-foreground">Total Tx</div>
                                            <div className="font-medium text-sm">{formatBytes(status.network?.total_tx)}</div>
                                        </div>
                                        <div className="p-3 bg-muted/30 rounded-lg">
                                            <div className="text-xs text-muted-foreground">TCP Connections</div>
                                            <div className="font-medium text-sm">{status.network?.tcp_connections}</div>
                                        </div>
                                        <div className="p-3 bg-muted/30 rounded-lg">
                                            <div className="text-xs text-muted-foreground">UDP Connections</div>
                                            <div className="font-medium text-sm">{status.network?.udp_connections}</div>
                                        </div>
                                    </div>
                                    
                                    <div className="mt-4">
                                        <h4 className="text-sm font-medium mb-2">Listen Ports</h4>
                                        <div className="flex flex-wrap gap-2">
                                            {status.network?.listen_ports?.map((port: number) => (
                                                <span key={port} className="px-2 py-1 bg-secondary text-secondary-foreground text-xs rounded-md font-mono">
                                                    {port}
                                                </span>
                                            ))}
                                            {(!status.network?.listen_ports || status.network.listen_ports.length === 0) && (
                                                <span className="text-muted-foreground text-sm">No ports detected</span>
                                            )}
                                        </div>
                                    </div>
                                </CardContent>
                            </Card>

                            <Card>
                                <CardHeader><CardTitle>网卡接口</CardTitle></CardHeader>
                                <CardContent>
                                    <div className="overflow-x-auto">
                                        <table className="w-full text-sm">
                                            <thead>
                                                <tr className="border-b">
                                                    <th className="text-left py-2 font-medium text-muted-foreground">Interface</th>
                                                    <th className="text-left py-2 font-medium text-muted-foreground">IP Address</th>
                                                    <th className="text-right py-2 font-medium text-muted-foreground">Rx</th>
                                                    <th className="text-right py-2 font-medium text-muted-foreground">Tx</th>
                                                </tr>
                                            </thead>
                                            <tbody>
                                                {status.network?.interfaces?.map((iface: any, i: number) => (
                                                    <tr key={i} className="border-b last:border-0 hover:bg-muted/50">
                                                        <td className="py-2 font-medium">{iface.name}</td>
                                                        <td className="py-2 font-mono text-muted-foreground">{iface.ip || '-'}</td>
                                                        <td className="py-2 text-right">{formatBytes(iface.rx)}</td>
                                                        <td className="py-2 text-right">{formatBytes(iface.tx)}</td>
                                                    </tr>
                                                ))}
                                            </tbody>
                                        </table>
                                    </div>
                                </CardContent>
                            </Card>
                        </div>
                    )}

                    {/* PROCESS TAB */}
                    {activeTab === 'process' && (
                        <Card className="h-full flex flex-col">
                            <CardHeader>
                                <div className="flex justify-between items-center">
                                    <CardTitle>Top Processes</CardTitle>
                                    <span className="text-sm text-muted-foreground">Total: {status.process?.total}</span>
                                </div>
                            </CardHeader>
                            <CardContent className="flex-1 overflow-auto">
                                <table className="w-full text-sm">
                                    <thead className="sticky top-0 bg-card z-10">
                                        <tr className="border-b">
                                            <th className="text-left py-2 font-medium text-muted-foreground w-16">PID</th>
                                            <th className="text-left py-2 font-medium text-muted-foreground">Name</th>
                                            <th className="text-left py-2 font-medium text-muted-foreground hidden md:table-cell">Path</th>
                                            <th className="text-right py-2 font-medium text-muted-foreground">CPU%</th>
                                            <th className="text-right py-2 font-medium text-muted-foreground">Mem%</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {status.process?.list?.map((proc: any) => (
                                            <tr key={proc.pid} className="border-b last:border-0 hover:bg-muted/50">
                                                <td className="py-2 font-mono text-muted-foreground">{proc.pid}</td>
                                                <td className="py-2 font-medium">{proc.name}</td>
                                                <td className="py-2 text-muted-foreground text-xs hidden md:table-cell truncate max-w-[200px]" title={proc.path}>
                                                    {proc.path}
                                                </td>
                                                <td className="py-2 text-right font-mono">{proc.cpu?.toFixed(1)}</td>
                                                <td className="py-2 text-right font-mono">{proc.mem?.toFixed(1)}</td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                            </CardContent>
                        </Card>
                    )}

                    {/* HARDWARE TAB */}
                    {activeTab === 'hardware' && (
                        <Card>
                            <CardHeader><CardTitle>硬件信息</CardTitle></CardHeader>
                            <CardContent>
                                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                    <div className="space-y-4">
                                        <div className="p-4 bg-muted/20 rounded-lg border">
                                            <div className="text-sm font-medium text-muted-foreground mb-1">Motherboard</div>
                                            <div className="font-medium">{status.hardware?.motherboard || 'Unknown'}</div>
                                            <div className="text-xs text-muted-foreground mt-1">{status.hardware?.baseboard}</div>
                                        </div>
                                        <div className="p-4 bg-muted/20 rounded-lg border">
                                            <div className="text-sm font-medium text-muted-foreground mb-1">BIOS</div>
                                            <div className="font-medium">{status.hardware?.bios || 'Unknown'}</div>
                                        </div>
                                    </div>
                                    <div className="space-y-4">
                                        <div className="p-4 bg-muted/20 rounded-lg border">
                                            <div className="text-sm font-medium text-muted-foreground mb-1">CPU Model</div>
                                            <div className="font-medium">{status.cpu?.model || 'Unknown'}</div>
                                        </div>
                                        <div className="p-4 bg-muted/20 rounded-lg border">
                                            <div className="text-sm font-medium text-muted-foreground mb-1">Memory Model</div>
                                            <div className="font-medium">{status.hardware?.memory_model || 'Unknown'}</div>
                                        </div>
                                        <div className="p-4 bg-muted/20 rounded-lg border">
                                            <div className="text-sm font-medium text-muted-foreground mb-1">Disk Model</div>
                                            <div className="font-medium">{status.hardware?.disk_model || 'Unknown'}</div>
                                        </div>
                                    </div>
                                </div>
                            </CardContent>
                        </Card>
                    )}
                </div>
            </div>
        </div>
    );
}
