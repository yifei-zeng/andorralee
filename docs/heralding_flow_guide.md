# Heralding 蜜罐工作流程详解

本文档详细描述了 Heralding 蜜罐在本项目后端中的完整生命周期，包括容器创建、端口映射、日志产生、日志收集与读取的流程。

## 1. 容器创建与端口映射 (Creation & Port Mapping)

### 1.1 前端请求
用户在前端（如 `debug-container.html`）发起创建容器请求：
- **输入**: 容器名称、镜像名称（包含 "heralding" 字样）、协议类型（如 "heralding"）。
- **API**: `POST /api/v1/container-instances`
- **Payload**:
  ```json
  {
    "name": "heralding-test",
    "image_name": "heralding:latest",
    "protocol": "heralding"
  }
  ```
  *注意：前端通常不发送 `port_mappings`，或者发送空值。*

### 1.2 后端处理 (`CreateContainerInstance`)
后端接收到请求后，执行以下逻辑：
1.  **识别 Heralding 镜像**: 检查 `image_name` 是否包含 "heralding" 且 `port_mappings` 为空。
2.  **自动注入端口**: 如果满足上述条件，后端自动注入 Heralding 支持的所有标准容器端口映射配置，并将宿主机端口设置为 "auto"（自动分配）：
    - `22/tcp` (SSH)
    - `23/tcp` (Telnet)
    - `80/tcp` (HTTP)
    - `110/tcp` (POP3)
    - `143/tcp` (IMAP)
    - `443/tcp` (HTTPS)
    - `3306/tcp` (MySQL)
    - `3389/tcp` (RDP)
    - `5900/tcp` (VNC)
    - `993/tcp` (IMAPS)
    - `995/tcp` (POP3S)
3.  **Docker 创建**: 调用 Docker API 创建容器，Docker Daemon 会为上述每个容器端口在宿主机上分配一个随机端口（例如 `0.0.0.0:32768 -> 22/tcp`）。
4.  **数据库记录**: 将分配好的端口映射关系保存到数据库 `honeypot_instances` 表中。

## 2. 访问与日志产生 (Access & Logging)

### 2.1 攻击/访问
攻击者或测试用户通过宿主机 IP 和随机分配的端口访问蜜罐。
- 例如：访问 `HostIP:32768` (映射到容器 22 端口)。
- Docker 将流量转发到容器内部。

### 2.2 容器内处理
容器内的 Heralding 进程监听这些端口，处理连接请求，并记录交互数据。
- **日志位置**: Heralding 将日志写入容器内的文件系统，通常位于 `/var/log/heralding/` 或 `/opt/heralding/logs/`。
- **日志文件**:
  - `auth.csv` / `log_auth.csv`: 记录认证尝试（用户名、密码、源IP等）。
  - `session.csv` / `log_session.csv`: 记录会话元数据（持续时间、协议等）。

## 3. 日志收集 (Log Collection)

### 3.1 触发收集
用户在前端点击 "拉取日志" 按钮，或系统定时任务触发。
- **API**: `POST /api/v1/heralding/pull-logs`
- **Payload**: `{"container_id": "..."}`

### 3.2 后端执行 (`PullHeraldingLogs`)
后端 `HeraldingService` 执行以下步骤：
1.  **定位文件**: 尝试从容器内的多个预定义路径查找日志文件（`log_auth.csv`, `/var/log/heralding/auth.csv` 等）。
2.  **复制文件**: 使用 `docker cp` (API: `CopyFromContainer`) 将日志文件从运行中的容器复制到内存中。
3.  **解析 CSV**: 解析 CSV 内容，提取字段（Timestamp, AuthID, SourceIP, Username, Password 等）。
4.  **去重与入库**:
    - 检查数据库中是否已存在相同的 `AuthID` 或 `SessionID`。
    - 过滤掉时间戳早于上次同步时间的记录。
    - 将新记录批量插入到 MySQL 数据库的 `heralding_auth_logs` 和 `heralding_session_logs` 表中。

## 4. 日志读取 (Log Reading)

### 4.1 查询请求
用户在前端查看日志列表或统计信息。
- **API**: `GET /api/v1/heralding/logs` 或 `GET /api/v1/heralding/statistics`

### 4.2 数据返回
后端从 MySQL 数据库查询结构化数据并返回 JSON 格式给前端展示。

---

## 总结
整个流程实现了**零配置启动**（自动端口映射）和**非侵入式日志收集**（通过 Docker API 复制容器内文件），使得 Heralding 蜜罐能够即开即用，并持久化存储攻击数据。

## 5. 如何访问 Heralding 蜜罐并触发日志（实用指南）

下面给出从“创建容器”到“触发并读取日志”的实操步骤与常见排查方法，包含命令示例，便于快速验证与调试。

### 5.1 为什么 `localhost:8080` 不能访问 Heralding？
- `localhost:8080` 通常是后端（管理 UI / API）监听的端口，而不是 Heralding 容器的任意服务端口。
- Heralding 容器内部监听的是一组服务端口（如 22/80/3306 等），后端在创建时会把这些容器端口映射到宿主机上的随机端口（当使用 `auto` 分配时）。因此需要先查出宿主机上分配到的具体端口，再用 `localhost:HOSTPORT` 进行访问。

### 5.2 如何查到宿主机上分配的端口（几种方法）

- 使用后端 API（推荐）:
  1. 创建实例返回的 JSON 中会包含 `port_mappings`，例如：
     ```json
     "port_mappings": { "22": "32768", "80": "32769" }
     ```
     这里的 key 是容器端口，value 是宿主机端口。
  2. 查询所有实例：`GET /api/v1/container-instances`，或单个实例 `GET /api/v1/container-instances/{id}`。

- 使用前端界面（如果你使用了 `debug-container.html` 或 `index.html`）:
  - 在容器列表中查看 `ports_pretty` 或 `port_mappings` 字段（界面会展示“hostPort protocol”的友好视图）。

- 使用 Docker CLI（仅当 Docker 可用并且有容器 ID 时）:
  - `docker ps` 找到容器 ID
  - `docker port <container-id>` 将列出映射关系，例如 `22/tcp -> 0.0.0.0:32768`

### 5.3 访问不同服务的示例命令
假设你的容器 `port_mappings` 为 `{"22":"32768","80":"32769","3306":"32770","3389":"32771","5900":"32772"}`，在宿主机上使用以下命令：

- HTTP (Web)：
  ```bash
  curl http://localhost:32769/
  # 或在浏览器打开 http://localhost:32769
  ```

- SSH：
  ```bash
  ssh -p 32768 user@localhost
  # 如果是自动化测试，使用密码或公钥按蜜罐配置尝试登录
  ```

- Telnet（若开放 23）：
  ```bash
  telnet localhost 32768_or_telnet_port
  # 或使用 nc 来发送/接收数据： nc localhost <hostport>
  ```

- MySQL：
  ```bash
  mysql -h 127.0.0.1 -P 32770 -u root -p
  # 注意：Heralding 的 mysql 协议端口用于诱捕，不一定是完整数据库功能
  ```

- RDP（远程桌面，端口 3389）：
  - Windows: 在远程桌面连接中输入 `localhost:32771` 或 `localhost:32771`（Windows 有时需要 `localhost:32771`）
  - Linux: 使用 `rdesktop` / `xfreerdp`: `xfreerdp /v:localhost:32771`

- VNC（5900）：
  ```bash
  # VNC客户端通常使用 :display 格式，若映射到 32772, 则 display = 32772 - 5900
  # 更可靠的方式是直接在VNC客户端中输入 Host=localhost, Port=32772
  vncviewer localhost:32772
  ```

### 5.4 触发日志（攻击/访问）后的采集步骤
1. 先确保容器状态为 `running`（可以在 API 或 `docker ps` 查看）。
2. 通过上述命令访问相应服务以生成认证尝试或会话（例如 SSH 登录尝试、HTTP 请求等）。
3. 等待一段时间让 Heralding 将交互写入容器内部日志文件（通常是 CSV 文件，见第 3 节）。
4. 调用后端拉取日志 API：
   ```bash
   curl -X POST "http://localhost:8080/api/v1/heralding/pull-logs" \
     -H "Content-Type: application/json" \
     -d '{"container_id":"<your-docker-container-id>"}'
   ```
5. 拉取后查询已保存的日志：
   ```bash
   curl "http://localhost:8080/api/v1/heralding/logs/container/<your-docker-container-id>"
   ```
   或者查询所有日志：`GET /api/v1/heralding/logs`。

### 5.5 前端快捷操作（如果使用 UI）
- 在 `Logs` 面板：选择 `Heralding`，从下拉菜单选择容器（或手动填写 Docker 容器 ID），点击 “拉取日志” 按钮。
- 拉取后点击 “查询” 即可在界面中查看最近的认证尝试和会话信息。

### 5.6 常见问题与排查建议
- 容器没有响应：检查容器是否处于 `running` 状态（`GET /api/v1/container-instances` 或 `docker ps`）。
- 找不到端口映射：确认创建时后端是否返回了 `port_mappings`，或在数据库/API 中读取实例详情查看 `port_mappings` 字段；若为空，说明创建时未注入端口（可能镜像名未包含 "heralding"）。
- 仍访问不了：确认宿主机防火墙/安全组是否允许该端口访问（macOS 本地一般允许 localhost）。
- 日志未生成：确认 Heralding 服务确实在容器内运行并写日志，或在容器内部手动检查日志路径（如果可以进入容器）：
  ```bash
  docker exec -it <container-id> /bin/sh -c "ls -la /var/log/heralding || ls -la /opt/heralding/logs"
  ```

### 5.7 示例完整流程（快速验证）
1. 创建容器（前端或 API）。
2. 从创建返回或通过 `GET /container-instances` 找到 `container_id` 与 `port_mappings`。
3. 访问其中一个映射端口，例如：`curl http://localhost:<hostPort>` 或 `ssh -p <hostPort> localhost`。
4. 触发一次认证/会话后，调用 `POST /heralding/pull-logs` 以拉取并入库日志。
5. `GET /heralding/logs/container/<container_id>` 查看结果。

---

如果您愿意，我可以：
- 帮您在前端增加一个“创建并显示端口映射”的弹窗，创建完成后自动打开映射端口列表；
- 或者帮您写一组自动化脚本（shell）来创建、访问（示例请求）并自动拉取日志用于快速验证。

