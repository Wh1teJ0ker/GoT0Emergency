# GoT0Emergency

🚀 基于 Wails 的竞赛/应急响应桌面应用 - Go + React 打造

![GitHub](https://img.shields.io/github/last-commit/Wh1teJ0ker/GoT0Emergency)
![License](https://img.shields.io/github/license/Wh1teJ0ker/GoT0Emergency)

---

## 📖 项目简介

GoT0Emergency 是一款面向竞赛和应急响应场景的桌面应用，支持：

- 🖥️ **远程主机管理** - SSH 连接、终端访问、文件传输
- 📊 **实时监控** - CPU、内存、磁盘、网络等系统指标
- 🔌 **插件系统** - 可扩展的 Node 代理，支持自定义功能
- 🌐 **多主机并发** - 同时管理多台远程主机

---

## ✨ 功能特性

### 远程管理
- SSH 连接管理（支持密码/密钥认证）
- 多标签终端（支持本地和远程）
- SFTP 文件上传/下载
- RDP 远程桌面（Windows 主机）
- 主机信息编辑

### 系统监控
- 实时 CPU/内存使用率
- 磁盘空间监控
- 网络流量统计
- 进程列表查看
- 历史数据图表（支持 1h/6h/24h/7d）

### Node 代理
- 跨平台 Node 编译（Windows/Linux/macOS）
- 一键部署到远程主机
- 自动端口转发
- 实时数据回传

### 其他特性
- 数据自定义保留时间
- 数据库路径自定义
- 日志查看和清理
- GitHub 链接快速跳转

---

## 🛠️ 技术栈

| 层级 | 技术 |
|------|------|
| **框架** | Wails v2 |
| **后端** | Go 1.23 |
| **前端** | React + TypeScript |
| **UI** | shadcn/ui + Tailwind CSS |
| **终端** | Xterm.js |
| **图表** | Recharts |
| **数据库** | SQLite3 (WAL 模式) |
| **SSH** | golang.org/x/crypto/ssh |
| **路由** | React Router v6 |

---

## 📦 安装

### 方式一：下载预编译版本（推荐）

1. 从 [Releases](https://github.com/Wh1teJ0ker/GoT0Emergency/releases) 下载最新版
2. 解压到任意文件夹
3. 双击运行 `GoT0Emergency.exe`

### 方式二：源码编译

#### 环境要求
- Go 1.23+
- Node.js 20+
- Wails CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

#### 编译步骤

```bash
# 克隆项目
git clone https://github.com/Wh1teJ0ker/GoT0Emergency.git
cd GoT0Emergency

# 安装前端依赖
cd frontend
npm install
cd ..

# 编译
wails build
```

编译后的可执行文件位于：`build/bin/GoT0Emergency.exe`

---

## 🚀 使用方法

### 1. 添加远程主机

1. 进入「远程管理」页面
2. 点击右上角「添加主机」
3. 填写主机信息：
   - 名称（自定义）
   - IP 地址
   - SSH 端口（默认 22）
   - 用户名
   - 认证方式（密码/密钥）
   - 密码或密钥路径

### 2. 连接主机

- 点击主机卡片或「查看详情」按钮
- 首次进入会自动尝试连接
- 连接成功后显示绿色「在线」状态

### 3. 使用终端

- 在主机详情页面点击「终端」标签
- 支持多开终端（点击「+」按钮）
- 每个终端独立标签页
- 支持本地终端（localhost）

### 4. 文件传输

1. 切换到「文件管理」标签
2. 左侧为本地文件，右侧为远程文件
3. 支持上传、下载、删除操作

### 5. 部署 Node 代理

1. 在主机详情页面点击「Node 部署」
2. 选择目标平台和架构
3. 点击「构建并部署」
4. 部署成功后 Node 会自动运行并回传数据

---

## 📁 项目结构

```
GoT0Emergency/
├── cmd/                   # Wails 入口
├── frontend/              # React 前端
│   ├── src/
│   │   ├── components/    # UI 组件
│   │   ├── pages/         # 页面组件
│   │   ├── lib/           # 工具函数
│   │   └── wailsjs/       # Wails 生成的 bindings
│   └── package.json
├── internal/
│   ├── app/               # Wails app 主逻辑
│   ├── infra/             # 基础设施层
│   │   ├── db/            # 数据库
│   │   ├── executor/      # 命令执行器
│   │   └── session/       # SSH 会话管理
│   ├── modules/           # 功能模块
│   │   ├── host/          # 主机管理
│   │   ├── monitor/       # 监控服务
│   │   ├── node/          # Node 代理
│   │   ├── settings/      # 设置服务
│   │   └── terminal/      # 终端管理
│   └── pkg/               # 公共包
│       ├── log/           # 日志
│       └── path/          # 路径管理
├── build/                 # 编译输出
└── data/                  # 运行时数据
    ├── db/                # SQLite 数据库
    ├── logs/              # 日志文件
    ├── nodes/             # Node 代理构建产物
    └── ssh/               # SSH 密钥文件
```

---

## ⚙️ 配置说明

### 数据库设置

在「设置」页面可以配置：

- **数据库路径** - 自定义 SQLite 数据库保存位置
- **数据保留时间** - 监控数据保留小时数（默认 24 小时）
- **初始化数据库** - 重置数据库结构（⚠️ 会清空所有数据）

### SSH 连接配置

支持两种认证方式：

| 认证方式 | 说明 |
|---------|------|
| 密码认证 | 输入 SSH 密码 |
| 密钥认证 | 选择 PEM 格式私钥文件 |

---

## 🔧 开发指南

### 实时开发模式

```bash
wails dev
```

- 前端热重载（Vite）
- 后端自动编译
- 浏览器开发模式访问：http://localhost:34115

### 生产构建

```bash
# 标准构建
wails build

# 带版本号的构建
wails build -o GoT0Emergency-v1.0.0.exe

# 交叉编译（需要配置）
wails build -platform windows/amd64,linux/amd64
```

---

## 📝 更新日志

### v1.0.0
- ✨ 初始版本发布
- 🖥️ 远程主机管理
- 📊 实时监控
- 🔌 Node 代理系统
- 🌐 多主机并发支持
- 📁 数据库路径自定义
- 🔄 SSH Keepalive 机制

---

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

---

## 📄 开源协议

MIT License - 详见 [LICENSE](LICENSE)

---

## 🙏 致谢

- [Wails](https://wails.io/) - Go 桌面应用框架
- [shadcn/ui](https://ui.shadcn.com/) - 美观的 UI 组件
- [Xterm.js](https://xtermjs.org/) - Web 终端模拟

---

## 📬 联系方式

- GitHub: [@Wh1teJ0ker](https://github.com/Wh1teJ0ker)
- 项目地址：[https://github.com/Wh1teJ0ker/GoT0Emergency](https://github.com/Wh1teJ0ker/GoT0Emergency)

---

<p align="center">⭐ 如果这个项目对你有帮助，请给一个 Star！</p>
