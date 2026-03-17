# GoT0Emergency

🚀 基于 Wails 的竞赛/应急响应桌面应用

***

## 功能特性

- 🖥️ **远程主机管理** - SSH 连接、多标签终端、文件传输
- 📊 **实时监控** - CPU/内存/磁盘/网络指标监控
- 🔌 **Node 代理** - 跨平台代理部署，自动数据回传
- 🌐 **多主机并发** - 同时管理多台远程主机

***

## 快速开始

### 下载安装

1. 从 [Releases](https://github.com/Wh1teJ0ker/GoT0Emergency/releases) 下载
2. 解压后运行 `GoT0Emergency.exe`

### 源码编译

```bash
git clone https://github.com/Wh1teJ0ker/GoT0Emergency.git
cd GoT0Emergency
wails build
```

输出：`build/bin/GoT0Emergency.exe`

***

## 使用指南

### 添加主机

1. 进入「远程管理」→ 点击「添加主机」
2. 填写：名称、IP、端口、用户名、密码/密钥

### 连接主机

- 点击主机卡片进入详情
- 自动尝试连接，绿色「在线」= 已连接

### 终端多开

- 点击「终端」标签
- 点「+」新增终端标签页

### 文件传输

- 切换到「文件管理」
- 左侧本地，右侧远程，支持上传/下载/删除

### Node 部署

1. 点击「Node 部署」标签
2. 选择平台 → 「构建并部署」

***

## 技术栈

| <br /> | <br />                  |
| ------ | ----------------------- |
| 框架     | Wails v2                |
| 后端     | Go 1.23                 |
| 前端     | React + TypeScript      |
| 数据库    | SQLite3 (WAL)           |
| SSH    | golang.org/x/crypto/ssh |

***

## 配置说明

**设置页面支持：**

- 数据库路径自定义
- 数据保留时间（小时）
- 日志查看/清理

***

## 开发

```bash
# 实时开发模式
wails dev

# 生产构建
wails build
```

***

<p align="center">⭐ 如果觉得有用，请给一个 Star！</p>
