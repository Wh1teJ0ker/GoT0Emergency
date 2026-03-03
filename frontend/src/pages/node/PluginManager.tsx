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
                    <Card key={p.name} className="hover:shadow-md transition-shadow">
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2 text-lg">
                                <Box size={20} className="text-primary" />
                                {p.name}
                            </CardTitle>
                            <CardDescription className="font-mono text-xs bg-muted px-2 py-1 rounded w-fit mt-1">
                                {p.tag}
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            <p className="text-sm text-muted-foreground leading-relaxed mb-4">
                                {p.description}
                            </p>
                            <div className="flex items-center gap-2 text-xs text-muted-foreground bg-muted/50 p-2 rounded border font-mono break-all mb-4">
                                <Code size={12} />
                                {p.path}
                            </div>
                            <Button variant="outline" size="sm" className="w-full" onClick={() => setEditingPlugin(p.name)}>
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
                <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
                    <Card className="w-[400px] bg-background shadow-xl border border-border">
                        <CardHeader>
                            <CardTitle>新建插件</CardTitle>
                            <CardDescription>创建一个新的 Node 功能模块。</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="space-y-2">
                                <label className="text-sm font-medium">插件名称 (英文)</label>
                                <input 
                                    className="w-full p-2 rounded-md border bg-background"
                                    placeholder="例如: redis_monitor"
                                    value={newPluginName}
                                    onChange={(e) => setNewPluginName(e.target.value)}
                                />
                            </div>
                            <div className="space-y-2">
                                <label className="text-sm font-medium">描述</label>
                                <textarea 
                                    className="w-full p-2 rounded-md border bg-background min-h-[80px]"
                                    placeholder="描述该插件的功能..."
                                    value={newPluginDesc}
                                    onChange={(e) => setNewPluginDesc(e.target.value)}
                                />
                            </div>
                            {createError && (
                                <p className="text-xs text-red-500">{createError}</p>
                            )}
                            <div className="flex justify-end gap-2 mt-4">
                                <Button variant="ghost" onClick={() => setShowCreate(false)}>取消</Button>
                                <Button onClick={handleCreate}>创建</Button>
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
