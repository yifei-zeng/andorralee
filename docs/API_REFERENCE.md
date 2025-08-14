# Andorralee API 参考 (v2.0)

Base URL: http://localhost:9090/api/v1
统一响应: { code: number, message: string, data: any }

## 健康
- GET /health | /ready | /live

## 恶意软件检测
- POST /malware/scan/file
  - form-data: file
  - 200 data: { scan_result: { is_infected: bool, detections: [] }, md5_hash, sha256_hash }
- POST /malware/scan/url
  - json: { url: string }
  - 200 data 同上，并附 source_url
- GET /malware/scan/history?limit=50
  - 200 data: { scan_history: ScanResult[], total }
- GET /malware/results/{hash}
- GET /malware/statistics
- GET /malware/signatures
- POST /malware/signatures
  - json: { name, pattern, type: hash|string|hex|regex, severity? }

## 威胁情报与会话
- POST /threat/intelligence
  - json: ThreatIntelligence 对象(最小: indicator_type, indicator_value, threat_type)
- GET /threat/intelligence?type=ip&value=1.2.3.4
- GET /threat/assessment
- POST /threat/sessions
  - json: { source_ip, destination_ip?, destination_port?, protocol?, description? }
- PUT /threat/sessions/{sessionId}/end
- GET /threat/sessions/{sessionId}

## 蜜签 (内存存储)
- POST /honeytokens
  - json: { name, type, content, description? }
- GET /honeytokens
- GET /honeytokens/{id}
- PUT /honeytokens/{id}
- DELETE /honeytokens/{id}
- POST /honeytokens/{id}/trigger
  - json: { action, details? }
- GET /honeytokens/triggers?token_id=1
- GET /honeytokens/statistics

## 容器实例 (临时)
- POST /temp-containers
  - json: { name, honeypot_name, image_name, protocol, interface_type, port_mappings, environment }
- GET /temp-containers
- GET /temp-containers/{id}
- DELETE /temp-containers/{id}
- POST /temp-containers/{id}/start|stop|restart
- POST /temp-containers/sync
- GET /temp-containers/id-status
- GET /temp-containers/{id}/scan (端口扫描)

## 容器日志分析
- POST /container-logs/parse
  - json: { container_id, log_lines: string[], log_type? }
- GET /container-logs/container/{container_id}
- GET /container-logs/time-range?start_time=...&end_time=...
- GET /container-logs/event-type/{event_type}
- GET /container-logs/source-ip/{source_ip}
- GET /container-logs/session/{session_id}/summary
- POST /container-logs/session/summary
- GET /container-logs/statistics
- GET /container-logs/analysis
- POST /container-logs/export (json/csv)

## 蜜罐/日志/规则 (节选)
- GET /honeypot/instances, GET /honeypot/logs
- GET /rules, POST /rules, GET /rules/logs
- GET /cowrie/logs, GET /headling/logs

---

# 常用示例 (PowerShell)

## 上传扫描文件
$resp = Invoke-RestMethod -Uri "http://localhost:9090/api/v1/malware/scan/file" -Method Post -Form @{ file = Get-Item .\sample.bin }
$resp | ConvertTo-Json -Depth 5

## 扫描远程URL
Invoke-RestMethod -Uri "http://localhost:9090/api/v1/malware/scan/url" -Method Post -Body '{"url":"https://example.com/mal.bin"}' -ContentType 'application/json' | ConvertTo-Json -Depth 5

## 获取IOC
Invoke-RestMethod -Uri "http://localhost:9090/api/v1/threat/intelligence?type=ip&value=1.2.3.4" | ConvertTo-Json -Depth 5

## 创建临时容器
$body = @{ name='demo'; honeypot_name='demo'; image_name='hello-world:latest'; protocol='http'; interface_type='network'; port_mappings=@{ '80'='auto' }; environment=@{ 'TEST'='true' } } | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:9090/api/v1/temp-containers" -Method Post -Body $body -ContentType 'application/json' | ConvertTo-Json -Depth 5

## 蜜签触发
Invoke-RestMethod -Uri "http://localhost:9090/api/v1/honeytokens/1/trigger" -Method Post -Body '{"action":"open","details":"click"}' -ContentType 'application/json'

---

# 前端测试页面
- frontend/debug-container.html 已修复部分过期端点
- 新增 frontend/api-playground.html 可直接测试：恶意软件、威胁情报、蜜签、Docker 与临时容器

# 注意
- 统一返回格式见上。
- DB_MODE=dameng 时，布尔过滤采用 0/1，已在服务层做好兼容与回退。
