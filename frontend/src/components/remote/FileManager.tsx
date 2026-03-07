import { useState, useEffect } from 'react';
import { createPortal } from 'react-dom';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';
import { Folder, File, ArrowUp, RefreshCw, Upload, Download, Home, FileText, CheckCircle, AlertCircle, Trash2 } from 'lucide-react';
// @ts-ignore
import { ListRemoteFiles, UploadFile, DownloadFile, SelectFile, SelectSaveFile, RemoveRemoteFile } from '../../../wailsjs/go/app/App';

interface FileInfo {
    name: string;
    size: number;
    is_dir: boolean;
    mod_time: string;
}

interface FileManagerProps {
    hostId: number;
}

// Simple Confirm Dialog Component
function ConfirmDialog({ open, title, content, onConfirm, onCancel, loading }: { 
    open: boolean, 
    title: string, 
    content: string, 
    onConfirm: () => void, 
    onCancel: () => void,
    loading?: boolean
}) {
    if (!open) return null;
    return createPortal(
        <div className="fixed inset-0 z-50 flex items-center justify-center">
            <div className="absolute inset-0 bg-black/50 backdrop-blur-sm animate-in fade-in duration-200" onClick={onCancel} />
            <div className="relative bg-background p-6 rounded-lg shadow-xl w-[400px] border z-50 animate-in zoom-in-95 duration-200">
                <h3 className="text-lg font-bold mb-2">{title}</h3>
                <p className="text-muted-foreground mb-6">{content}</p>
                <div className="flex justify-end gap-2">
                    <Button variant="ghost" onClick={onCancel} disabled={loading}>取消</Button>
                    <Button variant="destructive" onClick={onConfirm} disabled={loading}>
                        {loading ? <RefreshCw size={16} className="animate-spin mr-2" /> : null}
                        {loading ? '删除中...' : '删除'}
                    </Button>
                </div>
            </div>
        </div>,
        document.body
    );
}

export function FileManager({ hostId }: FileManagerProps) {
    const [currentPath, setCurrentPath] = useState('.');
    const [files, setFiles] = useState<FileInfo[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    
    // Action States
    const [uploadStatus, setUploadStatus] = useState<'idle' | 'uploading' | 'success' | 'error'>('idle');
    const [downloadStatus, setDownloadStatus] = useState<'idle' | 'downloading' | 'success' | 'error'>('idle');
    const [lastUploadedFile, setLastUploadedFile] = useState<string | null>(null);
    const [sortField, setSortField] = useState<'name' | 'size' | 'modified'>('name');
    const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('asc');
    
    // Delete Dialog State
    const [fileToDelete, setFileToDelete] = useState<FileInfo | null>(null);
    const [isDeleting, setIsDeleting] = useState(false);

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

            setUploadStatus('uploading');
            await UploadFile(hostId, localPath, remotePath || "", false);
            
            // Success state with artificial delay for better UX
            setUploadStatus('success');
            setTimeout(() => setUploadStatus('idle'), 2000);
            
            // Set highlighted file
            if (filename) {
                setLastUploadedFile(filename);
                // Clear highlight after 3s
                setTimeout(() => setLastUploadedFile(null), 3000);
            }

            // Refresh immediately
            loadFiles(currentPath); 
            // Refresh again after 1s to handle slow filesystem updates (race conditions)
            setTimeout(() => loadFiles(currentPath), 1000);
        } catch (err) {
            setUploadStatus('error');
            setError("Upload failed: " + err);
            setTimeout(() => setUploadStatus('idle'), 3000);
        }
    };

    const handleDownload = async (file: FileInfo) => {
        try {
            const localPath = await SelectSaveFile();
            if (!localPath) return;

            const remotePath = currentPath === '.' ? file.name : `${currentPath}/${file.name}`;
            
            setDownloadStatus('downloading');
            await DownloadFile(hostId, remotePath, localPath, false);
            
            setDownloadStatus('success');
            setTimeout(() => setDownloadStatus('idle'), 2000);
        } catch (err) {
            setDownloadStatus('error');
            setError("Download failed: " + err);
            setTimeout(() => setDownloadStatus('idle'), 3000);
        }
    };

    const handleRemoveClick = (file: FileInfo) => {
        setFileToDelete(file);
    };

    const confirmDelete = async () => {
        if (!fileToDelete) return;
        
        setIsDeleting(true);
        const remotePath = currentPath === '.' ? fileToDelete.name : `${currentPath}/${fileToDelete.name}`;
        
        try {
            await RemoveRemoteFile(hostId, remotePath);
            // Close dialog
            setFileToDelete(null);
            // Refresh list
            loadFiles(currentPath);
        } catch (err) {
            setError("删除失败: " + err);
        } finally {
            setIsDeleting(false);
        }
    };

    const handleSort = (field: 'name' | 'size' | 'modified') => {
        if (sortField === field) {
            setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
        } else {
            setSortField(field);
            setSortOrder('asc'); // Default to asc when switching fields
        }
    };

    const getSortedFiles = () => {
        // Clone array to avoid mutating state
        const sorted = [...files];
        
        // Separate ".." from rest
        const parentDir = sorted.find(f => f.name === '..');
        const others = sorted.filter(f => f.name !== '..');

        others.sort((a, b) => {
            // Always put directories first
            if (a.is_dir && !b.is_dir) return -1;
            if (!a.is_dir && b.is_dir) return 1;

            let comparison = 0;
            switch (sortField) {
                case 'name':
                    comparison = a.name.localeCompare(b.name);
                    break;
                case 'size':
                    comparison = a.size - b.size;
                    break;
                case 'modified':
                    // Parse dates, treat invalid dates as 0
                    const dateA = a.mod_time ? new Date(a.mod_time).getTime() : 0;
                    const dateB = b.mod_time ? new Date(b.mod_time).getTime() : 0;
                    comparison = dateA - dateB;
                    break;
            }
            return sortOrder === 'asc' ? comparison : -comparison;
        });

        if (parentDir) {
            return [parentDir, ...others];
        }
        return others;
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
                <Button 
                    onClick={handleUpload} 
                    disabled={uploadStatus !== 'idle'}
                    className={`gap-2 transition-all duration-300 min-w-[100px] 
                        ${uploadStatus === 'success' ? 'bg-green-600 hover:bg-green-700 text-white' : ''}
                        ${uploadStatus === 'error' ? 'bg-destructive text-destructive-foreground' : ''}
                    `}
                >
                    {uploadStatus === 'uploading' && <RefreshCw size={16} className="animate-spin" />}
                    {uploadStatus === 'success' && <CheckCircle size={16} />}
                    {uploadStatus === 'error' && <AlertCircle size={16} />}
                    {uploadStatus === 'idle' && <Upload size={16} />}
                    
                    {uploadStatus === 'uploading' && '上传中...'}
                    {uploadStatus === 'success' && '已上传'}
                    {uploadStatus === 'error' && '失败'}
                    {uploadStatus === 'idle' && '上传文件'}
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
                        <thead className="bg-muted text-muted-foreground sticky top-0 z-10">
                            <tr>
                                <th className="px-4 py-2 font-medium w-[40px]"></th>
                                <th 
                                    className="px-4 py-2 font-medium cursor-pointer hover:text-foreground hover:bg-muted/80 transition-colors"
                                    onClick={() => handleSort('name')}
                                >
                                    <div className="flex items-center gap-1">
                                        文件名
                                        {sortField === 'name' && (
                                            <ArrowUp size={12} className={`transition-transform ${sortOrder === 'desc' ? 'rotate-180' : ''}`} />
                                        )}
                                    </div>
                                </th>
                                <th 
                                    className="px-4 py-2 font-medium w-[100px] cursor-pointer hover:text-foreground hover:bg-muted/80 transition-colors"
                                    onClick={() => handleSort('size')}
                                >
                                    <div className="flex items-center gap-1">
                                        大小
                                        {sortField === 'size' && (
                                            <ArrowUp size={12} className={`transition-transform ${sortOrder === 'desc' ? 'rotate-180' : ''}`} />
                                        )}
                                    </div>
                                </th>
                                <th 
                                    className="px-4 py-2 font-medium w-[200px] cursor-pointer hover:text-foreground hover:bg-muted/80 transition-colors"
                                    onClick={() => handleSort('modified')}
                                >
                                    <div className="flex items-center gap-1">
                                        修改时间
                                        {sortField === 'modified' && (
                                            <ArrowUp size={12} className={`transition-transform ${sortOrder === 'desc' ? 'rotate-180' : ''}`} />
                                        )}
                                    </div>
                                </th>
                                <th className="px-4 py-2 font-medium w-[100px]">操作</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-border">
                            {getSortedFiles().map((file, i) => (
                                <tr 
                                    key={i} 
                                    className={`
                                        group transition-all duration-500
                                        ${file.name === lastUploadedFile 
                                            ? 'bg-green-500/20 hover:bg-green-500/30' 
                                            : 'hover:bg-muted/50'
                                        }
                                    `}
                                >
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
                                        {file.name !== '..' && (
                                            <div className="flex gap-2">
                                                {!file.is_dir && (
                                                    <Button 
                                                        variant="ghost" 
                                                        size="icon" 
                                                        className={`h-6 w-6 transition-all 
                                                            ${downloadStatus !== 'idle' ? 'opacity-100' : 'opacity-0 group-hover:opacity-100'}
                                                            ${downloadStatus === 'success' ? 'text-green-600' : ''}
                                                            ${downloadStatus === 'error' ? 'text-destructive' : ''}
                                                        `}
                                                        onClick={() => handleDownload(file)}
                                                        disabled={downloadStatus !== 'idle'}
                                                        title="下载"
                                                    >
                                                        {downloadStatus === 'downloading' ? (
                                                            <RefreshCw size={14} className="animate-spin" />
                                                        ) : downloadStatus === 'success' ? (
                                                            <CheckCircle size={14} />
                                                        ) : downloadStatus === 'error' ? (
                                                            <AlertCircle size={14} />
                                                        ) : (
                                                            <Download size={14} />
                                                        )}
                                                    </Button>
                                                )}
                                                <Button 
                                                    variant="ghost" 
                                                    size="icon" 
                                                    className="h-6 w-6 opacity-0 group-hover:opacity-100 transition-all text-muted-foreground hover:text-destructive"
                                                    onClick={() => handleRemoveClick(file)}
                                                    title="删除"
                                                >
                                                    <Trash2 size={14} />
                                                </Button>
                                            </div>
                                        )}
                                    </td>
                                </tr>
                            ))}
                            {files.length === 0 && !loading && (
                                <tr>
                                    <td colSpan={5} className="px-4 py-8 text-center text-muted-foreground">
                                        暂无文件或加载失败。
                                    </td>
                                </tr>
                            )}
                        </tbody>
                    </table>
                </div>
            </div>

            {/* Confirm Delete Dialog */}
            <ConfirmDialog 
                open={!!fileToDelete} 
                title="确认删除" 
                content={`确定要删除 ${fileToDelete?.name} 吗？此操作不可恢复。`}
                onConfirm={confirmDelete}
                onCancel={() => setFileToDelete(null)}
                loading={isDeleting}
            />
        </div>
    );
}
