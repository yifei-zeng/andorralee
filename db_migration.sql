-- 先删除外键约束
ALTER TABLE honeypot_instance DROP FOREIGN KEY honeypot_instance_ibfk_1;
ALTER TABLE honeypot_log DROP FOREIGN KEY honeypot_log_ibfk_1;
ALTER TABLE rule_log DROP FOREIGN KEY rule_log_ibfk_1;
ALTER TABLE rule_log DROP FOREIGN KEY IF EXISTS rule_log_ibfk_2;
ALTER TABLE bait DROP FOREIGN KEY bait_ibfk_1;

-- 修改蜜罐模板表的ID字段
ALTER TABLE honeypot_template MODIFY COLUMN id BIGINT UNSIGNED AUTO_INCREMENT;

-- 修改蜜罐实例表的ID和外键字段
ALTER TABLE honeypot_instance MODIFY COLUMN id BIGINT UNSIGNED AUTO_INCREMENT;
ALTER TABLE honeypot_instance MODIFY COLUMN template_id BIGINT UNSIGNED NOT NULL;

-- 修改蜜罐日志表的ID和外键字段
ALTER TABLE honeypot_log MODIFY COLUMN id BIGINT UNSIGNED AUTO_INCREMENT;
ALTER TABLE honeypot_log MODIFY COLUMN instance_id BIGINT UNSIGNED NOT NULL;

-- 修改安全规则表的ID字段
ALTER TABLE security_rule MODIFY COLUMN id BIGINT UNSIGNED AUTO_INCREMENT;

-- 修改规则日志表的ID和外键字段
ALTER TABLE rule_log MODIFY COLUMN id BIGINT UNSIGNED AUTO_INCREMENT;
ALTER TABLE rule_log MODIFY COLUMN rule_id BIGINT UNSIGNED NOT NULL;
ALTER TABLE rule_log MODIFY COLUMN instance_id BIGINT UNSIGNED;

-- 修改诱饵表的ID和外键字段
ALTER TABLE bait MODIFY COLUMN id BIGINT UNSIGNED AUTO_INCREMENT;
ALTER TABLE bait MODIFY COLUMN instance_id BIGINT UNSIGNED;

-- 重新添加外键约束
ALTER TABLE honeypot_instance 
    ADD CONSTRAINT honeypot_instance_ibfk_1 
    FOREIGN KEY (template_id) REFERENCES honeypot_template(id);

ALTER TABLE honeypot_log 
    ADD CONSTRAINT honeypot_log_ibfk_1 
    FOREIGN KEY (instance_id) REFERENCES honeypot_instance(id);

ALTER TABLE rule_log 
    ADD CONSTRAINT rule_log_ibfk_1 
    FOREIGN KEY (rule_id) REFERENCES security_rule(id);

ALTER TABLE rule_log 
    ADD CONSTRAINT rule_log_ibfk_2 
    FOREIGN KEY (instance_id) REFERENCES honeypot_instance(id);

ALTER TABLE bait 
    ADD CONSTRAINT bait_ibfk_1 
    FOREIGN KEY (instance_id) REFERENCES honeypot_instance(id); 