# 蜜罐系统功能使用指南

## 目录
1. [端口管理功能](#端口管理功能)
2. [病毒检测功能](#病毒检测功能)

---

# 端口管理功能

## 概述

新的端口管理功能解决了甲方提出的端口映射需求，支持：

1. **自动端口分配** - 系统自动分配空闲端口
2. **指定端口分配** - 支持指定特定端口
3. **智能端口范围** - 根据服务类型选择合适的端口范围
4. **端口冲突检测** - 避免端口冲突
5. **端口生命周期管理** - 自动释放容器删除时的端口

## 端口分配策略

### 服务类型端口范围

- **MySQL**: 13306-13399
- **SSH**: 12222-12299  
- **HTTP/Web**: 18000-18999
- **默认动态范围**: 10000-19999, 30000-39999

### 使用场景示例

#### 场景1: MySQL容器自动端口分配

```bash
# 创建MySQL容器，自动分配端口
curl -X POST http://localhost:8081/api/v1/container-instances \
  -H "Content-Type: application/json" \
  -d '{
    "name": "MySQL蜜罐",
    "honeypot_name": "mysql-honeypot",
    "image_name": "mysql:8.0",
    "protocol": "mysql",
    "port_mappings": {
      "3306": "auto"
    },
    "environment": {
      "MYSQL_ROOT_PASSWORD": "honeypot123"
    },
    "description": "MySQL蜜罐实例"
  }'
```

**响应示例:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "port_mappings": {
      "3306": "13306"
    },
    "requested_port_mappings": {
      "3306": "auto"
    }
  }
}
```

#### 场景2: 指定端口映射

```bash
# 创建容器，指定端口12345
curl -X POST http://localhost:8081/api/v1/container-instances \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Web蜜罐",
    "honeypot_name": "web-honeypot",
    "image_name": "nginx:latest",
    "protocol": "http",
    "port_mappings": {
      "80": "12345"
    },
    "description": "指定端口的Web蜜罐"
  }'
```

#### 场景3: 混合端口分配

```bash
# 混合使用自动分配和指定端口
curl -X POST http://localhost:8081/api/v1/container-instances \
  -H "Content-Type: application/json" \
  -d '{
    "name": "多服务蜜罐",
    "honeypot_name": "multi-service",
    "image_name": "custom-honeypot:latest",
    "protocol": "multi",
    "port_mappings": {
      "22": "auto",      // SSH自动分配
      "80": "auto",      // HTTP自动分配
      "3306": "auto",    // MySQL自动分配
      "8080": "18080"    // 指定端口
    },
    "description": "多服务蜜罐实例"
  }'
```

## 端口管理API

### 1. 自动分配端口

```bash
POST /api/v1/ports/allocate
{
  "container_id": "container-name",
  "service_type": "mysql",
  "description": "MySQL服务端口"
}
```

### 2. 分配指定端口

```bash
POST /api/v1/ports/allocate-specific
{
  "port": 12345,
  "container_id": "container-name", 
  "service_type": "http",
  "description": "Web服务端口"
}
```

### 3. 自动分配端口映射

```bash
POST /api/v1/ports/auto-allocate-mapping
{
  "container_id": "container-name",
  "port_mappings": {
    "22": "auto",
    "80": "auto", 
    "3306": "13306"
  }
}
```

### 4. 查询端口信息

```bash
# 查询特定端口
GET /api/v1/ports/12345

# 查询所有已分配端口
GET /api/v1/ports/allocated

# 查询容器的端口
GET /api/v1/ports/container/container-name

# 获取端口统计
GET /api/v1/ports/statistics
```

### 5. 释放端口

```bash
# 释放特定端口
DELETE /api/v1/ports/12345/release

# 释放容器的所有端口
DELETE /api/v1/ports/container/container-name/release
```

### 6. 检查端口可用性

```bash
GET /api/v1/ports/12345/check
```

### 7. 获取可用端口

```bash
POST /api/v1/ports/available
{
  "start": 10000,
  "end": 20000,
  "limit": 10
}
```

## PowerShell使用示例

```powershell
# 设置基础URL
$BaseUrl = "http://localhost:8081/api/v1"

# 创建自动端口分配的容器
$containerRequest = @{
    name = "MySQL蜜罐"
    honeypot_name = "mysql-test"
    image_name = "mysql:8.0"
    protocol = "mysql"
    port_mappings = @{
        "3306" = "auto"
    }
    environment = @{
        "MYSQL_ROOT_PASSWORD" = "honeypot123"
    }
    description = "自动端口分配的MySQL蜜罐"
} | ConvertTo-Json -Depth 3

$result = Invoke-RestMethod -Uri "$BaseUrl/container-instances" -Method POST -Body $containerRequest -ContentType "application/json"

Write-Host "容器创建成功，分配的端口: $($result.data.port_mappings.'3306')"
```

## 与Nginx流量转发集成

创建容器后，可以配置Nginx进行流量转发：

```nginx
# /etc/nginx/sites-available/honeypot-mysql
upstream mysql_honeypot {
    server 127.0.0.1:13306;  # 自动分配的端口
}

server {
    listen 3306;
    proxy_pass mysql_honeypot;
    proxy_timeout 1s;
    proxy_responses 1;
}
```

## 最佳实践

1. **使用自动分配** - 对于大多数场景，推荐使用"auto"自动分配端口
2. **服务类型标识** - 正确设置protocol字段，帮助系统选择合适的端口范围
3. **端口监控** - 定期查询端口统计信息，监控端口使用情况
4. **及时清理** - 删除容器时会自动释放端口，无需手动操作
5. **冲突处理** - 如果指定端口被占用，系统会返回错误，可以选择其他端口

## 故障排除

### 端口分配失败
- 检查端口范围是否有可用端口
- 确认指定端口未被占用
- 查看系统端口统计信息

### 容器创建失败
- 检查端口映射格式是否正确
- 确认Docker服务正常运行
- 查看系统日志获取详细错误信息

### 端口冲突
- 使用端口检查API确认端口状态
- 选择其他端口或使用自动分配
- 检查系统防火墙设置

## 测试脚本

使用提供的测试脚本验证功能：

```powershell
# 运行完整测试
.\test-port-management.ps1

# 详细输出模式
.\test-port-management.ps1 -Verbose

# 指定服务器地址
.\test-port-management.ps1 -BaseUrl "http://your-server:8081/api/v1"
```

## 技术实现

- **端口管理器**: 单例模式，线程安全
- **端口范围**: 可配置的端口分配策略
- **冲突检测**: 实时检查端口占用状态
- **生命周期**: 与容器生命周期绑定
- **持久化**: 内存管理，重启后重新扫描

---

# 病毒检测功能

## 概述

蜜罐系统集成了强大的病毒检测功能，能够实时检测和分析恶意软件，提供多层次的安全防护。

## 核心特性

### 检测方法
- **哈希匹配** - MD5/SHA256精确匹配
- **字符串特征** - 恶意代码特征字符串检测
- **模糊匹配** - 基于相似度的模糊检测
- **PE文件分析** - Windows可执行文件深度分析
- **启发式检测** - 基于行为特征的检测

### 支持的文件类型
- **PE文件** - Windows可执行文件(.exe, .dll)
- **脚本文件** - PowerShell, 批处理文件
- **压缩文件** - ZIP, RAR等压缩包
- **文档文件** - PDF, Office文档
- **其他文件** - 任意二进制文件

## API接口使用

### 1. 文件上传扫描

```bash
# 扫描上传的文件
curl -X POST http://localhost:8081/api/v1/malware/scan/file \
  -F "file=@suspicious_file.exe" \
  -F "source_ip=192.168.1.100" \
  -F "container_id=honeypot-web-1"
```

**响应示例:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "scan_result": {
      "file_info": {
        "file_name": "suspicious_file.exe",
        "file_size": 1024000,
        "file_type": "pe",
        "md5_hash": "d41d8cd98f00b204e9800998ecf8427e",
        "sha256_hash": "e3b0c44298fc1c149afbf4c8996fb924..."
      },
      "is_malware": true,
      "threat_level": "HIGH",
      "detected_by": ["Hash-SHA256", "PE-Analysis"],
      "signatures": [
        {
          "rule_name": "Trojan.Win32.Generic",
          "match_type": "HASH",
          "confidence": 1.0,
          "description": "已知恶意软件哈希匹配"
        }
      ],
      "pe_info": {
        "architecture": "x86",
        "suspicious_apis": ["CreateRemoteThread", "WriteProcessMemory"],
        "packer_detected": true
      },
      "scan_time": "150ms",
      "timestamp": "2024-07-17T10:30:00Z"
    },
    "message": "检测到恶意软件！威胁等级: HIGH"
  }
}
```

### 2. URL文件扫描

```bash
# 扫描URL指向的文件
curl -X POST http://localhost:8081/api/v1/malware/scan/url \
  -H "Content-Type: application/json" \
  -d '{
    "url": "http://malicious-site.com/malware.exe",
    "source_ip": "192.168.1.101",
    "container_id": "honeypot-ftp-1"
  }'
```

### 3. 批量文件扫描

```bash
# 批量扫描多个文件
curl -X POST http://localhost:8081/api/v1/malware/scan/batch \
  -H "Content-Type: application/json" \
  -d '{
    "file_paths": [
      "/tmp/file1.exe",
      "/tmp/file2.dll",
      "/tmp/script.ps1"
    ],
    "source_ip": "192.168.1.102",
    "container_id": "honeypot-smb-1"
  }'
```

### 4. 检测统计信息

```bash
# 获取检测统计
curl -X GET http://localhost:8081/api/v1/malware/statistics
```

**响应示例:**
```json
{
  "code": 200,
  "data": {
    "total_scans": 1500,
    "malware_detected": 75,
    "clean_files": 1425,
    "detection_rate": "5.0%",
    "threat_levels": {
      "critical": 15,
      "high": 25,
      "medium": 30,
      "low": 5
    },
    "last_update": "2024-07-17 10:30:00"
  }
}
```

## 病毒特征管理

### 1. 获取特征列表

```bash
curl -X GET "http://localhost:8081/api/v1/malware/signatures?page=1&size=20"
```

### 2. 添加病毒特征

```bash
curl -X POST http://localhost:8081/api/v1/malware/signatures \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Custom-Malware-Signature",
    "type": 0,
    "data": "MALICIOUS_STRING_PATTERN",
    "threat_family": "CustomTrojan",
    "severity": 2,
    "description": "自定义恶意软件特征"
  }'
```

### 3. 测试特征

```bash
curl -X POST http://localhost:8081/api/v1/malware/signatures/1/test \
  -F "file=@test_file.exe"
```

## 与蜜罐系统集成

### 容器文件监控

病毒检测功能已集成到容器创建流程中：

```bash
# 创建带病毒检测的容器
curl -X POST http://localhost:8081/api/v1/container-instances \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Web蜜罐",
    "honeypot_name": "web-honeypot",
    "image_name": "nginx:latest",
    "protocol": "http",
    "port_mappings": {
      "80": "auto"
    },
    "enable_malware_detection": true,
    "scan_uploads": true
  }'
```

### 实时流量检测

系统会自动检测通过蜜罐传输的文件：

1. **上传文件检测** - 自动扫描上传到蜜罐的文件
2. **下载文件检测** - 检测从蜜罐下载的文件
3. **实时告警** - 发现恶意文件时立即告警
4. **自动隔离** - 恶意文件自动隔离到安全目录

## PowerShell使用示例

```powershell
# 设置基础URL
$BaseUrl = "http://localhost:8081/api/v1"

# 扫描文件
$form = @{
    file = Get-Item "suspicious_file.exe"
    source_ip = "192.168.1.100"
    container_id = "test-container"
}

$result = Invoke-RestMethod -Uri "$BaseUrl/malware/scan/file" -Method POST -Form $form

if ($result.data.scan_result.is_malware) {
    Write-Host "检测到恶意软件！" -ForegroundColor Red
    Write-Host "威胁等级: $($result.data.scan_result.threat_level)" -ForegroundColor Yellow
    Write-Host "检测方法: $($result.data.scan_result.detected_by -join ', ')" -ForegroundColor Yellow
} else {
    Write-Host "文件安全" -ForegroundColor Green
}
```

## 威胁等级说明

- **CRITICAL** - 严重威胁，已知恶意软件
- **HIGH** - 高威胁，强特征匹配
- **MEDIUM** - 中等威胁，可疑行为
- **LOW** - 低威胁，轻微可疑

## 检测结果处理

### 自动处理
- **样本保存** - 恶意文件自动保存到样本库
- **文件隔离** - 危险文件移动到隔离目录
- **日志记录** - 详细检测日志记录
- **告警通知** - 实时安全告警

### 手动处理
- **误报处理** - 支持白名单机制
- **特征更新** - 支持自定义特征规则
- **结果审核** - 可疑结果人工确认

## 性能优化

### 缓存机制
- **哈希缓存** - 已扫描文件哈希缓存
- **结果缓存** - 检测结果缓存
- **特征缓存** - 病毒特征内存缓存

### 并发处理
- **多线程扫描** - 支持并发文件扫描
- **异步处理** - 大文件异步处理
- **队列管理** - 扫描任务队列管理

## 测试验证

使用提供的测试脚本验证功能：

```powershell
# 运行病毒检测测试
.\test-malware-detection.ps1

# 详细输出模式
.\test-malware-detection.ps1 -Verbose

# 指定服务器地址
.\test-malware-detection.ps1 -BaseUrl "http://your-server:8081/api/v1"
```

## 安全考虑

### 样本隔离
- **沙箱环境** - 恶意样本在隔离环境中分析
- **权限控制** - 最小权限原则运行
- **加密存储** - 样本文件加密存储

### 误报控制
- **多重验证** - 多种检测方法交叉验证
- **置信度评分** - 基于置信度的结果评估
- **白名单机制** - 已知安全文件白名单

这个病毒检测系统为蜜罐提供了强大的恶意软件检测能力，能够有效识别和分析各种威胁，提升整体安全防护水平。
