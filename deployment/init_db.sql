-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS andorralee;

-- 使用数据库
USE andorralee;

-- 创建数据表（如果不存在）
CREATE TABLE IF NOT EXISTS data_models (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    behavior VARCHAR(50),
    data TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci; 

-- 扫描结果表（支持前端历史列表字段）
CREATE TABLE IF NOT EXISTS scan_result (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    file_hash VARCHAR(64) NOT NULL COMMENT '文件SHA256',
    md5_hash VARCHAR(32) COMMENT '文件MD5',
    file_name VARCHAR(255) NOT NULL COMMENT '文件名',
    file_size BIGINT NOT NULL COMMENT '文件大小(字节)',
    scan_time DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '扫描时间',
    is_infected BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否感染',
    threat_level VARCHAR(20) COMMENT '威胁等级',
    detection_count INT DEFAULT 0 COMMENT '威胁数量',
    scan_duration_ms BIGINT COMMENT '扫描耗时(毫秒)',
    source_ip VARCHAR(45) COMMENT '来源IP',
    user_agent TEXT COMMENT '用户代理',
    file_path VARCHAR(500) COMMENT '服务器端保存路径',
    status VARCHAR(20) NOT NULL DEFAULT 'uploaded' COMMENT '扫描状态(uploaded/scanning/clean/infected/failed)',
    UNIQUE KEY uk_file_hash (file_hash),
    INDEX idx_scan_time (scan_time),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 检测结果表
CREATE TABLE IF NOT EXISTS detection_result (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    scan_result_id BIGINT UNSIGNED NOT NULL COMMENT '扫描结果ID',
    signature_id BIGINT UNSIGNED COMMENT '特征ID',
    signature_name VARCHAR(100) COMMENT '特征名称',
    match_type VARCHAR(20) COMMENT '匹配类型',
    match_content TEXT COMMENT '匹配内容',
    severity VARCHAR(20) COMMENT '严重程度',
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    INDEX idx_scan_result_id (scan_result_id),
    CONSTRAINT fk_detection_scan FOREIGN KEY (scan_result_id) REFERENCES scan_result(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;