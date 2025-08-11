# 达梦数据库结构与访问参考 (Dameng DB Reference)

本文档用于为其他项目组提供 Andorralee 达梦数据库 (DB_MODE=dameng) 的表结构、关键字段语义、典型查询及注意事项。

## 1. 连接信息
- 驱动: go (gorm.io) / JDBC 参考: dm.jdbc.driver.DmDriver
- 默认端口: 5236
- 示例 DSN (Go/GORM): dm://USER:PASSWORD@HOST:5236?schema=DOCKER_OPS
- 字符集: UTF-8

## 2. 通用约定
| 类型 | 说明 |
|------|------|
| 主键 | BIGINT 自增 (IDENTITY) |
| 布尔 | TINYINT/SMALLINT (0/1) |
| 长文本 | CLOB |
| JSON 逻辑 | 字符串存储 (CLOB/VARCHAR) 由应用层解析 |
| 时间 | TIMESTAMP(6) / DATETIME(6) 精度到微秒 |

## 3. 核心表与字段
### 3.1 恶意软件与检测
malware_signature(id,name,pattern,type,severity,is_active,description,create_time,update_time)
scan_result(id,file_hash,md5_hash,file_name,file_size,scan_time,is_infected,threat_level,detection_count,scan_duration_ms,source_ip,user_agent)
detection_result(id,scan_result_id,signature_id,signature_name,match_type,match_content,severity,created_at)

### 3.2 攻击会话与事件
attack_session(id,session_id,source_ip,source_port,destination_ip,destination_port,protocol,start_time,end_time,duration,status,auth_attempts,command_count,threat_level,container_id,container_name,user_agent,fingerprint)
attack_event(id,session_id,event_type,event_time,source_ip,protocol,username,password,command,payload,result,threat_level,is_blocked,raw_data)

### 3.3 威胁情报与蜜签
threat_intelligence(id,indicator_type,indicator_value,threat_type,confidence,severity,source,description,first_seen,last_seen,is_active,tags,created_at,updated_at)
honeytoken_event(id,token_id,token_type,token_name,trigger_time,source_ip,user_agent,request_path,request_data,response_code,location,description,threat_level,is_processed)

### 3.4 蜜罐实例与日志
honeypot_instance(id,name,honeypot_name,container_name,container_id,ip,honeypot_ip,port,protocol,interface_type,status,image_name,image_id,port_mappings,environment,create_time,update_time,description)
honeypot_log(id,instance_id,log_type,content,log_time)

### 3.5 规则与诱饵
security_rule(id,rule_name,trigger_conditions,actions,is_enabled)
rule_log(id,rule_id,rule_name,content,log_time)
bait(id,name,file_type,is_deployed,create_time,instance_id)

### 3.6 容器与镜像
docker_container(id,container_id,container_name,image_id,image_name,status,ports,environment,created_at,updated_at)
docker_image(id,image_id,repository,tag,digest,size,created_at,updated_at)
docker_image_log(id,image_id,image_name,operation,details,status,message,created_at)
container_log_segment(id,container_id,container_name,segment_type,content,timestamp,line_number,component,severity_level,created_at)

### 3.7 蜜罐协议日志
headling_auth_log(id,timestamp,auth_id,session_id,source_ip,source_port,destination_ip,destination_port,protocol,username,password,password_hash,container_id,container_name,created_at)
cowrie_log(id,event_time,auth_id,session_id,source_ip,source_port,destination_ip,destination_port,protocol,client_info,fingerprint,username,password,password_hash,command,command_found,raw_log,container_id,container_name,created_at)

## 4. 索引建议 (若未显式创建)
- *auth_id*, *session_id*, *source_ip*, *indicator_value*, *file_hash* 建议创建 BTREE 索引。
- 高频时间范围查询字段 (event_time, timestamp, scan_time, start_time) 可建立组合索引 (字段 + source_ip)。

## 5. 示例查询
获取激活的威胁情报：
SELECT id, indicator_type, indicator_value, severity FROM threat_intelligence WHERE is_active=1 ORDER BY updated_at DESC FETCH FIRST 50 ROWS ONLY;

按哈希聚合文件扫描结果：
SELECT file_hash, COUNT(*) AS total, MAX(scan_time) AS last_scan FROM scan_result GROUP BY file_hash ORDER BY last_scan DESC;

最近活跃攻击会话：
SELECT session_id, source_ip, start_time, status FROM attack_session WHERE status='active' ORDER BY start_time DESC FETCH FIRST 20 ROWS ONLY;

提取最近 24h Cowrie 命令：
SELECT command, COUNT(*) cnt FROM cowrie_log WHERE event_time > (SYSDATE - 1) AND command IS NOT NULL GROUP BY command ORDER BY cnt DESC FETCH FIRST 30 ROWS ONLY;

## 6. 迁移与兼容注意
1. is_active / is_blocked / is_processed 均为 0/1 数值列；应用侧可能存在回退逻辑（当列缺失时内存过滤）。
2. TEXT -> CLOB；若使用外部 ETL 需注意流式读取。
3. JSON 逻辑字段无需 DB 级别校验，写入前确保序列化。
4. 大批量写入建议使用分批事务 (500~1000 行/批)。

## 7. 权限最小化建议
- 运行账户仅授予 CONNECT, RESOURCE, 对业务 schema 的 SELECT/INSERT/UPDATE/DELETE。
- 禁止对系统库的 DROP / ALTER 权限。

## 8. 监控指标 (建议)
| 指标 | 描述 |
|------|------|
| active_sessions | 攻击会话活跃数量 |
| malware_signatures_total | 恶意特征总数 |
| threat_indicators_active | 激活威胁情报指标数量 |
| honeytoken_triggers_24h | 最近 24h 蜜签触发次数 |
| cowrie_commands_5m | 5分钟内命令事件数 |

## 9. 故障排查
| 现象 | 可能原因 | 排查步骤 |
|------|----------|----------|
| 查询 500 且日志含 invalid column IS_ACTIVE | 列未正确迁移或大小写差异 | 检查 dameng_migration.sql 与实际表结构；执行 DESCRIBE threat_intelligence; |
| 性能慢 | 缺少索引 / 大范围全表扫描 | 查看 V$SQL，添加必要索引 |
| 大字段写入失败 | CLOB 空间/权限限制 | 检查表空间剩余 / 账户权限 |

## 10. 变更记录
- 2025-08-12: 首版发布，涵盖全部核心表 & 查询示例。

---
维护团队: Andorralee 开发团队
