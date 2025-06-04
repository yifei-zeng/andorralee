-- 更新数据库表结构以支持更长的image_id字段
-- 执行此脚本前请备份数据库

-- 更新蜜罐实例表的image_id字段长度
ALTER TABLE honeypot_instance MODIFY COLUMN image_id VARCHAR(100) COMMENT 'Docker镜像ID';

-- 更新Docker镜像表的image_id字段长度
ALTER TABLE docker_image MODIFY COLUMN image_id VARCHAR(100) NOT NULL COMMENT '镜像ID';

-- 更新Docker镜像日志表的image_id字段长度
ALTER TABLE docker_image_log MODIFY COLUMN image_id VARCHAR(100) COMMENT '镜像ID';

-- 更新Docker容器表的image_id字段长度
ALTER TABLE docker_container MODIFY COLUMN image_id VARCHAR(100) COMMENT '关联的镜像ID';

-- 验证更改
SHOW COLUMNS FROM honeypot_instance LIKE 'image_id';
SHOW COLUMNS FROM docker_image LIKE 'image_id';
SHOW COLUMNS FROM docker_image_log LIKE 'image_id';
SHOW COLUMNS FROM docker_container LIKE 'image_id';
