import { useState, useEffect } from 'react';
import { PageContainer } from '../../components/layout/PageContainer';
import { PageHeader } from '../../components/layout/PageHeader';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '../../components/ui/Card';
import { Button } from '../../components/ui/Button';
import { Box, Plus, Code, Edit } from 'lucide-react';
import { PluginEditor } from '../../components/node/PluginEditor';
// @ts-ignore
import { GetNodeModules, CreatePlugin } from '../../../wailsjs/go/app/App';
import { node } from '../../../wailsjs/go/models';

export function PluginManager() {
    const [plugins, setPlugins] = useState<node.Plugin[]>([]);
    const [loading, setLoading] = useState(false);
    const [showCreate, setShowCreate] = useState(false);
    const [editingPlugin, setEditingPlugin] = useState<string | null>(null);
    
    // Create form state
    const [newPluginName, setNewPluginName] = useState('');
    const [newPluginDesc, setNewPluginDesc] = useState('');
    const [createError, setCreateError] = useState('');

    useEffect(() => {
        loadPlugins();
    }, []);

    const loadPlugins = async () => {
        setLoading(true);
        try {
            const result = await GetNodeModules();
            setPlugins(result || []);
        } catch (err) {
            console.error("Failed to load plugins:", err);
        } finally {
            setLoading(false);
        }
    };

    const handleCreate = async () => {
        if (!newPluginName) {
            setCreateError("Name is required");
            return;
        }
        setCreateError('');
        try {
            await CreatePlugin(newPluginName, newPluginDesc);
            setShowCreate(false);
            setNewPluginName('');
            setNewPluginDesc('');
            loadPlugins();
        } catch (err: any) {
            setCreateError(err.toString());
        }
    };

    return (
        <PageContainer>
            <PageHeader 
                title="插件管理" 
                description="查看、管理和开发 Node 功能插件。" 
                action={
                    <Button onClick={() => setShowCreate(true)}>
                        <Plus size={16} className="mr-2" />
                        新建插件
                    </Button>
                }
            />

            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                {plugins.map(p => (
                    <Card key={p.name} className="hover:shadow-lg hover:border-primary/50 transition-all duration-200 group">
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2 text-lg group-hover:text-primary transition-colors">
                                <Box size={20} className="text-muted-foreground group-hover:text-primary transition-colors" />
                                {p.name}
                            </CardTitle>
                            <CardDescription className="font-mono text-[10px] bg-muted px-2 py-0.5 rounded w-fit mt-1 uppercase tracking-wider">
                                {p.tag || 'LATEST'}
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            <p className="text-sm text-muted-foreground leading-relaxed mb-4 min-h-[40px] line-clamp-2">
                                {p.description || "暂无描述"}
                            </p>
                            <div className="flex items-center gap-2 text-xs text-muted-foreground bg-muted/30 p-2 rounded border border-dashed font-mono break-all mb-4 group-hover:bg-muted/50 transition-colors">
                                <Code size={12} className="shrink-0" />
                                <span className="truncate">{p.path}</span>
                            </div>
                            <Button variant="outline" size="sm" className="w-full group-hover:border-primary/50 group-hover:text-primary transition-colors" onClick={() => setEditingPlugin(p.name)}>
                                <Edit size={14} className="mr-2" />
                                编辑源码
                            </Button>
                        </CardContent>
                    </Card>
                ))}
                
                {plugins.length === 0 && !loading && (
                    <div className="col-span-full text-center py-12 text-muted-foreground">
                        未找到插件。点击右上角创建新插件。
                    </div>
                )}
            </div>

            {/* Create Dialog */}
            {showCreate && (
                <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm animate-in fade-in duration-200">
                    <Card className="w-[420px] bg-background shadow-2xl border border-border scale-100 animate-in zoom-in-95 duration-200">
                        <CardHeader>
                            <CardTitle className="text-xl">新建插件</CardTitle>
                            <CardDescription>创建一个新的 Node 功能模块 (Go语言)。</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-5">
                            <div className="space-y-2">
                                <label className="text-sm font-medium text-foreground/80">插件名称 (英文)</label>
                                <input 
                                    className="w-full px-3 py-2.5 rounded-md border border-input bg-background text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 transition-all"
                                    placeholder="例如: redis_monitor"
                                    value={newPluginName}
                                    onChange={(e) => setNewPluginName(e.target.value)}
                                    autoFocus
                                />
                                <p className="text-[10px] text-muted-foreground">只能包含小写字母、数字和下划线。</p>
                            </div>
                            <div className="space-y-2">
                                <label className="text-sm font-medium text-foreground/80">功能描述</label>
                                <textarea 
                                    className="w-full px-3 py-2.5 rounded-md border border-input bg-background text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 min-h-[100px] resize-none transition-all"
                                    placeholder="简要描述该插件的功能..."
                                    value={newPluginDesc}
                                    onChange={(e) => setNewPluginDesc(e.target.value)}
                                />
                            </div>
                            {createError && (
                                <div className="p-3 rounded-md bg-destructive/10 border border-destructive/20 text-xs text-destructive flex items-center gap-2">
                                    <span>⚠️</span> {createError}
                                </div>
                            )}
                            <div className="flex justify-end gap-3 pt-2">
                                <Button variant="ghost" onClick={() => setShowCreate(false)} className="hover:bg-muted">取消</Button>
                                <Button onClick={handleCreate} className="px-6">创建插件</Button>
                            </div>
                        </CardContent>
                    </Card>
                </div>
            )}

            {/* Plugin Editor */}
            {editingPlugin && (
                <PluginEditor 
                    pluginName={editingPlugin} 
                    onClose={() => setEditingPlugin(null)} 
                />
            )}
        </PageContainer>
    );
}
