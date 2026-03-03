import { useEffect, useRef, useState, useCallback } from 'react';
import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import { WebLinksAddon } from 'xterm-addon-web-links';
import { TerminalOpen, TerminalWrite, TerminalResize, TerminalClose } from '../../../wailsjs/go/app/App';
import { EventsOn, EventsOff } from '../../../wailsjs/runtime/runtime';

interface XtermTerminalProps {
    hostId: number;
    onClose?: () => void;
}

export function XtermTerminal({ hostId, onClose }: XtermTerminalProps) {
    const terminalRef = useRef<HTMLDivElement>(null);
    const [termId, setTermId] = useState<string | null>(null);
    const xtermRef = useRef<Terminal | null>(null);
    const fitAddonRef = useRef<FitAddon | null>(null);

    // Debounce function
    const debounce = (func: Function, wait: number) => {
        let timeout: any;
        return (...args: any[]) => {
            clearTimeout(timeout);
            timeout = setTimeout(() => func(...args), wait);
        };
    };

    useEffect(() => {
        if (!terminalRef.current) return;

        // Initialize xterm with performance optimizations
        const term = new Terminal({
            cursorBlink: true,
            theme: {
                background: '#0f0f0f',
                foreground: '#cccccc',
                cursor: '#ffffff',
                selectionBackground: '#444444',
                black: '#0f0f0f',
                red: '#c0392b',
                green: '#27ae60',
                yellow: '#f39c12',
                blue: '#2980b9',
                magenta: '#8e44ad',
                cyan: '#16a085',
                white: '#ffffff',
                brightBlack: '#7f8c8d',
                brightRed: '#e74c3c',
                brightGreen: '#2ecc71',
                brightYellow: '#f1c40f',
                brightBlue: '#3498db',
                brightMagenta: '#9b59b6',
                brightCyan: '#1abc9c',
                brightWhite: '#ecf0f1',
            },
            fontFamily: 'Menlo, Monaco, "Courier New", monospace',
            fontSize: 14,
            lineHeight: 1.2,
            letterSpacing: 0,
            scrollback: 5000, // Limit scrollback for memory optimization
            allowProposedApi: true,
            fastScrollModifier: 'alt',
            scrollOnUserInput: true,
        });

        const fitAddon = new FitAddon();
        term.loadAddon(fitAddon);
        term.loadAddon(new WebLinksAddon());
        
        fitAddonRef.current = fitAddon;

        term.open(terminalRef.current);
        fitAddon.fit();
        xtermRef.current = term;

        // Start backend terminal
        const startTerminal = async () => {
            try {
                // Initial fit
                fitAddon.fit();
                const rows = term.rows;
                const cols = term.cols;
                
                const id = await TerminalOpen(hostId, rows, cols);
                setTermId(id);

                // Handle data from backend
                EventsOn(`terminal:data:${id}`, (data: string) => {
                    // Decode base64
                    const text = atob(data);
                    term.write(text);
                });

                // Handle close event
                EventsOn(`terminal:closed:${id}`, () => {
                    term.write('\r\n\x1b[31m连接已断开。\x1b[0m\r\n');
                    if (onClose) onClose();
                });

                // Handle input
                term.onData((data) => {
                    TerminalWrite(id, data);
                });

                // Handle resize from terminal (if user somehow resizes via escape codes? unlikely but good to have)
                term.onResize((size) => {
                    TerminalResize(id, size.rows, size.cols);
                });
                
                // Optimized resize handler
                const handleResize = debounce(() => {
                    if (fitAddonRef.current && xtermRef.current) {
                        fitAddonRef.current.fit();
                        const { rows, cols } = xtermRef.current;
                        if (rows > 0 && cols > 0) {
                             TerminalResize(id, rows, cols);
                        }
                    }
                }, 100);

                window.addEventListener('resize', handleResize);
                
                // Container resize observer
                const resizeObserver = new ResizeObserver(() => {
                    handleResize();
                });
                
                if (terminalRef.current) {
                    resizeObserver.observe(terminalRef.current);
                }
                
                // Focus terminal
                term.focus();
                
                // Initial fit after short delay to ensure layout is computed
                setTimeout(() => {
                    fitAddon.fit();
                    TerminalResize(id, term.rows, term.cols);
                }, 100);

                return () => {
                    window.removeEventListener('resize', handleResize);
                    resizeObserver.disconnect();
                    EventsOff(`terminal:data:${id}`);
                    EventsOff(`terminal:closed:${id}`);
                    TerminalClose(id);
                };

            } catch (err) {
                term.write(`\r\n\x1b[31m启动终端失败：${err}\x1b[0m\r\n`);
            }
        };

        const cleanupPromise = startTerminal();

        return () => {
            cleanupPromise.then(cleanup => cleanup && cleanup());
            term.dispose();
        };
    }, [hostId]);

    return <div className="w-full h-full" ref={terminalRef} />;
}
