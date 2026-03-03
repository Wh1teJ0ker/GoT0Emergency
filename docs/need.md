技术栈：**Go + Wails + SQLite**

# 一、总体目标

构建一个**单机运行的本地运维管理工具**，具备：

- 本机管理
- 远程 SSH 管理
- 日志管理
- 基础设置

统一特性：

- 本机与远程复用同一套功能逻辑
- 所有运行期数据集中存储在 `data/`
- 启动自动初始化数据库与目录
- 不依赖外部服务

---

# 二、整体架构

分为四层：

```
UI（Wails 前端）

App 层（接口适配）

Service 层（业务逻辑）

Executor 抽象执行层
```

---

## 架构关系

```
UI
 ↓
App（Wails binding）
 ↓
Service（功能模块）
 ↓
Executor（Local / SSH）
 ↓
系统或远程主机
```

---

# 三、项目目录结构

```
project/
├─ cmd/
│  └─ app/
│     └─ main.go

├─ internal/

│  ├─ app/                 // Wails 绑定层
│  │  ├─ app.go
│  │  └─ ssh_api.go

│  ├─ common/
│  │  ├─ path/             // 路径统一管理
│  │  └─ log/

│  ├─ db/
│  │  ├─ init.go
│  │  └─ migrations/

│  ├─ executor/
│  │  ├─ executor.go
│  │  ├─ local.go
│  │  └─ ssh.go

│  ├─ service/
│  │  ├─ host/
│  │  ├─ process/
│  │  ├─ system/
│  │  └─ log/

│  ├─ session/
│  │  └─ ssh_session.go

├─ frontend/               // Wails 前端

├─ data/                   // 运行后自动创建
│  ├─ db/
│  ├─ logs/
│  ├─ cache/
│  ├─ ssh/
│  └─ tmp/

├─ go.mod
```

---

# 四、启动初始化流程

启动顺序必须固定：

1. 初始化路径管理
2. 创建 `data/` 目录
3. 创建子目录

```
data/db
data/logs
data/cache
data/ssh
data/tmp
```

4. 初始化 SQLite

```
data/db/app.db
```

5. 执行自动建表
6. 初始化日志系统
7. 启动 Wails

禁止跳过初始化阶段。

---

# 五、路径统一管理设计

模块：

```
internal/common/path
```

职责：

- 获取运行目录
- 确保 data 目录存在
- 返回各子目录绝对路径

核心接口：

```
GetDataDir()
GetDBPath()
GetLogDir()
GetSSHDir()
GetTmpDir()
```

禁止在任何模块硬编码路径。

---

# 六、数据库设计

数据库：

```
SQLite
```

位置：

```
data/db/app.db
```

---

## 核心表

### 1. 主机表

```
hosts
```

字段：

- id
- name
- ip
- port
- username
- auth\_type
- password / key\_path
- last\_connected\_at
- created\_at

---

### 2. 日志索引表（可选）

用于 UI 查询：

```
logs_index
```

---

# 七、执行层抽象（关键）

定义统一接口：

```
type Executor interface {
    Exec(cmd string) (string, error)
}
```

---

## 实现

### 本机执行

```
LocalExecutor
```

底层：

```
os/exec
```

---

### 远程执行

```
SSHExecutor
```

底层：

```
golang.org/x/crypto/ssh
```

---

# 八、功能模块复用机制

所有功能模块依赖：

```
Executor
```

示例：

```
func GetSystemInfo(exec Executor)
func GetProcessList(exec Executor)
```

结果：

- 本机管理 → LocalExecutor
- SSH 管理 → SSHExecutor

只实现一套业务逻辑。

---

# 九、SSH 会话管理

模块：

```
internal/session
```

职责：

- 保存已连接主机
- 复用连接
- 自动断线清理

结构：

```
SessionManager
```

---

# 十、UI 功能结构

左侧导航：

```
本机管理
远程管理
日志管理
基础设置
```

---

## 1. 本机管理

使用：

```
LocalExecutor
```

功能面板：

- 系统信息
- 进程
- 网络
- 端口

---

## 2. 远程管理（带下拉）

当前仅实现：

```
SSH 管理
```

功能：

- 主机列表
- 新建连接
- 保存连接
- 连接状态

连接后复用同一功能面板。

---

## 3. 日志管理

来源：

```
data/logs/
```

功能：

- 查看
- 过滤
- 按级别筛选

---

## 4. 基础设置

当前：

- 数据目录展示
- 日志级别
- 自动初始化开关

---

# 十一、日志系统

统一日志输出：

位置：

```
data/logs/app.log
```

要求：

- 结构化日志
- 包含模块字段
- 错误单独级别

---

# 十二、最小可运行里程碑

阶段一必须完成：

1. 路径管理模块
2. data 目录自动创建
3. SQLite 初始化
4. 主机表
5. Executor 抽象
6. LocalExecutor
7. SSHExecutor（基础连接）
8. 左侧 UI 框架
9. 主机增删改查