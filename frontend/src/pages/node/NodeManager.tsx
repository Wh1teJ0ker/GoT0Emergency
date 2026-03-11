import { useState, useEffect } from 'react';
import { PageContainer } from '../../components/layout/PageContainer';
import { PageHeader } from '../../components/layout/PageHeader';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '../../components/ui/Card';
import { Button } from '../../components/ui/Button';
import { Box, Terminal, Server, Download, Trash2, Activity, Wifi } from 'lucide-react';
// @ts-ignore
import { BuildNode, GetNodeModules, GetBuiltNodes } from '../../../wailsjs/go/app/App';
// @ts-ignore
import { EventsOn } from '../../../wailsjs/runtime/runtime';
import { node } from '../../../wailsjs/go/models';

export function NodeManager() {
    const [activeTab, setActiveTab] = useState<'build' | 'artifacts' | 'live'>('build');
    
    // Build State
    const [plugins, setPlugins] = useState<node.Plugin[]>([]);
    const [selectedTags, setSelectedTags] = useState<Record<string, boolean>>({});
    const [targetOS, setTargetOS] = useState('linux');
    const [targetArch, setTargetArch] = useState('amd64');
    const [building, setBuilding] = useState(false);
    const [output, setOutput] = useState<string>('');

    // Artifacts State
    const [artifacts, setArtifacts] = useState<string[]>([]);

    // Live Nodes State
    const [activeNodes, setActiveNodes] = useState<Record<string, any>>({});

    useEffect(() => {
        loadPlugins();
        loadArtifacts();

        // Subscribe to node callback
        // @ts-ignore
        const cleanup = EventsOn("node:callback", (data: any) => {
            setActiveNodes(prev => {
                const newState = { ...prev };
                // Use stable key: hostname-os-ip_without_port to avoid duplicates
                if (data && data.remote_ip) {
                    const ipWithoutPort = data.remote_ip.split(':')[0];
                    const key = data.hostname
                        ? `${data.hostname}-${data.os}-${ipWithoutPort}`
                        : `unknown-${data.os}-${Date.now()}`;
                    newState[key] = { ...data, last_seen: Date.now(), display_ip: ipWithoutPort };
                }
                return newState;
            });
        });

        // Cleanup mechanism for stale nodes (visual only, real cleanup happens on update)
        const interval = setInterval(() => {
            setActiveNodes(prev => ({...prev})); // Force re-render to update "last seen" or status color
        }, 5000);
        
        return () => {
            if (cleanup) cleanup();
            clearInterval(interval);
        };
    }, []);

    const loadPlugins = async () => {
        try {
            const result = await GetNodeModules();
            if (result) {
                setPlugins(result);
                // Default enable host_monitor
                const initialTags: Record<string, boolean> = {};
                result.forEach((p: node.Plugin) => {
                    initialTags[p.tag] = p.name === 'host_monitor';
                });
                setSelectedTags(initialTags);
            }
        } catch (err) {
            console.error(err);
        }
    };

    const loadArtifacts = async () => {
        try {
            const result = await GetBuiltNodes();
            setArtifacts(result || []);
        } catch (err) {
            console.error(err);
        }
    };

    const togglePlugin = (tag: string) => {
        setSelectedTags(prev => ({ ...prev, [tag]: !prev[tag] }));
    };

    const handleBuild = async () => {
        setBuilding(true);
        setOutput(`Building node for ${targetOS}/${targetArch}...\n`);
        
        const features = Object.entries(selectedTags)
            .filter(([_, enabled]) => enabled)
            .map(([tag]) => tag);

        try {
            const result = await BuildNode(features, targetOS, targetArch);
            setOutput(prev => prev + result + '\n');
            loadArtifacts(); // Refresh artifacts list
        } catch (err: any) {
            setOutput(prev => prev + 'Error: ' + err + '\n');
        } finally {
            setBuilding(false);
        }
    };

    return (
        <PageContainer>
            <PageHeader 
                title="Node 管理" 
                description="定制编译 Node，管理已构建的版本。" 
            />

            <div className="flex border-b mb-6">
                <button
                    className={`px-4 py-2 font-medium text-sm transition-colors border-b-2 ${activeTab === 'build' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'}`}
                    onClick={() => setActiveTab('build')}
                >
                    定制编译
                </button>
                <button
                    className={`px-4 py-2 font-medium text-sm transition-colors border-b-2 ${activeTab === 'artifacts' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'}`}
                    onClick={() => setActiveTab('artifacts')}
                >
                    构建产物
                </button>
                <button
                    className={`px-4 py-2 font-medium text-sm transition-colors border-b-2 ${activeTab === 'live' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground hover:text-foreground'}`}
                    onClick={() => setActiveTab('live')}
                >
                    实时监控
                </button>
            </div>

            {activeTab === 'build' && (
                <div className="grid gap-6 md:grid-cols-2">
                    <Card className="flex flex-col h-full overflow-hidden">
                        <CardHeader>
                            <CardTitle>编译配置</CardTitle>
                            <CardDescription>选择目标平台和插件</CardDescription>
                        </CardHeader>
                        <CardContent className="flex-1 overflow-y-auto space-y-6">
                            <div className="space-y-3">
                                <h3 className="text-sm font-medium">目标平台</h3>
                                <div className="grid grid-cols-2 gap-4">
                                    <div className="space-y-1">
                                        <label className="text-xs text-muted-foreground">操作系统</label>
                                        <select 
                                            className="w-full p-2 rounded-md border bg-background"
                                            value={targetOS}
                                            onChange={(e) => setTargetOS(e.target.value)}
                                        >
                                            <option value="linux">Linux</option>
                                            <option value="darwin">macOS</option>
                                            <option value="windows">Windows</option>
                                        </select>
                                    </div>
                                    <div className="space-y-1">
                                        <label className="text-xs text-muted-foreground">架构</label>
                                        <select 
                                            className="w-full p-2 rounded-md border bg-background"
                                            value={targetArch}
                                            onChange={(e) => setTargetArch(e.target.value)}
                                        >
                                            <option value="amd64">AMD64 (x86_64)</option>
                                            <option value="arm64">ARM64</option>
                                        </select>
                                    </div>
                                </div>
                            </div>

                            <div className="space-y-3">
                                <h3 className="text-sm font-medium">选择插件</h3>
                                <div className="space-y-2">
                                    {plugins.map(plugin => (
                                        <div 
                                            key={plugin.tag}
                                            className={`flex items-center justify-between p-3 border rounded-lg cursor-pointer transition-colors ${selectedTags[plugin.tag] ? 'bg-primary/5 border-primary/30' : 'hover:bg-muted/50'}`}
                                            onClick={() => togglePlugin(plugin.tag)}
                                        >
                                            <div className="flex items-center gap-3">
                                                <Box size={16} className="text-muted-foreground" />
                                                <span className="text-sm font-medium">{plugin.name}</span>
                                            </div>
                                            <input 
                                                type="checkbox" 
                                                checked={!!selectedTags[plugin.tag]}
                                                onChange={() => {}}
                                                className="h-4 w-4"
                                            />
                                        </div>
                                    ))}
                                </div>
                            </div>

                            <Button 
                                className="w-full" 
                                onClick={handleBuild} 
                                disabled={building}
                            >
                                {building ? '编译中...' : '开始编译'}
                            </Button>
                        </CardContent>
                    </Card>

                    <Card className="flex flex-col h-full overflow-hidden min-h-[400px]">
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2">
                                <Terminal size={18} />
                                构建日志
                            </CardTitle>
                        </CardHeader>
                        <CardContent className="flex-1 bg-black p-4 font-mono text-xs text-green-400 overflow-auto rounded-b-lg">
                            <pre className="whitespace-pre-wrap">{output || '> 等待构建...'}</pre>
                        </CardContent>
                    </Card>
                </div>
            )}

            {activeTab === 'artifacts' && (
                <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                    {artifacts.map(node => (
                        <Card key={node} className="hover:shadow-md transition-shadow">
                            <CardHeader className="pb-2">
                                <CardTitle className="text-base flex items-center gap-2">
                                    <Server size={18} className="text-primary" />
                                    {node}
                                </CardTitle>
                            </CardHeader>
                            <CardContent>
                                <div className="flex justify-between items-center mt-2">
                                    <span className="text-xs text-muted-foreground bg-muted px-2 py-1 rounded">
                                        已编译
                                    </span>
                                    {/* Placeholder for future actions like delete/download */}
                                </div>
                            </CardContent>
                        </Card>
                    ))}
                    {artifacts.length === 0 && (
                        <div className="col-span-full text-center py-12 text-muted-foreground">
                            暂无构建产物。请前往“定制编译”页签进行构建。
                        </div>
                    )}
                </div>
            )}

            {activeTab === 'live' && (
                <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                    {Object.entries(activeNodes).map(([key, node]: [string, any]) => {
                        const isOnline = Date.now() - node.last_seen < 10000; // 10s threshold
                        return (
                            <Card key={key} className={`hover:shadow-md transition-shadow ${!isOnline ? 'opacity-60 grayscale' : ''}`}>
                                <CardHeader className="pb-2">
                                    <CardTitle className="text-base flex items-center justify-between">
                                        <div className="flex items-center gap-2">
                                            <Activity size={18} className={isOnline ? "text-green-500" : "text-gray-400"} />
                                            {node.hostname || 'Unknown Host'}
                                        </div>
                                        <span className="text-xs font-mono bg-muted px-1.5 py-0.5 rounded">{node.display_ip || key}</span>
                                    </CardTitle>
                                    <CardDescription className="text-xs">
                                        {node.os} / {node.arch}
                                    </CardDescription>
                                </CardHeader>
                                <CardContent>
                                    <div className="grid grid-cols-2 gap-2 text-sm mt-2">
                                        <div className="flex flex-col">
                                            <span className="text-xs text-muted-foreground">CPU</span>
                                            <span className="font-medium">
                                                {node.modules?.host_monitor?.cpu_usage
                                                    ? node.modules.host_monitor.cpu_usage.toFixed(1) + '%'
                                                    : '0%'}
                                            </span>
                                        </div>
                                        <div className="flex flex-col">
                                            <span className="text-xs text-muted-foreground">Memory</span>
                                            <span className="font-medium">
                                                {node.modules?.host_monitor?.memory_percent
                                                    ? node.modules.host_monitor.memory_percent.toFixed(1) + '%'
                                                    : '0%'}
                                            </span>
                                        </div>
                                        <div className="flex flex-col col-span-2 mt-1">
                                            <span className="text-xs text-muted-foreground">Uptime</span>
                                            <span className="font-medium">{formatUptime(node.uptime)}</span>
                                        </div>
                                    </div>
                                    <div className="mt-3 flex items-center gap-1 text-xs text-muted-foreground">
                                        <Wifi size={12} />
                                        Last seen: {new Date(node.last_seen).toLocaleTimeString()}
                                    </div>
                                </CardContent>
                            </Card>
                        );
                    })}
                    {Object.keys(activeNodes).length === 0 && (
                        <div className="col-span-full text-center py-12 text-muted-foreground">
                            暂无在线节点。请运行 node agent 并指定回调地址。
                            <br />
                            <code className="text-xs bg-muted px-1 py-0.5 rounded mt-2 inline-block">./node -callback http://YOUR_IP:36911/api/callback -interval 5s</code>
                        </div>
                    )}
                </div>
            )}

        </PageContainer>
    );
}

function formatUptime(seconds: number): string {
    if (!seconds) return '0s';
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    
    if (days > 0) return `${days}d ${hours}h`;
    if (hours > 0) return `${hours}h ${minutes}m`;
    return `${minutes}m ${seconds % 60}s`;
}
