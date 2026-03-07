import { useState, useEffect } from 'react';
import { createPortal } from 'react-dom';
import { useParams, useNavigate } from 'react-router-dom';
import { PageContainer } from '../components/layout/PageContainer';
import { PageHeader } from '../components/layout/PageHeader';
import { Button } from '../components/ui/Button';
import { Input } from '../components/ui/Input';
// @ts-ignore
import { GetHosts, CreateHost, DeleteHost, DeployNode, SelectFile, ConnectSSH, IsConnected } from '../../wailsjs/go/app/App';
// @ts-ignore
import { host } from '../../wailsjs/go/models';
import { 
    Trash2, Plus, Server, Key, Zap, MonitorSmartphone, 
    Monitor, Settings, ChevronDown, ChevronRight, CheckCircle, Terminal, Power, PowerOff
} from 'lucide-react';
import { SessionView } from '../components/remote/SessionView';
import { Loading } from '../components/ui/Loading';

function Modal({ open, onClose, title, children }: { open: boolean, onClose: () => void, title: string, children: React.ReactNode }) {
    if (!open) return null;
    return createPortal(
        <div className="fixed inset-0 z-50 flex items-center justify-center">
            <div className="absolute inset-0 bg-black/50 backdrop-blur-sm" onClick={onClose} />
            <div className="relative bg-background p-6 rounded-lg shadow-xl w-[450px] border z-50 max-h-[90vh] overflow-y-auto">
                <div className="flex justify-between items-center mb-4">
                    <h3 className="text-lg font-bold">{title}</h3>
                    <button onClick={onClose} className="text-muted-foreground hover:text-foreground">✕</button>
                </div>
                {children}
            </div>
        </div>,
        document.body
    );
}

export function RemotePage() {
    const { id } = useParams();
    const navigate = useNavigate();
    const [hosts, setHosts] = useState<host.Host[]>([]);
    const [loading, setLoading] = useState(false);
    const [connectionStatus, setConnectionStatus] = useState<Record<number, boolean>>({});
    const [connectingHosts, setConnectingHosts] = useState<number[]>([]); // Track connecting state
    
    // Main Page State
    const [activeTab, setActiveTab] = useState<'hosts' | 'batch'>('hosts');
    // Removed expandedHostId since we use routing now
    
    // Create Host State
    const [isCreating, setIsCreating] = useState(false);
    const [newHost, setNewHost] = useState<any>({
        name: '', ip: '', port: 22, username: '', password: '', auth_type: 'password', key_path: ''
    });

    // Batch Ops State
    const [selectedHosts, setSelectedHosts] = useState<number[]>([]);
    const [batchProcessing, setBatchProcessing] = useState(false);

    useEffect(() => {
        loadHosts();
    }, []);

    const loadHosts = async () => {
        // Only show full loading if we don't have hosts yet
        if (hosts.length === 0) setLoading(true);
        try {
            const result = await GetHosts();
            setHosts(result || []);
            // Check connections after loading hosts
            if (result && result.length > 0) {
                checkConnections(result);
            }
        } catch (err) {
            console.error("Failed to load hosts:", err);
        } finally {
            setLoading(false);
        }
    };

    const checkConnections = async (hostList: host.Host[]) => {
        const statuses: Record<number, boolean> = {};
        for (const h of hostList) {
            try {
                const connected = await IsConnected(h.id);
                statuses[h.id] = connected;
            } catch (e) {
                statuses[h.id] = false;
            }
        }
        setConnectionStatus(prev => ({ ...prev, ...statuses }));
    };

    const handleConnect = async (hostId: number, e?: React.MouseEvent) => {
        if (e) e.stopPropagation();
        
        setConnectingHosts(prev => [...prev, hostId]);
        try {
            await ConnectSSH(hostId);
            setConnectionStatus(prev => ({ ...prev, [hostId]: true }));
        } catch (err) {
            console.error("Connection failed:", err);
            alert("连接失败: " + err);
            setConnectionStatus(prev => ({ ...prev, [hostId]: false }));
        } finally {
            setConnectingHosts(prev => prev.filter(id => id !== hostId));
        }
    };

    // --- Host Management ---
    
    const handleCreate = async () => {
        try {
            if (!newHost.name || !newHost.ip || !newHost.username) return alert("请填写必要信息");
            await CreateHost(newHost);
            setIsCreating(false);
            setNewHost({ name: '', ip: '', port: 22, username: '', password: '', auth_type: 'password', key_path: '' });
            loadHosts();
        } catch (err) {
            alert("创建失败: " + err);
        }
    };

    const handleDelete = async (hostId: number, e: React.MouseEvent) => {
        e.stopPropagation();
        e.preventDefault();
        // Use window.confirm for now as it's reliable
        if (window.confirm('确定删除此主机?')) {
            try {
                await DeleteHost(hostId);
                // If we are currently viewing this host, go back to list
                if (id && parseInt(id) === hostId) {
                    navigate('/remote');
                }
                loadHosts();
            } catch (err) {
                console.error("Delete failed:", err);
                alert("删除失败: " + err);
            }
        }
    };

    const startCreate = (type: 'linux' | 'windows') => {
        setNewHost({
            name: '', ip: '', 
            port: type === 'windows' ? 3389 : 22,
            username: type === 'linux' ? 'root' : 'Administrator',
            password: '', auth_type: 'password', key_path: ''
        });
        setIsCreating(true);
    };

    const handleSelectKey = async () => {
        const path = await SelectFile();
        if (path) setNewHost((prev: any) => ({ ...prev, key_path: path }));
    };

    // --- Batch Operations ---

    const toggleHostSelection = (id: number) => {
        setSelectedHosts(prev => prev.includes(id) ? prev.filter(x => x !== id) : [...prev, id]);
    };

    const handleSelectAll = () => {
        if (selectedHosts.length === hosts.length) setSelectedHosts([]);
        else setSelectedHosts(hosts.map(h => h.id));
    };

    const handleBatchDeploy = async () => {
        if (selectedHosts.length === 0) return alert("请先选择主机");
        const input = prompt("请输入目标系统 (格式: OS/Arch, 如 linux/amd64):", "linux/amd64");
        if (!input) return;
        const [os, arch] = input.split('/');
        
        setBatchProcessing(true);
        try {
            for (const id of selectedHosts) {
                try {
                    await DeployNode(id, os, arch);
                } catch (e) {
                    console.error(`Deploy failed for host ${id}:`, e);
                }
            }
            alert("批量部署完成 (查看日志详情)");
        } finally {
            setBatchProcessing(false);
        }
    };

    // --- Render Details View if ID is present ---
    if (id) {
        const hostId = parseInt(id);
        const currentHost = hosts.find(h => h.id === hostId);
        
        // Even if hosts are not loaded yet, we can render the SessionView 
        // (SessionView fetches its own monitor data, but hostName might be missing initially)
        return (
            <div className="h-full bg-background animate-in fade-in duration-200">
                <SessionView 
                    hostId={hostId} 
                    hostName={currentHost?.name}
                    onClose={() => navigate('/remote')} 
                />
            </div>
        );
    }

    return (
        <PageContainer className="h-full flex flex-col space-y-0">
            <div className="flex items-center justify-between mb-4">
                <PageHeader title="远程管理" description="管理主机、批量操作及部署 Node 节点。" />
                <div className="flex gap-2">
                    <div className="flex bg-muted rounded-md p-1">
                        <button
                            onClick={() => setActiveTab('hosts')}
                            className={`px-3 py-1 text-sm font-medium rounded-sm transition-all ${activeTab === 'hosts' ? 'bg-background shadow-sm text-foreground' : 'text-muted-foreground hover:text-foreground'}`}
                        >
                            主机列表
                        </button>
                        <button
                            onClick={() => setActiveTab('batch')}
                            className={`px-3 py-1 text-sm font-medium rounded-sm transition-all ${activeTab === 'batch' ? 'bg-background shadow-sm text-foreground' : 'text-muted-foreground hover:text-foreground'}`}
                        >
                            批量操作
                        </button>
                    </div>
                </div>
            </div>

            <div className="flex-1 overflow-y-auto p-1">
                {activeTab === 'hosts' && (
                    <div className="space-y-4">
                        <div className="flex justify-end gap-2">
                            <Button size="sm" onClick={() => startCreate('linux')} className="gap-2">
                                <Plus size={16} /> Linux 主机
                            </Button>
                        </div>

                        {loading && hosts.length === 0 ? <Loading /> : (
                            <div className="border rounded-md bg-card">
                                {hosts.length === 0 ? (
                                    <div className="p-8 text-center text-muted-foreground">
                                        <Server size={48} className="mx-auto mb-4 opacity-50" />
                                        <p>暂无主机，请点击右上角添加</p>
                                    </div>
                                ) : (
                                    <div className="divide-y">
                                        {hosts.map(h => (
                                            <div key={h.id} className="group">
                                                {/* Host Row */}
                                                <div 
                                                    className="p-4 flex items-center justify-between cursor-pointer hover:bg-muted/50 transition-colors"
                                                    onClick={() => navigate(`/remote/${h.id}`)}
                                                >
                                                    <div className="flex items-center gap-4">
                                                        <div className={`p-2 rounded-full ${h.port === 3389 ? 'bg-blue-100 text-blue-600' : 'bg-green-100 text-green-600'}`}>
                                                            {h.port === 3389 ? <MonitorSmartphone size={20} /> : <Server size={20} />}
                                                        </div>
                                                        <div>
                                                            <div className="font-medium flex items-center gap-2">
                                                                {h.name}
                                                                {connectingHosts.includes(h.id) ? (
                                                                    <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-yellow-100 text-yellow-800 animate-pulse">
                                                                        <Zap size={10} className="mr-1" /> 连接中...
                                                                    </span>
                                                                ) : connectionStatus[h.id] ? (
                                                                    <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-green-100 text-green-800">
                                                                        <CheckCircle size={10} className="mr-1" /> 在线
                                                                    </span>
                                                                ) : (
                                                                    <button 
                                                                        className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-600 hover:bg-gray-200 transition-colors"
                                                                        onClick={(e) => handleConnect(h.id, e)}
                                                                        title="点击连接"
                                                                    >
                                                                        <PowerOff size={10} className="mr-1" /> 离线 (点击连接)
                                                                    </button>
                                                                )}
                                                            </div>
                                                            <div className="text-sm text-muted-foreground flex items-center gap-4 mt-1">
                                                                <span className="flex items-center gap-1"><Monitor size={12} /> {h.ip}:{h.port}</span>
                                                                <span className="flex items-center gap-1"><Key size={12} /> {h.username}</span>
                                                            </div>
                                                        </div>
                                                    </div>
                                                    <div className="flex items-center gap-2">
                                                        <Button variant="ghost" size="icon" onClick={(e) => handleDelete(h.id, e)} className="text-muted-foreground hover:text-destructive z-10 relative">
                                                            <Trash2 size={16} />
                                                        </Button>
                                                        <div className="transition-transform duration-200 -rotate-90">
                                                            <ChevronDown size={20} className="text-muted-foreground" />
                                                        </div>
                                                    </div>
                                                </div>
                                            </div>
                                        ))}
                                    </div>
                                )}
                            </div>
                        )}
                    </div>
                )}

                {activeTab === 'batch' && (
                    <div className="space-y-4">
                        <div className="bg-muted/30 p-4 rounded-lg flex items-center justify-between">
                            <label className="flex items-center gap-2 font-medium cursor-pointer">
                                <input type="checkbox" 
                                    checked={selectedHosts.length === hosts.length && hosts.length > 0} 
                                    onChange={handleSelectAll}
                                    className="w-4 h-4 rounded border-gray-300"
                                />
                                全选 ({selectedHosts.length})
                            </label>
                            <div className="flex gap-2">
                                <Button size="sm" onClick={handleBatchDeploy} disabled={batchProcessing || selectedHosts.length === 0}>
                                    <Zap size={16} className="mr-2" />
                                    批量部署 Node
                                </Button>
                                <Button size="sm" variant="outline" disabled={batchProcessing || selectedHosts.length === 0} onClick={() => alert("功能开发中")}>
                                    <Terminal size={16} className="mr-2" />
                                    执行命令
                                </Button>
                            </div>
                        </div>

                        <div className="border rounded-md bg-card">
                            <table className="w-full text-sm text-left">
                                <thead className="bg-muted/50 border-b">
                                    <tr>
                                        <th className="p-3 w-10"></th>
                                        <th className="p-3">主机名称</th>
                                        <th className="p-3">地址</th>
                                        <th className="p-3">系统</th>
                                        <th className="p-3">状态</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {hosts.map(h => (
                                        <tr key={h.id} className="border-b last:border-0 hover:bg-muted/20">
                                            <td className="p-3 text-center">
                                                <input type="checkbox" 
                                                    checked={selectedHosts.includes(h.id)}
                                                    onChange={() => toggleHostSelection(h.id)}
                                                    className="w-4 h-4 rounded border-gray-300"
                                                />
                                            </td>
                                            <td className="p-3 font-medium">{h.name}</td>
                                            <td className="p-3 text-muted-foreground">{h.ip}</td>
                                            <td className="p-3">{h.port === 3389 ? 'Windows' : 'Linux/Unix'}</td>
                                            <td className="p-3">
                                                {connectionStatus[h.id] ? (
                                                    <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-green-100 text-green-800">在线</span>
                                                ) : (
                                                    <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-600">离线</span>
                                                )}
                                            </td>
                                        </tr>
                                    ))}
                                    {hosts.length === 0 && (
                                        <tr>
                                            <td colSpan={5} className="p-8 text-center text-muted-foreground">暂无主机</td>
                                        </tr>
                                    )}
                                </tbody>
                            </table>
                        </div>
                    </div>
                )}
            </div>

            {/* Create Host Modal */}
            <Modal 
                open={isCreating} 
                onClose={() => setIsCreating(false)} 
                title="添加主机"
            >
                <div className="space-y-4">
                    <div className="space-y-1">
                        <label className="text-sm font-medium">主机名称</label>
                        <Input 
                            value={newHost.name} 
                            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setNewHost({...newHost, name: e.target.value})} 
                            placeholder="My Server"
                        />
                    </div>
                    <div className="grid grid-cols-2 gap-4">
                        <div className="space-y-1">
                            <label className="text-sm font-medium">IP 地址</label>
                            <Input 
                                value={newHost.ip} 
                                onChange={(e: React.ChangeEvent<HTMLInputElement>) => setNewHost({...newHost, ip: e.target.value})} 
                                placeholder="192.168.1.1"
                            />
                        </div>
                        <div className="space-y-1">
                            <label className="text-sm font-medium">端口</label>
                            <Input 
                                type="number"
                                value={newHost.port} 
                                onChange={(e: React.ChangeEvent<HTMLInputElement>) => setNewHost({...newHost, port: parseInt(e.target.value)})} 
                            />
                        </div>
                    </div>
                    <div className="space-y-1">
                        <label className="text-sm font-medium">用户名</label>
                        <Input 
                            value={newHost.username} 
                            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setNewHost({...newHost, username: e.target.value})} 
                        />
                    </div>
                    
                    <div className="space-y-2">
                        <label className="text-sm font-medium">认证方式</label>
                        <div className="flex gap-4">
                            <label className="flex items-center gap-2 cursor-pointer">
                                <input 
                                    type="radio" 
                                    name="auth_type"
                                    checked={newHost.auth_type === 'password'}
                                    onChange={() => setNewHost({...newHost, auth_type: 'password'})}
                                />
                                <span>密码</span>
                            </label>
                            <label className="flex items-center gap-2 cursor-pointer">
                                <input 
                                    type="radio" 
                                    name="auth_type"
                                    checked={newHost.auth_type === 'key'}
                                    onChange={() => setNewHost({...newHost, auth_type: 'key'})}
                                />
                                <span>密钥</span>
                            </label>
                        </div>
                    </div>

                    {newHost.auth_type === 'password' ? (
                        <div className="space-y-1">
                            <label className="text-sm font-medium">密码</label>
                            <Input 
                                type="password"
                                value={newHost.password} 
                                onChange={(e: React.ChangeEvent<HTMLInputElement>) => setNewHost({...newHost, password: e.target.value})} 
                            />
                        </div>
                    ) : (
                        <div className="space-y-1">
                            <label className="text-sm font-medium">密钥路径</label>
                            <div className="flex gap-2">
                                <Input 
                                    value={newHost.key_path} 
                                    readOnly 
                                    placeholder="请选择密钥文件" 
                                />
                                <Button onClick={handleSelectKey} variant="outline">选择</Button>
                            </div>
                        </div>
                    )}

                    <div className="flex justify-end gap-2 pt-4">
                        <Button variant="ghost" onClick={() => setIsCreating(false)}>取消</Button>
                        <Button onClick={handleCreate}>添加</Button>
                    </div>
                </div>
            </Modal>
        </PageContainer>

    );
}