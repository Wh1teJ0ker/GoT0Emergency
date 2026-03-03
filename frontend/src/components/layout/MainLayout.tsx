import { Outlet } from 'react-router-dom';
import { Sidebar } from './Sidebar';

export function MainLayout() {
    return (
        <div className="h-screen bg-background text-foreground flex overflow-hidden">
            <Sidebar />
            <main className="ml-64 flex-1 h-full overflow-hidden flex flex-col relative">
                <Outlet />
            </main>
        </div>
    );
}
