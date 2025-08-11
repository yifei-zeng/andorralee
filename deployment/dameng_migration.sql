-- Dameng 数据库初始化与迁移脚本
-- 该脚本综合了原 create_tables.sql 与 db_migration.sql 的结构需求
-- 语法适配：
-- * AUTO_INCREMENT  ->  IDENTITY
-- * (BIG)INT UNSIGNED -> BIGINT (去掉 UNSIGNED)
-- * TINYINT -> SMALLINT
-- * DATETIME / DATETIME(3) -> TIMESTAMP / TIMESTAMP(3)
-- * TEXT -> CLOB
-- * NOW() / CURRENT_TIMESTAMP(3) -> CURRENT_TIMESTAMP
-- * 去除 ENGINE / CHARSET / COLLATE / ON UPDATE CURRENT_TIMESTAMP(3) MySQL 专有写法
-- * 索引用单独 CREATE INDEX / UNIQUE INDEX 方式

-- 注意：Dameng 8 支持 IF NOT EXISTS

CREATE TABLE IF NOT EXISTS honeypot_template (
    id BIGINT IDENTITY PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    protocol VARCHAR(20) NOT NULL,
    import_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deploy_count INT DEFAULT 0
);

CREATE TABLE IF NOT EXISTS honeypot_instance (
    id BIGINT IDENTITY PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    container_name VARCHAR(50),
    ip VARCHAR(45),
    port INT,
    protocol VARCHAR(20),
    status VARCHAR(20),
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    template_id BIGINT,
    CONSTRAINT fk_instance_template FOREIGN KEY (template_id) REFERENCES honeypot_template(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS honeypot_log (
    id BIGINT IDENTITY PRIMARY KEY,
    instance_id BIGINT NOT NULL,
    log_type VARCHAR(20),
    content CLOB,
    log_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_log_instance FOREIGN KEY (instance_id) REFERENCES honeypot_instance(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS bait (
    id BIGINT IDENTITY PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    file_type VARCHAR(10),
    is_deployed SMALLINT DEFAULT 0,
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    instance_id BIGINT,
    CONSTRAINT fk_bait_instance FOREIGN KEY (instance_id) REFERENCES honeypot_instance(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS security_rule (
    id BIGINT IDENTITY PRIMARY KEY,
    rule_name VARCHAR(50) NOT NULL,
    trigger_conditions CLOB,
    actions CLOB,
    is_enabled SMALLINT DEFAULT 1
);

CREATE TABLE IF NOT EXISTS rule_log (
    id BIGINT IDENTITY PRIMARY KEY,
    rule_id BIGINT NOT NULL,
    rule_name VARCHAR(50),
    content CLOB,
    log_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_rulelog_rule FOREIGN KEY (rule_id) REFERENCES security_rule(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS docker_image (
    id BIGINT IDENTITY PRIMARY KEY,
    image_id VARCHAR(64) NOT NULL,
    repository VARCHAR(100),
    tag VARCHAR(50),
    digest VARCHAR(100),
    size BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_docker_image_image_id UNIQUE (image_id)
);

CREATE TABLE IF NOT EXISTS docker_image_log (
    id BIGINT IDENTITY PRIMARY KEY,
    image_id VARCHAR(64),
    image_name VARCHAR(200),
    operation VARCHAR(20) NOT NULL,
    details CLOB,
    status VARCHAR(20) NOT NULL,
    message CLOB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS container_log_segment (
    id BIGINT IDENTITY PRIMARY KEY,
    container_id VARCHAR(64) NOT NULL,
    container_name VARCHAR(100),
    segment_type VARCHAR(20) NOT NULL,
    content CLOB NOT NULL,
    timestamp TIMESTAMP,
    line_number INT,
    component VARCHAR(50),
    severity_level VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ======================== 追加：缺失的业务/安全/日志/情报/病毒检测等表 ========================

-- Docker 容器表
CREATE TABLE IF NOT EXISTS docker_container (
    id BIGINT IDENTITY PRIMARY KEY,
    container_id VARCHAR(64) NOT NULL,
    container_name VARCHAR(100) NOT NULL,
    image_id VARCHAR(100),
    image_name VARCHAR(200),
    status VARCHAR(20),
    ports CLOB,
    environment CLOB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(container_id)
);

-- Headling 认证日志
CREATE TABLE IF NOT EXISTS headling_auth_log (
    id BIGINT IDENTITY PRIMARY KEY,
    timestamp TIMESTAMP NOT NULL,
    auth_id VARCHAR(36) NOT NULL,
    session_id VARCHAR(36) NOT NULL,
    source_ip VARCHAR(45) NOT NULL,
    source_port INT NOT NULL,
    destination_ip VARCHAR(45) NOT NULL,
    destination_port INT NOT NULL,
    protocol VARCHAR(20) NOT NULL,
    username VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255),
    container_id VARCHAR(64),
    container_name VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(auth_id)
);
CREATE INDEX IF NOT EXISTS idx_headling_session ON headling_auth_log(session_id);
CREATE INDEX IF NOT EXISTS idx_headling_srcip ON headling_auth_log(source_ip);
CREATE INDEX IF NOT EXISTS idx_headling_protocol ON headling_auth_log(protocol);
CREATE INDEX IF NOT EXISTS idx_headling_container ON headling_auth_log(container_id);

-- Cowrie 日志
CREATE TABLE IF NOT EXISTS cowrie_log (
    id BIGINT IDENTITY PRIMARY KEY,
    event_time TIMESTAMP NOT NULL,
    auth_id VARCHAR(36) NOT NULL,
    session_id VARCHAR(36) NOT NULL,
    source_ip VARCHAR(15) NOT NULL,
    source_port INT NOT NULL,
    destination_ip VARCHAR(15) NOT NULL,
    destination_port INT NOT NULL,
    protocol VARCHAR(20) NOT NULL,
    client_info VARCHAR(255),
    fingerprint VARCHAR(64),
    username VARCHAR(255),
    password VARCHAR(255),
    password_hash VARCHAR(255),
    command CLOB,
    command_found SMALLINT,
    raw_log CLOB NOT NULL,
    container_id VARCHAR(64),
    container_name VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(auth_id)
);
CREATE INDEX IF NOT EXISTS idx_cowrie_session ON cowrie_log(session_id);
CREATE INDEX IF NOT EXISTS idx_cowrie_srcip ON cowrie_log(source_ip);
CREATE INDEX IF NOT EXISTS idx_cowrie_protocol ON cowrie_log(protocol);
CREATE INDEX IF NOT EXISTS idx_cowrie_container ON cowrie_log(container_id);

-- 病毒特征
CREATE TABLE IF NOT EXISTS malware_signature (
    id BIGINT IDENTITY PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    pattern CLOB NOT NULL,
    type VARCHAR(20) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    description CLOB,
    create_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMP,
    is_active SMALLINT DEFAULT 1
);

-- 扫描结果
CREATE TABLE IF NOT EXISTS scan_result (
    id BIGINT IDENTITY PRIMARY KEY,
    file_hash VARCHAR(64) NOT NULL,
    md5_hash VARCHAR(32),
    file_name VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL,
    scan_time TIMESTAMP NOT NULL,
    is_infected SMALLINT NOT NULL,
    threat_level VARCHAR(20),
    detection_count INT DEFAULT 0,
    scan_duration_ms BIGINT,
    source_ip VARCHAR(45),
    user_agent CLOB
);
CREATE INDEX IF NOT EXISTS idx_scan_filehash ON scan_result(file_hash);

-- 检测结果
CREATE TABLE IF NOT EXISTS detection_result (
    id BIGINT IDENTITY PRIMARY KEY,
    scan_result_id BIGINT NOT NULL,
    signature_id BIGINT NOT NULL,
    signature_name VARCHAR(100),
    match_type VARCHAR(20),
    match_content CLOB,
    severity VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- 达梦不强制外键，视需要可加: ALTER TABLE detection_result ADD CONSTRAINT fk_detection_scan FOREIGN KEY (scan_result_id) REFERENCES scan_result(id);

-- 攻击会话
CREATE TABLE IF NOT EXISTS attack_session (
    id BIGINT IDENTITY PRIMARY KEY,
    session_id VARCHAR(36) NOT NULL,
    source_ip VARCHAR(45) NOT NULL,
    source_port INT,
    destination_ip VARCHAR(45) NOT NULL,
    destination_port INT NOT NULL,
    protocol VARCHAR(20) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    duration BIGINT,
    status VARCHAR(20) DEFAULT 'active',
    auth_attempts INT DEFAULT 0,
    command_count INT DEFAULT 0,
    threat_level VARCHAR(20),
    container_id VARCHAR(64),
    container_name VARCHAR(100),
    user_agent CLOB,
    fingerprint VARCHAR(64)
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_attack_session ON attack_session(session_id);
CREATE INDEX IF NOT EXISTS idx_attack_session_srcip ON attack_session(source_ip);

-- 攻击事件
CREATE TABLE IF NOT EXISTS attack_event (
    id BIGINT IDENTITY PRIMARY KEY,
    session_id VARCHAR(36) NOT NULL,
    event_type VARCHAR(30) NOT NULL,
    event_time TIMESTAMP NOT NULL,
    source_ip VARCHAR(45) NOT NULL,
    protocol VARCHAR(20) NOT NULL,
    username VARCHAR(255),
    password VARCHAR(255),
    command CLOB,
    payload CLOB,
    result CLOB,
    threat_level VARCHAR(20),
    is_blocked SMALLINT DEFAULT 0,
    raw_data CLOB
);
CREATE INDEX IF NOT EXISTS idx_attack_event_session ON attack_event(session_id);
CREATE INDEX IF NOT EXISTS idx_attack_event_srcip ON attack_event(source_ip);

-- 威胁情报
CREATE TABLE IF NOT EXISTS threat_intelligence (
    id BIGINT IDENTITY PRIMARY KEY,
    indicator_type VARCHAR(20) NOT NULL,
    indicator_value VARCHAR(255) NOT NULL,
    threat_type VARCHAR(50),
    confidence INT,
    severity VARCHAR(20),
    source VARCHAR(100),
    description CLOB,
    first_seen TIMESTAMP,
    last_seen TIMESTAMP,
    is_active SMALLINT DEFAULT 1,
    tags CLOB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_threat_indicator ON threat_intelligence(indicator_value);

-- 蜜签事件
CREATE TABLE IF NOT EXISTS honeytoken_event (
    id BIGINT IDENTITY PRIMARY KEY,
    token_id VARCHAR(36) NOT NULL,
    token_type VARCHAR(50) NOT NULL,
    token_name VARCHAR(100),
    trigger_time TIMESTAMP NOT NULL,
    source_ip VARCHAR(45) NOT NULL,
    user_agent CLOB,
    request_path VARCHAR(500),
    request_data CLOB,
    response_code INT,
    location VARCHAR(255),
    description CLOB,
    threat_level VARCHAR(20),
    is_processed SMALLINT DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_honeytoken_token ON honeytoken_event(token_id);
CREATE INDEX IF NOT EXISTS idx_honeytoken_srcip ON honeytoken_event(source_ip);

-- ==================================================================================
CREATE INDEX IF NOT EXISTS idx_cls_container_id ON container_log_segment(container_id);
CREATE INDEX IF NOT EXISTS idx_cls_segment_type ON container_log_segment(segment_type);
CREATE INDEX IF NOT EXISTS idx_cls_timestamp ON container_log_segment(timestamp);

-- 额外业务表（来自 AutoMigrate 模型）
-- 对于未在原始 SQL 中定义但在 GORM 模型中的表，可按需补充；此处留出扩展位
-- 例如：
-- CREATE TABLE IF NOT EXISTS malware_signature (...);
-- 根据实际模型定义补充

-- 完成
