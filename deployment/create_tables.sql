-- 创建蜜罐模板表
CREATE TABLE IF NOT EXISTS honeypot_template (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL COMMENT '模板名称',
    protocol VARCHAR(20) NOT NULL COMMENT '协议类型(SSH/HTTP/FTP/MySQL等)',
    import_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '导入时间',
    deploy_count INT DEFAULT 0 COMMENT '已部署数量'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='蜜罐模板管理';

-- 创建蜜罐实例表
CREATE TABLE IF NOT EXISTS honeypot_instance (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL COMMENT '实例名称',
    container_name VARCHAR(50) COMMENT 'Docker容器名称',
    ip VARCHAR(15) COMMENT 'IP地址',
    port INT COMMENT '端口号',
    protocol VARCHAR(20) COMMENT '协议类型',
    status VARCHAR(10) COMMENT '状态(running/stopped/failed)',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    template_id INT COMMENT '关联的模板ID',
    FOREIGN KEY (template_id) REFERENCES honeypot_template(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='蜜罐实例管理';

-- 创建蜜罐日志表
CREATE TABLE IF NOT EXISTS honeypot_log (
    id INT AUTO_INCREMENT PRIMARY KEY,
    instance_id INT NOT NULL COMMENT '关联的蜜罐实例ID',
    log_type VARCHAR(20) COMMENT '日志类型(warning/info/error)',
    content TEXT COMMENT '日志内容',
    log_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '记录时间',
    FOREIGN KEY (instance_id) REFERENCES honeypot_instance(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='蜜罐运行日志';

-- 创建诱饵(蜜签)表
CREATE TABLE IF NOT EXISTS bait (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL COMMENT '诱饵名称',
    file_type VARCHAR(10) COMMENT '文件类型(TXT/PDF/DOCX等)',
    is_deployed TINYINT DEFAULT 0 COMMENT '是否已部署(0-未部署,1-已部署)',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    instance_id INT COMMENT '部署的蜜罐实例ID',
    FOREIGN KEY (instance_id) REFERENCES honeypot_instance(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='诱饵管理';

-- 创建安全规则表
CREATE TABLE IF NOT EXISTS security_rule (
    id INT AUTO_INCREMENT PRIMARY KEY,
    rule_name VARCHAR(50) NOT NULL COMMENT '规则名称',
    trigger_conditions TEXT COMMENT '触发条件',
    actions TEXT COMMENT '执行动作',
    is_enabled TINYINT DEFAULT 1 COMMENT '是否启用(0-禁用,1-启用)'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='安全规则管理';

-- 创建规则执行日志表
CREATE TABLE IF NOT EXISTS rule_log (
    id INT AUTO_INCREMENT PRIMARY KEY,
    rule_id INT NOT NULL COMMENT '关联的规则ID',
    rule_name VARCHAR(50) COMMENT '规则名称',
    content TEXT COMMENT '执行内容',
    log_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '执行时间',
    FOREIGN KEY (rule_id) REFERENCES security_rule(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='规则执行日志';


-- 创建Docker镜像表
CREATE TABLE IF NOT EXISTS docker_image (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    image_id VARCHAR(64) NOT NULL COMMENT '镜像ID',
    repository VARCHAR(100) COMMENT '仓库名称',
    tag VARCHAR(50) COMMENT '标签',
    digest VARCHAR(100) COMMENT '摘要',
    size BIGINT COMMENT '镜像大小(字节)',
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    UNIQUE KEY (image_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Docker镜像管理';

-- 创建Docker镜像操作日志表
CREATE TABLE IF NOT EXISTS docker_image_log (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    image_id VARCHAR(64) COMMENT '镜像ID',
    image_name VARCHAR(200) COMMENT '镜像名称(包含仓库和标签)',
    operation VARCHAR(20) NOT NULL COMMENT '操作类型(pull/delete/tag/inspect)',
    details TEXT COMMENT '操作详情',
    status VARCHAR(10) NOT NULL COMMENT '操作状态(success/failed)',
    message TEXT COMMENT '状态消息',
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Docker镜像操作日志';

-- 创建容器日志分析表
CREATE TABLE IF NOT EXISTS container_log_segment (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    container_id VARCHAR(64) NOT NULL COMMENT '容器ID',
    container_name VARCHAR(100) COMMENT '容器名称',
    segment_type VARCHAR(20) NOT NULL COMMENT '日志段类型(error/warning/info/debug)',
    content TEXT NOT NULL COMMENT '日志内容',
    timestamp DATETIME(3) COMMENT '日志时间戳',
    line_number INT COMMENT '行号',
    component VARCHAR(50) COMMENT '组件名称',
    severity_level VARCHAR(10) COMMENT '严重程度',
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '分析时间',
    INDEX idx_container_id (container_id),
    INDEX idx_segment_type (segment_type),
    INDEX idx_timestamp (timestamp)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='容器日志语义分析结果';

-- ======================== 追加：缺失的业务/安全/日志/情报/病毒检测等表 ========================

-- Docker 运行容器表
CREATE TABLE IF NOT EXISTS docker_container (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    container_id VARCHAR(64) NOT NULL COMMENT 'Docker容器ID',
    container_name VARCHAR(100) NOT NULL COMMENT '容器名称',
    image_id VARCHAR(100) COMMENT '关联的镜像ID',
    image_name VARCHAR(200) COMMENT '镜像名称',
    status VARCHAR(20) COMMENT '容器状态',
    ports JSON COMMENT '端口映射信息',
    environment JSON COMMENT '环境变量',
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '创建时间',
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6) COMMENT '更新时间',
    UNIQUE KEY uq_docker_container_id (container_id),
    KEY idx_docker_container_name (container_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Docker容器管理';

-- Headling 认证日志表
CREATE TABLE IF NOT EXISTS headling_auth_log (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    timestamp DATETIME(6) NOT NULL COMMENT '捕获时间',
    auth_id VARCHAR(36) NOT NULL COMMENT '认证唯一ID',
    session_id VARCHAR(36) NOT NULL COMMENT '会话ID',
    source_ip VARCHAR(45) NOT NULL COMMENT '攻击者IP',
    source_port INT NOT NULL COMMENT '攻击者端口',
    destination_ip VARCHAR(45) NOT NULL COMMENT '目标IP',
    destination_port INT NOT NULL COMMENT '目标端口',
    protocol VARCHAR(20) NOT NULL COMMENT '协议',
    username VARCHAR(255) NOT NULL COMMENT '用户名',
    password VARCHAR(255) NOT NULL COMMENT '密码',
    password_hash VARCHAR(255) COMMENT '密码Hash',
    container_id VARCHAR(64) COMMENT '容器ID',
    container_name VARCHAR(100) COMMENT '容器名称',
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '记录创建时间',
    UNIQUE KEY uq_headling_auth_id (auth_id),
    KEY idx_headling_session (session_id),
    KEY idx_headling_srcip (source_ip),
    KEY idx_headling_protocol (protocol),
    KEY idx_headling_container (container_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Headling认证日志';

-- Cowrie 蜜罐日志表
CREATE TABLE IF NOT EXISTS cowrie_log (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    event_time DATETIME(6) NOT NULL COMMENT '事件时间',
    auth_id VARCHAR(36) NOT NULL COMMENT '认证唯一ID',
    session_id VARCHAR(36) NOT NULL COMMENT '会话ID',
    source_ip VARCHAR(15) NOT NULL COMMENT '攻击者IP',
    source_port INT NOT NULL COMMENT '攻击者端口',
    destination_ip VARCHAR(15) NOT NULL COMMENT '目标IP',
    destination_port INT NOT NULL COMMENT '目标端口',
    protocol VARCHAR(20) NOT NULL COMMENT '协议类型',
    client_info VARCHAR(255) COMMENT '客户端信息',
    fingerprint VARCHAR(64) COMMENT '指纹',
    username VARCHAR(255) COMMENT '用户名',
    password VARCHAR(255) COMMENT '密码',
    password_hash VARCHAR(255) COMMENT '密码哈希',
    command TEXT COMMENT '执行命令',
    command_found TINYINT COMMENT '命令识别',
    raw_log TEXT NOT NULL COMMENT '原始日志',
    container_id VARCHAR(64) COMMENT '容器ID',
    container_name VARCHAR(100) COMMENT '容器名称',
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '创建时间',
    UNIQUE KEY uq_cowrie_auth_id (auth_id),
    KEY idx_cowrie_session (session_id),
    KEY idx_cowrie_srcip (source_ip),
    KEY idx_cowrie_protocol (protocol),
    KEY idx_cowrie_container (container_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Cowrie蜜罐日志';

-- 病毒特征表
CREATE TABLE IF NOT EXISTS malware_signature (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL COMMENT '特征名称',
    pattern TEXT NOT NULL COMMENT '特征模式',
    type VARCHAR(20) NOT NULL COMMENT '类型',
    severity VARCHAR(20) NOT NULL COMMENT '严重程度',
    description TEXT COMMENT '描述',
    create_time DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '创建时间',
    update_time DATETIME(6) DEFAULT NULL COMMENT '更新时间',
    is_active TINYINT NOT NULL DEFAULT 1 COMMENT '是否激活'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='病毒特征';

-- 扫描结果表
CREATE TABLE IF NOT EXISTS scan_result (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    file_hash VARCHAR(64) NOT NULL COMMENT 'SHA256',
    md5_hash VARCHAR(32) COMMENT 'MD5',
    file_name VARCHAR(255) NOT NULL COMMENT '文件名',
    file_size BIGINT NOT NULL COMMENT '大小',
    scan_time DATETIME(6) NOT NULL COMMENT '扫描时间',
    is_infected TINYINT NOT NULL COMMENT '是否感染',
    threat_level VARCHAR(20) COMMENT '威胁等级',
    detection_count INT DEFAULT 0 COMMENT '检测威胁数量',
    scan_duration_ms BIGINT COMMENT '扫描耗时(毫秒)',
    source_ip VARCHAR(45) COMMENT '来源IP',
    user_agent TEXT COMMENT 'UA',
    KEY idx_scan_filehash (file_hash)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文件扫描结果';

-- 检测结果表
CREATE TABLE IF NOT EXISTS detection_result (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    scan_result_id BIGINT UNSIGNED NOT NULL COMMENT '扫描结果ID',
    signature_id BIGINT UNSIGNED NOT NULL COMMENT '特征ID',
    signature_name VARCHAR(100) COMMENT '特征名称',
    match_type VARCHAR(20) COMMENT '匹配类型',
    match_content TEXT COMMENT '匹配内容',
    severity VARCHAR(20) COMMENT '严重程度',
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '创建时间',
    CONSTRAINT fk_detection_scan FOREIGN KEY (scan_result_id) REFERENCES scan_result(id) ON DELETE CASCADE,
    CONSTRAINT fk_detection_signature FOREIGN KEY (signature_id) REFERENCES malware_signature(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='检测结果';

-- 攻击会话表
CREATE TABLE IF NOT EXISTS attack_session (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    session_id VARCHAR(36) NOT NULL COMMENT '会话ID',
    source_ip VARCHAR(45) NOT NULL COMMENT '攻击者IP',
    source_port INT COMMENT '攻击者端口',
    destination_ip VARCHAR(45) NOT NULL COMMENT '目标IP',
    destination_port INT NOT NULL COMMENT '目标端口',
    protocol VARCHAR(20) NOT NULL COMMENT '协议',
    start_time DATETIME(6) NOT NULL COMMENT '开始时间',
    end_time DATETIME(6) DEFAULT NULL COMMENT '结束时间',
    duration BIGINT COMMENT '持续时间(秒)',
    status VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '状态',
    auth_attempts INT DEFAULT 0 COMMENT '认证次数',
    command_count INT DEFAULT 0 COMMENT '命令次数',
    threat_level VARCHAR(20) COMMENT '威胁等级',
    container_id VARCHAR(64) COMMENT '容器ID',
    container_name VARCHAR(100) COMMENT '容器名称',
    user_agent TEXT COMMENT 'UA',
    fingerprint VARCHAR(64) COMMENT '指纹',
    UNIQUE KEY uq_attack_session (session_id),
    KEY idx_attack_session_srcip (source_ip)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='攻击会话';

-- 攻击事件表
CREATE TABLE IF NOT EXISTS attack_event (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    session_id VARCHAR(36) NOT NULL COMMENT '关联会话ID',
    event_type VARCHAR(30) NOT NULL COMMENT '事件类型',
    event_time DATETIME(6) NOT NULL COMMENT '事件时间',
    source_ip VARCHAR(45) NOT NULL COMMENT '攻击者IP',
    protocol VARCHAR(20) NOT NULL COMMENT '协议',
    username VARCHAR(255) COMMENT '用户名',
    password VARCHAR(255) COMMENT '密码',
    command TEXT COMMENT '命令',
    payload TEXT COMMENT '载荷',
    result TEXT COMMENT '执行结果',
    threat_level VARCHAR(20) COMMENT '威胁等级',
    is_blocked TINYINT DEFAULT 0 COMMENT '是否阻断',
    raw_data TEXT COMMENT '原始数据',
    KEY idx_attack_event_session (session_id),
    KEY idx_attack_event_srcip (source_ip),
    CONSTRAINT fk_event_session FOREIGN KEY (session_id) REFERENCES attack_session(session_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='攻击事件';

-- 威胁情报表
CREATE TABLE IF NOT EXISTS threat_intelligence (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    indicator_type VARCHAR(20) NOT NULL COMMENT '指标类型',
    indicator_value VARCHAR(255) NOT NULL COMMENT '指标值',
    threat_type VARCHAR(50) COMMENT '威胁类型',
    confidence INT COMMENT '置信度',
    severity VARCHAR(20) COMMENT '严重程度',
    source VARCHAR(100) COMMENT '来源',
    description TEXT COMMENT '描述',
    first_seen DATETIME(6) COMMENT '首次发现',
    last_seen DATETIME(6) COMMENT '最后发现',
    is_active TINYINT DEFAULT 1 COMMENT '是否激活',
    tags JSON COMMENT '标签',
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '创建时间',
    updated_at DATETIME(6) DEFAULT NULL COMMENT '更新时间',
    KEY idx_threat_indicator (indicator_value)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='威胁情报';

-- 蜜签事件表
CREATE TABLE IF NOT EXISTS honeytoken_event (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    token_id VARCHAR(36) NOT NULL COMMENT '蜜签ID',
    token_type VARCHAR(50) NOT NULL COMMENT '蜜签类型',
    token_name VARCHAR(100) COMMENT '蜜签名称',
    trigger_time DATETIME(6) NOT NULL COMMENT '触发时间',
    source_ip VARCHAR(45) NOT NULL COMMENT '触发IP',
    user_agent TEXT COMMENT 'UA',
    request_path VARCHAR(500) COMMENT '请求路径',
    request_data TEXT COMMENT '请求数据',
    response_code INT COMMENT '响应码',
    location VARCHAR(255) COMMENT '位置',
    description TEXT COMMENT '描述',
    threat_level VARCHAR(20) COMMENT '威胁等级',
    is_processed TINYINT DEFAULT 0 COMMENT '是否处理',
    KEY idx_honeytoken_token (token_id),
    KEY idx_honeytoken_srcip (source_ip)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='蜜签事件';

-- ==================================================================================