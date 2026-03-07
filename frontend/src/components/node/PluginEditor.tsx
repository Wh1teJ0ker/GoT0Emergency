import { useState, useEffect } from 'react';
import Editor from 'react-simple-code-editor';
// @ts-ignore
import { highlight, languages } from 'prismjs';
import 'prismjs/components/prism-clike';
import 'prismjs/components/prism-go';
import 'prismjs/components/prism-javascript';
import 'prismjs/components/prism-typescript';
import 'prismjs/themes/prism-solarizedlight.css'; // Light theme with good contrast
import { Button } from '../ui/Button';
import { Save, X, Loader2, FileCode, ZoomIn, ZoomOut, CheckCircle } from 'lucide-react';
// @ts-ignore
import { GetPluginSource, SavePluginSource } from '../../../wailsjs/go/app/App';

interface PluginEditorProps {
    pluginName: string;
    onClose: () => void;
}

export function PluginEditor({ pluginName, onClose }: PluginEditorProps) {
    const [code, setCode] = useState('');
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [fontSize, setFontSize] = useState(14);

    const [saved, setSaved] = useState(false);

    useEffect(() => {
        // Load font size preference
        const savedSize = localStorage.getItem('plugin-editor-fontsize');
        if (savedSize) {
            setFontSize(parseInt(savedSize, 10));
        }
        loadSource();
    }, [pluginName]);

    const changeFontSize = (delta: number) => {
        const newSize = Math.max(10, Math.min(32, fontSize + delta));
        setFontSize(newSize);
        localStorage.setItem('plugin-editor-fontsize', newSize.toString());
    };

    const loadSource = async () => {
        setLoading(true);
        setError(null);
        try {
            const content = await GetPluginSource(pluginName);
            setCode(content);
        } catch (err) {
            setError("Failed to load source: " + err);
        } finally {
            setLoading(false);
        }
    };

    const handleSave = async () => {
        setSaving(true);
        try {
            await SavePluginSource(pluginName, code);
            // Artificial delay to show saving state (UX)
            await new Promise(resolve => setTimeout(resolve, 600));
            setSaved(true);
            setTimeout(() => setSaved(false), 2000);
        } catch (err) {
            setError("Failed to save: " + err);
        } finally {
            setSaving(false);
        }
    };

    return (
        <div className="absolute inset-0 z-50 flex flex-col bg-background">
            {/* Header */}
            <div className="flex items-center justify-between px-6 py-4 border-b bg-card shadow-sm">
                <div className="flex items-center gap-3">
                    <div className="p-2 bg-blue-500/10 rounded-md text-blue-500">
                        <FileCode size={20} />
                    </div>
                    <div>
                        <h2 className="text-base font-semibold leading-tight">{pluginName}</h2>
                        <p className="text-xs text-muted-foreground mt-0.5">编辑插件源代码 (Go)</p>
                    </div>
                    {loading && <Loader2 className="animate-spin text-muted-foreground ml-2" size={16} />}
                </div>
                <div className="flex items-center gap-3">
                    <Button 
                        variant={saved ? "secondary" : "default"} 
                        size="sm" 
                        onClick={handleSave} 
                        disabled={loading || saving} 
                        className={`h-9 px-4 transition-all duration-300 ${saved ? "bg-green-600 hover:bg-green-700 text-white" : ""}`}
                    >
                        {saved ? (
                            <>
                                <CheckCircle size={16} className="mr-2" />
                                已保存
                            </>
                        ) : (
                            <>
                                <Save size={16} className="mr-2" />
                                {saving ? '保存中...' : '保存'}
                            </>
                        )}
                    </Button>
                    <Button variant="ghost" size="icon" onClick={onClose} className="h-9 w-9 text-muted-foreground hover:text-foreground">
                        <X size={20} />
                    </Button>
                </div>
            </div>

            {/* Editor Area */}
            <div className="flex-1 relative bg-muted/10">
                {/* Floating Font Size Control */}
                <div className="absolute top-4 right-6 z-20 flex items-center gap-1.5 bg-background/80 backdrop-blur-sm shadow-sm border rounded-md p-1 opacity-50 hover:opacity-100 transition-opacity">
                    <button 
                        onClick={() => changeFontSize(-1)} 
                        disabled={fontSize <= 10}
                        className="h-6 w-6 flex items-center justify-center rounded hover:bg-muted disabled:opacity-30 transition-all text-foreground"
                        title="缩小字号"
                    >
                        <ZoomOut size={12} />
                    </button>
                    <span className="w-6 text-center font-mono text-xs select-none">{fontSize}</span>
                    <button 
                        onClick={() => changeFontSize(1)} 
                        disabled={fontSize >= 32}
                        className="h-6 w-6 flex items-center justify-center rounded hover:bg-muted disabled:opacity-30 transition-all text-foreground"
                        title="放大字号"
                    >
                        <ZoomIn size={12} />
                    </button>
                </div>
                
                <div className="absolute inset-0 overflow-auto selection:bg-blue-500/20">
                    {error ? (
                        <div className="m-6 p-4 border border-destructive/20 rounded-lg bg-destructive/5 text-destructive text-sm flex items-start gap-3">
                            <div className="mt-0.5">⚠️</div>
                            <div>
                                <p className="font-semibold mb-1">加载失败</p>
                                <p className="opacity-80">{error}</p>
                            </div>
                        </div>
                    ) : (
                        <div className="min-h-full font-mono text-sm leading-relaxed p-6">
                            <div className="max-w-5xl mx-auto rounded-lg border border-border/50 bg-background shadow-lg overflow-hidden ring-1 ring-black/5">
                                <Editor
                                    value={code}
                                    onValueChange={code => setCode(code)}
                                    highlight={code => highlight(code, languages.go || languages.js, 'go')}
                                    padding={24}
                                    style={{
                                        fontFamily: '"Consolas", "Monaco", "Menlo", monospace',
                                        fontSize: fontSize,
                                        backgroundColor: '#ffffff',
                                        minHeight: '600px',
                                        lineHeight: '1.5',
                                        color: '#000000',
                                        fontWeight: 500,
                                    }}
                                    textareaClassName="focus:outline-none caret-black"
                                />
                            </div>
                        </div>
                    )}
                </div>
            </div>
            
            {/* Footer / Status Bar */}
            <div className="px-6 py-2 border-t bg-card text-[10px] text-muted-foreground flex justify-between items-center select-none">
                <div className="flex gap-4">
                    <span>Language: Go</span>
                    <span>UTF-8</span>
                </div>
                <div>
                    {code.length} characters
                </div>
            </div>
        </div>
    );
}
