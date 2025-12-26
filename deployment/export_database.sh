#!/bin/bash

# 导出MySQL数据库脚本
# 用途：在本地Mac上导出andorralee数据库的数据

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKUP_FILE="${SCRIPT_DIR}/andorralee_data.sql"

echo "=== 开始导出MySQL数据库 ==="
echo "备份文件: ${BACKUP_FILE}"

# 导出数据库（不包含CREATE DATABASE语句，因为init_db.sql已经包含了表结构）
# 跳过视图，只导出表数据
mysqldump -h 127.0.0.1 -P 3306 -u root -pmac123456 \
  --databases andorralee \
  --single-transaction \
  --quick \
  --lock-tables=false \
  --no-create-db \
  --no-create-info \
  --skip-triggers \
  --ignore-table=andorralee.v_headling_auth_statistics \
  --set-gtid-purged=OFF \
  > "${BACKUP_FILE}"

# 压缩备份文件
gzip -f "${BACKUP_FILE}"

echo ""
echo "=== 数据库导出完成 ==="
echo "备份文件: ${BACKUP_FILE}.gz"
ls -lh "${BACKUP_FILE}.gz"
echo ""
echo "提示: 此文件包含所有数据，将与 init_db.sql 一起使用"
