import { useState, useEffect } from "react";
import { PageContainer } from "../components/layout/PageContainer";
import { PageHeader } from "../components/layout/PageHeader";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../components/ui/Card";
import { Button } from "../components/ui/Button";
import { Database, RefreshCw } from "lucide-react";
import { useToast } from "../components/ui/ToastProvider";
import { ConfirmDialog } from "../components/ui/ConfirmDialog";
// @ts-ignore
import { InitDB, GetDBPath, GetRetentionHours, SetRetentionHours } from "../../wailsjs/go/app/App";

export function SettingsPage() {
    const toast = useToast();
    const [dbPath, setDbPath] = useState("");
    const [retentionHours, setRetentionHours] = useState(24);
    const [loading, setLoading] = useState(false);
    const [saving, setSaving] = useState(false);
    const [isInitDialogOpen, setIsInitDialogOpen] = useState(false);

    useEffect(() => {
        loadDBPath();
        loadRetention();
    }, []);

    const loadRetention = async () => {
        try {
            const hours = await GetRetentionHours();
            setRetentionHours(hours);
        } catch (err) {
            console.error("Failed to load retention hours:", err);
        }
    };

    const handleSaveRetention = async () => {
        setSaving(true);
        try {
            await SetRetentionHours(retentionHours);
            toast.success('设置已保存');
        } catch (err) {
            console.error("Failed to save retention hours:", err);
            toast.error('设置保存失败', String(err));
        } finally {
            setSaving(false);
        }
    };

    const loadDBPath = async () => {
        try {
            const path = await GetDBPath();
            setDbPath(path);
        } catch (err) {
            console.error("Failed to load DB path:", err);
        }
    };

    const handleInitDB = async () => {
        setLoading(true);
        try {
            await InitDB();
            toast.success('数据库初始化完成');
        } catch (err) {
            console.error("Init DB failed:", err);
            toast.error('数据库初始化失败', String(err));
        } finally {
            setLoading(false);
            setIsInitDialogOpen(false);
        }
    };

    return (
        <PageContainer>
            <PageHeader
                title="设置"
                description="管理您的偏好和应用设置。"
            />

            <ConfirmDialog
                open={isInitDialogOpen}
                title="确认初始化数据库？"
                content="此操作将重置数据库结构，所有现有数据将丢失！"
                onConfirm={handleInitDB}
                onCancel={() => setIsInitDialogOpen(false)}
                confirmText="确认初始化"
                loadingText="初始化中..."
            />

            <div className="grid gap-6">
                <Card>
                    <CardHeader>
                        <CardTitle>数据保留策略</CardTitle>
                        <CardDescription>配置历史数据的保留时间。</CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="flex flex-col gap-4">
                            <div className="flex items-center gap-4">
                                <label className="text-sm font-medium whitespace-nowrap">监测数据保留时间（小时）</label>
                                <input
                                    type="number"
                                    min="1"
                                    className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                                    value={retentionHours}
                                    onChange={(e) => setRetentionHours(parseInt(e.target.value) || 24)}
                                />
                            </div>
                            <Button
                                onClick={handleSaveRetention}
                                disabled={saving}
                                className="w-full sm:w-auto self-start"
                            >
                                {saving ? "保存中..." : "保存设置"}
                            </Button>
                        </div>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle>数据库设置</CardTitle>
                        <CardDescription>管理本地数据库配置。</CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="flex flex-col gap-2">
                            <label className="text-sm font-medium">数据库路径</label>
                            <div className="p-2 bg-secondary rounded text-sm font-mono break-all">
                                {dbPath || "加载中..."}
                            </div>
                        </div>

                        <div className="pt-4 border-t">
                            <Button
                                variant="destructive"
                                onClick={() => setIsInitDialogOpen(true)}
                                disabled={loading}
                                className="w-full sm:w-auto"
                            >
                                <RefreshCw className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
                                初始化数据库
                            </Button>
                            <p className="text-xs text-muted-foreground mt-2">
                                警告：这将使用内置的 schema.sql 文件重置数据库结构。
                            </p>
                        </div>
                    </CardContent>
                </Card>
            </div>
        </PageContainer>
    );
}
