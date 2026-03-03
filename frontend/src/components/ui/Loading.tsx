import { Loader2 } from 'lucide-react';
import { cn } from '../../lib/utils';

interface LoadingProps {
    className?: string;
    text?: string;
    size?: number;
    fullScreen?: boolean;
}

export function Loading({
    className,
    text = "Loading...",
    size = 24,
    fullScreen = false
}: LoadingProps) {

    const iconSize = fullScreen ? size * 1.5 : size;

    const Content = (
        <div className={cn(
            "flex flex-col items-center justify-center",
            fullScreen ? "" : "p-4",
            className
        )}>
            <Loader2
                className="animate-spin text-primary"
                style={{ width: iconSize, height: iconSize }}
            />
            {text && (
                <p className="text-sm text-muted-foreground mt-2 animate-pulse">
                    {text}
                </p>
            )}
        </div>
    );

    if (!fullScreen) return Content;

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center
                        bg-background/80 backdrop-blur-sm">
            {Content}
        </div>
    );
}
