import { createContext, useContext, useState, useCallback, ReactNode } from 'react';
import { Toast, ToastType } from './Toast';

interface ToastContextType {
    showToast: (type: ToastType, title: string, message?: string) => void;
    success: (title: string, message?: string) => void;
    error: (title: string, message?: string) => void;
    info: (title: string, message?: string) => void;
    warning: (title: string, message?: string) => void;
}

const ToastContext = createContext<ToastContextType | undefined>(undefined);

export function ToastProvider({ children }: { children: ReactNode }) {
    const [toast, setToast] = useState<{
        open: boolean;
        type: ToastType;
        title: string;
        message?: string;
    }>({ open: false, type: 'info', title: '' });

    const showToast = useCallback((type: ToastType, title: string, message?: string) => {
        setToast({ open: true, type, title, message });
    }, []);

    const handleClose = useCallback(() => {
        setToast(prev => ({ ...prev, open: false }));
    }, []);

    const contextValue = {
        showToast,
        success: (title: string, message?: string) => showToast('success', title, message),
        error: (title: string, message?: string) => showToast('error', title, message),
        info: (title: string, message?: string) => showToast('info', title, message),
        warning: (title: string, message?: string) => showToast('warning', title, message),
    };

    return (
        <ToastContext.Provider value={contextValue}>
            {children}
            <Toast
                open={toast.open}
                type={toast.type}
                title={toast.title}
                message={toast.message}
                onClose={handleClose}
            />
        </ToastContext.Provider>
    );
}

export function useToast() {
    const context = useContext(ToastContext);
    if (context === undefined) {
        throw new Error('useToast must be used within a ToastProvider');
    }
    return context;
}
