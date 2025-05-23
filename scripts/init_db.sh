#!/bin/bash

# MySQL连接信息
MYSQL_USER="root"
MYSQL_PASSWORD="123456"
MYSQL_HOST="127.0.0.1"
MYSQL_PORT="3306"

# 运行SQL脚本
mysql -h$MYSQL_HOST -P$MYSQL_PORT -u$MYSQL_USER -p$MYSQL_PASSWORD < init_db.sql

echo "数据库初始化完成！" 