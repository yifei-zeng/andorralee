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