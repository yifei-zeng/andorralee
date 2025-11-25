-- 检查 scan_result 表结构
SHOW CREATE TABLE scan_result;

-- 检查是否有数据
SELECT COUNT(*) as total_count FROM scan_result;

-- 检查最近的记录
SELECT id, file_hash, file_name, scan_time FROM scan_result ORDER BY scan_time DESC LIMIT 5;

-- 检查表的自增值
SELECT AUTO_INCREMENT FROM information_schema.TABLES 
WHERE TABLE_SCHEMA = 'andorralee' AND TABLE_NAME = 'scan_result';
