# Andorralee 蜜罐管理系统 - 银河麒麟Docker部署指南

## 项目简介

Andorralee 是一个专业的蜜罐管理系统，专门针对银河麒麟V10操作系统进行了深度优化。本系统支持容器化部署，能够在银河麒麟系统上稳定运行，为国产化环境提供完整的蜜罐管理解决方案。

## 系统特性

- ✅ **银河麒麟原生支持** - 专门针对银河麒麟V10系统优化
- ✅ **多架构兼容** - 支持x86_64和ARM64架构
- ✅ **容器化部署** - 基于Docker的完整容器化解决方案
- ✅ **国产数据库支持** - 集成达梦数据库和MySQL
- ✅ **安全加固** - 符合国产化安全要求
- ✅ **高可用设计** - 支持集群部署和负载均衡

## 系统要求

### 硬件要求
- **CPU**: 2核心及以上 (推荐4核心)
- **内存**: 4GB及以上 (推荐8GB)
- **存储**: 20GB及以上可用空间 (推荐50GB)
- **网络**: 千兆网卡

### 软件要求
- **操作系统**: 银河麒麟V10 SP1/SP2/SP3
- **架构支持**: x86_64 或 ARM64 (aarch64)
- **Docker**: 19.03.0及以上版本
- **Docker Compose**: 1.25.0及以上版本
- **内核版本**: 4.19及以上

### 网络端口
- **8081**: 应用服务端口
- **3306**: MySQL数据库端口
- **5236**: 达梦数据库端口
- **6379**: Redis缓存端口 (可选)

## 安装前准备

### 1. 系统环境检查

```bash
# 检查操作系统版本
cat /etc/kylin-release

# 检查系统架构
uname -m

# 检查内核版本
uname -r

# 检查内存和磁盘
free -h
df -h
```

### 2. Docker环境安装

#### 在线安装 (推荐)
```bash
# 配置Docker官方源
sudo yum install -y yum-utils
sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo

# 安装Docker CE
sudo yum install -y docker-ce docker-ce-cli containerd.io

# 启动Docker服务
sudo systemctl start docker
sudo systemctl enable docker

# 安装Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/1.29.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

#### 离线安装
如果系统无法连接互联网，请参考 `scripts/install-docker-offline.sh` 脚本进行离线安装。

### 3. 系统优化配置

```bash
# 关闭防火墙 (或配置相应端口)
sudo systemctl stop firewalld
sudo systemctl disable firewalld

# 关闭SELinux (或配置相应策略)
sudo setenforce 0
sudo sed -i 's/SELINUX=enforcing/SELINUX=disabled/' /etc/selinux/config

# 配置Docker守护进程
sudo mkdir -p /etc/docker
sudo tee /etc/docker/daemon.json <<EOF
{
    "registry-mirrors": [
        "https://docker.mirrors.ustc.edu.cn",
        "https://hub-mirror.c.163.com"
    ],
    "storage-driver": "overlay2",
    "log-driver": "json-file",
    "log-opts": {
        "max-size": "100m",
        "max-file": "3"
    },
    "data-root": "/var/lib/docker"
}
EOF

# 重启Docker服务
sudo systemctl restart docker
```

## 部署方式

### 方式一：自动化部署 (推荐)

#### Linux环境部署
```bash
# 下载项目代码
git clone <项目地址>
cd andorralee

# 给部署脚本执行权限
chmod +x deploy-kylin.sh

# 执行自动化部署
./deploy-kylin.sh
```

#### Windows环境远程部署
```powershell
# 使用PowerShell脚本远程部署到银河麒麟系统
.\deploy-kylin.ps1 -KylinHost "192.168.1.100" -KylinUser "admin" -KylinPassword "password"

# 或使用SSH密钥认证
.\deploy-kylin.ps1 -KylinHost "192.168.1.100" -KylinUser "admin" -UseSSHKey -SSHKeyPath "C:\Users\user\.ssh\id_rsa"
```

### 方式二：手动部署

#### 1. 准备项目文件
```bash
# 创建项目目录
mkdir -p andorralee-kylin
cd andorralee-kylin

# 创建必要的目录结构
mkdir -p data/{mysql,dameng,redis}
mkdir -p logs config mysql/conf.d dameng/scripts redis scripts
```

#### 2. 配置文件准备
```bash
# 复制项目文件
cp Dockerfile.kylin ./
cp docker-compose.kylin.yml ./docker-compose.yml
cp -r cmd internal pkg routers static scripts ./

# 创建MySQL配置文件
cat > mysql/conf.d/kylin.cnf << 'EOF'
[mysqld]
character-set-server=utf8mb4
collation-server=utf8mb4_unicode_ci
max_connections=1000
innodb_buffer_pool_size=512M
lower_case_table_names=1
default_authentication_plugin=mysql_native_password
EOF
```

#### 3. 构建和启动
```bash
# 构建应用镜像
docker build -f Dockerfile.kylin -t andorralee:kylin-latest .

# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps
```

## 配置说明

### 环境变量配置

主要环境变量及其说明：

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| MYSQL_HOST | mysql | MySQL服务器地址 |
| MYSQL_PASSWORD | Kylin123456! | MySQL密码 |
| DAMENG_HOST | dameng | 达梦数据库地址 |
| DAMENG_PASSWORD | Kylin123456! | 达梦数据库密码 |
| GIN_MODE | release | Gin框架运行模式 |
| LOG_LEVEL | info | 日志级别 |

### 数据库配置

#### MySQL配置优化
```sql
-- 银河麒麟系统MySQL优化配置
SET GLOBAL max_connections = 1000;
SET GLOBAL innodb_buffer_pool_size = 536870912; -- 512MB
SET GLOBAL query_cache_size = 67108864; -- 64MB
```

#### 达梦数据库配置
```sql
-- 达梦数据库初始化配置
ALTER SYSTEM SET MAX_SESSIONS = 1000;
ALTER SYSTEM SET BUFFER_POOL_SIZE = 512;
```

## 验证部署

### 1. 服务状态检查
```bash
# 检查容器状态
docker-compose ps

# 检查服务日志
docker-compose logs -f andorralee

# 检查端口监听
netstat -tlnp | grep -E ':(8081|3306|5236)'
```

### 2. 功能验证
```bash
# 健康检查
curl http://localhost:8081/api/v1/health

# API接口测试
curl http://localhost:8081/api/v1/docker/images

# 数据库连接测试
docker-compose exec mysql mysql -u root -pKylin123456! -e "SHOW DATABASES;"
```

### 3. 性能测试
```bash
# 内存使用情况
docker stats

# 磁盘使用情况
du -sh data/

# 网络连接测试
curl -w "@curl-format.txt" -o /dev/null -s http://localhost:8081/api/v1/health
```

## 运维管理

### 日常管理命令

```bash
# 启动服务
docker-compose up -d

# 停止服务
docker-compose down

# 重启服务
docker-compose restart

# 查看日志
docker-compose logs -f [服务名]

# 进入容器
docker-compose exec andorralee bash

# 备份数据
docker-compose exec mysql mysqldump -u root -pKylin123456! andorralee > backup.sql
```

### 监控和告警

```bash
# 监控脚本示例
#!/bin/bash
# 检查服务状态
if ! curl -f http://localhost:8081/api/v1/health > /dev/null 2>&1; then
    echo "服务异常，发送告警"
    # 发送告警逻辑
fi

# 检查磁盘空间
DISK_USAGE=$(df / | awk 'NR==2{print $5}' | sed 's/%//')
if [ $DISK_USAGE -gt 80 ]; then
    echo "磁盘空间不足：${DISK_USAGE}%"
fi
```

## 故障排除

### 常见问题及解决方案

#### 1. Docker服务无法启动
```bash
# 检查Docker服务状态
sudo systemctl status docker

# 查看Docker日志
sudo journalctl -u docker.service

# 重启Docker服务
sudo systemctl restart docker
```

#### 2. 容器启动失败
```bash
# 查看容器日志
docker-compose logs [服务名]

# 检查端口占用
netstat -tlnp | grep [端口号]

# 检查磁盘空间
df -h
```

#### 3. 数据库连接失败
```bash
# 检查数据库容器状态
docker-compose ps mysql

# 测试数据库连接
docker-compose exec mysql mysql -u root -p

# 重置数据库密码
docker-compose exec mysql mysql -u root -e "ALTER USER 'root'@'%' IDENTIFIED BY 'Kylin123456!';"
```

#### 4. 性能问题
```bash
# 检查系统资源
top
free -h
iostat -x 1

# 优化Docker配置
# 编辑 /etc/docker/daemon.json
{
    "storage-opts": ["overlay2.size=20G"],
    "default-ulimits": {
        "nofile": {
            "Name": "nofile",
            "Hard": 64000,
            "Soft": 64000
        }
    }
}
```

## 安全加固

### 1. 容器安全
```bash
# 使用非root用户运行
USER appuser

# 限制容器权限
--security-opt no-new-privileges:true

# 只读文件系统
--read-only --tmpfs /tmp
```

### 2. 网络安全
```bash
# 配置防火墙规则
sudo firewall-cmd --permanent --add-port=8081/tcp
sudo firewall-cmd --reload

# 配置SSL/TLS
# 在nginx配置中启用HTTPS
```

### 3. 数据安全
```bash
# 数据库加密
# 启用MySQL透明数据加密
ALTER INSTANCE ROTATE INNODB MASTER KEY;

# 备份加密
gpg --cipher-algo AES256 --compress-algo 1 --s2k-cipher-algo AES256 --s2k-digest-algo SHA512 --s2k-mode 3 --s2k-count 65536 --symmetric backup.sql
```

## 升级和维护

### 版本升级
```bash
# 备份当前数据
docker-compose exec mysql mysqldump -u root -pKylin123456! andorralee > backup-$(date +%Y%m%d).sql

# 拉取新版本
git pull origin main

# 重新构建镜像
docker-compose build --no-cache

# 滚动更新
docker-compose up -d
```

### 定期维护
```bash
# 清理无用镜像
docker image prune -f

# 清理无用容器
docker container prune -f

# 清理无用卷
docker volume prune -f

# 数据库优化
docker-compose exec mysql mysql -u root -pKylin123456! -e "OPTIMIZE TABLE andorralee.*;"
```

## 技术支持

### 联系方式
- 技术支持邮箱: support@andorralee.com
- 文档中心: https://docs.andorralee.com
- 问题反馈: https://github.com/andorralee/issues

### 社区支持
- 官方论坛: https://forum.andorralee.com
- QQ技术群: 123456789
- 微信技术群: 扫描二维码加入

---

**注意**: 本部署指南专门针对银河麒麟V10系统编写，如在其他操作系统上部署，请参考通用部署指南。
