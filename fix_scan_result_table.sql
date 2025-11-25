-- 修复 scan_result 表的自增ID问题
-- 1. 先检查当前表结构
SHOW CREATE TABLE scan_result;

-- 2. 如果ID字段不是自增的，需要修改
-- ALTER TABLE scan_result MODIFY COLUMN id BIGINT UNSIGNED AUTO_INCREMENT;

-- 3. 检查并删除可能存在的孤儿记录
DELETE FROM detection_result WHERE scan_result_id = 0 OR scan_result_id NOT IN (SELECT id FROM scan_result);

-- 4. 重置自增ID（可选，如果表为空）
-- ALTER TABLE scan_result AUTO_INCREMENT = 1;

-- 5. 验证修复
SELECT 
    COLUMN_NAME,
    DATA_TYPE,
    COLUMN_KEY,
    EXTRA
FROM INFORMATION_SCHEMA.COLUMNS
WHERE TABLE_SCHEMA = 'andorralee' 
AND TABLE_NAME = 'scan_result'
AND COLUMN_NAME = 'id';
