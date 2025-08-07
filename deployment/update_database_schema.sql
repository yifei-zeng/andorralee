-- 数据库结构更新脚本
-- 用于将现有数据库结构更新为支持Docker镜像管理和日志分析的完整结构

-- 1. 更新现有表结构，统一字段类型和约束

-- 更新 security_rule 表
ALTER TABLE `security_rule` 
MODIFY COLUMN `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
MODIFY COLUMN `create_time` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
MODIFY COLUMN `update_time` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3);

-- 更新 rule_log 表
ALTER TABLE `rule_log` 
MODIFY COLUMN `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
MODIFY COLUMN `rule_id` BIGINT UNSIGNED NOT NULL,
MODIFY COLUMN `log_time` DATETIME(3) NOT NULL,
MODIFY COLUMN `create_time` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3);

-- 更新 honeypot_template 表
ALTER TABLE `honeypot_template` 
MODIFY COLUMN `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
MODIFY COLUMN `import_time` DATETIME(3) NOT NULL,
MODIFY COLUMN `create_time` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
MODIFY COLUMN `update_time` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3);

-- 更新 honeypot_instance 表
ALTER TABLE `honeypot_instance`
MODIFY COLUMN `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
MODIFY COLUMN `template_id` BIGINT UNSIGNED DEFAULT NULL,
MODIFY COLUMN `create_time` DATETIME(3) NOT NULL,
MODIFY COLUMN `update_time` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
ADD COLUMN IF NOT EXISTS `honeypot_name` VARCHAR(100) NOT NULL COMMENT '蜜罐名称' AFTER `name`,
ADD COLUMN IF NOT EXISTS `container_id` VARCHAR(64) COMMENT 'Docker容器ID' AFTER `container_name`,
ADD COLUMN IF NOT EXISTS `honeypot_ip` VARCHAR(45) COMMENT '蜜罐IP地址' AFTER `ip`,
ADD COLUMN IF NOT EXISTS `interface_type` VARCHAR(50) COMMENT '蜜罐接口类型' AFTER `protocol`,
ADD COLUMN IF NOT EXISTS `image_name` VARCHAR(200) COMMENT 'Docker镜像名称' AFTER `status`,
ADD COLUMN IF NOT EXISTS `image_id` VARCHAR(64) COMMENT 'Docker镜像ID' AFTER `image_name`,
ADD COLUMN IF NOT EXISTS `port_mappings` JSON COMMENT '端口映射配置' AFTER `image_id`,
ADD COLUMN IF NOT EXISTS `environment` JSON COMMENT '环境变量配置' AFTER `port_mappings`,
MODIFY COLUMN `ip` VARCHAR(45) NOT NULL COMMENT 'IP地址',
MODIFY COLUMN `status` VARCHAR(20) NOT NULL DEFAULT 'created' COMMENT '部署状态';

-- 添加索引
ALTER TABLE `honeypot_instance`
ADD INDEX IF NOT EXISTS `idx_container_id` (`container_id`),
ADD INDEX IF NOT EXISTS `idx_image_id` (`image_id`),
ADD INDEX IF NOT EXISTS `idx_status` (`status`),
ADD INDEX IF NOT EXISTS `idx_honeypot_name` (`honeypot_name`);

-- 更新 honeypot_log 表
ALTER TABLE `honeypot_log` 
MODIFY COLUMN `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
MODIFY COLUMN `instance_id` BIGINT UNSIGNED NOT NULL,
MODIFY COLUMN `log_time` DATETIME(3) NOT NULL,
MODIFY COLUMN `create_time` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3);

-- 2. 创建新的Docker相关表（如果不存在）

-- 创建Docker镜像表
CREATE TABLE IF NOT EXISTS `docker_image` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `image_id` VARCHAR(64) NOT NULL COMMENT '镜像ID',
    `repository` VARCHAR(100) COMMENT '仓库名称',
    `tag` VARCHAR(50) COMMENT '标签',
    `digest` VARCHAR(100) COMMENT '摘要',
    `size` BIGINT COMMENT '镜像大小(字节)',
    `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_image_id` (`image_id`),
    INDEX `idx_repository` (`repository`),
    INDEX `idx_tag` (`tag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Docker镜像管理';

-- 创建Docker镜像操作日志表
CREATE TABLE IF NOT EXISTS `docker_image_log` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `image_id` VARCHAR(64) COMMENT '镜像ID',
    `image_name` VARCHAR(200) COMMENT '镜像名称(包含仓库和标签)',
    `operation` VARCHAR(20) NOT NULL COMMENT '操作类型(pull/delete/tag/inspect)',
    `details` TEXT COMMENT '操作详情',
    `status` VARCHAR(10) NOT NULL COMMENT '操作状态(success/failed)',
    `message` TEXT COMMENT '状态消息',
    `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    PRIMARY KEY (`id`),
    INDEX `idx_image_id` (`image_id`),
    INDEX `idx_operation` (`operation`),
    INDEX `idx_status` (`status`),
    INDEX `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Docker镜像操作日志';

-- 创建容器日志分析表
CREATE TABLE IF NOT EXISTS `container_log_segment` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `container_id` VARCHAR(64) NOT NULL COMMENT '容器ID',
    `container_name` VARCHAR(100) COMMENT '容器名称',
    `segment_type` VARCHAR(20) NOT NULL COMMENT '日志段类型(error/warning/info/debug)',
    `content` TEXT NOT NULL COMMENT '日志内容',
    `timestamp` DATETIME(3) COMMENT '日志时间戳',
    `line_number` INT COMMENT '行号',
    `component` VARCHAR(50) COMMENT '组件名称',
    `severity_level` VARCHAR(10) COMMENT '严重程度',
    `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '分析时间',
    PRIMARY KEY (`id`),
    INDEX `idx_container_id` (`container_id`),
    INDEX `idx_segment_type` (`segment_type`),
    INDEX `idx_timestamp` (`timestamp`),
    INDEX `idx_severity_level` (`severity_level`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='容器日志语义分析结果';

-- 创建容器管理表
CREATE TABLE IF NOT EXISTS `docker_container` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `container_id` VARCHAR(64) NOT NULL COMMENT 'Docker容器ID',
    `container_name` VARCHAR(100) NOT NULL COMMENT '容器名称',
    `image_id` VARCHAR(64) COMMENT '关联的镜像ID',
    `image_name` VARCHAR(200) COMMENT '镜像名称',
    `status` VARCHAR(20) COMMENT '容器状态(running/stopped/exited等)',
    `ports` JSON COMMENT '端口映射信息',
    `environment` JSON COMMENT '环境变量',
    `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_container_id` (`container_id`),
    INDEX `idx_container_name` (`container_name`),
    INDEX `idx_image_id` (`image_id`),
    INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Docker容器管理';

-- 3. 添加外键约束（如果需要）
-- 注意：由于Docker容器和镜像可能会被删除，这里使用软关联，不设置严格的外键约束

-- 4. 创建视图，方便查询
CREATE OR REPLACE VIEW `v_container_with_image` AS
SELECT 
    c.id,
    c.container_id,
    c.container_name,
    c.status,
    c.ports,
    c.environment,
    c.created_at as container_created_at,
    i.repository,
    i.tag,
    i.size as image_size,
    i.created_at as image_created_at
FROM docker_container c
LEFT JOIN docker_image i ON c.image_id = i.image_id;

-- 创建日志统计视图
CREATE OR REPLACE VIEW `v_log_statistics` AS
SELECT
    container_id,
    container_name,
    segment_type,
    COUNT(*) as log_count,
    MIN(timestamp) as first_log_time,
    MAX(timestamp) as last_log_time,
    MAX(created_at) as last_analysis_time
FROM container_log_segment
GROUP BY container_id, container_name, segment_type;

-- 创建headling认证日志表
CREATE TABLE IF NOT EXISTS `headling_auth_log` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `timestamp` DATETIME(6) NOT NULL COMMENT '捕获到认证行为的时间戳',
    `auth_id` VARCHAR(36) NOT NULL COMMENT '此次认证行为的唯一ID',
    `session_id` VARCHAR(36) NOT NULL COMMENT '所属会话ID',
    `source_ip` VARCHAR(45) NOT NULL COMMENT '攻击者IP',
    `source_port` INT UNSIGNED NOT NULL COMMENT '攻击者使用的端口',
    `destination_ip` VARCHAR(45) NOT NULL COMMENT '被攻击的蜜罐容器IP',
    `destination_port` INT UNSIGNED NOT NULL COMMENT '目标端口',
    `protocol` VARCHAR(20) NOT NULL COMMENT '使用的协议',
    `username` VARCHAR(255) NOT NULL COMMENT '攻击者输入的用户名',
    `password` VARCHAR(255) NOT NULL COMMENT '攻击者输入的密码',
    `password_hash` VARCHAR(255) COMMENT '密码hash值',
    `container_id` VARCHAR(64) COMMENT '关联的容器ID',
    `container_name` VARCHAR(100) COMMENT '容器名称',
    `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '记录创建时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_auth_id` (`auth_id`),
    INDEX `idx_timestamp` (`timestamp`),
    INDEX `idx_session_id` (`session_id`),
    INDEX `idx_source_ip` (`source_ip`),
    INDEX `idx_destination_ip` (`destination_ip`),
    INDEX `idx_protocol` (`protocol`),
    INDEX `idx_container_id` (`container_id`),
    INDEX `idx_username` (`username`),
    INDEX `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Headling认证日志';

-- 创建headling认证统计视图
CREATE OR REPLACE VIEW `v_headling_auth_statistics` AS
SELECT
    DATE(timestamp) as log_date,
    protocol,
    COUNT(*) as total_attempts,
    COUNT(DISTINCT source_ip) as unique_ips,
    COUNT(DISTINCT username) as unique_usernames,
    COUNT(DISTINCT session_id) as unique_sessions,
    MIN(timestamp) as first_attempt,
    MAX(timestamp) as last_attempt
FROM headling_auth_log
GROUP BY DATE(timestamp), protocol;

-- 创建攻击者IP统计视图
CREATE OR REPLACE VIEW `v_attacker_ip_statistics` AS
SELECT
    source_ip,
    COUNT(*) as total_attempts,
    COUNT(DISTINCT protocol) as protocols_used,
    COUNT(DISTINCT username) as usernames_tried,
    COUNT(DISTINCT destination_port) as ports_targeted,
    MIN(timestamp) as first_seen,
    MAX(timestamp) as last_seen,
    TIMESTAMPDIFF(MINUTE, MIN(timestamp), MAX(timestamp)) as attack_duration_minutes
FROM headling_auth_log
GROUP BY source_ip
ORDER BY total_attempts DESC;

-- 创建Cowrie蜜罐日志表
CREATE TABLE IF NOT EXISTS `cowrie_log` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `event_time` DATETIME(6) NOT NULL COMMENT '事件发生的精确时间戳',
    `auth_id` VARCHAR(36) NOT NULL COMMENT '认证行为的唯一ID',
    `session_id` VARCHAR(36) NOT NULL COMMENT '会话ID',
    `source_ip` VARCHAR(15) NOT NULL COMMENT '攻击者IP',
    `source_port` SMALLINT UNSIGNED NOT NULL COMMENT '攻击者使用的端口',
    `destination_ip` VARCHAR(15) NOT NULL COMMENT '蜜罐容器IP',
    `destination_port` SMALLINT UNSIGNED NOT NULL COMMENT '目标端口',
    `protocol` ENUM('http','ssh','telnet','ftp','smb','other') NOT NULL COMMENT '使用的协议类型',
    `client_info` VARCHAR(255) COMMENT '客户端信息',
    `fingerprint` VARCHAR(64) COMMENT '客户端指纹',
    `username` VARCHAR(255) COMMENT '攻击者输入的用户名',
    `password` VARCHAR(255) COMMENT '攻击者输入的密码',
    `password_hash` VARCHAR(255) COMMENT '密码哈希值',
    `command` TEXT COMMENT '攻击者执行的命令内容',
    `command_found` BOOLEAN COMMENT '命令是否被系统识别',
    `raw_log` TEXT NOT NULL COMMENT '原始日志内容',
    `container_id` VARCHAR(64) COMMENT '关联的容器ID',
    `container_name` VARCHAR(100) COMMENT '容器名称',
    `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '记录创建时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_auth_id` (`auth_id`),
    INDEX `idx_event_time` (`event_time`),
    INDEX `idx_session_id` (`session_id`),
    INDEX `idx_source_ip` (`source_ip`),
    INDEX `idx_destination_ip` (`destination_ip`),
    INDEX `idx_protocol` (`protocol`),
    INDEX `idx_container_id` (`container_id`),
    INDEX `idx_username` (`username`),
    INDEX `idx_command_found` (`command_found`),
    INDEX `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Cowrie蜜罐日志';

-- 创建Cowrie日志统计视图
CREATE OR REPLACE VIEW `v_cowrie_statistics` AS
SELECT
    DATE(event_time) as log_date,
    protocol,
    COUNT(*) as total_events,
    COUNT(DISTINCT source_ip) as unique_ips,
    COUNT(DISTINCT session_id) as unique_sessions,
    COUNT(CASE WHEN username IS NOT NULL THEN 1 END) as auth_attempts,
    COUNT(CASE WHEN command IS NOT NULL THEN 1 END) as command_attempts,
    COUNT(CASE WHEN command_found = TRUE THEN 1 END) as valid_commands,
    MIN(event_time) as first_event,
    MAX(event_time) as last_event
FROM cowrie_log
GROUP BY DATE(event_time), protocol;

-- 创建Cowrie攻击者行为统计视图
CREATE OR REPLACE VIEW `v_cowrie_attacker_behavior` AS
SELECT
    source_ip,
    COUNT(*) as total_events,
    COUNT(DISTINCT protocol) as protocols_used,
    COUNT(DISTINCT session_id) as sessions_created,
    COUNT(CASE WHEN username IS NOT NULL THEN 1 END) as auth_attempts,
    COUNT(CASE WHEN command IS NOT NULL THEN 1 END) as commands_executed,
    COUNT(CASE WHEN command_found = TRUE THEN 1 END) as valid_commands,
    COUNT(DISTINCT username) as usernames_tried,
    COUNT(DISTINCT fingerprint) as unique_fingerprints,
    MIN(event_time) as first_seen,
    MAX(event_time) as last_seen,
    TIMESTAMPDIFF(MINUTE, MIN(event_time), MAX(event_time)) as activity_duration_minutes
FROM cowrie_log
GROUP BY source_ip
ORDER BY total_events DESC;

-- 创建Cowrie命令统计视图
CREATE OR REPLACE VIEW `v_cowrie_command_statistics` AS
SELECT
    command,
    COUNT(*) as usage_count,
    COUNT(DISTINCT source_ip) as unique_ips,
    COUNT(DISTINCT session_id) as unique_sessions,
    command_found,
    MIN(event_time) as first_used,
    MAX(event_time) as last_used
FROM cowrie_log
WHERE command IS NOT NULL AND command != ''
GROUP BY command, command_found
ORDER BY usage_count DESC;
