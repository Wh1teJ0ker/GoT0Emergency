import { useState, useEffect } from 'react';
import Editor from 'react-simple-code-editor';
// @ts-ignore
import { highlight, languages } from 'prismjs/components/prism-core';
import 'prismjs/components/prism-clike';
import 'prismjs/components/prism-go';
import 'prismjs/themes/prism-tomorrow.css'; // Dark theme
import { Button } from '../ui/Button';
import { Save, X, Loader2 } from 'lucide-react';
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

    useEffect(() => {
        loadSource();
    }, [pluginName]);

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
            alert("Saved successfully!");
        } catch (err) {
            alert("Failed to save: " + err);
        } finally {
            setSaving(false);
        }
    };

    return (
        <div className="absolute inset-0 z-50 flex flex-col bg-background">
            {/* Header */}
            <div className="flex items-center justify-between p-4 border-b bg-background">
                <div className="flex items-center gap-4">
                    <h2 className="text-lg font-medium">Editing: {pluginName}</h2>
                    {loading && <Loader2 className="animate-spin text-muted-foreground" size={16} />}
                </div>
                <div className="flex items-center gap-2">
                    <Button variant="default" size="sm" onClick={handleSave} disabled={loading || saving}>
                        <Save size={16} className="mr-2" />
                        {saving ? 'Saving...' : 'Save'}
                    </Button>
                    <Button variant="ghost" size="sm" onClick={onClose}>
                        <X size={16} className="mr-2" />
                        Close
                    </Button>
                </div>
            </div>

            {/* Editor Area */}
            <div className="flex-1 overflow-auto p-4 bg-[#1d1f21]"> {/* Matches tomorrow theme somewhat */}
                {error ? (
                    <div className="text-destructive p-4 border border-destructive/50 rounded bg-destructive/10">
                        {error}
                    </div>
                ) : (
                    <div className="min-h-full font-mono text-sm">
                        <Editor
                            value={code}
                            onValueChange={code => setCode(code)}
                            highlight={code => highlight(code, languages.go, 'go')}
                            padding={10}
                            style={{
                                fontFamily: '"Fira code", "Fira Mono", monospace',
                                fontSize: 14,
                                backgroundColor: 'transparent',
                                minHeight: '100%'
                            }}
                            textareaClassName="focus:outline-none"
                        />
                    </div>
                )}
            </div>
        </div>
    );
}
