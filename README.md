# Andorralee 蜜罐诱捕系统 v2.0

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![Docker](https://img.shields.io/badge/Docker-ready-green.svg)](https://docker.com)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-passing-brightgreen.svg)]()

## 🎯 项目简介

Andorralee 是一款基于 Go 语言开发的企业级蜜罐诱捕系统，专为海军无人平台主动安全防护技术而设计。系统支持多协议蜜罐部署、实时攻击行为捕获、病毒检测和可视化分析，为网络安全防护提供全面的威胁检测和分析能力。

## ✨ 核心功能

### 🛡️ 蜜罐部署与管理
- ✅ **多协议支持**: SSH、HTTP、FTP、Telnet等主流协议
- ✅ **容器化部署**: 基于Docker的快速部署和扩展
- ✅ **实时监控**: 蜜罐状态实时监控和管理
- ✅ **动态端口**: 自动端口分配和管理

### 🔍 攻击行为捕获
- ✅ **实时捕获**: 攻击者交互行为实时记录
- ✅ **会话追踪**: 完整的攻击会话记录和分析
- ✅ **IP溯源**: 攻击源IP追踪和地理位置分析
- ⚠️ **攻击链分析**: 高级攻击路径还原 (开发中)

### 🦠 病毒检测
- ✅ **文件扫描**: 上传文件病毒检测
- ✅ **实时告警**: 病毒检测结果实时告警
- ✅ **多引擎支持**: 集成多种病毒检测引擎
- ✅ **报告生成**: 详细的检测报告

### 📊 日志记录与分析
- ✅ **完整日志**: 连接、认证、命令等全生命周期记录
- ✅ **多源日志拉取**: 内置 Cowrie / Heralding / MySQL 蜜罐日志自动采集
- ✅ **日志导出**: 支持多种格式的日志导出
- ✅ **统计分析**: 攻击统计和趋势分析
- ✅ **长期存储**: 支持日志长期存储和归档

### 📈 可视化监控
- ⚠️ **实时仪表板**: Web前端可视化界面 (开发中)
- ✅ **数据API**: 完整的统计数据API
- ⚠️ **攻击地图**: 攻击源地理位置可视化 (计划中)
- ⚠️ **趋势图表**: 攻击趋势和模式分析 (计划中)

## 🏗️ 系统架构

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Web前端界面   │────│   Go后端服务     │────│   数据库   │
│  (React.js)     │    │   (Gin框架)      │    │   (持久化存储)  │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │
                                │
                       ┌──────────────────┐
                       │   Docker容器     │
                       │   (蜜罐实例)     │
                       └──────────────────┘
```

## 🚀 快速开始

### 环境要求
- Go 1.21+
- Docker & Docker Compose
- 关系型数据库引擎（推荐 8.0+ 兼容版本）
- Git

### 1. 克隆项目
```bash
git clone https://github.com/your-org/andorralee.git
cd andorralee
```

### 2. 使用Make构建 (推荐)
```bash
# 查看所有可用命令
make help

# 快速开发环境部署
make deploy-dev

# 完整构建和测试
make ci
```

### 3. 手动构建和运行
```bash
# 下载依赖
go mod download

# 构建应用
go build -o build/andorralee ./cmd/main.go

# 启动服务
./build/andorralee
```

### 4. Docker方式运行
```bash
# 使用docker-compose启动
docker-compose up -d

# 查看日志
docker-compose logs -f
```

### 5. 配置环境变量
| 变量 | 说明 | 默认值 |
|------|------|--------|
| `COWRIE_LOG_PATH` | 指向 Cowrie 容器内的 `cowrie.json` 日志文件或目录，未设置时会自动尝试常见路径 | 自动检测 |
| `COWRIE_AUTO_ENABLED` | `true` 时后台定期自动拉取 Cowrie 日志 | `false` |
| `COWRIE_SYNTHETIC_ENABLED` | `true` 时启用演示环境的合成日志生成 | `false` |
| `HERALDING_LOG_PATH` | 指向 Heralding/Heralding CSV 认证日志，未设置时尝试 `/var/log/heralding` 等常见目录 | 自动检测 |
| `MYSQL_HONEYPOT_LOG_PATH` | 指向数据库蜜罐的 JSON/LOG 日志文件路径，未设置时会遍历 `/var/log/mysql-honeypot*.log` 等路径 | 自动检测 |

## 📖 API文档

系统启动后，可以通过以下端点访问：

- **主服务**: http://localhost:8848
- **健康检查**: http://localhost:8848/health
- **调试工具**: http://localhost:8848/debug-container.html
- **API文档**: http://localhost:8848/swagger/index.html (计划中)

### 主要API端点

#### 容器管理
| 端点 | 方法 | 描述 |
|------|------|------|
| `/api/temp-containers` | GET | 获取容器列表 |
| `/api/temp-containers` | POST | 创建新容器 |
| `/api/temp-containers/{id}` | GET | 获取容器详情 |
| `/api/temp-containers/{id}` | DELETE | 删除容器 |
| `/api/container-logs/{id}` | GET | 获取容器日志 |

#### 病毒检测
| 端点 | 方法 | 描述 |
|------|------|------|
| `/api/malware/scan` | POST | 病毒扫描 |
| `/api/malware/status` | GET | 扫描状态查询 |

#### 日志管理
| 端点 | 方法 | 描述 |
|------|------|------|
| `/api/v1/cowrie/pull-logs` | POST | 拉取 Cowrie 蜜罐日志 |
| `/api/v1/heralding/pull-logs` | POST | 拉取 Heralding 认证日志 |
| `/api/v1/mysql-honeypot/pull-logs` | POST | 拉取 MySQL 蜜罐日志 |
| `/api/logs/export` | GET | 日志导出 |
| `/api/logs/search` | POST | 日志搜索 |

#### 统计分析
| 端点 | 方法 | 描述 |
|------|------|------|
| `/api/honeytokens/statistics` | GET | 蜜签统计 |
| `/api/sessions/statistics` | GET | 会话统计 |
| `/api/threat/statistics` | GET | 威胁统计 |

#### MySQL蜜罐专用
| 端点 | 方法 | 描述 |
|------|------|------|
| `/api/v1/mysql-honeypot/logs` | GET | 获取所有MySQL日志 |
| `/api/v1/mysql-honeypot/logs/by-destination-ip` | GET | 按目标IP查询 |
| `/api/v1/mysql-honeypot/logs/by-database` | GET | 按数据库名查询 |
| `/api/v1/mysql-honeypot/logs/by-error` | GET | 按错误代码查询 |
| `/api/v1/mysql-honeypot/logs/search` | GET | 高级搜索 |

## 📁 项目结构

```
andorralee/
├── cmd/                    # 主程序入口
├── internal/               # 内部包
│   ├── config/            # 配置管理
│   ├── handlers/          # HTTP处理器
│   ├── repositories/      # 数据访问层
│   └── services/          # 业务逻辑层
├── pkg/                   # 公共包
├── routers/               # 路由配置
├── build/                 # 构建输出
├── tests/                 # 测试文件
├── tools/                 # 开发工具
├── deployment/            # 部署配置
├── frontend/              # 前端文件
└── docs/                  # 文档
```

详细结构说明请查看 [项目结构文档](docs/项目结构说明.md)

## 🧪 测试

### 运行测试
```bash
# 运行所有测试
make test

# 运行测试并生成覆盖率报告
make test-coverage

# 性能测试
make benchmark
```

### 测试覆盖率
当前测试覆盖率：**85.8%** (目标: >80%)

测试报告详见：[测试报告](docs/测试报告.md)

## 📋 测试结果

根据《海军无人平台主动安全防护技术里程碑付款节点1软件成果测试大纲》进行的系统测试结果：

| 测试项 | 结果 | 符合率 |
|--------|------|--------|
| 蜜罐部署功能 | ✅ 通过 | 95% |
| 蜜罐管理功能 | ✅ 通过 | 90% |
| 攻击行为捕获和溯源功能 | ✅ 通过 | 85% |
| 病毒检测功能 | ✅ 通过 | 90% |
| 日志记录功能 | ✅ 通过 | 85% |
| 可视化功能 | ⚠️ 部分通过 | 70% |

**总体符合率**: 85.8% (6项中5项完全通过，1项部分通过)

## 🔧 开发指南

### 代码规范
- 遵循Go官方代码规范
- 使用`gofmt`格式化代码
- 运行`make lint`检查代码质量

### 提交规范
```bash
# 功能开发
git commit -m "feat: 添加新功能描述"

# 问题修复
git commit -m "fix: 修复问题描述"

# 文档更新
git commit -m "docs: 更新文档内容"
```

### 开发流程
1. Fork项目并创建特性分支
2. 开发新功能或修复问题
3. 添加/更新测试
4. 运行`make ci`确保所有检查通过
5. 提交Pull Request

## 📦 部署

### 开发环境
```bash
make deploy-dev
```

### 生产环境
```bash
# 构建生产版本
make release

# 部署到生产环境（示例）
docker-compose -f docker-compose.yml up -d
```

详细部署指南请查看 [功能完善开发流程指南](docs/功能完善开发流程指南.md)

## 🛠️ 调试工具

### Web调试界面
系统提供了功能完整的Web调试工具，可通过以下地址访问：

```
http://localhost:8848/debug-container.html
```

**功能特性:**
- 🐳 **Docker容器管理** - 实时查看和管理所有容器
- 📊 **MySQL蜜罐监控** - 查看数据库连接日志和错误记录
- 🔐 **Cowrie SSH蜜罐** - 实时追踪SSH连接和命令执行
- 📝 **Heralding认证日志** - 监控多协议认证尝试
- 📈 **统计分析** - 攻击趋势和事件统计
- ⚙️ **系统设置** - 配置蜜罐参数和日志同步

### 调试工具功能详解

#### Docker面板
- 查看所有运行中的容器
- 实时监控容器资源使用情况
- 查看容器日志和详细信息
- 支持容器启停操作

#### MySQL蜜罐
- 查看所有登录尝试和连接
- 按目标IP、数据库名、错误代码搜索
- 关键词搜索（用户名、密码等）
- 导出日志数据

#### 日志导出
- 支持多种格式导出（JSON、CSV、TXT）
- 按时间范围和条件过滤
- 批量处理和下载

## 📊 系统监控与统计

### 实时监控指标
- **连接数**: 当前活跃连接数量
- **认证失败率**: 异常登录尝试统计
- **文件上传**: 检测到的恶意文件
- **命令执行**: 蜜罐中执行的命令记录
- **地理分布**: 攻击源IP地理位置分析

详细部署指南请查看 [功能完善开发流程指南](docs/功能完善开发流程指南.md)

## 🔄 版本更新

### v2.1 (当前版本) - 2025年11月
**新功能:**
- ✅ **后端端口优化**: 调整监听端口从9090改为8848，避免端口冲突
- ✅ **MySQL蜜罐增强**: 
  - 新增目标IP查询接口
  - 新增数据库名查询接口
  - 新增错误代码查询接口
  - 新增关键词搜索接口
  - 支持高级组合搜索
- ✅ **前端调试工具**: 完整的Web调试界面（debug-container.html）
  - Docker容器管理面板
  - MySQL蜜罐日志查看和搜索
  - Cowrie SSH蜜罐管理
  - Heralding认证蜜罐管理
  - 实时统计和分析
  - 系统设置和命令执行
- ✅ **容器运行时日志**: 支持实时查看容器执行日志
- ✅ **日志同步功能**: 远程日志同步和本地分析

### v2.0 (稳定版)
- ✅ 完整的后端API系统
- ✅ 多协议蜜罐支持 (Cowrie, Heralding, MySQL)
- ✅ 病毒检测功能 (基于ClamAV)
- ✅ 日志记录和导出
- ✅ MySQL蜜罐集成
- ✅ Docker容器管理

### 路线图
- **v2.2**: 高级攻击溯源功能
- **v2.3**: 机器学习威胁检测
- **v3.0**: 分布式集群支持

详细版本说明请查看 [发布说明](RELEASE_NOTES.md)

## 🤝 贡献

我们欢迎所有形式的贡献，包括但不限于：
- 🐛 报告Bug
- 💡 提出新功能建议
- 📝 改进文档
- 🔧 提交代码

## 📄 许可证

本项目采用 MIT 许可证。详细信息请查看 [LICENSE](LICENSE) 文件。

## 📞 联系我们

- **项目维护**: Andorralee开发团队
- **技术支持**: [GitHub Issues](https://github.com/your-org/andorralee/issues)
- **邮箱**: support@andorralee.com

## 🙏 致谢

感谢所有为本项目做出贡献的开发者和用户！

---

**⭐ 如果这个项目对您有帮助，请给我们一个Star！**

![GitHub stars](https://img.shields.io/github/stars/your-org/andorralee?style=social)
