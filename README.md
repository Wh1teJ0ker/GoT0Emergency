# GoT0Emergency

> 🚀 基于 Wails 的竞赛/应急响应桌面应用 - Go + React 打造

***

## 📖 项目描述

GoT0Emergency 是一款面向**竞赛和应急响应场景**的桌面应用，帮助你快速部署和管理远程主机监控系统。

### 技术选型

| 技术                     | 原因                                   |
| ---------------------- | ------------------------------------ |
| **Wails v2**           | 使用 Go 后端 + Web 前端，无需学习 Electron 的复杂性 |
| **Go 1.23**            | 高性能、跨平台、SSH 库成熟                      |
| **React + TypeScript** | 现代化的前端开发体验，类型安全                      |
| **SQLite3**            | 轻量级、零配置、WAL 模式支持并发读取                 |
| **Xterm.js**           | 标准的 Web 终端模拟                         |

### 已实现的功能

- ✅ SSH 连接管理（密码/密钥认证）
- ✅ 多标签终端（支持本地和远程）
- ✅ 实时系统监控（CPU/内存/磁盘/网络）
- ✅ 文件上传/下载（SFTP）
- ✅ Node 代理部署和回调
- ✅ 数据库路径自定义
- ✅ SSH Keepalive 长连接

***

## 📑 目录

- [安装和运行](#-安装和运行)
- [使用方法](#-使用方法)
- [配置说明](#-配置说明)
- [项目结构](#-项目结构)
- [开发指南](#-开发指南)
- [鸣谢](#-鸣谢)
- [许可](#-许可)

***

## 📦 安装和运行

### 方式一：下载预编译版本（推荐）

1. 访问 [Releases](https://github.com/Wh1teJ0ker/GoT0Emergency/releases)
2. 下载对应系统的版本：
   - Windows: `GoT0Emergency-windows-amd64.exe`
   - Linux: `GoT0Emergency-linux-amd64`
   - macOS: `GoT0Emergency-macos-universal.zip`
3. 解压到任意文件夹
4. 双击运行

### 方式二：源码编译

#### 环境要求

| 依赖        | 版本    |
| --------- | ----- |
| Go        | 1.23+ |
| Node.js   | 20+   |
| Wails CLI | 最新    |

#### 安装步骤

```bash
# 1. 克隆项目
git clone https://github.com/Wh1teJ0ker/GoT0Emergency.git
cd GoT0Emergency

# 2. 安装前端依赖
cd frontend
npm install
cd ..

# 3. 安装 Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 4. 编译
wails build
```

编译后的文件位于：`build/bin/GoT0Emergency.exe`

***

## 🚀 使用方法

### 1. 添加远程主机

1. 进入「远程管理」页面
2. 点击右上角「添加主机」
3. 填写主机信息：

| 字段    | 说明      | 示例              |
| ----- | ------- | --------------- |
| 名称    | 自定义主机名称 | `目标服务器 -01`     |
| IP 地址 | 远程主机 IP | `192.168.1.100` |
| 端口    | SSH 端口  | `22`            |
| 用户名   | SSH 用户名 | `root`          |
| 认证方式  | 密码或密钥   | 密码              |
| 密码    | SSH 密码  | `your_password` |

### 2. 连接主机

- 点击主机卡片或「查看详情」
- 首次进入会自动尝试连接
- 绿色「在线」= 已连接，黄色「连接中」= 正在连接

### 3. 使用终端

- 切换到「终端」标签
- 点击「+」新增终端标签页
- 支持同时打开多个终端

### 4. 文件传输

1. 切换到「文件管理」标签
2. 左侧 = 本地文件，右侧 = 远程文件
3. 支持操作：
   - 上传（本地 → 远程）
   - 下载（远程 → 本地）
   - 删除（远程文件）

### 5. 部署 Node 代理

1. 切换到「Node 部署」标签
2. 选择目标平台和架构
3. 点击「构建并部署」
4. Node 会自动运行并回传监控数据

***

## ⚙️ 配置说明

### 数据库设置

在「设置」页面可以配置：

| 配置项    | 说明               | 默认值                |
| ------ | ---------------- | ------------------ |
| 数据库路径  | SQLite 数据库保存位置   | `./data/db/app.db` |
| 数据保留时间 | 监控数据保留小时数        | `24` 小时            |
| 初始化数据库 | 重置数据库结构（⚠️ 清空数据） | -                  |

### SSH 认证方式

**密码认证**

- 适用于大多数 Linux 服务器
- 直接输入 SSH 密码

**密钥认证**

- 适用于使用 SSH 密钥的服务器
- 选择 PEM 格式私钥文件

***

## 📁 项目结构

```
GoT0Emergency/
├── cmd/                    # Wails 入口
├── frontend/               # React 前端
│   ├── src/
│   │   ├── components/     # UI 组件
│   │   ├── pages/          # 页面组件
│   │   ├── lib/            # 工具函数
│   │   └── wailsjs/        # Wails 生成的 bindings
│   └── package.json
├── internal/
│   ├── app/                # Wails app 主逻辑
│   ├── infra/              # 基础设施层
│   │   ├── db/             # 数据库
│   │   ├── executor/       # 命令执行器
│   │   └── session/        # SSH 会话管理
│   ├── modules/            # 功能模块
│   │   ├── host/           # 主机管理
│   │   ├── monitor/        # 监控服务
│   │   ├── node/           # Node 代理
│   │   ├── settings/       # 设置服务
│   │   └── terminal/       # 终端管理
│   └── pkg/                # 公共包
│       ├── log/            # 日志
│       └── path/           # 路径管理
├── build/                  # 编译输出
├── data/                   # 运行时数据
│   ├── db/                 # SQLite 数据库
│   ├── logs/               # 日志文件
│   ├── nodes/              # Node 代理构建产物
│   └── ssh/                # SSH 密钥文件
└── README.md
```

***

## 🔧 开发指南

### 实时开发模式

```bash
wails dev
```

- 前端热重载（Vite）
- 后端自动编译
- 浏览器调试访问：<http://localhost:34115>

### 生产构建

```bash
# 标准构建
wails build

# 带版本号
wails build -o GoT0Emergency-v1.0.0.exe

# 多平台（需要配置交叉编译）
wails build -platform windows/amd64,linux/amd64
```

### 运行测试

```bash
go test ./...
```

***

🙏 鸣谢

### 团队成员

| 角色  | GitHub                                       |
| --- | -------------------------------------------- |
| 开发者 | [@Wh1teJ0ker](https://github.com/Wh1teJ0ker) |

### 感谢以下开源项目

- [Wails](https://wails.io/) - Go 桌面应用框架
- [shadcn/ui](https://ui.shadcn.com/) - 美观的 UI 组件
- [Xterm.js](https://xtermjs.org/) - Web 终端模拟
- [Recharts](https://recharts.org/) - React 图表库
- [React Router](https://reactrouter.com/) - React 路由

### 参考资料

- [Wails 官方文档](https://wails.io/docs/gettingstarted/installation)
- [Go 语言官方教程](https://go.dev/tour/)

***

## 📄 许可

本项目采用 **MIT 许可** - 详见 [LICENSE](LICENSE) 文件

**你可以：**

- ✅ 商业使用
- ✅ 修改代码
- ✅ 分发
- ✅ 私有使用

**你需要：**

- 📝 保留许可声明和版权声明

***

## 📬 联系方式

- **GitHub**: [@Wh1teJ0ker](https://github.com/Wh1teJ0ker)
- **项目地址**: <https://github.com/Wh1teJ0ker/GoT0Emergency>
- **Issues**: <https://github.com/Wh1teJ0ker/GoT0Emergency/issues>

***

<p align="center">⭐ 如果这个项目对你有帮助，请给一个 Star！</p>
