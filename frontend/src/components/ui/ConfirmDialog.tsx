import { createPortal } from 'react-dom';
import { Button } from './Button';
import { RefreshCw } from 'lucide-react';

interface ConfirmDialogProps {
    open: boolean;
    title: string;
    content: string;
    onConfirm: () => void;
    onCancel: () => void;
    loading?: boolean;
    confirmText?: string;
    loadingText?: string;
}

export function ConfirmDialog({ 
    open, 
    title, 
    content, 
    onConfirm, 
    onCancel, 
    loading,
    confirmText = "删除",
    loadingText = "删除中..."
}: ConfirmDialogProps) {
    if (!open) return null;
    return createPortal(
        <div className="fixed inset-0 z-50 flex items-center justify-center">
            <div className="absolute inset-0 bg-black/50 backdrop-blur-sm animate-in fade-in duration-200" onClick={loading ? undefined : onCancel} />
            <div className="relative bg-background p-6 rounded-lg shadow-xl w-[400px] border z-50 animate-in zoom-in-95 duration-200">
                <h3 className="text-lg font-bold mb-2">{title}</h3>
                <p className="text-muted-foreground mb-6">{content}</p>
                <div className="flex justify-end gap-2">
                    <Button variant="ghost" onClick={onCancel} disabled={loading}>取消</Button>
                    <Button variant="destructive" onClick={onConfirm} disabled={loading}>
                        {loading ? <RefreshCw size={16} className="animate-spin mr-2" /> : null}
                        {loading ? loadingText : confirmText}
                    </Button>
                </div>
            </div>
        </div>,
        document.body
    );
}
