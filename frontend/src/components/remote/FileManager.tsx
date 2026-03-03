import { useState, useEffect } from 'react';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';
import { Folder, File, ArrowUp, RefreshCw, Upload, Download, Home, FileText } from 'lucide-react';
// @ts-ignore
import { ListRemoteFiles, UploadFile, DownloadFile, SelectFile, SelectSaveFile } from '../../../wailsjs/go/app/App';

interface FileInfo {
    name: string;
    size: number;
    is_dir: boolean;
    mod_time: string;
}

interface FileManagerProps {
    hostId: number;
}

export function FileManager({ hostId }: FileManagerProps) {
    const [currentPath, setCurrentPath] = useState('.');
    const [files, setFiles] = useState<FileInfo[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        loadFiles(currentPath);
    }, [hostId]);

    const loadFiles = async (path: string) => {
        setLoading(true);
        setError(null);
        try {
            const result = await ListRemoteFiles(hostId, path);
            setFiles(result || []);
            setCurrentPath(path);
        } catch (err) {
            console.error("Failed to list files:", err);
            setError(String(err));
        } finally {
            setLoading(false);
        }
    };

    const handleNavigate = (file: FileInfo) => {
        if (file.name === '..') {
            // Simple parent directory logic (imperfect for complex paths but works for basic unix)
            if (currentPath === '.' || currentPath === '/') return; // Already at root-ish
            const parts = currentPath.split('/');
            parts.pop();
            const newPath = parts.join('/') || '/';
            loadFiles(newPath);
        } else if (file.is_dir) {
            const newPath = currentPath === '.' || currentPath === '/' 
                ? file.name 
                : `${currentPath}/${file.name}`; // Simple concatenation, ideally handle trailing slashes
            loadFiles(newPath);
        }
    };

    const handleUpload = async () => {
        try {
            const localPath: string = await SelectFile();
            if (!localPath) return;

            // Extract filename from path (naive)
            const filename = localPath.split(/[\\/]/).pop();
            const remotePath = currentPath === '.' ? filename : `${currentPath}/${filename}`;

            setLoading(true); // Show loading state
            await UploadFile(hostId, localPath, remotePath || "", false);
            alert("Upload successful!");
            loadFiles(currentPath); // Refresh
        } catch (err) {
            alert("Upload failed: " + err);
        } finally {
            setLoading(false);
        }
    };

    const handleDownload = async (file: FileInfo) => {
        try {
            const localPath = await SelectSaveFile();
            if (!localPath) return;

            const remotePath = currentPath === '.' ? file.name : `${currentPath}/${file.name}`;
            
            setLoading(true);
            await DownloadFile(hostId, remotePath, localPath, false);
            alert("Download successful!");
        } catch (err) {
            alert("Download failed: " + err);
        } finally {
            setLoading(false);
        }
    };

    const formatSize = (bytes: number) => {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    };

    return (
        <div className="flex flex-col h-full gap-4">
            <div className="flex gap-2 items-center">
                <Button variant="outline" size="icon" onClick={() => loadFiles('.')}>
                    <Home size={16} />
                </Button>
                <Button variant="outline" size="icon" onClick={() => loadFiles(currentPath + '/..')} disabled={currentPath === '.' || currentPath === '/'}>
                    <ArrowUp size={16} />
                </Button>
                <Input 
                    value={currentPath} 
                    onChange={(e) => setCurrentPath(e.target.value)} 
                    onKeyDown={(e) => e.key === 'Enter' && loadFiles(currentPath)}
                    className="flex-1 font-mono text-sm"
                />
                <Button variant="outline" size="icon" onClick={() => loadFiles(currentPath)}>
                    <RefreshCw size={16} className={loading ? "animate-spin" : ""} />
                </Button>
                <Button onClick={handleUpload} className="gap-2">
                    <Upload size={16} /> Upload
                </Button>
            </div>

            {error && (
                <div className="bg-destructive/10 text-destructive p-2 rounded text-sm">
                    {error}
                </div>
            )}

            <div className="flex-1 border rounded-md overflow-hidden bg-card">
                <div className="overflow-auto h-full">
                    <table className="w-full text-sm text-left">
                        <thead className="bg-muted text-muted-foreground sticky top-0">
                            <tr>
                                <th className="px-4 py-2 font-medium w-[40px]"></th>
                                <th className="px-4 py-2 font-medium">Name</th>
                                <th className="px-4 py-2 font-medium w-[100px]">Size</th>
                                <th className="px-4 py-2 font-medium w-[200px]">Modified</th>
                                <th className="px-4 py-2 font-medium w-[100px]">Actions</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-border">
                            {files.map((file, i) => (
                                <tr key={i} className="hover:bg-muted/50 group transition-colors">
                                    <td className="px-4 py-2 text-center">
                                        {file.is_dir ? <Folder size={16} className="text-blue-400" /> : <FileText size={16} className="text-gray-400" />}
                                    </td>
                                    <td className="px-4 py-2">
                                        {file.is_dir ? (
                                            <button 
                                                className="hover:underline font-medium text-blue-400 text-left"
                                                onClick={() => handleNavigate(file)}
                                            >
                                                {file.name}
                                            </button>
                                        ) : (
                                            <span className="text-foreground">{file.name}</span>
                                        )}
                                    </td>
                                    <td className="px-4 py-2 text-muted-foreground font-mono text-xs">
                                        {file.is_dir ? '-' : formatSize(file.size)}
                                    </td>
                                    <td className="px-4 py-2 text-muted-foreground text-xs whitespace-nowrap">
                                        {file.mod_time ? new Date(file.mod_time).toLocaleString() : '-'}
                                    </td>
                                    <td className="px-4 py-2">
                                        {!file.is_dir && file.name !== '..' && (
                                            <Button 
                                                variant="ghost" 
                                                size="icon" 
                                                className="h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity"
                                                onClick={() => handleDownload(file)}
                                                title="Download"
                                            >
                                                <Download size={14} />
                                            </Button>
                                        )}
                                    </td>
                                </tr>
                            ))}
                            {files.length === 0 && !loading && (
                                <tr>
                                    <td colSpan={5} className="px-4 py-8 text-center text-muted-foreground">
                                        Empty directory or failed to load.
                                    </td>
                                </tr>
                            )}
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    );
}
