#!/bin/bash

# 银河麒麟系统Docker部署脚本
# Andorralee蜜罐管理系统 - 银河麒麟V10专用部署脚本
# 支持x86_64和ARM64架构

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_debug() {
    echo -e "${BLUE}[DEBUG]${NC} $1"
}

# 检查是否为root用户
check_root() {
    if [[ $EUID -eq 0 ]]; then
        log_warn "检测到root用户，建议使用普通用户运行"
        read -p "是否继续？(y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
}

# 检测系统信息
detect_system() {
    log_info "检测系统信息..."
    
    # 检测操作系统
    if [ -f /etc/kylin-release ]; then
        OS_TYPE="Kylin"
        OS_VERSION=$(cat /etc/kylin-release)
        log_info "检测到银河麒麟系统: $OS_VERSION"
    elif [ -f /etc/os-release ]; then
        OS_TYPE=$(grep "^ID=" /etc/os-release | cut -d'=' -f2 | tr -d '"')
        OS_VERSION=$(grep "^VERSION=" /etc/os-release | cut -d'=' -f2 | tr -d '"')
        log_info "检测到操作系统: $OS_TYPE $OS_VERSION"
    else
        log_error "无法检测操作系统类型"
        exit 1
    fi
    
    # 检测架构
    ARCH=$(uname -m)
    log_info "系统架构: $ARCH"
    
    # 检测内核版本
    KERNEL=$(uname -r)
    log_info "内核版本: $KERNEL"
    
    # 检测Docker是否安装
    if command -v docker &> /dev/null; then
        DOCKER_VERSION=$(docker --version)
        log_info "Docker已安装: $DOCKER_VERSION"
    else
        log_error "Docker未安装，请先安装Docker"
        exit 1
    fi
    
    # 检测Docker Compose是否安装
    if command -v docker-compose &> /dev/null; then
        COMPOSE_VERSION=$(docker-compose --version)
        log_info "Docker Compose已安装: $COMPOSE_VERSION"
    else
        log_error "Docker Compose未安装，请先安装Docker Compose"
        exit 1
    fi
}

# 检查系统要求
check_requirements() {
    log_info "检查系统要求..."
    
    # 检查内存
    MEMORY_GB=$(free -g | awk '/^Mem:/{print $2}')
    if [ "$MEMORY_GB" -lt 4 ]; then
        log_warn "系统内存少于4GB，可能影响性能"
    else
        log_info "内存检查通过: ${MEMORY_GB}GB"
    fi
    
    # 检查磁盘空间
    DISK_GB=$(df -BG . | awk 'NR==2{print $4}' | sed 's/G//')
    if [ "$DISK_GB" -lt 10 ]; then
        log_warn "可用磁盘空间少于10GB，可能不足"
    else
        log_info "磁盘空间检查通过: ${DISK_GB}GB可用"
    fi
    
    # 检查端口占用
    check_port() {
        local port=$1
        if netstat -tuln | grep ":$port " > /dev/null; then
            log_error "端口 $port 已被占用"
            return 1
        fi
        return 0
    }
    
    log_info "检查端口占用..."
    check_port 8081 || exit 1
    check_port 3306 || exit 1
    check_port 5236 || exit 1
    log_info "端口检查通过"
}

# 创建必要的目录
create_directories() {
    log_info "创建必要的目录..."
    
    mkdir -p data/{mysql,dameng,redis}
    mkdir -p logs
    mkdir -p config
    mkdir -p mysql/conf.d
    mkdir -p dameng/scripts
    mkdir -p redis
    
    log_info "目录创建完成"
}

# 创建配置文件
create_configs() {
    log_info "创建配置文件..."
    
    # MySQL配置
    cat > mysql/conf.d/kylin.cnf << EOF
[mysqld]
# 银河麒麟系统优化配置
character-set-server=utf8mb4
collation-server=utf8mb4_unicode_ci
max_connections=1000
innodb_buffer_pool_size=512M
innodb_log_file_size=128M
innodb_flush_log_at_trx_commit=2
sync_binlog=0
sql_mode=STRICT_TRANS_TABLES,NO_ZERO_DATE,NO_ZERO_IN_DATE,ERROR_FOR_DIVISION_BY_ZERO

# 银河麒麟系统兼容性设置
lower_case_table_names=1
default_authentication_plugin=mysql_native_password
EOF

    # Redis配置
    cat > redis/redis.conf << EOF
# Redis配置 - 银河麒麟系统优化
bind 0.0.0.0
port 6379
timeout 300
tcp-keepalive 60
maxmemory 256mb
maxmemory-policy allkeys-lru
save 900 1
save 300 10
save 60 10000
EOF

    # 银河麒麟MySQL初始化脚本
    cat > scripts/kylin_mysql_init.sql << EOF
-- 银河麒麟系统MySQL初始化脚本
USE andorralee;

-- 创建银河麒麟系统特定配置表
CREATE TABLE IF NOT EXISTS kylin_system_config (
    id INT AUTO_INCREMENT PRIMARY KEY,
    config_key VARCHAR(100) NOT NULL UNIQUE,
    config_value TEXT,
    description VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 插入银河麒麟系统配置
INSERT INTO kylin_system_config (config_key, config_value, description) VALUES
('system_type', 'Kylin', '操作系统类型'),
('docker_support', 'true', 'Docker支持状态'),
('architecture', 'x86_64', '系统架构'),
('deployment_mode', 'production', '部署模式')
ON DUPLICATE KEY UPDATE 
config_value = VALUES(config_value),
updated_at = CURRENT_TIMESTAMP;

-- 创建系统监控表
CREATE TABLE IF NOT EXISTS system_monitoring (
    id INT AUTO_INCREMENT PRIMARY KEY,
    metric_name VARCHAR(100) NOT NULL,
    metric_value DECIMAL(10,2),
    unit VARCHAR(20),
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_metric_timestamp (metric_name, timestamp)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
EOF

    log_info "配置文件创建完成"
}

# 构建镜像
build_images() {
    log_info "构建Docker镜像..."
    
    # 设置构建参数
    if [ "$ARCH" = "aarch64" ]; then
        PLATFORM="linux/arm64"
    else
        PLATFORM="linux/amd64"
    fi
    
    log_info "目标平台: $PLATFORM"
    
    # 构建应用镜像
    docker build \
        --platform $PLATFORM \
        -f Dockerfile.kylin \
        -t andorralee:kylin-latest \
        --build-arg BUILDPLATFORM=$PLATFORM \
        --build-arg TARGETPLATFORM=$PLATFORM \
        .
    
    if [ $? -eq 0 ]; then
        log_info "镜像构建成功"
    else
        log_error "镜像构建失败"
        exit 1
    fi
}

# 启动服务
start_services() {
    log_info "启动服务..."
    
    # 使用银河麒麟专用配置启动
    docker-compose -f docker-compose.kylin.yml up -d
    
    if [ $? -eq 0 ]; then
        log_info "服务启动成功"
    else
        log_error "服务启动失败"
        exit 1
    fi
}

# 等待服务就绪
wait_for_services() {
    log_info "等待服务就绪..."
    
    # 等待MySQL就绪
    log_info "等待MySQL服务..."
    for i in {1..60}; do
        if docker-compose -f docker-compose.kylin.yml exec mysql mysqladmin ping -h localhost -u root -pKylin123456! &> /dev/null; then
            log_info "MySQL服务就绪"
            break
        fi
        sleep 2
        if [ $i -eq 60 ]; then
            log_error "MySQL服务启动超时"
            exit 1
        fi
    done
    
    # 等待应用就绪
    log_info "等待应用服务..."
    for i in {1..60}; do
        if curl -f http://localhost:8081/api/v1/health &> /dev/null; then
            log_info "应用服务就绪"
            break
        fi
        sleep 2
        if [ $i -eq 60 ]; then
            log_error "应用服务启动超时"
            exit 1
        fi
    done
}

# 验证部署
verify_deployment() {
    log_info "验证部署..."
    
    # 检查容器状态
    log_info "检查容器状态:"
    docker-compose -f docker-compose.kylin.yml ps
    
    # 检查服务健康状态
    log_info "检查服务健康状态:"
    curl -s http://localhost:8081/api/v1/health | jq . || echo "健康检查API响应异常"
    
    log_info "部署验证完成"
}

# 显示部署信息
show_deployment_info() {
    log_info "部署完成！"
    echo
    echo "=========================================="
    echo "  Andorralee蜜罐管理系统 - 银河麒麟版"
    echo "=========================================="
    echo "系统信息:"
    echo "  操作系统: $OS_TYPE $OS_VERSION"
    echo "  架构: $ARCH"
    echo "  内核: $KERNEL"
    echo
    echo "服务地址:"
    echo "  应用服务: http://localhost:8081"
    echo "  MySQL: localhost:3306"
    echo "  达梦数据库: localhost:5236"
    echo "  Redis: localhost:6379"
    echo
    echo "管理命令:"
    echo "  查看日志: docker-compose -f docker-compose.kylin.yml logs -f"
    echo "  停止服务: docker-compose -f docker-compose.kylin.yml down"
    echo "  重启服务: docker-compose -f docker-compose.kylin.yml restart"
    echo "=========================================="
}

# 主函数
main() {
    log_info "开始部署Andorralee蜜罐管理系统 - 银河麒麟版"
    
    check_root
    detect_system
    check_requirements
    create_directories
    create_configs
    build_images
    start_services
    wait_for_services
    verify_deployment
    show_deployment_info
    
    log_info "部署完成！"
}

# 执行主函数
main "$@"
