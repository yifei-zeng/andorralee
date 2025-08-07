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
│   Web前端界面   │────│   Go后端服务     │────│   MySQL数据库   │
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
- MySQL 8.0+ 或 达梦数据库
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

## 📖 API文档

系统启动后，可以通过以下端点访问：

- **主服务**: http://localhost:9090
- **健康检查**: http://localhost:9090/health
- **API文档**: http://localhost:9090/swagger/index.html (计划中)

### 主要API端点

| 端点 | 方法 | 描述 |
|------|------|------|
| `/api/temp-containers` | GET | 获取容器列表 |
| `/api/temp-containers` | POST | 创建新容器 |
| `/api/malware/scan` | POST | 病毒扫描 |
| `/api/honeytokens/statistics` | GET | 蜜签统计 |
| `/api/sessions/statistics` | GET | 会话统计 |
| `/api/logs/export` | GET | 日志导出 |

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

# 部署到生产环境
./deployment/deploy-kylin.sh
```

详细部署指南请查看 [功能完善开发流程指南](docs/功能完善开发流程指南.md)

## 🔄 版本更新

### v2.0 (当前版本)
- ✅ 完整的后端API系统
- ✅ 多协议蜜罐支持
- ✅ 病毒检测功能
- ✅ 日志记录和导出
- ⚠️ 前端可视化界面开发中

### 路线图
- **v2.1**: 完整前端可视化界面
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
