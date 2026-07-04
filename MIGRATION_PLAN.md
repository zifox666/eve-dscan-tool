# EVE DScan Tool v1 功能总结与 Gin 迁移计划

## 1. 当前项目概览

当前项目是一个基于 FastAPI 的 EVE Online DScan 分析工具，主要用于解析两类扫描数据：

- 本地频道成员列表：识别角色、军团、联盟关系，并生成可分享的分析页面。
- 舰船 DScan 结果：按舰船、旗舰、建筑、其他物品分类统计，并尝试识别扫描所在星系。

项目当前技术栈：

- Web 框架：FastAPI + Starlette middleware
- 模板渲染：Jinja2
- 数据库：SQLite，使用 SQLAlchemy async
- SDE 数据源：Fuzzwork SQLite dump，只读查询
- 外部 API：EVE ESI API
- 缓存：进程内 Python dict TTL cache
- 前端：Tailwind CDN、原生 JavaScript、html2canvas
- 部署：Docker + uvicorn

当前分支规划：

- `v1`：当前 FastAPI 版本存档分支
- `dev`：Gin + PostgreSQL + Redis 迁移开发分支

## 2. 现有功能总结

### 2.1 首页与提交入口

相关文件：

- `main.py`
- `router/index.py`
- `templates/index.html`
- `templates/redirect_form.html.jinja2`

功能：

- 提供统一 DScan 文本输入框。
- 自动识别输入类型：
  - 每行只有一个字段时识别为本地频道成员列表。
  - 否则识别为舰船 DScan。
- 根据识别结果将表单转发到：
  - `POST /c/process`
  - `POST /v/process`
- 支持“仅统计有距离信息的扫描结果”选项，主要用于舰船扫描。
- 支持手动添加侦查舰数量，并在提交前追加到 DScan 文本中：
  - Huginn
  - Lachesis
  - Rook
  - Curse

### 2.2 本地频道成员分析

相关文件：

- `router/local_dscan.py`
- `api/eve_api.py`
- `utils/helpers.py`
- `utils/character_cache.py`
- `templates/local_dscan.html.jinja2`

接口：

- `POST /c/process`
- `GET /c/{short_id}`

处理流程：

1. 将本地频道文本按行解析为角色名列表。
2. 调用 ESI `/universe/ids/` 将角色名转换为角色 ID。
3. 调用 ESI `/characters/affiliation/` 查询角色所属军团和联盟。
4. 调用 ESI `/universe/names/` 查询角色、军团、联盟名称。
5. 整理为角色、军团、联盟三类结构。
6. 生成短链接 ID。
7. 保存原始数据、处理结果、客户端 IP、访问次数。
8. 浏览器请求重定向到结果页，JSON 请求返回短链接信息。

结果页功能：

- 展示总角色数、公司数量、联盟数量、无联盟角色数。
- 三列展示联盟、公司、角色。
- 显示 EVE 官方图片资源：
  - 联盟 logo
  - 军团 logo
  - 角色头像
- 支持 hover/click 过滤关联实体。
- 支持复制当前分享链接。

### 2.3 舰船扫描分析

相关文件：

- `router/ship_dscan.py`
- `db/sde.py`
- `utils/sqlite_helper.py`
- `utils/system_list.py`
- `templates/ship_dscan.html.jinja2`

接口：

- `POST /v/process`
- `GET /v/{short_id}`

处理流程：

1. 按行解析 DScan 文本，字段格式大致为：
   - `type_id`
   - `name`
   - `type_name`
   - `distance`
2. 根据 type ID 查询 EVE SDE SQLite。
3. 同时生成中文和英文两份处理结果。
4. 按类型分类统计：
   - 普通舰船：category ID `6`
   - 旗舰：group ID 在 `[30, 485, 547, 659, 883, 1538, 4594]`
   - 建筑：category ID 在 `[41, 46, 65, 66]`
   - 其他物品
5. 支持过滤无距离信息的扫描项。
6. 从非舰船物品名称中尝试提取星系名。
7. 中文星系名通过 `utils/system_list.py` 映射到 system ID。
8. 英文星系名通过 SDE 查询 system ID。
9. 根据 system ID 查询星座、星域、安全等级。
10. 保存原始数据、处理中英文结果、客户端 IP、访问次数。

结果页功能：

- 展示总数量、舰船数量、旗舰数量、建筑数量。
- 展示识别到的星系、安全等级、星域。
- 分类展示普通舰船、旗舰、建筑、其他物品。
- 支持类型与名称之间的 hover 高亮和过滤。
- 支持复制分享链接。
- 支持中英文结果切换。

### 2.4 短链接与数据持久化

相关文件：

- `db/models/dscan.py`
- `utils/helpers.py`

当前表：

- `local_dscan`
- `ship_dscan`

主要字段：

- `id`
- `short_id`
- `raw_data`
- `processed_data`
- `client_ip`
- `view_count`
- `created_at`
- `updated_at`

短链接生成：

- 使用大小写字母和数字生成随机字符串。
- 默认长度为 10。
- 当前没有显式冲突重试逻辑，依赖数据库唯一约束兜底。

### 2.5 请求日志与真实 IP

相关文件：

- `utils/middleware.py`
- `db/models/log.py`

功能：

- 支持 Cloudflare `CF-Connecting-IP` 头，将真实 IP 写入 request client。
- 每次请求后记录请求日志：
  - client IP
  - path
  - method
  - process time
  - status code
  - created at

### 2.6 缓存

相关文件：

- `utils/cache.py`
- `utils/character_cache.py`

当前缓存均为进程内缓存：

- `dscan_cache`
  - 缓存扫描结果页数据。
  - 默认最大 128 条，TTL 3600 秒。
- `character_cache`
  - 缓存角色名到 ID。
  - 缓存角色 affiliation。
  - 缓存实体 ID 到名称。
  - 默认 TTL 7 天。

迁移后应将这些缓存统一迁入 Redis，以支持多实例部署和重启后保留热点数据。

### 2.7 国际化与前端能力

相关文件：

- `static/js/i18n.js`
- `static/js/translations.js`
- `templates/base.html`

功能：

- 支持中文和英文。
- 语言优先级：
  - localStorage
  - cookie
  - 浏览器语言
  - 默认中文
- 切换语言时写入 cookie 和 localStorage，并刷新页面。
- 支持暗黑模式。
- 支持整页截图下载。
- 使用 Tailwind CDN 和 html2canvas CDN。
- 有 Google Analytics 脚本。

## 3. 当前架构问题与迁移关注点

1. 进程内缓存不适合多实例。
2. SQLite 主业务库不适合并发写入和长期扩展。
3. SDE 查询仍依赖本地 SQLite 文件，和主业务数据割裂。
4. `Base.metadata.create_all` 缺少正式迁移管理。
5. 短链接生成没有应用层冲突重试。
6. 请求日志同步写数据库，可能拖慢请求响应。
7. 业务逻辑、路由、数据整理耦合较高，迁移时应拆成 service 层。
8. 前端模板逻辑较重，后续可以先保持服务端渲染，再逐步拆前端。
9. HTTP ESI 请求重试日志中引用 `response`，异常发生在 response 创建前时存在潜在问题。
10. `utils/sqlite_helper.py` 查询的 `types` 表结构和 `db/sde.py` 查询的 Fuzzwork 原始表结构不一致，迁移前需要确认实际运行依赖哪一套 SDE 文件。

## 4. 目标技术栈

- Go Web 框架：Gin
- 主数据库：PostgreSQL
- 缓存：Redis
- 数据访问：
  - 主业务库使用 GORM + PostgreSQL driver
  - Redis 推荐 `go-redis`
- 模板：
  - 第一阶段保留服务端渲染，可使用 Go `html/template`
  - 第二阶段再评估是否拆成独立前端
- 数据迁移：
  - 使用 GORM AutoMigrate 创建当前阶段业务表
- 部署：
  - Docker Compose 提供 app + postgres + redis

## 5. 建议的新模块结构

```text
cmd/server/main.go
internal/config/
internal/http/
internal/http/handlers/
internal/http/middleware/
internal/service/dscan/
internal/service/esi/
internal/service/sde/
internal/store/postgres/
internal/store/redis/
internal/model/
internal/templates/
internal/static/
migrations/
scripts/
```

职责拆分：

- `handlers`：解析请求、返回 HTML/JSON。
- `service/dscan`：本地扫描和舰船扫描的核心业务逻辑。
- `service/esi`：EVE ESI API 客户端、重试、限流、缓存封装。
- `service/sde`：SDE 类型、分组、分类、星系查询。
- `store/postgres`：短链接、扫描记录、请求日志等持久化。
- `store/redis`：扫描结果缓存、角色缓存、ESI 响应缓存。
- `middleware`：Cloudflare IP、请求日志、panic recovery、request ID。

## 6. PostgreSQL 数据模型建议

### 6.1 扫描记录

可以先保留两张表，降低迁移风险：

- `local_dscans`
- `ship_dscans`

建议字段：

- `id BIGSERIAL PRIMARY KEY`
- `short_id VARCHAR(16) UNIQUE NOT NULL`
- `raw_data TEXT NOT NULL`
- `processed_data JSONB NOT NULL`
- `client_ip INET`
- `view_count BIGINT NOT NULL DEFAULT 0`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT now()`

后续也可以合并为一张 `dscans`：

- `scan_type TEXT CHECK (scan_type IN ('local', 'ship'))`
- `processed_data JSONB`

第一阶段不建议立即合并，避免影响 API 兼容。

### 6.2 请求日志

表名：`request_logs`

建议字段：

- `id BIGSERIAL PRIMARY KEY`
- `client_ip INET`
- `request_path TEXT NOT NULL`
- `request_method VARCHAR(10) NOT NULL`
- `process_time_ms INTEGER NOT NULL`
- `status_code INTEGER NOT NULL`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`

建议：

- 请求日志可以异步写入，避免阻塞响应。
- 增加按时间分区或定期清理策略。

### 6.3 SDE 数据

- 采用独立维护的pgsql数据库实例 使用和data一致的数据库账密地址但是连接的数据库不同

## 7. Redis 缓存设计建议

### 7.1 扫描结果缓存

Key：

- `dscan:local:{short_id}`
- `dscan:ship:{lang}:{short_id}`

TTL：

- 1 小时或更长，按访问量调整。

注意：

- `view_count` 不建议完全依赖缓存值。
- 查看详情时 PostgreSQL 原子自增，然后把最新 view count 回填响应。

### 7.2 ESI 缓存

Key：

- `esi:character:id:{name_hash}`
- `esi:character:affiliation:{character_id}`
- `esi:entity:name:{entity_id}`

TTL：

- 角色、军团、联盟名称：7 天到 30 天。
- affiliation：建议 1 天到 7 天，根据实时性需求取舍。

### 7.3 短链接防冲突

可选：

- 使用 PostgreSQL 唯一索引 + 插入失败重试。
- 或 Redis `SETNX shortid:{id}` 做短时间预占位。

第一阶段推荐数据库唯一索引 + 最多重试 5 次，简单可靠。

## 8. API 兼容计划

迁移后建议保持以下 URL 不变：

- `GET /`
- `POST /submit`
- `POST /c/process`
- `GET /c/:short_id`
- `POST /v/process`
- `GET /v/:short_id`

响应兼容：

- HTML 默认响应。
- 请求头包含 `Accept: application/json` 时返回 JSON。
- 保持当前 JSON 结构：
  - `code`
  - `msg`
  - `data`

语言兼容：

- 继续读取 `lang` cookie。
- 支持 `zh` 和 `en`。

## 9. 分阶段迁移计划

### Phase 0：冻结 v1 与建立迁移基线

状态：已完成分支准备。

- 创建 `v1` 分支存档当前版本。
- 创建并切换到 `dev` 分支。
- 新增本文档作为迁移蓝图。

### Phase 1：Go 项目骨架

状态：已完成。

目标：Gin 服务能启动，并提供基础页面和健康检查。

任务：

- [x] 初始化 Go module。
- [x] 引入 Gin、GORM、go-redis。
- [x] 建立 `cmd/server` 和 `internal` 目录。
- [x] 增加配置读取：
  - `DATABASE_URL`
  - `REDIS_ADDR`
  - `ESI_BASE_URL`
  - `SDE` 相关配置
- [x] 增加 `/healthz`。
- [x] 增加 `/readyz`，检查 PostgreSQL 和 Redis。
- [x] 增加 Go Dockerfile 和 docker-compose。

验收：

- [x] `go test ./...` 通过。
- [x] `go build ./cmd/server` 通过。
- [x] `docker compose config` 通过。
- [ ] 本地 `docker compose up` 启动 app、PostgreSQL、Redis。

### Phase 2：PostgreSQL schema 与存储层

状态：已完成。

目标：替换主业务 SQLite。

任务：

- [x] 使用 GORM AutoMigrate 创建业务表。
- [x] 实现 local dscan、ship dscan、request log repository。
- [x] 实现短链接生成和唯一冲突重试。
- [x] 实现 view count 事务内原子自增。
- [x] 服务启动时加载 `.env.local` / `.env`。
- [x] `/readyz` 校验 PostgreSQL 和 Redis。

验收：

- [x] `go test ./...` 通过。
- [x] 本地临时启动服务后 `/readyz` 返回 PostgreSQL 和 Redis 均为 ok。
- [x] AutoMigrate 已在本地启动路径执行。
- [ ] 后续接入业务 handler 后验证真实扫描记录写入和读取。

### Phase 3：Redis 缓存层

状态：已完成。

目标：替换进程内缓存。

任务：

- [x] 实现 Redis JSON cache interface。
- [x] 迁移扫描结果缓存封装。
- [x] 迁移角色名、affiliation、实体名缓存封装。
- [x] 添加 JSON 序列化和 TTL 管理。
- [x] 增加缓存 key 单元测试。
- [x] 增加 `.env.example` TTL 配置。

验收：

- [x] Redis cache wrapper 通过 `(found, error)` 暴露 miss/error，业务层可忽略缓存错误并降级查询 PostgreSQL/ESI。
- [x] 多实例可共享同一 Redis key 空间。
- [x] `go test ./...` 通过。
- [x] `go build ./cmd/server` 通过。
- [x] 本地临时启动服务后 `/readyz` 返回 PostgreSQL 和 Redis 均为 ok。

### Phase 4：ESI 客户端迁移

状态：已完成。

目标：复刻本地频道分析能力。

任务：

- [x] 实现 ESI HTTP client。
- [x] 支持 timeout、重试、批处理。
- [x] 接入 Redis cache，缓存失败不阻断 ESI 请求。
- [x] 实现：
  - `POST /universe/ids/`
  - `POST /characters/affiliation/`
  - `POST /universe/names/`
- [x] 实现本地频道成员分析 service。
- [x] 加入错误处理和单元测试。

验收：

- [x] fake ESI 单元测试覆盖角色、军团、联盟整理逻辑。
- [x] `go test ./...` 通过。

### Phase 5：SDE 数据迁移

状态：已完成。

目标：复刻舰船扫描分析能力。

任务：

- [x] 采用独立维护的 PostgreSQL SDE 实例，通过 `SDE_URL` 连接。
- [x] 主业务库继续 AutoMigrate；SDE 库只读连接，不迁移 schema。
- [x] 实现 type/group/category translation 查询。
- [x] 实现星系、星座、星域、安全等级查询。
- [x] 实现舰船 DScan 解析与分类统计：
  - 普通舰船
  - 旗舰
  - 建筑
  - 其他物品
- [x] 实现星系名提取和 SDE 查询。
- [x] `/readyz` 增加 SDE 连接检查。

验收：

- [x] fake SDE 单元测试覆盖舰船分类、距离过滤、星系识别。
- [x] `RUN_SDE_INTEGRATION=1 go test -run TestIntegrationSDEQueries ./internal/service/sde` 通过。
- [x] 本地临时启动服务后 `/readyz` 返回 PostgreSQL、Redis、SDE 均为 ok。

### Phase 6：模板与前端迁移

目标：保持用户体验基本不变。

任务：

- 将 Jinja2 模板迁移为 Go `html/template`。
- 保留现有静态 JS 行为：
  - i18n
  - 暗黑模式
  - 截图
  - 复制链接
  - hover/click 筛选
  - 手动添加侦查舰
- 评估 Tailwind CDN 是否改为构建产物。

验收：

- 首页、local 结果页、ship 结果页视觉和交互可用。
- 中英文切换可用。

### Phase 7：兼容性测试与数据迁移

目标：迁移上线前可对比验证。

任务：

- 准备典型测试样本：
  - local scan
  - ship scan
  - 带距离过滤 ship scan
  - 中文星系
  - 英文星系
  - 无法识别 type ID
- 为解析、整理、短链接、缓存、repository 添加测试。
- 编写 SQLite 到 PostgreSQL 的历史数据迁移脚本。
- 对比 v1 与 Go 版 JSON 输出。

验收：

- 核心样本输出一致或差异已记录。
- 历史短链接可继续访问。

### Phase 8：上线与回滚

目标：可控切换。

任务：

- 先灰度部署 Go 服务。
- PostgreSQL 和 Redis 做备份策略。
- 保留 v1 镜像和 `v1` 分支作为回滚点。
- 增加应用日志、慢请求日志、ESI 错误率监控。

验收：

- 新服务稳定处理真实流量。
- 出现问题可切回 v1。

## 10. 优先级建议

第一优先级：

- 分支归档。
- Go 项目骨架。
- PostgreSQL schema。
- 本地频道分析链路。
- 舰船扫描核心分类链路。

第二优先级：

- Redis 缓存。
- SDE PostgreSQL 导入。
- 请求日志异步化。
- API 兼容测试。

第三优先级：

- 前端构建体系优化。
- 管理后台或历史记录。
- API key 认证。
- 更多语言支持。

## 11. 迁移时建议保留的兼容行为

- 短链接路径继续使用 `/c/{short_id}` 和 `/v/{short_id}`。
- JSON 响应继续使用 `code/msg/data`。
- HTML 和 JSON 共用同一业务数据结构。
- ship processed data 继续保存 `zh` 和 `en` 双语结构。
- `lang` cookie 继续控制 ship 结果页语言。
- `filter_distance` 继续影响统计结果。
- 当前短链接长度默认保持 10。

## 12. 需要在开工前确认的问题

1. 是否需要迁移已有 `dscan.sqlite` 历史短链接数据？ 不迁移
2. SDE 是否必须完全导入 PostgreSQL，还是允许短期继续读 SQLite？ sde采用独立维护 本项目只考虑连接给出的地址
3. 新版是否继续服务端渲染，还是同时拆前端 SPA？前端单独挂载到nginx
4. 请求日志是否需要长期保留？如果需要，保留多久？ 30days
5. ESI affiliation 缓存 TTL 期望是更实时还是更省请求？ 7days in redis / 90days in pgsql (分期过期: 高可信 -> 紧急调用)
6. 是否需要新增 API key 认证，还是等 Go 迁移完成后再做？ 后续再加
