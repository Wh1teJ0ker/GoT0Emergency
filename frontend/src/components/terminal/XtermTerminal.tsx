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
                background: '#1e1e1e', // Updated to VS Code style dark
                foreground: '#cccccc',
                cursor: '#ffffff',
                selectionBackground: '#444444',
                black: '#000000',
                red: '#cd3131',
                green: '#0dbc79',
                yellow: '#e5e510',
                blue: '#2472c8',
                magenta: '#bc3fbc',
                cyan: '#11a8cd',
                white: '#e5e5e5',
                brightBlack: '#666666',
                brightRed: '#f14c4c',
                brightGreen: '#23d18b',
                brightYellow: '#f5f543',
                brightBlue: '#3b8eea',
                brightMagenta: '#d670d6',
                brightCyan: '#29b8db',
                brightWhite: '#e5e5e5',
            },
            fontFamily: 'Menlo, Monaco, "Courier New", monospace',
            fontSize: 14,
            fontWeight: 'bold', // Forced bold
            fontWeightBold: '900', // Extra bold
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
                    // Decode base64 to binary string
                    const binaryString = atob(data);
                    // Convert binary string to Uint8Array
                    const bytes = new Uint8Array(binaryString.length);
                    for (let i = 0; i < binaryString.length; i++) {
                        bytes[i] = binaryString.charCodeAt(i);
                    }
                    // Decode UTF-8
                    const text = new TextDecoder('utf-8').decode(bytes);
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

    return (
        <div className="w-full h-full relative bg-[#1e1e1e]" ref={terminalRef}>
            <style>{`
                .xterm-viewport::-webkit-scrollbar {
                    width: 12px;
                }
                .xterm-viewport::-webkit-scrollbar-track {
                    background: rgba(255, 255, 255, 0.05);
                }
                .xterm-viewport::-webkit-scrollbar-thumb {
                    background-color: #888;
                    border-radius: 6px;
                    border: 2px solid #1e1e1e;
                }
                .xterm-viewport::-webkit-scrollbar-thumb:hover {
                    background-color: #bbb;
                }
            `}</style>
        </div>
    );
}
