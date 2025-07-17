# Andorralee 蜜罐管理系统

一个基于 Go 语言开发的现代化蜜罐管理系统，集成了智能端口管理和病毒检测功能。

## 🚀 核心特性

### 🐳 容器管理
- **Docker 镜像管理** - 拉取、构建、删除镜像
- **容器实例管理** - 创建、启动、停止、删除蜜罐容器
- **内存容器支持** - 虚拟容器实例，无需真实Docker环境

### 🔌 智能端口管理 (NEW!)
- **自动端口分配** - 智能分配空闲端口，避免冲突
- **服务类型识别** - 根据服务类型(MySQL/SSH/HTTP)选择合适端口范围
- **端口生命周期管理** - 容器删除时自动释放端口
- **端口冲突检测** - 实时检测端口占用状态

### 🛡️ 病毒检测系统 (NEW!)
- **多重检测引擎** - 哈希匹配、字符串特征、模糊匹配
- **PE文件分析** - Windows可执行文件深度分析
- **启发式检测** - 基于行为特征的智能检测
- **实时威胁感知** - 自动检测上传/下载的恶意文件

### 🔍 安全功能
- **端口扫描** - 内置端口扫描和检测
- **攻击溯源** - 记录攻击来源和传播路径
- **实时监控** - 蜜罐活动实时监控
- **日志管理** - 完整的操作和安全日志

## 🚀 快速开始

### 环境要求
- Go 1.19+
- Docker (可选，支持无Docker模式)
- MySQL 或达梦数据库 (可选)

### 一键启动
```bash
# 1. 克隆项目
git clone https://github.com/yifei-zeng/andorralee.git
cd andorralee

# 2. 安装依赖
go mod tidy

# 3. 编译运行
go build cmd/main.go
./main

# 4. 访问系统
# http://localhost:8081
```

## 🎯 新功能演示

### 智能端口管理
```bash
# 自动分配端口
curl -X POST http://localhost:8081/api/v1/container-instances \
  -d '{
    "name": "MySQL蜜罐",
    "image_name": "mysql:8.0",
    "port_mappings": {
      "3306": "auto"  # 自动分配MySQL服务端口
    }
  }'

# 响应: 自动分配到13306端口
{
  "port_mappings": {"3306": "13306"},
  "message": "端口自动分配成功"
}
```

### 病毒检测API
```bash
# 扫描上传文件
curl -X POST http://localhost:8081/api/v1/malware/scan/file \
  -F "file=@suspicious.exe" \
  -F "source_ip=192.168.1.100"

# 响应: 检测结果
{
  "scan_result": {
    "is_malware": true,
    "threat_level": "HIGH",
    "detected_by": ["Hash-SHA256", "PE-Analysis"],
    "signatures": [...]
  }
}
```

## 📚 主要API接口

### 端口管理
- `POST /api/v1/ports/allocate` - 自动分配端口
- `POST /api/v1/ports/allocate-specific` - 分配指定端口
- `GET /api/v1/ports/statistics` - 获取端口统计
- `DELETE /api/v1/ports/{port}/release` - 释放端口

### 病毒检测
- `POST /api/v1/malware/scan/file` - 文件扫描
- `POST /api/v1/malware/scan/url` - URL扫描
- `POST /api/v1/malware/scan/batch` - 批量扫描
- `GET /api/v1/malware/statistics` - 检测统计

### 容器管理
- `GET/POST/PUT/DELETE /api/v1/container-instances` - 容器实例管理
- `GET/POST/DELETE /api/v1/memory-containers` - 内存容器管理
- `GET/POST/DELETE /api/v1/images` - 镜像管理

## 🧪 功能测试

### 端口管理测试
```powershell
# 运行端口管理测试
.\test-port-management.ps1

# 详细输出
.\test-port-management.ps1 -Verbose
```

### 病毒检测测试
```powershell
# 运行病毒检测测试
.\test-malware-detection.ps1

# 指定服务器
.\test-malware-detection.ps1 -BaseUrl "http://your-server:8081/api/v1"
```

## 📁 项目结构

```
andorralee/
├── cmd/main.go                           # 主程序入口
├── internal/
│   ├── handlers/                         # HTTP处理器
│   │   ├── malware_detection_handler.go  # 病毒检测API
│   │   └── port_manager_handler.go       # 端口管理API
│   ├── services/                         # 业务逻辑
│   │   ├── malware_detection_service.go  # 病毒检测引擎
│   │   ├── port_manager_service.go       # 端口管理服务
│   │   ├── string_matcher.go             # 字符串匹配器
│   │   └── pe_analyzer.go                # PE文件分析器
│   └── config/                           # 配置管理
├── docs/                                 # 文档
│   ├── port_management_guide.md          # 功能使用指南
│   ├── malware_detection_design.md       # 病毒检测设计
│   └── malware_detection_detailed_guide.md # 病毒检测详细指南
├── test-port-management.ps1              # 端口管理测试
├── test-malware-detection.ps1            # 病毒检测测试
└── routers/router.go                     # 路由配置
```

## 🌟 使用场景

### 1. 智能蜜罐部署
自动分配端口，避免冲突，支持多服务蜜罐

### 2. 恶意文件检测
实时检测上传文件，自动识别威胁

### 3. 攻击溯源分析
记录攻击来源，分析恶意文件特征

## 🔒 安全特性

- **端口冲突自动检测**
- **多层恶意软件检测**
- **实时威胁感知**
- **自动样本隔离**

## 📖 详细文档

- [功能使用指南](docs/port_management_guide.md)
- [病毒检测设计](docs/malware_detection_design.md)
- [病毒检测详细指南](docs/malware_detection_detailed_guide.md)

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 发起 Pull Request

## 📄 许可证

本项目采用 MIT 许可证。

---

**Andorralee** - 让蜜罐管理更智能、更安全！ 🍯🛡️
