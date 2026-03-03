export namespace host {
	
	export class Host {
	    id: number;
	    name: string;
	    ip: string;
	    port: number;
	    username: string;
	    auth_type: string;
	    password: string;
	    key_path: string;
	    last_connected_at: string;
	    created_at: string;
	
	    static createFrom(source: any = {}) {
	        return new Host(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.ip = source["ip"];
	        this.port = source["port"];
	        this.username = source["username"];
	        this.auth_type = source["auth_type"];
	        this.password = source["password"];
	        this.key_path = source["key_path"];
	        this.last_connected_at = source["last_connected_at"];
	        this.created_at = source["created_at"];
	    }
	}

}

export namespace monitor {
	
	export class CPUInfo {
	    model: string;
	    cores_logical: number;
	    cores_physical: number;
	    usage_total: number;
	    usage_per_core: number[];
	    load_avg: string;
	    frequency: number;
	
	    static createFrom(source: any = {}) {
	        return new CPUInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.model = source["model"];
	        this.cores_logical = source["cores_logical"];
	        this.cores_physical = source["cores_physical"];
	        this.usage_total = source["usage_total"];
	        this.usage_per_core = source["usage_per_core"];
	        this.load_avg = source["load_avg"];
	        this.frequency = source["frequency"];
	    }
	}
	export class PartitionInfo {
	    path: string;
	    fstype: string;
	    total: number;
	    used: number;
	    usage: number;
	
	    static createFrom(source: any = {}) {
	        return new PartitionInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.fstype = source["fstype"];
	        this.total = source["total"];
	        this.used = source["used"];
	        this.usage = source["usage"];
	    }
	}
	export class DiskInfo {
	    partitions: PartitionInfo[];
	    total: number;
	    used: number;
	    usage: number;
	    read_bytes: number;
	    write_bytes: number;
	    read_ops: number;
	    write_ops: number;
	
	    static createFrom(source: any = {}) {
	        return new DiskInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.partitions = this.convertValues(source["partitions"], PartitionInfo);
	        this.total = source["total"];
	        this.used = source["used"];
	        this.usage = source["usage"];
	        this.read_bytes = source["read_bytes"];
	        this.write_bytes = source["write_bytes"];
	        this.read_ops = source["read_ops"];
	        this.write_ops = source["write_ops"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class HardwareInfo {
	    motherboard: string;
	    bios: string;
	    baseboard: string;
	    chassis: string;
	    memory_model: string;
	    disk_model: string;
	
	    static createFrom(source: any = {}) {
	        return new HardwareInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.motherboard = source["motherboard"];
	        this.bios = source["bios"];
	        this.baseboard = source["baseboard"];
	        this.chassis = source["chassis"];
	        this.memory_model = source["memory_model"];
	        this.disk_model = source["disk_model"];
	    }
	}
	export class ProcessItem {
	    name: string;
	    pid: number;
	    ppid: number;
	    path: string;
	    cpu: number;
	    mem: number;
	
	    static createFrom(source: any = {}) {
	        return new ProcessItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.pid = source["pid"];
	        this.ppid = source["ppid"];
	        this.path = source["path"];
	        this.cpu = source["cpu"];
	        this.mem = source["mem"];
	    }
	}
	export class ProcessInfo {
	    total: number;
	    list: ProcessItem[];
	
	    static createFrom(source: any = {}) {
	        return new ProcessInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.list = this.convertValues(source["list"], ProcessItem);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class InterfaceInfo {
	    name: string;
	    ip: string;
	    rx: number;
	    tx: number;
	
	    static createFrom(source: any = {}) {
	        return new InterfaceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.ip = source["ip"];
	        this.rx = source["rx"];
	        this.tx = source["tx"];
	    }
	}
	export class NetworkInfo {
	    interfaces: InterfaceInfo[];
	    total_rx: number;
	    total_tx: number;
	    tcp_connections: number;
	    udp_connections: number;
	    listen_ports: number[];
	
	    static createFrom(source: any = {}) {
	        return new NetworkInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.interfaces = this.convertValues(source["interfaces"], InterfaceInfo);
	        this.total_rx = source["total_rx"];
	        this.total_tx = source["total_tx"];
	        this.tcp_connections = source["tcp_connections"];
	        this.udp_connections = source["udp_connections"];
	        this.listen_ports = source["listen_ports"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class MemoryInfo {
	    total: number;
	    used: number;
	    free: number;
	    usage: number;
	    swap_total: number;
	    swap_used: number;
	
	    static createFrom(source: any = {}) {
	        return new MemoryInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.used = source["used"];
	        this.free = source["free"];
	        this.usage = source["usage"];
	        this.swap_total = source["swap_total"];
	        this.swap_used = source["swap_used"];
	    }
	}
	export class SystemInfo {
	    hostname: string;
	    os: string;
	    platform: string;
	    kernel_arch: string;
	    boot_time: number;
	    uptime: number;
	    uptime_str: string;
	    current_user: string;
	
	    static createFrom(source: any = {}) {
	        return new SystemInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hostname = source["hostname"];
	        this.os = source["os"];
	        this.platform = source["platform"];
	        this.kernel_arch = source["kernel_arch"];
	        this.boot_time = source["boot_time"];
	        this.uptime = source["uptime"];
	        this.uptime_str = source["uptime_str"];
	        this.current_user = source["current_user"];
	    }
	}
	export class HostStatus {
	    system: SystemInfo;
	    cpu: CPUInfo;
	    memory: MemoryInfo;
	    disk: DiskInfo;
	    network: NetworkInfo;
	    process: ProcessInfo;
	    hardware: HardwareInfo;
	
	    static createFrom(source: any = {}) {
	        return new HostStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.system = this.convertValues(source["system"], SystemInfo);
	        this.cpu = this.convertValues(source["cpu"], CPUInfo);
	        this.memory = this.convertValues(source["memory"], MemoryInfo);
	        this.disk = this.convertValues(source["disk"], DiskInfo);
	        this.network = this.convertValues(source["network"], NetworkInfo);
	        this.process = this.convertValues(source["process"], ProcessInfo);
	        this.hardware = this.convertValues(source["hardware"], HardwareInfo);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	export class MetricPoint {
	    // Go type: time
	    timestamp: any;
	    cpu_usage: number;
	    memory_usage: number;
	    disk_usage: number;
	
	    static createFrom(source: any = {}) {
	        return new MetricPoint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.cpu_usage = source["cpu_usage"];
	        this.memory_usage = source["memory_usage"];
	        this.disk_usage = source["disk_usage"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	

}

export namespace node {
	
	export class Plugin {
	    name: string;
	    description: string;
	    tag: string;
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new Plugin(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.tag = source["tag"];
	        this.path = source["path"];
	    }
	}

}

export namespace session {
	
	export class FileInfo {
	    name: string;
	    size: number;
	    is_dir: boolean;
	    mod_time: string;
	
	    static createFrom(source: any = {}) {
	        return new FileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.size = source["size"];
	        this.is_dir = source["is_dir"];
	        this.mod_time = source["mod_time"];
	    }
	}

}

