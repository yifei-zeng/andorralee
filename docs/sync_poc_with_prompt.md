# 远端后端推送到中央后端 同步方案（PoC）说明

本文档说明最小可行的 PoC：远端后端周期性将本地新增数据推送到中央后端，前端仅访问中央（例如 `127.0.0.1`）。不部署 agent、不改前端。该 PoC 由代码补丁、迁移脚本与一个轻量远端脚本组成（已放在仓库 `tools/`）。

**目标**
- 在内网环境下，快速实现可运行的增量同步 PoC，使前端只读取中央数据库的数据视图。
- 今日能验证（单台机器或在 VM 上模拟多台机器）。

---

## 一、前提与约束
- 项目语言/框架：Go + Gin；数据库：MySQL（现有 schema 可用）。
- 不部署额外 agent，远端通过运行一个小脚本定期把新增行 POST 到中央。
- 使用静态同步令牌（HTTP 头 `X-SYNC-TOKEN`）做最小认证（内网场景）。
- PoC 不处理大二进制样本传输，仅同步表行与元数据。

---

## 二、架构概览（PoC）
- 中央后端（Central）
  - 新增接口：`POST /api/v1/sync/events`，用于接收远端批量 JSON 行并写入中央 DB。
  - 在业务表中添加 `source_host` 与 `remote_id` 字段，用于去重与溯源。
  - 失败记录写入 `sync_errors` 表，便于排查。
- 远端后端（Remote）
  - 保持现有后端与本地 DB；运行 `tools/remote_sync.py` 定时任务，查询本地自上次同步后的新增行并 POST 到 Central。
- 前端
  - 继续只连接 Central（无需改动）。

---

## 三、中央端最小改动（已实现）
1. 新增接口 `POST /api/v1/sync/events`（`internal/handlers/sync_handler.go`）：
   - 校验 `X-SYNC-TOKEN` against `SYNC_TOKEN`（配置在 `internal/config/config.go` 环境变量 `SYNC_TOKEN`）。
   - 支持 `table` 为 `honeypot_instances` 或 `honeypot_logs`（可扩展）。
   - 对每行按 `source_host + remote_id` 去重后写入；失败则写入 `sync_errors`。
2. 模型修改（`internal/repositories/models.go`）：为 `honeypot_instance` 与 `honeypot_log` 增加 `source_host` 与 `remote_id` 字段；新增 `SyncError` 模型。
3. Migration SQL：`deployment/migrations/20251129_sync_poc.sql`（增加列、索引、`sync_errors` 表）。

---

## 四、远端同步脚本（已放仓库 `tools/remote_sync.py`）
- 功能：从远端 MySQL 查询自上次同步后的新行，构建 JSON payload，POST 到 Central `POST /api/v1/sync/events`，成功则更新本地 `sync_state.json`（记录每表最大id）。
- 配置示例：`tools/remote_sync_config.example.json`（包含 `source_host`、`central_url`、`sync_token`、`mysql` 连接信息、同步表信息）。
- 运行方式：一次性运行或 `--loop` 持续运行；可通过 `cron` 或 `systemd` 调度。

---

## 五、数据库迁移（手动步骤）
1. 在 Central MySQL 上执行：
   - `deployment/migrations/20251129_sync_poc.sql` 中的 SQL。
2. 可选：重启 Central 服务并设置环境变量 `SYNC_TOKEN`：
   ```powershell
   $env:SYNC_TOKEN = "sync-dev-token"
   ```

---

## 六、单台设备实验（最小验证流程）
场景 A：你现在只有一台设备
- 你可以在该台机器同时运行 "中央" + "远端"（把 `central_url` 指向 `http://127.0.0.1:8080/api/v1/sync/events`），以模拟远端向中央推送。
- 可选：在本机用 Docker 启动一个 MySQL 容器作为本地 DB（远端脚本使用本地 DB 连接），或直接用项目自带 MySQL。推荐快速步骤：

快速步骤（单机，全在一台机器测试）：
1. 在 MySQL 中执行迁移 SQL（见上）并确认表结构变更。 
2. 启动 Central 后端（确保 `SYNC_TOKEN` 环境变量和 `config` 一致）。
3. 准备 `tools/remote_sync_config.example.json` 为 `remote_sync_config.json`，把 `central_url` 指向 `http://127.0.0.1:8080/api/v1/sync/events`，`source_host` 写 `127.0.0.2`（或任意标识），配置 `mysql` 为当前机器的 DB。 
4. 运行远端脚本（手动一次）：
   ```bash
   python tools/remote_sync.py --config tools/remote_sync_config.json
   ```
5. 在 Central DB 执行验证 SQL：
   ```sql
   SELECT source_host, remote_id FROM honeypot_instance WHERE source_host='127.0.0.2' LIMIT 10;
   SELECT * FROM sync_errors ORDER BY created_at DESC LIMIT 10;
   ```
6. 若要模拟多台机器，可在本机复制多个 config（不同 `source_host` 字段）并并行运行脚本，或在 VM/容器中跑脚本以模拟独立节点。

场景 B：使用虚拟机或容器模拟多台机器
- 可使用 VirtualBox/Hyper-V 或 Docker Compose（每个容器运行 Python 脚本并连接各自的 MySQL），或把脚本放到不同端口的服务中运行来模拟。对 PoC 来说单机模式已经足够验证功能。

---

## 七、验收标准（PoC）
- 远端脚本成功 POST 后，中央 DB 能看到 `honeypot_instance` 或 `honeypot_log` 的新记录，且 `source_host` 与 `remote_id` 值正确。
- 重复 POST 相同数据不会重复写入（去重成功）。
- 当 Central 返回非 2xx 时，远端脚本不更新 `sync_state.json`，下次重试仍会尝试发送该批次数据。

---

## 八、最可能遇到的问题与快速解决办法
1. ID 冲突（remote_id 与 central 自增主键冲突）
   - 方案：PoC 中中央使用 `source_host+remote_id` 做去重，中央表仍由自身生成主键。请确保 `remote_id` 字段写入 `remote_id` 列而非替换中央主键。
2. Token 配置错误（401）
   - 检查 Central 的 `SYNC_TOKEN` 与 `remote_sync_config.json` 中的 `sync_token` 是否一致。
3. 批次过大或网络超时
   - 调小 `batch_size`，调整 `timeout`，或增大脚本运行频率。
4. 远端 DB 字段缺失导致 JSON 解码失败
   - 检查 `sync_errors` 表中错误信息，修复字段或调整脚本字段映射后重试。
5. 同步延迟或重复
   - 检查 `sync_state.json` 中的 last-id 是否更新；检查 Central 去重索引是否存在并正常工作。

---

## 九、我使用给更高级模型（如 GPT-5.1codex）的 prompt（原文）

下面是我用于实现该 PoC 的原始 prompt（原样保留，方便你未来直接提交给更强模型）。

```
目标
- 实现一个最小可行的“远端后端推送到中央后端”的同步方案（PoC）。
- 要求：不部署 agent，不改前端；前端只访问中央（127.0.0.1）。远端（127.0.0.2..127.0.10）各自运行现有后端和数据库，但周期性把新增数据推送到中央。

约束与前提
- 环境：Go 项目（使用 gin），数据库为 MySQL（现有 schema 可用）。
- 不改现有业务逻辑过多，只新增最小后端接收端与远端推送脚本/功能。
- 使用内网，安全用静态 token（HTTP 头 X-SYNC-TOKEN）。
- 要求今天内能做出可运行 PoC。

交付物（明确）
1. Central 后端：
   - 新增 HTTP 接口 POST /sync/events，接收批量增量数据并写入中央 DB。
   - 接收格式（JSON）示例：
     {
       "source_host": "127.0.0.2",
       "table": "honeypot_instances",    // 或 "honeypot_logs"
       "rows": [
         {"remote_id": 123, "created_at":"2025-11-29T12:00:00Z", "data": { ... 原始字段 ... }},
         ...
       ]
     }
   - 行为：
     - 验证 HTTP 头 X-SYNC-TOKEN（与配置中的预共享 token 比对），无 token 返回 401。
     - 对每个行，写入到对应中央表（honeypot_instances 或 honeypot_logs）：
       - 若中央表没有 source_host 和 remote_id 字段，新增两个字段：source_host VARCHAR(64), remote_id BIGINT NULL。
       - 插入时用 ON DUPLICATE KEY 或先检查唯一键（source_host + remote_id）以去重，避免重复写入。
       - 若业务表已有自增主键冲突逻辑，请把远端记录写入保留源 id 的字段 remote_id，并让中央生成自己的主键。
     - 返回 200 并携带写入统计（插入数 / 忽略数 / 错误数）。
   - 提供数据库迁移 SQL（最小化变更）：
     - ALTER TABLE honeypot_instances ADD COLUMN source_host VARCHAR(64) DEFAULT NULL, ADD COLUMN remote_id BIGINT DEFAULT NULL;
     - CREATE UNIQUE INDEX ux_honeypot_instances_source_remote ON honeypot_instances(source_host, remote_id);
     - 同理处理 honeypot_logs（或把日志单独归档表）。
   - 日志与错误处理：将错误行记录到 central 的 sync_errors 表（包含 source_host、payload、error_message、ts）。
   - 单元/集成测试：提供 curl 示例请求和期望响应；提供本地 DB 插入验证步骤。

2. 远端推送脚本（可作为独立小工具或集成到远端后端）
   - 小工具功能：
     - 从本地 DB 查询自上次同步时间或自增 id 之后的新记录（使用 last_sync_timestamp 或 last_sync_id 保存在本地文件或 DB 表）。
     - 批量打包为上述 JSON，POST 到中央 http://127.0.0.1:PORT/sync/events（实际用中央 IP），加 header X-SYNC-TOKEN: <token>。
     - 成功（HTTP 200）后更新本地 last_sync（按最大 remote_id 或最大 created_at）。
     - 失败时写日志并按指数退避重试，下次继续重试（简单实现：若失败，等待下次定时再次尝试，不删除 last_sync）。
   - 可实现成两种形式任选其一（实现者决定）：
     - Python 脚本（简单、易跑），带配置文件（DB 连接、中央 URL、TOKEN、batch size、interval）。
     - Go 小程序（与主项目同语言，便于集成）。
   - 提供运行说明：如何设置定时（cron 示例或 systemd timer）。

3. 验收标准（必须满足）
   - 从远端手动或脚本触发一次同步后，中央数据库表（honeypot_instances 或 honeypot_logs）可查询到来自远端的数据，且 source_host 和 remote_id 值正确。
   - 重复 POST 相同数据不会导致重复插入（去重生效）。
   - 若 central 返回错误（非 2xx），远端脚本不更新 last_sync，因此会在下一次重试仍然尝试发送。
   - 提供简明测试步骤（在本机模拟远端/中央，包含 curl 示例与 SQL 验证语句）。

实现细节与优先级（最小化改动）
- 优先实现 central 的 /sync/events 与 DB migration。
- 同步脚本先做 Python 版 PoC（一页脚本足够），便于在所有远端快速运行。
- 远端用 created_at 或自增 id 做增量过滤；若两者都可用，优先用自增 id（更可靠）。
- Security：在请求头 X-SYNC-TOKEN 里校验静态 token（token 从中央配置或环境变量读取）。
- 不实现大文件上传/样本传输：对于大二进制样本，仅同步 metadata（file_path 或 sha256），若需要后续再扩展。

需要你返回的内容（给我/给后端工程师）
- 对 central 的具体代码补丁（Go, gin），包括 handler、路由注册、DB migration SQL、以及单元测试或手动测试步骤。
- 远端推送脚本（Python 或 Go），以及如何在远端运行（示例 cron entry 或 systemd unit）。
- 简短部署说明：如何在几台机器上配置 token、如何初始化 central DB migration、如何启动远端脚本进行第一次同步。
- 最后给出 5 条最可能遇到的问题及解决方案（例如：ID 冲突、网络超时、token 配置错误、批次太大导致超时、重复数据导致前端显示异常）。

额外说明（供实现者考虑，但非必须）
- 若实现者认为无需修改业务表，可选择创建镜像表（honeypot_instances_remote）把原始远端数据作为 JSON 存储，随后再做映射；但优先方案是尽量写入业务表并加 source_host/remote_id。
- 时间敏感：实现者应优先产出可运行 PoC 并给出明确手动测试脚本。

结束语
- 目标是“今天能运行”的 PoC，请按最小改动与清晰测试路径交付。实现完成后列出如何在 9 台远端机器上快速部署（复制脚本+配置 token+定时任务）。
```

---

## 十、仓库中相关文件（便于查阅）
- `internal/handlers/sync_handler.go`（Central 接收逻辑）
- `internal/repositories/models.go`（模型变更：source_host, remote_id, SyncError）
- `internal/config/config.go`（新增 `SYNC_TOKEN`）
- `routers/router.go`（路由注册）
- `deployment/migrations/20251129_sync_poc.sql`（迁移 SQL）
- `tools/remote_sync.py`（远端推送脚本）
- `tools/remote_sync_config.example.json`（示例配置）
- `docs/sync_poc.md`（PoC 使用说明）


---

如果你现在想直接在一台机器上验证，我可以给出一步步的命令（MySQL 导入、如何填充示例数据、如何运行 remote_sync.py、如何在 Central 上查询验证）。要不要我现在把那份一步步命令（适用于 Windows PowerShell）写到这个文档末尾？
