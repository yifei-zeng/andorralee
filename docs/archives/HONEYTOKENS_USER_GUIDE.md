# 🍯 蜜签(HoneyTokens)功能使用指南

## 📋 功能概述

蜜签是一种欺骗性安全技术，通过在系统中放置虚假的敏感信息（如假凭证、假文件路径、假API密钥等），当攻击者尝试使用这些信息时，系统会立即检测到并记录攻击行为。

### 🎯 核心价值
- **早期威胁检测**: 攻击者一旦触碰蜜签，立即暴露其存在
- **零误报**: 正常用户不会接触到蜜签，触发即为攻击
- **攻击溯源**: 记录攻击者IP、时间、行为等详细信息
- **威胁情报**: 了解攻击者的目标和手法

## 🔧 功能特性

### 支持的蜜签类型
- **credential**: 虚假凭证（用户名密码、API密钥等）
- **file**: 虚假文件路径（敏感文档、配置文件等）
- **url**: 虚假URL地址
- **email**: 虚假邮箱地址

### 核心功能
- ✅ 蜜签创建和管理
- ✅ 触发事件记录
- ✅ 实时警报通知
- ✅ 统计分析报告
- ✅ 攻击者行为追踪

## 🚀 快速开始

### 1. 启动服务
```bash
go run cmd/main.go
```
服务启动后会自动初始化3个默认蜜签。

### 2. 查看现有蜜签
```bash
curl http://localhost:9090/api/v1/honeytokens
```

### 3. 创建新蜜签
```bash
curl -X POST http://localhost:9090/api/v1/honeytokens \
  -H "Content-Type: application/json" \
  -d '{
    "name": "API密钥",
    "type": "credential", 
    "content": "sk-1234567890abcdef",
    "description": "虚假的API密钥"
  }'
```

## 📚 API接口详解

### 基础管理接口

#### 1. 获取所有蜜签
```http
GET /api/v1/honeytokens
```

**响应示例:**
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "管理员凭证",
      "type": "credential",
      "content": "admin:123456",
      "description": "虚假的管理员账号密码",
      "is_active": true,
      "create_time": "2025-07-23T15:46:07+08:00",
      "update_time": "2025-07-23T15:46:07+08:00",
      "trigger_count": 0
    }
  ]
}
```

#### 2. 创建蜜签
```http
POST /api/v1/honeytokens
Content-Type: application/json

{
  "name": "蜜签名称",
  "type": "credential|file|url|email",
  "content": "蜜签内容",
  "description": "描述信息"
}
```

**参数说明:**
- `name`: 蜜签名称（必填）
- `type`: 蜜签类型（必填）
- `content`: 蜜签内容（必填）
- `description`: 描述信息（可选）

#### 3. 获取单个蜜签
```http
GET /api/v1/honeytokens/{id}
```

#### 4. 更新蜜签
```http
PUT /api/v1/honeytokens/{id}
Content-Type: application/json

{
  "name": "新名称",
  "description": "新描述"
}
```

#### 5. 删除蜜签
```http
DELETE /api/v1/honeytokens/{id}
```

### 触发和监控接口

#### 6. 触发蜜签（模拟攻击）
```http
POST /api/v1/honeytokens/{id}/trigger
Content-Type: application/json

{
  "action": "unauthorized_access",
  "details": "攻击者尝试使用虚假凭证登录"
}
```

**参数说明:**
- `action`: 触发动作（必填）
- `details`: 详细信息（可选）

**响应示例:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "蜜签触发记录成功",
    "token_name": "管理员凭证",
    "trigger_id": 1,
    "trigger_time": "2025-07-23T15:47:49+08:00"
  }
}
```

#### 7. 获取触发记录
```http
GET /api/v1/honeytokens/triggers
```

**响应示例:**
```json
{
  "code": 0,
  "message": "success", 
  "data": [
    {
      "id": 1,
      "token_id": 1,
      "token_name": "管理员凭证",
      "source_ip": "192.168.1.100",
      "user_agent": "Mozilla/5.0...",
      "trigger_time": "2025-07-23T15:47:49+08:00",
      "action": "unauthorized_access",
      "details": "攻击者尝试使用虚假凭证登录"
    }
  ]
}
```

#### 8. 获取统计信息
```http
GET /api/v1/honeytokens/statistics
```

**响应示例:**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total_tokens": 5,
    "active_tokens": 5,
    "total_triggers": 2,
    "tokens_by_type": {
      "credential": 3,
      "file": 2
    },
    "recent_triggers": [...]
  }
}
```

## 💡 使用场景示例

### 场景1: 数据库凭证蜜签
```bash
# 创建虚假数据库连接字符串
curl -X POST http://localhost:9090/api/v1/honeytokens \
  -H "Content-Type: application/json" \
  -d '{
    "name": "生产数据库连接",
    "type": "credential",
    "content": "mysql://root:prod_password@db.company.com:3306/users",
    "description": "虚假的生产数据库连接字符串"
  }'
```

### 场景2: 敏感文件路径蜜签
```bash
# 创建虚假敏感文件路径
curl -X POST http://localhost:9090/api/v1/honeytokens \
  -H "Content-Type: application/json" \
  -d '{
    "name": "客户数据文件",
    "type": "file", 
    "content": "/var/backups/customer_data_2024.sql",
    "description": "虚假的客户数据备份文件路径"
  }'
```

### 场景3: API密钥蜜签
```bash
# 创建虚假API密钥
curl -X POST http://localhost:9090/api/v1/honeytokens \
  -H "Content-Type: application/json" \
  -d '{
    "name": "支付网关API密钥",
    "type": "credential",
    "content": "pk_live_51234567890abcdef",
    "description": "虚假的支付网关API密钥"
  }'
```

## 🚨 实时监控

### 服务器日志监控
当蜜签被触发时，服务器会输出实时警报：
```
🚨 蜜签触发警报: 管理员凭证 (ID:1) 被 192.168.1.100 触发，动作: unauthorized_access
```

### 触发事件包含信息
- **蜜签信息**: ID、名称、类型、内容
- **攻击者信息**: 源IP地址、User-Agent
- **触发信息**: 触发时间、动作、详细描述
- **统计信息**: 触发次数、频率分析

## 📊 数据分析

### 统计维度
- **按类型统计**: credential、file、url、email
- **按时间统计**: 小时、天、周、月趋势
- **按攻击者统计**: IP地址、User-Agent分析
- **按蜜签统计**: 最常被触发的蜜签

### 威胁情报
- **攻击模式**: 攻击者的行为模式分析
- **攻击来源**: 地理位置和网络分析
- **攻击时间**: 攻击活跃时段分析
- **攻击目标**: 攻击者最感兴趣的信息类型

## 🔒 安全建议

### 蜜签部署策略
1. **多样化部署**: 使用不同类型的蜜签
2. **真实性**: 蜜签内容要看起来真实可信
3. **隐蔽性**: 蜜签不应被正常用户发现
4. **更新频率**: 定期更新蜜签内容

### 响应策略
1. **立即响应**: 蜜签触发后立即调查
2. **隔离措施**: 必要时隔离可疑IP
3. **证据保全**: 保存完整的攻击证据
4. **威胁狩猎**: 主动搜索相关威胁

## 🛠️ 故障排除

### 常见问题

**Q: 蜜签创建失败？**
A: 检查必填字段是否完整，type字段是否为支持的类型。

**Q: 触发记录为空？**
A: 确认蜜签状态为active，检查触发请求格式是否正确。

**Q: 统计数据不准确？**
A: 蜜签数据存储在内存中，服务重启后会重置。

### 调试命令
```bash
# 检查服务状态
curl http://localhost:9090/api/v1/health

# 查看所有蜜签
curl http://localhost:9090/api/v1/honeytokens

# 查看统计信息
curl http://localhost:9090/api/v1/honeytokens/statistics
```

## 📞 技术支持

- **API文档**: `/static/swagger` 目录
- **项目地址**: https://github.com/yifei-zeng/andorralee
- **问题反馈**: GitHub Issues

---

**注意**: 蜜签功能使用内存存储，服务重启后数据会重置。生产环境建议配置数据库持久化存储。
