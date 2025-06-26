#!/bin/bash

# 银河麒麟系统Docker离线安装脚本
# 适用于银河麒麟V10系统的Docker CE离线安装

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

# 检查是否为root用户
check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "此脚本需要root权限运行"
        exit 1
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
    else
        log_error "此脚本仅适用于银河麒麟系统"
        exit 1
    fi
    
    # 检测架构
    ARCH=$(uname -m)
    log_info "系统架构: $ARCH"
    
    # 设置Docker包架构
    if [ "$ARCH" = "x86_64" ]; then
        DOCKER_ARCH="x86_64"
    elif [ "$ARCH" = "aarch64" ]; then
        DOCKER_ARCH="aarch64"
    else
        log_error "不支持的架构: $ARCH"
        exit 1
    fi
    
    # 检测内核版本
    KERNEL=$(uname -r)
    log_info "内核版本: $KERNEL"
    
    # 检查内核版本是否满足要求
    KERNEL_MAJOR=$(echo $KERNEL | cut -d. -f1)
    KERNEL_MINOR=$(echo $KERNEL | cut -d. -f2)
    
    if [ "$KERNEL_MAJOR" -lt 4 ] || ([ "$KERNEL_MAJOR" -eq 4 ] && [ "$KERNEL_MINOR" -lt 19 ]); then
        log_warn "内核版本可能过低，建议升级到4.19或更高版本"
    fi
}

# 检查Docker是否已安装
check_existing_docker() {
    if command -v docker &> /dev/null; then
        EXISTING_VERSION=$(docker --version)
        log_warn "检测到已安装的Docker: $EXISTING_VERSION"
        read -p "是否卸载现有Docker并重新安装？(y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            uninstall_docker
        else
            log_info "保留现有Docker安装"
            exit 0
        fi
    fi
}

# 卸载现有Docker
uninstall_docker() {
    log_info "卸载现有Docker..."
    
    # 停止Docker服务
    systemctl stop docker || true
    systemctl disable docker || true
    
    # 卸载Docker包
    yum remove -y docker \
        docker-client \
        docker-client-latest \
        docker-common \
        docker-latest \
        docker-latest-logrotate \
        docker-logrotate \
        docker-engine \
        docker-ce \
        docker-ce-cli \
        containerd.io || true
    
    # 清理Docker数据
    rm -rf /var/lib/docker
    rm -rf /var/lib/containerd
    
    log_info "Docker卸载完成"
}

# 下载Docker离线包
download_docker_packages() {
    log_info "准备Docker离线安装包..."
    
    # 创建下载目录
    DOWNLOAD_DIR="/tmp/docker-offline"
    mkdir -p $DOWNLOAD_DIR
    cd $DOWNLOAD_DIR
    
    # Docker版本
    DOCKER_VERSION="20.10.21"
    
    # 下载地址
    DOCKER_URL="https://download.docker.com/linux/static/stable/${DOCKER_ARCH}/docker-${DOCKER_VERSION}.tgz"
    COMPOSE_URL="https://github.com/docker/compose/releases/download/v2.12.2/docker-compose-linux-${DOCKER_ARCH}"
    
    log_info "下载Docker二进制包..."
    if [ -f "docker-${DOCKER_VERSION}.tgz" ]; then
        log_info "Docker包已存在，跳过下载"
    else
        if command -v wget &> /dev/null; then
            wget $DOCKER_URL
        elif command -v curl &> /dev/null; then
            curl -L -o "docker-${DOCKER_VERSION}.tgz" $DOCKER_URL
        else
            log_error "需要wget或curl工具下载文件"
            log_info "请手动下载以下文件到 $DOWNLOAD_DIR 目录:"
            log_info "1. $DOCKER_URL"
            log_info "2. $COMPOSE_URL"
            exit 1
        fi
    fi
    
    log_info "下载Docker Compose..."
    if [ -f "docker-compose-linux-${DOCKER_ARCH}" ]; then
        log_info "Docker Compose已存在，跳过下载"
    else
        if command -v wget &> /dev/null; then
            wget $COMPOSE_URL
        elif command -v curl &> /dev/null; then
            curl -L -o "docker-compose-linux-${DOCKER_ARCH}" $COMPOSE_URL
        fi
    fi
}

# 安装Docker
install_docker() {
    log_info "安装Docker..."
    
    cd $DOWNLOAD_DIR
    
    # 解压Docker包
    tar -xzf docker-${DOCKER_VERSION}.tgz
    
    # 复制二进制文件
    cp docker/* /usr/bin/
    
    # 设置执行权限
    chmod +x /usr/bin/docker*
    chmod +x /usr/bin/containerd*
    chmod +x /usr/bin/ctr
    chmod +x /usr/bin/runc
    
    log_info "Docker二进制文件安装完成"
}

# 安装Docker Compose
install_docker_compose() {
    log_info "安装Docker Compose..."
    
    cd $DOWNLOAD_DIR
    
    # 复制Docker Compose
    cp "docker-compose-linux-${DOCKER_ARCH}" /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose
    
    # 创建符号链接
    ln -sf /usr/local/bin/docker-compose /usr/bin/docker-compose
    
    log_info "Docker Compose安装完成"
}

# 配置Docker服务
configure_docker_service() {
    log_info "配置Docker服务..."
    
    # 创建Docker用户组
    groupadd docker || true
    
    # 创建Docker服务文件
    cat > /etc/systemd/system/docker.service << 'EOF'
[Unit]
Description=Docker Application Container Engine
Documentation=https://docs.docker.com
After=network-online.target firewalld.service containerd.service
Wants=network-online.target
Requires=docker.socket containerd.service

[Service]
Type=notify
ExecStart=/usr/bin/dockerd -H fd:// --containerd=/run/containerd/containerd.sock
ExecReload=/bin/kill -s HUP $MAINPID
TimeoutSec=0
RestartSec=2
Restart=always
StartLimitBurst=3
StartLimitInterval=60s
LimitNOFILE=infinity
LimitNPROC=infinity
LimitCORE=infinity
TasksMax=infinity
Delegate=yes
KillMode=process
OOMScoreAdjust=-500

[Install]
WantedBy=multi-user.target
EOF

    # 创建Docker socket文件
    cat > /etc/systemd/system/docker.socket << 'EOF'
[Unit]
Description=Docker Socket for the API

[Socket]
ListenStream=/var/run/docker.sock
SocketMode=0660
SocketUser=root
SocketGroup=docker

[Install]
WantedBy=sockets.target
EOF

    # 创建containerd服务文件
    cat > /etc/systemd/system/containerd.service << 'EOF'
[Unit]
Description=containerd container runtime
Documentation=https://containerd.io
After=network.target local-fs.target

[Service]
ExecStartPre=-/sbin/modprobe overlay
ExecStart=/usr/bin/containerd
Type=notify
Delegate=yes
KillMode=process
Restart=always
RestartSec=5
LimitNPROC=infinity
LimitCORE=infinity
LimitNOFILE=infinity
TasksMax=infinity
OOMScoreAdjust=-999

[Install]
WantedBy=multi-user.target
EOF

    # 重新加载systemd
    systemctl daemon-reload
    
    log_info "Docker服务配置完成"
}

# 配置Docker守护进程
configure_docker_daemon() {
    log_info "配置Docker守护进程..."
    
    # 创建Docker配置目录
    mkdir -p /etc/docker
    
    # 创建Docker守护进程配置
    cat > /etc/docker/daemon.json << 'EOF'
{
    "registry-mirrors": [
        "https://docker.mirrors.ustc.edu.cn",
        "https://hub-mirror.c.163.com",
        "https://registry.docker-cn.com"
    ],
    "storage-driver": "overlay2",
    "storage-opts": [
        "overlay2.override_kernel_check=true"
    ],
    "log-driver": "json-file",
    "log-opts": {
        "max-size": "100m",
        "max-file": "3"
    },
    "data-root": "/var/lib/docker",
    "exec-opts": ["native.cgroupdriver=systemd"],
    "live-restore": true,
    "userland-proxy": false,
    "experimental": false,
    "icc": false,
    "default-ulimits": {
        "nofile": {
            "Name": "nofile",
            "Hard": 64000,
            "Soft": 64000
        }
    }
}
EOF

    log_info "Docker守护进程配置完成"
}

# 启动Docker服务
start_docker_service() {
    log_info "启动Docker服务..."
    
    # 启用并启动containerd
    systemctl enable containerd
    systemctl start containerd
    
    # 启用并启动Docker
    systemctl enable docker.socket
    systemctl enable docker.service
    systemctl start docker.socket
    systemctl start docker.service
    
    # 等待服务启动
    sleep 5
    
    # 检查服务状态
    if systemctl is-active --quiet docker; then
        log_info "Docker服务启动成功"
    else
        log_error "Docker服务启动失败"
        systemctl status docker
        exit 1
    fi
}

# 验证安装
verify_installation() {
    log_info "验证Docker安装..."
    
    # 检查Docker版本
    DOCKER_VERSION=$(docker --version)
    log_info "Docker版本: $DOCKER_VERSION"
    
    # 检查Docker Compose版本
    COMPOSE_VERSION=$(docker-compose --version)
    log_info "Docker Compose版本: $COMPOSE_VERSION"
    
    # 运行测试容器
    log_info "运行测试容器..."
    if docker run --rm hello-world > /dev/null 2>&1; then
        log_info "Docker测试成功"
    else
        log_warn "Docker测试失败，但基本功能可能正常"
    fi
    
    # 显示Docker信息
    log_info "Docker系统信息:"
    docker info | head -20
}

# 清理安装文件
cleanup() {
    log_info "清理安装文件..."
    rm -rf $DOWNLOAD_DIR
    log_info "清理完成"
}

# 显示安装完成信息
show_completion_info() {
    log_info "Docker安装完成！"
    echo
    echo "=========================================="
    echo "  Docker离线安装完成 - 银河麒麟版"
    echo "=========================================="
    echo "Docker版本: $(docker --version)"
    echo "Docker Compose版本: $(docker-compose --version)"
    echo
    echo "常用命令:"
    echo "  启动Docker: systemctl start docker"
    echo "  停止Docker: systemctl stop docker"
    echo "  重启Docker: systemctl restart docker"
    echo "  查看状态: systemctl status docker"
    echo "  查看日志: journalctl -u docker.service"
    echo
    echo "用户管理:"
    echo "  添加用户到docker组: usermod -aG docker \$USER"
    echo "  重新登录后生效"
    echo "=========================================="
}

# 主函数
main() {
    log_info "开始Docker离线安装 - 银河麒麟版"
    
    check_root
    detect_system
    check_existing_docker
    download_docker_packages
    install_docker
    install_docker_compose
    configure_docker_service
    configure_docker_daemon
    start_docker_service
    verify_installation
    cleanup
    show_completion_info
    
    log_info "安装完成！"
}

# 执行主函数
main "$@"
