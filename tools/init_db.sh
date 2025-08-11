#!/bin/bash
set -e

DB_MODE=${DB_MODE:-mysql}
echo "[init_db] DB_MODE=$DB_MODE"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

if [ "$DB_MODE" = "dameng" ]; then
	echo "[init_db] 使用达梦数据库脚本 deployment/dameng_migration.sql (全量)"
	DM_HOST=${DAMENG_HOST:-127.0.0.1}
	DM_PORT=${DAMENG_PORT:-5236}
	DM_USER=${DAMENG_USER:-SYSDBA}
	DM_PWD=${DAMENG_PASSWORD:-Dm123456}
	DM_DB=${DAMENG_DATABASE:-DOCKER_OPS}

	SQL_FILE="$ROOT_DIR/deployment/dameng_migration.sql"
	if [ ! -f "$SQL_FILE" ]; then
		echo "[init_db] 未找到达梦初始化脚本: $SQL_FILE" >&2
		exit 1
	fi

	# dmsql/dmcli 具体命令视镜像工具而定，这里假设使用 disql
	if ! command -v disql >/dev/null 2>&1; then
		echo "[init_db] 未找到 disql 客户端，请在达梦容器内执行或安装客户端" >&2
		exit 1
	fi

	echo "[init_db] 开始初始化达梦数据库 ($DM_HOST:$DM_PORT / $DM_USER)"
	disql "$DM_USER/$DM_PWD@$DM_HOST:$DM_PORT" < "$SQL_FILE"
	echo "[init_db] 达梦数据库初始化完成"
else
	echo "[init_db] 使用 MySQL 初始化脚本 deployment/init_db.sql + deployment/create_tables.sql (全量)"
	MYSQL_USER=${MYSQL_USER:-root}
	MYSQL_PASSWORD=${MYSQL_PASSWORD:-123456}
	MYSQL_HOST=${MYSQL_HOST:-127.0.0.1}
	MYSQL_PORT=${MYSQL_PORT:-13306}
	MYSQL_DB=${MYSQL_DATABASE:-andorralee}

	INIT_SQL="$ROOT_DIR/deployment/init_db.sql"
	TABLE_SQL="$ROOT_DIR/deployment/create_tables.sql"
	if [ ! -f "$INIT_SQL" ]; then echo "[init_db] 未找到 $INIT_SQL" >&2; exit 1; fi
	if [ ! -f "$TABLE_SQL" ]; then echo "[init_db] 未找到 $TABLE_SQL" >&2; exit 1; fi

	echo "[init_db] 运行基础结构脚本"
	mysql -h$MYSQL_HOST -P$MYSQL_PORT -u$MYSQL_USER -p$MYSQL_PASSWORD < "$INIT_SQL"
	echo "[init_db] 运行表结构脚本 ($MYSQL_DB)"
	mysql -h$MYSQL_HOST -P$MYSQL_PORT -u$MYSQL_USER -p$MYSQL_PASSWORD $MYSQL_DB < "$TABLE_SQL"
	echo "[init_db] MySQL 数据库初始化完成"
fi

echo "[init_db] 全部完成"