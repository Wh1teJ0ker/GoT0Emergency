import { Monitor, Cloud, FileText, Settings, Github, Box, ChevronDown, Server } from 'lucide-react';
import { NavLink, useLocation, useNavigate } from 'react-router-dom';
import { cn } from '../../lib/utils';
import { Button } from '../ui/Button';
import { useState, useEffect } from 'react';
// @ts-ignore
import { GetHosts } from '../../../wailsjs/go/app/App';
// @ts-ignore
import { host } from '../../../wailsjs/go/models';

type NavItem = {
    icon: any;
    label: string;
    path: string;
    children?: {
        label: string;
        path: string;
        icon?: any;
    }[];
    dynamicChildren?: boolean;
};

const staticNavItems: NavItem[] = [
    { icon: Monitor, label: '本地管理', path: '/local' },
    { 
        icon: Cloud, 
        label: '远程管理', 
        path: '/remote',
        dynamicChildren: true,
        children: [
            { label: '主机概览', path: '/remote', icon: Cloud }
        ]
    },
    { 
        icon: Box, 
        label: 'Node管理', 
        path: '/node',
        children: [
            { label: '插件管理', path: '/node/plugins' },
            { label: 'Node 管理', path: '/node/manager' },
        ]
    },
    { icon: FileText, label: '日志管理', path: '/logs' },
    { icon: Settings, label: '基础设置', path: '/settings' },
];

export function Sidebar() {
    const location = useLocation();
    const navigate = useNavigate();
    const [expandedItems, setExpandedItems] = useState<string[]>(['/node', '/remote']);
    const [hosts, setHosts] = useState<host.Host[]>([]);
    const [navItems, setNavItems] = useState<NavItem[]>(staticNavItems);

    useEffect(() => {
        loadHosts();
        // Set up an interval to refresh host list periodically (e.g., every 10s) 
        // or just rely on manual refresh/event bus if available.
        // For now, simple poll or just mount.
        const interval = setInterval(loadHosts, 10000);
        return () => clearInterval(interval);
    }, []);

    // Update navItems when hosts change
    useEffect(() => {
        const newItems = staticNavItems.map(item => {
            if (item.dynamicChildren && item.path === '/remote') {
                return {
                    ...item,
                    children: [
                        { label: '主机概览', path: '/remote', icon: Cloud },
                        ...hosts.map(h => ({
                            label: h.name,
                            path: `/remote/${h.id}`,
                            icon: Server
                        }))
                    ]
                };
            }
            return item;
        });
        setNavItems(newItems);
    }, [hosts]);

    const loadHosts = async () => {
        try {
            const result = await GetHosts();
            if (result) {
                setHosts(result);
            }
        } catch (err) {
            console.error("Failed to load hosts for sidebar:", err);
        }
    };

    const toggleExpand = (path: string) => {
        setExpandedItems(prev => 
            prev.includes(path) 
                ? prev.filter(p => p !== path) 
                : [...prev, path]
        );
    };

    const isExpanded = (path: string) => expandedItems.includes(path);
    
    // Check if active. For /remote/123, parent /remote should be active
    const isActive = (path: string) => location.pathname === path;
    const isChildActive = (item: NavItem) => {
        if (!item.children) return false;
        return item.children.some(child => location.pathname === child.path);
    };

    return (
        <aside className="w-64 h-screen bg-card text-card-foreground fixed left-0 top-0 flex flex-col border-r border-border z-40">
            <div className="p-6 border-b border-border">
                <h1 className="text-xl font-bold flex items-center gap-2">
                    <span className="text-primary">GoT0</span>Emergency
                </h1>
            </div>
            
            <nav className="flex-1 p-4 space-y-2 overflow-y-auto scrollbar-hide">
                {navItems.map((item) => {
                    if (item.children) {
                        const expanded = isExpanded(item.path);
                        const active = isChildActive(item) || isActive(item.path);
                        
                        return (
                            <div key={item.path} className="space-y-1">
                                <button
                                    onClick={() => toggleExpand(item.path)}
                                    className={cn(
                                        "w-full flex items-center justify-between px-4 py-3 rounded-md transition-colors duration-200 text-sm font-medium",
                                        "hover:bg-accent hover:text-accent-foreground text-muted-foreground",
                                        active && "text-foreground"
                                    )}
                                >
                                    <div className="flex items-center gap-3">
                                        <item.icon size={18} />
                                        <span>{item.label}</span>
                                    </div>
                                    <ChevronDown 
                                        size={16} 
                                        className={cn("transition-transform duration-200", expanded ? "" : "-rotate-90")} 
                                    />
                                </button>
                                
                                {expanded && (
                                    <div className="pl-4 space-y-1">
                                        {item.children.map((child) => (
                                            <NavLink
                                                key={child.path}
                                                to={child.path}
                                                end={child.path === '/remote'} // Exact match for root
                                                className={({ isActive }) =>
                                                    cn(
                                                        "flex items-center gap-3 px-4 py-2 rounded-md transition-colors duration-200 text-sm font-medium",
                                                        "hover:bg-accent hover:text-accent-foreground text-muted-foreground",
                                                        isActive && "bg-secondary text-secondary-foreground"
                                                    )
                                                }
                                            >
                                                {child.icon ? <child.icon size={16} /> : <span className="w-4" />}
                                                <span className="truncate">{child.label}</span>
                                            </NavLink>
                                        ))}
                                    </div>
                                )}
                            </div>
                        );
                    }

                    return (
                        <NavLink
                            key={item.path}
                            to={item.path}
                            className={({ isActive }) =>
                                cn(
                                    "flex items-center gap-3 px-4 py-3 rounded-md transition-colors duration-200 text-sm font-medium",
                                    "hover:bg-accent hover:text-accent-foreground text-muted-foreground",
                                    isActive && "bg-secondary text-secondary-foreground"
                                )
                            }
                        >
                            <item.icon size={18} />
                            <span>{item.label}</span>
                        </NavLink>
                    );
                })}
            </nav>

            <div className="p-4 border-t border-border space-y-4">
                <Button variant="outline" className="w-full justify-start gap-2" size="sm">
                    <Github size={16} />
                    <span>GitHub</span>
                </Button>
                <div className="bg-muted/50 rounded-md p-3">
                    <p className="text-xs text-muted-foreground font-mono">v1.0.0</p>
                </div>
            </div>
        </aside>
    );
}
