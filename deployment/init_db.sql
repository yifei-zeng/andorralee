-- Andorralee: All-in-one idempotent initialization script
-- 文件位置: deployment/init_db.sql
-- 目的: 初始化 Andorralee 系统所需的数据库和表结构。
-- 注意：此脚本会创建数据库（如果不存在）和所有必要的表。

-- 配置：数据库名为 andorralee
SET NAMES utf8mb4;
SET @DBNAME = 'andorralee';

-- 创建数据库（含字符集）
CREATE DATABASE IF NOT EXISTS `andorralee` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `andorralee`;

-- =====================================================
-- 基础表：honeypot_template / honeypot_instance / honeypot_log / bait / security_rule / rule_log
-- =====================================================

CREATE TABLE IF NOT EXISTS honeypot_template (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL COMMENT '模板名称',
    protocol VARCHAR(20) NOT NULL COMMENT '协议类型(SSH/HTTP/FTP/MySQL等)',
    import_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '导入时间',
    deploy_count INT DEFAULT 0 COMMENT '已部署数量'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='蜜罐模板管理';

CREATE TABLE IF NOT EXISTS honeypot_instance (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL COMMENT '实例名称',
    container_name VARCHAR(100) COMMENT 'Docker容器名称',
    ip VARCHAR(45) COMMENT 'IP地址',
    port INT COMMENT '端口号',
    protocol VARCHAR(50) COMMENT '协议类型',
    status VARCHAR(20) NOT NULL DEFAULT 'created' COMMENT '部署状态',
    create_time DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    template_id BIGINT UNSIGNED DEFAULT NULL,
    honeypot_name VARCHAR(100) COMMENT '蜜罐名称',
    container_id VARCHAR(100) COMMENT 'Docker容器ID',
    honeypot_ip VARCHAR(45) COMMENT '蜜罐IP地址',
    interface_type VARCHAR(50) COMMENT '蜜罐接口类型',
    image_name VARCHAR(200) COMMENT 'Docker镜像名称',
    image_id VARCHAR(100) COMMENT 'Docker镜像ID',
    port_mappings JSON COMMENT '端口映射配置',
    environment JSON COMMENT '环境变量配置',
    FOREIGN KEY (template_id) REFERENCES honeypot_template(id) ON DELETE SET NULL,
    INDEX idx_container_id (container_id),
    INDEX idx_image_id (image_id),
    INDEX idx_status (status),
    INDEX idx_honeypot_name (honeypot_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='蜜罐实例管理';

CREATE TABLE IF NOT EXISTS honeypot_log (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    instance_id BIGINT UNSIGNED NOT NULL COMMENT '关联的蜜罐实例ID',
    log_type VARCHAR(20) COMMENT '日志类型(warning/info/error)',
    content TEXT COMMENT '日志内容',
    log_time DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '记录时间',
    FOREIGN KEY (instance_id) REFERENCES honeypot_instance(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='蜜罐运行日志';

CREATE TABLE IF NOT EXISTS bait (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL COMMENT '诱饵名称',
    file_type VARCHAR(10) COMMENT '文件类型(TXT/PDF/DOCX等)',
    is_deployed TINYINT DEFAULT 0 COMMENT '是否已部署(0-未部署,1-已部署)',
    create_time DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    instance_id BIGINT UNSIGNED COMMENT '部署的蜜罐实例ID',
    FOREIGN KEY (instance_id) REFERENCES honeypot_instance(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='诱饵管理';

CREATE TABLE IF NOT EXISTS security_rule (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    rule_name VARCHAR(100) NOT NULL COMMENT '规则名称',
    trigger_conditions TEXT COMMENT '触发条件',
    actions TEXT COMMENT '执行动作',
    is_enabled TINYINT DEFAULT 1 COMMENT '是否启用(0-禁用,1-启用)',
    create_time DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='安全规则管理';

CREATE TABLE IF NOT EXISTS rule_log (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    rule_id BIGINT UNSIGNED NOT NULL COMMENT '关联的规则ID',
    rule_name VARCHAR(100) COMMENT '规则名称',
    content TEXT COMMENT '执行内容',
    log_time DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '执行时间',
    FOREIGN KEY (rule_id) REFERENCES security_rule(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='规则执行日志';

-- =====================================================
-- Docker / 容器 / 日志 表与视图
-- =====================================================

CREATE TABLE IF NOT EXISTS docker_image (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    image_id VARCHAR(100) NOT NULL COMMENT '镜像ID',
    repository VARCHAR(100) COMMENT '仓库名称',
    tag VARCHAR(50) COMMENT '标签',
    digest VARCHAR(100) COMMENT '摘要',
    size BIGINT COMMENT '镜像大小(字节)',
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_image_id (image_id),
    INDEX idx_repository (repository),
    INDEX idx_tag (tag)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Docker镜像管理';

CREATE TABLE IF NOT EXISTS docker_image_log (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    image_id VARCHAR(100) COMMENT '镜像ID',
    image_name VARCHAR(200) COMMENT '镜像名称(包含仓库和标签)',
    operation VARCHAR(20) NOT NULL COMMENT '操作类型(pull/delete/tag/inspect)',
    details TEXT COMMENT '操作详情',
    status VARCHAR(10) NOT NULL COMMENT '操作状态(success/failed)',
    message TEXT COMMENT '状态消息',
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    PRIMARY KEY (id),
    INDEX idx_image_id (image_id),
    INDEX idx_operation (operation),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Docker镜像操作日志';

CREATE TABLE IF NOT EXISTS container_log_segment (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    container_id VARCHAR(100) NOT NULL COMMENT '容器ID',
    container_name VARCHAR(100) COMMENT '容器名称',
    segment_type VARCHAR(20) NOT NULL COMMENT '日志段类型(error/warning/info/debug)',
    content TEXT NOT NULL COMMENT '日志内容',
    timestamp DATETIME(3) COMMENT '日志时间戳',
    line_number INT COMMENT '行号',
    component VARCHAR(50) COMMENT '组件名称',
    severity_level VARCHAR(10) COMMENT '严重程度',
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '分析时间',
    PRIMARY KEY (id),
    INDEX idx_container_id (container_id),
    INDEX idx_segment_type (segment_type),
    INDEX idx_timestamp (timestamp),
    INDEX idx_severity_level (severity_level)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='容器日志语义分析结果';

CREATE TABLE IF NOT EXISTS docker_container (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    container_id VARCHAR(100) NOT NULL COMMENT 'Docker容器ID',
    container_name VARCHAR(100) NOT NULL COMMENT '容器名称',
    image_id VARCHAR(100) COMMENT '关联的镜像ID',
    image_name VARCHAR(200) COMMENT '镜像名称',
    status VARCHAR(20) COMMENT '容器状态(running/stopped/exited等)',
    ports JSON COMMENT '端口映射信息',
    environment JSON COMMENT '环境变量',
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_container_id (container_id),
    INDEX idx_container_name (container_name),
    INDEX idx_image_id (image_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Docker容器管理';

-- =====================================================
-- Heralding / Cowrie 表与视图
-- =====================================================

CREATE TABLE IF NOT EXISTS heralding_auth_log (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `timestamp` DATETIME(6) NOT NULL COMMENT '捕获到认证行为的时间戳',
    auth_id VARCHAR(36) NOT NULL COMMENT '此次认证行为的唯一ID',
    session_id VARCHAR(36) NOT NULL COMMENT '所属会话ID',
    source_ip VARCHAR(45) NOT NULL COMMENT '攻击者IP',
    source_port INT UNSIGNED NOT NULL COMMENT '攻击者使用的端口',
    destination_ip VARCHAR(45) NOT NULL COMMENT '被攻击的蜜罐容器IP',
    destination_port INT UNSIGNED NOT NULL COMMENT '目标端口',
    protocol VARCHAR(20) NOT NULL COMMENT '使用的协议',
    username VARCHAR(255) NOT NULL COMMENT '攻击者输入的用户名',
    password VARCHAR(255) NOT NULL COMMENT '攻击者输入的密码',
    password_hash VARCHAR(255) COMMENT '密码hash值',
    container_id VARCHAR(100) COMMENT '关联的容器ID',
    container_name VARCHAR(100) COMMENT '容器名称',
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '记录创建时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_auth_id (auth_id),
    INDEX idx_timestamp (`timestamp`),
    INDEX idx_session_id (session_id),
    INDEX idx_source_ip (source_ip),
    INDEX idx_destination_ip (destination_ip),
    INDEX idx_protocol (protocol),
    INDEX idx_container_id (container_id),
    INDEX idx_username (username),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Heralding认证日志';

CREATE TABLE IF NOT EXISTS heralding_session_log (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `timestamp` DATETIME(6) NOT NULL COMMENT '会话开始时间',
    duration BIGINT COMMENT '会话持续时间(毫秒)',
    session_id VARCHAR(36) NOT NULL COMMENT '会话ID',
    source_ip VARCHAR(45) NOT NULL COMMENT '源IP',
    source_port INT UNSIGNED NOT NULL COMMENT '源端口',
    destination_ip VARCHAR(45) NOT NULL COMMENT '目标IP',
    destination_port INT UNSIGNED NOT NULL COMMENT '目标端口',
    protocol VARCHAR(20) NOT NULL COMMENT '协议',
    num_auth_attempts INT DEFAULT 0 COMMENT '认证尝试次数',
    container_id VARCHAR(64) COMMENT '关联容器ID',
    container_name VARCHAR(100) COMMENT '容器名称',
    PRIMARY KEY (id),
    UNIQUE KEY uk_session_id (session_id),
    INDEX idx_timestamp (`timestamp`),
    INDEX idx_source_ip (source_ip),
    INDEX idx_container_id (container_id),
    INDEX idx_protocol (protocol)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Heralding会话日志';

CREATE TABLE IF NOT EXISTS cowrie_log (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    event_time DATETIME(6) NOT NULL COMMENT '事件发生的精确时间戳',
    auth_id VARCHAR(36) NOT NULL COMMENT '认证行为的唯一ID',
    session_id VARCHAR(36) NOT NULL COMMENT '会话ID',
    source_ip VARCHAR(45) NOT NULL COMMENT '攻击者IP',
    source_port SMALLINT UNSIGNED NOT NULL COMMENT '攻击者使用的端口',
    destination_ip VARCHAR(45) NOT NULL COMMENT '蜜罐容器IP',
    destination_port SMALLINT UNSIGNED NOT NULL COMMENT '目标端口',
    protocol VARCHAR(20) NOT NULL COMMENT '使用的协议类型',
    client_info VARCHAR(255) COMMENT '客户端信息',
    fingerprint VARCHAR(64) COMMENT '客户端指纹',
    username VARCHAR(255) COMMENT '攻击者输入的用户名',
    password VARCHAR(255) COMMENT '攻击者输入的密码',
    password_hash VARCHAR(255) COMMENT '密码哈希值',
    command TEXT COMMENT '攻击者执行的命令内容',
    command_found BOOLEAN COMMENT '命令是否被系统识别',
    raw_log TEXT NOT NULL COMMENT '原始日志内容',
    container_id VARCHAR(100) COMMENT '关联的容器ID',
    container_name VARCHAR(100) COMMENT '容器名称',
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '记录创建时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_auth_id_cowrie (auth_id),
    INDEX idx_event_time (event_time),
    INDEX idx_session_id (session_id),
    INDEX idx_source_ip (source_ip),
    INDEX idx_destination_ip (destination_ip),
    INDEX idx_protocol (protocol),
    INDEX idx_container_id (container_id),
    INDEX idx_username (username),
    INDEX idx_command_found (command_found),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Cowrie蜜罐日志';

-- =====================================================
-- MySQL 蜜罐日志表
-- =====================================================

CREATE TABLE IF NOT EXISTS mysql_honeypot_log (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    event_id VARCHAR(64) COMMENT '事件唯一ID',
    event_time DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '事件时间',
    container_id VARCHAR(100) COMMENT '容器ID',
    container_name VARCHAR(100) COMMENT '容器名称',
    source_ip VARCHAR(45) COMMENT '攻击源IP',
    source_port INT UNSIGNED COMMENT '攻击源端口',
    destination_ip VARCHAR(45) COMMENT '目标IP',
    destination_port INT UNSIGNED COMMENT '目标端口',
    username VARCHAR(255) COMMENT '登录用户名',
    password VARCHAR(255) COMMENT '登录密码',
    database_name VARCHAR(255) COMMENT '访问的数据库',
    query TEXT COMMENT '执行的SQL语句',
    error_code VARCHAR(50) COMMENT '错误码',
    raw_log TEXT COMMENT '原始日志',
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '记录创建时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_mysql_honeypot_event (event_id),
    INDEX idx_mysql_honeypot_event_time (event_time),
    INDEX idx_mysql_honeypot_container (container_id),
    INDEX idx_mysql_honeypot_source_ip (source_ip),
    INDEX idx_mysql_honeypot_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='MySQL蜜罐日志';

-- =====================================================
-- 视图（创建或替换）
-- =====================================================

CREATE OR REPLACE VIEW v_container_with_image AS
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

CREATE OR REPLACE VIEW v_log_statistics AS
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

CREATE OR REPLACE VIEW v_heralding_auth_statistics AS
SELECT
    DATE(`timestamp`) as log_date,
    protocol,
    COUNT(*) as total_attempts,
    COUNT(DISTINCT source_ip) as unique_ips,
    COUNT(DISTINCT username) as unique_usernames,
    COUNT(DISTINCT session_id) as unique_sessions,
    MIN(`timestamp`) as first_attempt,
    MAX(`timestamp`) as last_attempt
FROM heralding_auth_log
GROUP BY DATE(`timestamp`), protocol;

CREATE OR REPLACE VIEW v_attacker_ip_statistics AS
SELECT
    source_ip,
    COUNT(*) as total_attempts,
    COUNT(DISTINCT protocol) as protocols_used,
    COUNT(DISTINCT username) as usernames_tried,
    COUNT(DISTINCT destination_port) as ports_targeted,
    MIN(`timestamp`) as first_seen,
    MAX(`timestamp`) as last_seen,
    TIMESTAMPDIFF(MINUTE, MIN(`timestamp`), MAX(`timestamp`)) as attack_duration_minutes
FROM heralding_auth_log
GROUP BY source_ip
ORDER BY total_attempts DESC;

CREATE OR REPLACE VIEW v_cowrie_statistics AS
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

CREATE OR REPLACE VIEW v_cowrie_attacker_behavior AS
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

CREATE OR REPLACE VIEW v_cowrie_command_statistics AS
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

-- =====================================================
-- 完成提示
-- =====================================================
SELECT 'init_db: 数据库初始化完成' AS result_message;
