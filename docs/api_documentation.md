# Andorralee RESTful API 文档

## 概述

Andorralee 是一个基于 Go 语言开发的蜜罐系统，集成了 Docker 容器管理、多种数据库支持以及 AI 分析能力。本文档提供了 Andorralee 系统的 RESTful API 接口说明。

## API 版本

当前 API 版本：v1

所有 API 请求都应该发送到基础 URL：`/api`

## 认证

所有 API 请求需要在 HTTP 头部包含认证信息（具体认证方式待实现）。

## 通用响应格式

所有 API 响应都使用 JSON 格式，并包含以下字段：

成功响应：
```json
{
  "data": {}, // 响应数据，可能是对象或数组
  "message": "操作成功信息"
}
```

错误响应：
```json
{
  "error": "错误信息"
}
```

## 通用状态码

| 状态码 | 描述 |
| ------ | ---- |
| 200 | 请求成功 |
| 400 | 请求参数错误 |
| 401 | 未授权 |
| 403 | 禁止访问 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

## 蜜罐管理 API

### 部署蜜罐

- **URL**: `/api/honeypot/deploy`
- **方法**: `POST`
- **描述**: 部署一个新的蜜罐

**请求参数**:

```json
{
  "type": "string" // 蜜罐类型，必填，可选值："SSH", "HTTP", "MySQL", "Redis", "Custom"
}
```

**响应示例**:

```json
{
  "data": {
    "container_id": "abc123def456"
  },
  "message": "蜜罐部署成功"
}
```

### 停止蜜罐

- **URL**: `/api/honeypot/stop/{id}`
- **方法**: `POST`
- **描述**: 停止指定 ID 的蜜罐

**路径参数**:
- `id`: 容器 ID，必填

**响应示例**:

```json
{
  "message": "蜜罐停止成功"
}
```

### 获取蜜罐状态

- **URL**: `/api/honeypot/status/{id}`
- **方法**: `GET`
- **描述**: 获取指定 ID 的蜜罐状态

**路径参数**:
- `id`: 容器 ID，必填

**响应示例**:

```json
{
  "data": {
    "id": "abc123def456",
    "status": "running",
    "type": "SSH",
    "created_at": "2023-01-01T12:00:00Z",
    "ports": ["22:2222"]
  }
}
```

### 列出所有蜜罐

- **URL**: `/api/honeypot/list`
- **方法**: `GET`
- **描述**: 获取所有蜜罐的列表

**查询参数**:
- `type`: 蜜罐类型，可选，用于过滤特定类型的蜜罐

**响应示例**:

```json
{
  "data": [
    {
      "id": "abc123def456",
      "status": "running",
      "type": "SSH",
      "created_at": "2023-01-01T12:00:00Z"
    },
    {
      "id": "def456abc789",
      "status": "stopped",
      "type": "HTTP",
      "created_at": "2023-01-02T12:00:00Z"
    }
  ]
}
```

### 获取蜜罐日志

- **URL**: `/api/honeypot/logs/{id}`
- **方法**: `GET`
- **描述**: 获取指定 ID 的蜜罐日志

**路径参数**:
- `id`: 容器 ID，必填

**查询参数**:
- `lines`: 返回的日志行数，可选，默认为 100

**响应示例**:

```json
{
  "data": {
    "logs": "日志内容..."
  }
}
```

## 流量管理 API

### 添加重定向规则

- **URL**: `/api/traffic/redirect/add`
- **方法**: `POST`
- **描述**: 添加流量重定向规则

**请求参数**:

```json
{
  "source_port": "string", // 源端口，必填
  "target_port": "string"  // 目标端口，必填
}
```

**响应示例**:

```json
{
  "message": "重定向规则添加成功"
}
```

### 删除重定向规则

- **URL**: `/api/traffic/redirect/remove`
- **方法**: `POST`
- **描述**: 删除流量重定向规则

**请求参数**:

```json
{
  "source_port": "string", // 源端口，必填
  "target_port": "string"  // 目标端口，必填
}
```

**响应示例**:

```json
{
  "message": "重定向规则删除成功"
}
```

### 添加过滤规则

- **URL**: `/api/traffic/filter/add`
- **方法**: `POST`
- **描述**: 添加流量过滤规则

**请求参数**:

```json
{
  "source_ip": "string",   // 源 IP，必填
  "target_port": "string"  // 目标端口，必填
}
```

**响应示例**:

```json
{
  "message": "过滤规则添加成功"
}
```

### 删除过滤规则

- **URL**: `/api/traffic/filter/remove`
- **方法**: `POST`
- **描述**: 删除流量过滤规则

**请求参数**:

```json
{
  "source_ip": "string",   // 源 IP，必填
  "target_port": "string"  // 目标端口，必填
}
```

**响应示例**:

```json
{
  "message": "过滤规则删除成功"
}
```

### 列出所有规则

- **URL**: `/api/traffic/rules`
- **方法**: `GET`
- **描述**: 获取所有流量规则

**响应示例**:

```json
{
  "data": {
    "rules": [
      {
        "type": "redirect",
        "source_port": "8080",
        "target_port": "80"
      },
      {
        "type": "filter",
        "source_ip": "192.168.1.1",
        "target_port": "22"
      }
    ]
  }
}
```

### 保存规则

- **URL**: `/api/traffic/rules/save`
- **方法**: `POST`
- **描述**: 保存当前所有流量规则

**响应示例**:

```json
{
  "message": "规则保存成功"
}
```

### 恢复规则

- **URL**: `/api/traffic/rules/restore`
- **方法**: `POST`
- **描述**: 恢复之前保存的流量规则

**响应示例**:

```json
{
  "message": "规则恢复成功"
}
```

## 蜜签管理 API

### 创建蜜签

- **URL**: `/api/bait/create`
- **方法**: `POST`
- **描述**: 创建一个新的蜜签

**请求参数**:

```json
{
  "name": "string",       // 蜜签名称，必填
  "type": "string",       // 蜜签类型，必填
  "content": "string",    // 蜜签内容，必填
  "location": "string",   // 蜜签位置，必填
  "description": "string" // 蜜签描述，可选
}
```

**响应示例**:

```json
{
  "message": "蜜签创建成功"
}
```

### 获取蜜签信息

- **URL**: `/api/bait/{id}`
- **方法**: `GET`
- **描述**: 获取指定 ID 的蜜签信息

**路径参数**:
- `id`: 蜜签 ID，必填

**响应示例**:

```json
{
  "data": {
    "id": "abc123",
    "name": "测试蜜签",
    "type": "文档",
    "content": "机密内容",
    "location": "/var/www/html",
    "description": "测试用蜜签",
    "created_at": "2023-01-01T12:00:00Z"
  }
}
```

### 列出所有蜜签

- **URL**: `/api/bait/list`
- **方法**: `GET`
- **描述**: 获取所有蜜签的列表

**查询参数**:
- `type`: 蜜签类型，可选，用于过滤特定类型的蜜签

**响应示例**:

```json
{
  "data": [
    {
      "id": "abc123",
      "name": "测试蜜签1",
      "type": "文档",
      "location": "/var/www/html",
      "created_at": "2023-01-01T12:00:00Z"
    },
    {
      "id": "def456",
      "name": "测试蜜签2",
      "type": "凭证",
      "location": "/etc/passwd",
      "created_at": "2023-01-02T12:00:00Z"
    }
  ]
}
```

### 删除蜜签

- **URL**: `/api/bait/{id}`
- **方法**: `DELETE`
- **描述**: 删除指定 ID 的蜜签

**路径参数**:
- `id`: 蜜签 ID，必填

**响应示例**:

```json
{
  "message": "蜜签删除成功"
}
```

### 监控蜜签

- **URL**: `/api/bait/{id}/monitor`
- **方法**: `POST`
- **描述**: 开始监控指定 ID 的蜜签

**路径参数**:
- `id`: 蜜签 ID，必填

**响应示例**:

```json
{
  "message": "蜜签监控启动成功"
}
```

## 监控和告警 API

### 创建告警

- **URL**: `/api/monitor/alert`
- **方法**: `POST`
- **描述**: 创建一个新的告警

**请求参数**:

```json
{
  "type": "string",    // 告警类型，必填
  "level": "string",   // 告警级别，必填，可选值："low", "medium", "high", "critical"
  "source": "string",  // 告警来源，必填
  "message": "string", // 告警消息，必填
  "details": "string"  // 告警详情，可选
}
```

**响应示例**:

```json
{
  "message": "告警创建成功"
}
```

### 解决告警

- **URL**: `/api/monitor/alert/{id}/resolve`
- **方法**: `POST`
- **描述**: 将指定 ID 的告警标记为已解决

**路径参数**:
- `id`: 告警 ID，必填

**请求参数**:

```json
{
  "resolved_by": "string" // 解决人，必填
}
```

**响应示例**:

```json
{
  "message": "告警已解决"
}
```

### 列出所有告警

- **URL**: `/api/monitor/alerts`
- **方法**: `GET`
- **描述**: 获取所有告警的列表

**查询参数**:
- `type`: 告警类型，可选，用于过滤特定类型的告警
- `level`: 告警级别，可选，用于过滤特定级别的告警
- `status`: 告警状态，可选，可选值："active", "resolved"

**响应示例**:

```json
{
  "data": [
    {
      "id": "abc123",
      "type": "入侵检测",
      "level": "high",
      "source": "SSH蜜罐",
      "message": "检测到暴力破解尝试",
      "details": "来自 IP 192.168.1.1 的多次失败登录尝试",
      "created_at": "2023-01-01T12:00:00Z",
      "status": "active"
    },
    {
      "id": "def456",
      "type": "蜜签访问",
      "level": "medium",
      "source": "文件蜜签",
      "message": "检测到蜜签文件被访问",
      "details": "机密文档被用户 user1 访问",
      "created_at": "2023-01-02T12:00:00Z",
      "status": "resolved",
      "resolved_at": "2023-01-02T13:00:00Z",
      "resolved_by": "admin"
    }
  ]
}
```

### 获取告警信息

- **URL**: `/api/monitor/alert/{id}`
- **方法**: `GET`
- **描述**: 获取指定 ID 的告警信息

**路径参数**:
- `id`: 告警 ID，必填

**响应示例**:

```json
{
  "data": {
    "id": "abc123",
    "type": "入侵检测",
    "level": "high",
    "source": "SSH蜜罐",
    "message": "检测到暴力破解尝试",
    "details": "来自 IP 192.168.1.1 的多次失败登录尝试",
    "created_at": "2023-01-01T12:00:00Z",
    "status": "active"
  }
}
```

### 监控蜜罐

- **URL**: `/api/monitor/honeypot/{id}`
- **方法**: `POST`
- **描述**: 开始监控指定 ID 的蜜罐

**路径参数**:
- `id`: 蜜罐 ID，必填

**响应示例**:

```json
{
  "message": "蜜罐监控启动成功"
}
```

### 监控蜜签

- **URL**: `/api/monitor/bait/{id}`
- **方法**: `POST`
- **描述**: 开始监控指定 ID 的蜜签

**路径参数**:
- `id`: 蜜签 ID，必填

**响应示例**:

```json
{
  "message": "蜜签监控启动成功"
}
```

### 监控流量

- **URL**: `/api/monitor/traffic`
- **方法**: `POST`
- **描述**: 开始监控网络流量

**请求参数**:

```json
{
  "interface": "string", // 网络接口，必填
  "filter": "string"     // 过滤规则，可选
}
```

**响应示例**:

```json
{
  "message": "流量监控启动成功"
}
```

## 错误码说明

| 错误码 | 描述 |
| ------ | ---- |
| 1001 | 蜜罐部署失败 |
| 1002 | 蜜罐停止失败 |
| 1003 | 蜜罐不存在 |
| 2001 | 流量规则添加失败 |
| 2002 | 流量规则删除失败 |
| 3001 | 蜜签创建失败 |
| 3002 | 蜜签不存在 |
| 4001 | 告警创建失败 |
| 4002 | 告警不存在 |

## 数据模型

### 蜜罐 (Honeypot)

```json
{
  "id": "string",         // 容器 ID
  "type": "string",       // 蜜罐类型
  "status": "string",     // 状态：running, stopped, error
  "created_at": "string", // 创建时间
  "ports": ["string"],    // 端口映射
  "ip": "string"          // IP 地址
}
```

### 流量规则 (TrafficRule)

```json
{
  "id": "string",          // 规则 ID
  "type": "string",        // 规则类型：redirect, filter
  "source_port": "string", // 源端口（重定向规则）
  "target_port": "string", // 目标端口
  "source_ip": "string",   // 源 IP（过滤规则）
  "created_at": "string"   // 创建时间
}
```

### 蜜签 (Bait)

```json
{
  "id": "string",           // 蜜签 ID
  "name": "string",         // 蜜签名称
  "type": "string",         // 蜜签类型
  "content": "string",      // 蜜签内容
  "location": "string",     // 蜜签位置
  "description": "string",  // 蜜签描述
  "created_at": "string",   // 创建时间
  "accessed": "boolean",    // 是否被访问
  "last_accessed": "string" // 最后访问时间
}
```

### 告警 (Alert)

```json
{
  "id": "string",           // 告警 ID
  "type": "string",         // 告警类型
  "level": "string",        // 告警级别
  "source": "string",       // 告警来源
  "message": "string",      // 告警消息
  "details": "string",      // 告警详情
  "created_at": "string",   // 创建时间
  "status": "string",       // 状态：active, resolved
  "resolved_at": "string",  // 解决时间
  "resolved_by": "string"   // 解决人
}
```