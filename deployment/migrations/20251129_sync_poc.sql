-- Adds source tracking columns for remote sync PoC
ALTER TABLE `honeypot_instance`
	ADD COLUMN `source_host` VARCHAR(64) NULL DEFAULT NULL AFTER `description`,
	ADD COLUMN `remote_id` BIGINT NULL DEFAULT NULL AFTER `source_host`;

CREATE UNIQUE INDEX `ux_honeypot_instance_source_remote`
	ON `honeypot_instance` (`source_host`, `remote_id`);

ALTER TABLE `honeypot_log`
	ADD COLUMN `source_host` VARCHAR(64) NULL DEFAULT NULL AFTER `log_time`,
	ADD COLUMN `remote_id` BIGINT NULL DEFAULT NULL AFTER `source_host`;

CREATE UNIQUE INDEX `ux_honeypot_log_source_remote`
	ON `honeypot_log` (`source_host`, `remote_id`);

CREATE TABLE IF NOT EXISTS `sync_errors` (
	`id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
	`source_host` VARCHAR(64) NOT NULL,
	`table_name` VARCHAR(64) NOT NULL,
	`payload` LONGTEXT NOT NULL,
	`error_message` LONGTEXT NOT NULL,
	`created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (`id`),
	KEY `idx_sync_source` (`source_host`),
	KEY `idx_sync_table` (`table_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
