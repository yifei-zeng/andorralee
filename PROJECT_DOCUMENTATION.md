# Andorralee 蜜罐系统项目文档

## 📋 项目概述

**Andorralee v2.0** 是一款基于Go语言开发的企业级蜜罐诱捕系统，专为网络安全威胁检测和攻击行为分析而设计。系统采用微服务架构，支持多协议蜜罐部署、实时攻击行为捕获、病毒检测、威胁情报管理和数据可视化分析。

### 核心特性
- 🛡️ 多协议蜜罐支持 (SSH、HTTP、FTP、Telnet等)
- 🔍 实时攻击行为捕获和会话追踪  
- 🦠 病毒检测和威胁评级系统
- 📊 完整的日志记录和分析
- 🐳 容器化部署和管理
- 💾 双数据库支持 (MySQL + DamengDB)

---

## 🏗️ 项目结构详解

### 根目录文件

| 文件名 | 用途 | 状态 |
|--------|------|------|
| `go.mod` / `go.sum` | Go模块依赖管理 | ✅ 完整 |
| `go.work` / `go.work.sum` | Go工作区配置 | ✅ 完整 |
| `docker-compose.yml` | 统一容器编排配置 | ✅ 完整 |
| `Dockerfile` | 容器镜像构建文件 | ✅ 完整 |
| `deployment/` | 数据库脚本与部署资源 | ✅ 完整 |
| `README.md` | 项目介绍和使用说明 | ✅ 完整 |
| `RELEASE_NOTES.md` | 版本发布说明 | ✅ 完整 |

> 注: v2.0 起已移除 Kylin/ARM 专用 compose、Dockerfile 与部署脚本，统一使用 `docker-compose.yml` 搭配自定义环境变量完成部署。

### 核心目录结构

```
andorralee/
├── cmd/                    # 应用程序入口
├── internal/               # 内部核心业务逻辑
│   ├── config/            # 配置管理
│   ├── handlers/          # HTTP处理器
│   ├── repositories/      # 数据访问层
│   └── services/          # 业务服务层
├── pkg/                   # 公共包和工具
├── routers/               # 路由配置
├── scripts/               # 数据库脚本和工具
├── static/                # 静态文件
└── docs/                  # 项目文档
```

---

## 🔧 核心模块详解

### 1. 配置管理 (`internal/config/`)

#### `config.go`
**功能**: 系统核心配置管理和数据库初始化
- 数据库连接配置 (MySQL + DamengDB)
- 配置文件加载和验证
- 数据库表自动迁移
- 环境变量处理

**关键函数**:
- `LoadConfig()`: 加载系统配置
- `InitMySQL()`: 初始化MySQL数据库
- `InitDameng()`: 初始化达梦数据库
- `InitMySQLTables()`: MySQL表结构迁移
- `InitDamengTables()`: 达梦数据库表结构迁移

#### `honeypot_config.go`
**功能**: 蜜罐相关配置管理
- 蜜罐类型定义
- 协议配置
- 端口分配策略

### 2. 数据访问层 (`internal/repositories/`)

#### `models.go`
**功能**: 数据模型定义，包含系统所有数据结构
- 基础蜜罐模型 (HoneypotInstance, HoneypotLog等)
- 安全检测模型 (MalwareSignature, ScanResult, DetectionResult)
- 攻击分析模型 (AttackSession, AttackEvent, ThreatIntelligence)
- 容器管理模型 (ContainerInstance等)

**重要模型**:
- `MalwareSignature`: 病毒特征库
- `AttackSession`: 攻击会话记录
- `ThreatIntelligence`: 威胁情报数据
- `HoneytokenEvent`: 蜜签触发事件

#### `repository_interface.go`
**功能**: 定义数据访问接口规范

#### `mysql_repo.go` / `dameng_repo.go`
**功能**: 具体数据库实现
- MySQL数据库操作实现
- 达梦数据库操作实现

### 3. 业务服务层 (`internal/services/`)

#### `database_services.go`
**功能**: 统一数据库服务接口
- 双数据库支持的服务层封装
- 数据库切换和负载均衡
- 事务管理和错误处理

**服务接口**:
- `DatabaseService`: 统一数据库操作接口
- `MySQLService`: MySQL专用服务实现
- `DamengService`: 达梦数据库专用服务实现

#### `docker_services.go`
**功能**: Docker容器管理服务
- 容器生命周期管理
- 镜像操作和版本管理
- 网络和存储管理

#### `port_manager_service.go`
**功能**: 端口分配和管理服务
- 动态端口分配
- 端口冲突检测
- 端口映射管理

### 4. HTTP处理器 (`internal/handlers/`)

#### 核心处理器说明

| 处理器文件 | 功能 | API路径前缀 | 实现状态 |
|------------|------|-------------|----------|
| `health_handler.go` | 系统健康检查 | `/health`, `/ready`, `/live` | ✅ 完整 |
| `docker.go` | Docker容器管理 | `/api/v1/docker` | ✅ 完整 |
| `honeypot_handler.go` | 蜜罐实例管理 | `/api/v1/honeypots` | ✅ 完整 |
| `malware_handler.go` | 病毒检测和扫描 | `/api/v1/malware` | ✅ 完整 |
| `threat_handler.go` | 威胁情报管理 | `/api/v1/threats` | ✅ 完整 |
| `attack_capture_handler.go` | 攻击事件捕获 | `/api/v1/attack-capture` | ✅ 完整 |
| `heralding_handler.go` | 认证日志管理 | `/api/v1/heralding` | ✅ 完整 |
| `cowrie_handler.go` | Cowrie蜜罐日志 | `/api/v1/cowrie` | ✅ 完整 |
| `mysql_honeypot_handler.go` | MySQL蜜罐日志 | `/api/v1/mysql-honeypot` | ✅ 完整 |
| `honeytokens_handler.go` | 蜜签管理 | `/api/v1/honeytokens` | ✅ 完整 |
| `port_manager_handler.go` | 端口管理 | `/api/v1/ports` | ✅ 完整 |
| `data.go` | 通用数据操作 | `/api/v1/data` | ⚠️ **占位符** |

#### 重要提醒: `data.go` 状态说明
**当前状态**: 仅包含占位符函数，所有接口返回 "功能暂未实现" 消息

**占位符函数列表**:
- `QueryData()`: 数据查询
- `CreateData()`: 数据创建  
- `UpdateData()`: 数据更新
- `DeleteData()`: 数据删除
- `GetDataByID()`: 根据ID获取数据
- `GetDataByName()`: 根据名称获取数据
- `SaveData()`: 数据保存

**建议**: 这些是通用数据操作接口，可以根据实际需求实现具体业务逻辑，或者如果已有专门的处理器，可以考虑移除这些占位符。

---

## 🌐 API接口概览

### 系统管理接口
- `GET /health` - 简单健康检查
- `GET /api/v1/health` - 详细健康状态
- `GET /api/v1/ready` - 就绪状态检查
- `GET /api/v1/live` - 存活状态检查

### Docker管理接口 (25+ 接口)
- `POST /api/v1/docker/pull` - 拉取镜像
- `GET /api/v1/docker/images` - 镜像列表
- `POST /api/v1/docker/start` - 启动容器
- `POST /api/v1/docker/stop` - 停止容器
- 更多容器生命周期管理接口...

### 蜜罐管理接口 (30+ 接口)
- `GET /api/v1/honeypots/instances` - 获取所有蜜罐实例
- `POST /api/v1/honeypots/instances/:id/deploy` - 部署蜜罐实例
- `GET /api/v1/honeypots/logs` - 获取蜜罐日志
- 更多蜜罐实例和日志管理接口...

### 安全检测接口 (15+ 接口)
- `POST /api/v1/malware/scan` - 病毒扫描
- `GET /api/v1/malware/results` - 获取扫描结果
- `POST /api/v1/threats/intelligence` - 威胁情报管理
- 更多安全分析接口...

### 攻击捕获接口 (10+ 接口)  
- `POST /api/v1/attack-capture/events` - 捕获攻击事件
- `GET /api/v1/attack-capture/sessions` - 获取攻击会话
- `GET /api/v1/attack-capture/statistics` - 攻击统计
- 更多攻击分析接口...

### 日志分析接口 (40+ 接口)
- **Heralding认证日志**: 15+ 接口 (`/api/v1/heralding`)
- **Cowrie蜜罐日志**: 20+ 接口 (`/api/v1/cowrie`)
- **MySQL蜜罐日志**: 10+ 接口 (`/api/v1/mysql-honeypot`)
- **容器运行日志**: 多个日志管理接口

### 蜜签管理接口 (8 接口)
- `POST /api/v1/honeytokens` - 创建蜜签
- `GET /api/v1/honeytokens` - 获取所有蜜签
- `POST /api/v1/honeytokens/:id/trigger` - 触发蜜签
- 更多蜜签管理接口...

### 端口管理接口 (10+ 接口)
- `POST /api/v1/ports/allocate` - 自动分配端口
- `POST /api/v1/ports/allocate-specific` - 分配指定端口
- `POST /api/v1/port-scan` - 端口扫描
- 更多端口管理接口...

---

## 📊 数据库架构

### 支持的数据库
1. **MySQL 8.0+** (主要数据库)
2. **达梦数据库** (国产化支持)

### 核心数据表

#### 安全检测相关
- `malware_signatures` - 病毒特征库
- `scan_results` - 扫描结果记录
- `detection_results` - 检测结果详情

#### 攻击分析相关  
- `attack_sessions` - 攻击会话
- `attack_events` - 攻击事件
- `threat_intelligence` - 威胁情报

#### 蜜罐管理相关
- `honeypot_instances` - 蜜罐实例
- `honeypot_logs` - 蜜罐日志
- `container_instances` - 容器实例

#### 蜜签管理相关
- `honeytoken_events` - 蜜签触发事件

---

## 🔧 开发和部署

### 本地开发环境
```bash
# 克隆项目
git clone <repository-url>
cd andorralee

# 安装依赖
go mod download

# 编译项目
go build -o build/andorralee ./cmd/main.go

# 运行服务
./build/andorralee
```

### Docker部署
```bash
# 开发环境
docker-compose up -d

# 生产环境示例（按需调整端口/变量）
docker-compose -f docker-compose.yml up -d
```

### 数据库初始化
```bash
# MySQL初始化
mysql -u root -p < scripts/init_db.sql

# 使用初始化脚本
./scripts/init_db.sh
```

---

## 🚀 系统特色功能

### 1. 双数据库支持
- 无缝切换MySQL和达梦数据库
- 支持读写分离和负载均衡
- 自动表结构迁移

### 2. 病毒检测系统  
- 文件哈希计算和病毒特征匹配
- 威胁等级评估 (Low/Medium/High/Critical)
- 检测结果数据库持久化

### 3. 攻击行为分析
- 会话级攻击追踪
- IP地理位置分析  
- 攻击模式识别

### 4. 容器化蜜罐
- Docker容器自动化部署
- 多协议蜜罐支持
- 动态端口分配

### 5. 完整的日志系统
- 认证日志 (Heralding)
- 交互日志 (Cowrie)  
- 数据库欺骗日志 (MySQL Honeypot)
- 容器运行日志
- 系统操作日志

---

## 📈 API统计总览

| 模块 | 接口数量 | 主要功能 |
|------|----------|----------|
| 健康检查 | 4 | 系统状态监控 |
| Docker管理 | 25+ | 容器生命周期管理 |
| 蜜罐管理 | 30+ | 蜜罐实例和日志管理 |
| 安全检测 | 15+ | 病毒检测和威胁分析 |
| 攻击捕获 | 10+ | 攻击事件和会话管理 |
| 日志分析 | 40+ | 多类型日志处理 |
| 蜜签管理 | 8 | 蜜签生命周期管理 |
| 端口管理 | 10+ | 端口分配和扫描 |
| **总计** | **140+** | **完整的蜜罐系统** |

---

## ⚠️ 注意事项

### 待实现功能
- `internal/handlers/data.go` 中的通用数据操作接口当前为占位符
- 部分高级分析功能仍在开发中

### 建议改进
1. 考虑是否需要实现 `data.go` 中的通用接口，或移除占位符
2. 完善API文档和接口测试用例
3. 增加更多的安全检测引擎集成

### 系统状态
- ✅ 核心功能完整实现
- ✅ 数据库集成成功
- ✅ API接口测试通过  
- ✅ Docker容器化部署就绪
- ⚠️ 部分高级功能开发中

---

**Andorralee v2.0** 是一个功能完整、架构清晰的企业级蜜罐系统，为网络安全防护提供了强大的威胁检测和分析能力。
