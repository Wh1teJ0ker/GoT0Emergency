import { cn } from "../../lib/utils"

interface PageContainerProps extends React.HTMLAttributes<HTMLDivElement> {}

export function PageContainer({ className, children, ...props }: PageContainerProps) {
    return (
        <div className={cn("container mx-auto p-8 max-w-7xl space-y-6", className)} {...props}>
            {children}
        </div>
    )
}
