import { createPortal } from 'react-dom';
import { X, CheckCircle, AlertCircle, Info, XOctagon } from 'lucide-react';
import { useEffect, useState } from 'react';

export type ToastType = 'success' | 'error' | 'info' | 'warning';

export interface ToastProps {
    open: boolean;
    type: ToastType;
    title: string;
    message?: string;
    duration?: number;
    onClose: () => void;
}

const toastConfig = {
    success: { icon: CheckCircle, bgColor: 'bg-green-500', borderColor: 'border-green-600' },
    error: { icon: XOctagon, bgColor: 'bg-red-500', borderColor: 'border-red-600' },
    info: { icon: Info, bgColor: 'bg-blue-500', borderColor: 'border-blue-600' },
    warning: { icon: AlertCircle, bgColor: 'bg-yellow-500', borderColor: 'border-yellow-600' },
};

export function Toast({ open, type, title, message, duration = 3000, onClose }: ToastProps) {
    const [isVisible, setIsVisible] = useState(open);
    const [isLeaving, setIsLeaving] = useState(false);

    useEffect(() => {
        if (open) {
            setIsVisible(true);
            setIsLeaving(false);
            const timer = setTimeout(() => {
                setIsLeaving(true);
                setTimeout(() => {
                    setIsVisible(false);
                    onClose();
                }, 300);
            }, duration);
            return () => clearTimeout(timer);
        }
    }, [open, duration, onClose]);

    if (!isVisible) return null;

    const config = toastConfig[type];
    const Icon = config.icon;

    return createPortal(
        <div
            className={`fixed top-4 right-4 z-50 min-w-[320px] max-w-md bg-background border ${config.borderColor} shadow-lg rounded-lg overflow-hidden transition-all duration-300 ${
                isLeaving ? 'opacity-0 translate-x-full' : 'opacity-100 translate-x-0'
            }`}
        >
            <div className="flex items-start gap-3 p-4">
                <div className={`flex-shrink-0 w-8 h-8 ${config.bgColor} rounded-full flex items-center justify-center`}>
                    <Icon size={16} className="text-white" />
                </div>
                <div className="flex-1 min-w-0">
                    <h4 className="font-semibold text-sm text-foreground mb-0.5">{title}</h4>
                    {message && <p className="text-sm text-muted-foreground break-words">{message}</p>}
                </div>
                <button
                    onClick={() => {
                        setIsLeaving(true);
                        setTimeout(onClose, 300);
                    }}
                    className="flex-shrink-0 text-muted-foreground hover:text-foreground transition-colors"
                >
                    <X size={16} />
                </button>
            </div>
            <div className={`h-1 ${config.bgColor} w-full animate-progress`} />
        </div>,
        document.body
    );
}

// Simple hook-like utility functions for imperative usage
let toastCallback: ((type: ToastType, title: string, message?: string) => void) | null = null;

export function setToastCallback(callback: (type: ToastType, title: string, message?: string) => void) {
    toastCallback = callback;
}

export function toastSuccess(title: string, message?: string) {
    toastCallback?.('success', title, message);
}

export function toastError(title: string, message?: string) {
    toastCallback?.('error', title, message);
}

export function toastInfo(title: string, message?: string) {
    toastCallback?.('info', title, message);
}

export function toastWarning(title: string, message?: string) {
    toastCallback?.('warning', title, message);
}
