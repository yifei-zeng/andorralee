# 0826 离线部署脚本使用说明

本目录提供一个更易控的 ARM64 离线一键部署脚本：`load-and-start-0826.sh`。
默认关闭 Cowrie 自动拉取日志，你可以通过参数开启并指定容器与日志路径。

## 目录内容
- load-and-start-0826.sh：部署脚本（ARM64，仅需安装 docker）。
- README.md：使用说明。

可复用 send/honeypotV1.0 下的这些文件（放到与脚本同一目录）：
- app_andorralee_honeypotV1_0_arm64.tar（后端镜像）
- mysql_preseed_andorralee_8_0_arm64.tar（可选，预置数据的 MySQL 镜像）
- mysql_8_0_arm64.tar（当没有预置镜像时备用）
- create_tables.sql / init_db_extra.sql / andorralee_dump.sql（可选，INIT_SQL 模式用于初始化）
- offline-images/ 目录（离线依赖镜像，若存在 cowrie.tar 将被优先加载）

## 快速开始
1. 将脚本和离线包放在同一目录（例如 send/honeypotV1.0）。
2. 赋权并运行（需要 root）：
   sudo chmod +x ./load-and-start-0826.sh
   sudo ./load-and-start-0826.sh --app-port 9091
3. 健康检查：
   curl -I http://<服务器IP>:9091/health  # 预期 200

> 提示：端口 9091 是主机端口，容器内部服务固定监听 9090。

## 参数说明
- --force：清理现有 andorralee / andorralee-mysql 容器及 mysql_data 卷后重建。
- --app-port <port>：后端对外端口（默认 9090，建议指定为 9091 以免冲突）。
- --mysql-port <port>：MySQL 对外端口（默认 3306）。
- --cowrie-auto on|off：是否启用 Cowrie 自动拉取（默认 off）。
- --cowrie-container <name-or-id>：指定 Cowrie 容器名/ID（启用自动拉取时建议填写）。
- --cowrie-log-path <path>：指定 Cowrie 日志路径（容器内路径，示例 /var/log/cowrie）。
- --cowrie-target-digest <sha256>：按镜像 digest 精确锁定目标 Cowrie 容器（示例 sha256:7c9d...）。
- --cowrie-stop-others yes|no：与 --cowrie-target-digest 联用，找到目标后可选择停止其他 cowrie 容器（默认 no）。

脚本行为要点：
- 自动加载本地离线镜像 tar；若存在 offline-images/cowrie.tar，会先加载该文件。
- 自动创建 docker 网络 honeypot-net；MySQL 使用持久化卷 mysql_data。
- MySQL 就绪后：若本地 andorralee_dump.sql 存在且三张核心表总行数为 0，将尝试导入。
- 启动后端容器时会自动检测并挂载宿主机 docker.sock（rootful/rootless 皆可）。
- COWRIE_* 环境变量：
  - COWRIE_AUTO_ENABLED=true/false 由 --cowrie-auto 控制。
  - COWRIE_DEFAULT_CONTAINER 来自 --cowrie-container。
  - COWRIE_LOG_PATH 来自 --cowrie-log-path（也支持服务内置的多个常见路径扫描）。

## 常见用法
- 仅部署（不启用自动拉取）：
  sudo ./load-and-start-0826.sh --app-port 9091

- 启用自动拉取并指定容器与日志路径：
  sudo ./load-and-start-0826.sh --app-port 9091 --cowrie-auto on \
       --cowrie-container cowrie-arm64 --cowrie-log-path /var/log/cowrie

- 通过镜像 digest 精确定位当前可产生日志的容器，并可选择停止其它 cowrie 容器：
  sudo ./load-and-start-0826.sh --app-port 9091 --cowrie-auto on \
       --cowrie-target-digest sha256:7c9d390c2ef26920b99716ef11aa0521d46679979678c646d49e2f386343d9e2 \
       --cowrie-log-path /var/log/cowrie --cowrie-stop-others yes

- 强制清理并重建：
  sudo ./load-and-start-0826.sh --force --app-port 9091

## 验证
- 容器状态：
  docker ps
- 应用日志：
  docker logs -f andorralee
- MySQL 日志：
  docker logs -f andorralee-mysql
- API 健康：
  curl -I http://<服务器IP>:9091/health
- Cowrie 日志接口：
  curl http://<服务器IP>:9091/api/v1/cowrie/logs

## 故障排查
- 端口占用：脚本会在启动前检查 --app-port 是否被占用，必要时改为 9091。
- MySQL 导入失败：andorralee_dump.sql 若编码/注释异常可能报错，脚本会继续运行，你也可以手动清理注释后再导入。
- 取不到 Cowrie 日志：
  - 确认已启用 --cowrie-auto on，并正确设置 --cowrie-container 与 --cowrie-log-path。
  - 若未指定日志路径，服务会尝试常见目录：/cowrie/cowrie-git/var/log/cowrie、/var/log/cowrie、/home/cowrie/cowrie-git/var/log/cowrie。
  - 确认后端容器已挂载 docker.sock（脚本会自动处理），并且 Cowrie 容器正在运行。

## 访问前端
- 请使用 HTTP 方式访问，不要使用 HTTPS。
- 示例：浏览器访问 http://<服务器IP>:9091/
- 若自定义前端调用，请注意跨域与 OPTIONS 预检，后端已添加 CORS 处理并支持 204 预检响应（如仍遇到问题，请反馈具体路径与错误）。

## 清理
- 停止并删除容器：
  docker rm -f andorralee andorralee-mysql
- 删除持久化卷（包含数据库数据，谨慎）：
  docker volume rm mysql_data
