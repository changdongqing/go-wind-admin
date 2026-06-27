# 数据库脚本版本化升级方案 — 设计文档（spec）

| 字段 | 值 |
|---|---|
| 文档版本 | v1.0 |
| 撰写日期 | 2026-06-26 |
| 作者 | ttshang（架构）|
| 范围 | GoWind Admin（go-wind-admin）后端 + React 前端 |
| 状态 | 已 brainstorm 通过，待复核 |
| 目标数据库 | PostgreSQL（>= 14；当前生产为 PG，无需向后兼容 MySQL/SQLite 作为生产目标） |

> 本文档只描述**方案设计**，不含实现代码。实现路线图、PR 拆解、任务清单将由 `writing-plans` 阶段产出的 implementation plan 承担。

---

## 0 决策快照（Q1–Q9）

本设计建立在如下九个已对齐的架构决策之上。后续每个章节的细节都从这些决策衍生，修改任何一条都意味着方案需要重做：

| 编号 | 决策 | 结论 |
|---|---|---|
| Q1 | 与 Ent auto-migrate 的关系 | **完全替代**：关闭 `client.Schema.Create()`；所有 DDL 走版本化 `.sql` |
| Q2 | Ent schema 在新体系中的角色 | **A1**：Ent schema 仍是源头；Atlas 在开发期对 schema 做 diff 生成 `.sql` |
| Q3 | 数据库方言范围 | **Postgres-only**（MySQL/SQLite 在生产部署中下线，仅保留 Ent codegen 测试通道） |
| Q4 | Go 端种子如何融入版本化 | **B**：`.sql` 与 Go-fn 双形态，共用同一张版本表与同一时间线 |
| Q5 | 迁移运行时底座 | **B**：Atlas（开发期生成）+ pressly/goose（运行时执行）；以 Go 库内嵌方式集成，不调子进程 |
| Q6 | 迁移触发模型 | **C**：默认 CLI / UI 手动触发；启动期仅做兼容性校验（fail-fast）；可通过 `auto_migrate: true` 开启 dev 自动 apply |
| Q7 | v1 基线策略 | **B**：`pg_dump --schema-only` 整理为幂等 DDL（`IF NOT EXISTS`、`DO $$ … $$`），新旧环境一律真实执行，不使用 fake-apply |
| Q8 | 回滚契约边界 | **B**：三态契约 `reversible / irreversible / best_effort`，CI lint 强制声明 |
| Q9 | UI 落地范围 | **C**：仅在 `frontend/admin/react` 实现（Vue Element / Vue Vben 不在本期范围） |

---

## 1 目标与非目标

### 1.1 目标（Goals）

1. **统一时间线**：所有 schema DDL、数据修复 DML、系统预置种子（含密码哈希等需 Go 逻辑的步骤）按 `yyyymmddhhmmss_<desc>` 单调时间戳排成同一条版本线。
2. **双形态脚本**：同一时间线既能容纳 `.sql` 文件，也能容纳 `.go` Go-fn 迁移；由同一 runner 调度。
3. **三态回滚契约**：每个版本明示 `reversible / irreversible / best_effort`；UI 直观显示风险，不可逆迁移的回滚按钮置灰并显示原因。
4. **可视化运维台**：在 React Admin 中新增"数据库升级"模块，可查看版本列表、状态、checksum、应用人、耗时、错误堆栈、操作审计；具备 `sys:platform_admin` 权限的人可在 UI 触发 up / down / repair。
5. **CLI 同源**：`gow migrate <subcommand>` 与 UI 共用同一份 Go 库代码，不通过 shell out，灾备/无 UI 场景仍可工作。
6. **启动期解耦**：应用启动**不**自动 apply，只做"DB 已应用版本 ≥ 应用代码要求最低版本"的兼容性校验，校验失败 fail-fast，校验通过即可正常起服务；`dev/local` 环境可通过开关恢复"启动自动 apply"。
7. **现有库平滑过渡**：基线脚本采用幂等 DDL，新旧环境一律走 `goose up` 真实执行，无须 fake-apply。
8. **现有 Go 种子（`default_data.go` + `seed/*.go`）成为一等公民迁移**：每个种子拆为独立版本号，失败可精确定位、可重试、可回滚（best-effort）。

### 1.2 非目标（Non-goals）

- ❌ 不支持 MySQL / SQLite 作为**生产**目标。Ent codegen 层 SQLite 单元测试通道保留。
- ❌ 不在本期为 Vue Element 与 Vue Vben 实现 UI。
- ❌ 不接管租户级数据搬运 / PITR / 物理备份；不可逆迁移前的备份由 SRE runbook 保证。
- ❌ 不实现"在线零停 DDL"；需要时由迁移作者用 `CREATE INDEX CONCURRENTLY` 等手段并显式声明 `tx: false`。
- ❌ 不引入新的事件总线 / 消息队列；UI 进度采用 BFF SSE + 短轮询兜底。

### 1.3 术语表（Glossary）

| 术语 | 含义 |
|---|---|
| **Migration** | 一个版本号下的单次变更，载体可以是 `.sql` 或 `.go` |
| **Version ID** | `yyyymmddhhmmss` 14 位时间戳（UTC），单调递增，全局唯一 |
| **Kind** | `sql` ｜ `go` |
| **Direction** | `up` ｜ `down` ｜ `repair` |
| **Status** | `pending` ｜ `running` ｜ `applied` ｜ `failed` ｜ `rolled_back` ｜ `skipped` |
| **Reversibility** | `reversible` ｜ `irreversible` ｜ `best_effort` |
| **Baseline / v1** | 第一份基线脚本，由 `pg_dump --schema-only` 整理为幂等 DDL |
| **Runner** | 执行迁移的 Go 包 `pkg/migrate`，封装 goose 库 |
| **Orchestrator BFF** | `MigrationService` BFF 服务，UI 与 CLI 都通过它操作 Runner |
| **Actor** | 触发本次操作的主体，格式 `cli:user@host` / `ui:user-id=42` / `startup-autoapply` / `ci` |

---

## 2 总体架构

### 2.1 端到端数据流

```
┌───────────────────────────────────────────────────────────────────────────┐
│  Developer Workflow                                                       │
│                                                                           │
│  edit ent/schema/*.go          gow migrate new "add foo"                  │
│         │                              │                                  │
│         │ ent generate                 │                                  │
│         ▼                              ▼                                  │
│  api/.../ent/schema.go         migrations/yyyymmddhhmmss_<desc>.go        │
│         │                                                                 │
│         │ gow migrate diff "<desc>"  (Atlas — only at dev time)           │
│         ▼                                                                 │
│  migrations/yyyymmddhhmmss_<desc>.sql + .down.sql                         │
│         │                                                                 │
│         ▼                                                                 │
│  git commit + PR  →  CI: up→down→up replay test + lint                    │
└───────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌───────────────────────────────────────────────────────────────────────────┐
│  Runtime                                                                  │
│                                                                           │
│   ┌────────────────────────────────────────────────────────────────────┐  │
│   │  pkg/migrate (Go library — single source of truth)                 │  │
│   │  ┌──────────────────────────────────────────────────────────┐      │  │
│   │  │   Runner: wraps goose                                    │      │  │
│   │  │   - Discovers .sql + Go-fn migrations from embed.FS      │      │  │
│   │  │   - pg_advisory_lock(<bigint hash>)  ── multi-replica   │      │  │
│   │  │   - Writes goose_db_version + schema_migration_audit     │      │  │
│   │  │   - Emits structured logs + Prometheus metrics           │      │  │
│   │  └──────────────────────────────────────────────────────────┘      │  │
│   └───┬────────────────────┬─────────────────────────────────┬──────────┘  │
│       │                    │                                 │             │
│       ▼                    ▼                                 ▼             │
│  ┌─────────┐         ┌──────────────┐                 ┌──────────────┐    │
│  │  gow    │         │ Kratos       │                 │ MigrationSvc │    │
│  │ migrate │         │ startup hook │                 │  BFF / gRPC  │    │
│  │  CLI    │         │ (version     │                 │  + REST      │    │
│  │         │         │  compat only)│                 └──────┬───────┘    │
│  └─────────┘         └──────────────┘                        │            │
│                                                              ▼            │
│                                                  ┌────────────────────┐   │
│                                                  │ React Admin UI     │   │
│                                                  │ /system/migrations │   │
│                                                  └────────────────────┘   │
└───────────────────────────────────────────────────────────────────────────┘
```

### 2.2 关键不变量

- **No shell-out**：`gow migrate` CLI、Kratos 启动钩子、`MigrationService` BFF 三个入口共享 `pkg/migrate` 包；不存在"UI 调 shell 调 binary"这类多进程路径。
- **Single advisory lock**：所有写入口（up/down/repair）先 `SELECT pg_try_advisory_lock(L)`（`L = hashtext('gow_migrate')::bigint`），失败立即返回"另一个迁移正在进行"。
- **Atlas only at dev time**：运行时**不依赖** Atlas 二进制；Atlas 仅在开发机和 CI 上参与 `.sql` 生成与 replay 校验。
- **goose 作为内部库**：通过 `github.com/pressly/goose/v3` 的 Go API 嵌入 `pkg/migrate`，不调 goose CLI。
- **Three entries, one path**：CLI / startup hook / BFF 调用同一个 `Runner` 方法；不同入口只是身份（actor）与权限（authz）不同。

---

## 3 仓库布局

```
backend/
├── migrations/                         ← 唯一的迁移源目录（embed.FS 打包进 binary）
│   ├── README.md                       ← 命名规则 / 提交流程 / lint 规则
│   ├── 20260626120000_baseline.sql        ← v1：pg_dump 整理后的幂等 DDL
│   ├── 20260626120000_baseline.down.sql   ← v1 反向（best_effort，DROP IF EXISTS 全套）
│   ├── 20260626120100_seed_languages.go   ← v2：语言种子（Go-fn）
│   ├── 20260626120200_seed_permission_groups.go
│   ├── 20260626120300_seed_permissions.go
│   ├── 20260626120400_seed_menus.go
│   ├── 20260626120500_seed_roles.go
│   ├── 20260626120600_seed_admin_user.go
│   ├── 20260626120700_seed_thingmodel_units.go
│   ├── 20260626120800_seed_thingmodel_features.go
│   ├── 20260626120900_create_partial_index_unit_base.sql
│   └── 20260626120900_create_partial_index_unit_base.down.sql
│
├── pkg/migrate/                        ← Runner 库（CLI / BFF / startup hook 共用）
│   ├── runner.go                       ← Up / Down / UpTo / DownTo / Status / Repair
│   ├── registry.go                     ← Go-fn 注册表（按 version_id 索引）
│   ├── lock.go                         ← pg_advisory_lock 封装
│   ├── audit.go                        ← 写 schema_migration_audit 旁表
│   ├── lint.go                         ← 静态扫描 .sql / .go 违规（CI 入口）
│   ├── replay_test.go                  ← CI: up→down→up 全套 replay
│   └── embed.go                        ← go:embed migrations/*
│
├── cmd/gowmigrate/                     ← gow CLI 子命令载体
│   ├── main.go
│   └── commands.go                     ← new / diff / up / down / upto / downto / status / repair / verify
│
└── app/admin/service/internal/
    ├── service/migration_service.go    ← MigrationService BFF（gRPC + REST）
    └── data/ent_client.go              ← 改造：去掉 client.Schema.Create，改为 runner.VerifyCompat(...)

backend/api/protos/admin/service/v1/
└── i_migration.proto                   ← MigrationService BFF 契约

frontend/admin/react/src/
├── api/hooks/useMigrations.ts          ← 由 make ts 生成 client 后封装
├── views/system/migrations/
│   ├── MigrationsListPage.tsx
│   ├── MigrationDetailDrawer.tsx
│   ├── ConfirmUpToDialog.tsx
│   ├── ConfirmRollbackDialog.tsx       ← 含三态契约的红/黄/绿提示
│   └── AuditTimeline.tsx
└── router/routes/system.tsx            ← 新增路由 `/system/migrations`
```

**关于 `pkg/constants/default_data.go` 与 `internal/data/seed/*.go`**：
保留作为**被 Go-fn 迁移调用的纯数据源**。迁移文件本身只写"调用 + 幂等 upsert"的薄壳，不复制粘贴种子数据本身。`MenuService.init() / PermissionService.init() / RoleService.init() / UserService.init() / LanguageService.init()` 等启动期注入路径将下线（见 §5.3）。

---

## 4 迁移文件规范

### 4.1 命名（CI 强制）

```
yyyymmddhhmmss_<snake_case_desc>.{sql,go}
yyyymmddhhmmss_<snake_case_desc>.down.sql      （仅 .sql 形态需要伴生 down 文件）
```

- 时间戳 UTC，由 `gow migrate new` / `gow migrate diff` 自动注入；不允许人工命名。
- `<desc>` 必须 snake_case，长度 ≤ 60，禁止中文、空格、连字符以外的标点。
- 同一时间戳**只能出现一个版本**（同一秒新建两个时 CLI 自动 +1 秒）。

### 4.2 `.sql` 文件头（CI 强制注解）

```sql
-- migrate:version            20260626120900
-- migrate:author              <git-user.email>
-- migrate:kind                sql
-- migrate:reversibility       reversible | irreversible | best_effort
-- migrate:irreversible_reason ""                       (irreversible 必填)
-- migrate:data_loss           false | true             (best_effort 时若 true 必须填一段说明)
-- migrate:tx                  true | false             (默认 true；CREATE INDEX CONCURRENTLY 等必须 false)
-- migrate:depends_on          []                       (可选，显式依赖的 version_id 列表)
-- migrate:description         "create unit_base partial index"

-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX IF NOT EXISTS uix_thingmodel_unit_base
  ON thingmodel_units (tenant_id, category_id)
  WHERE is_base = true;
-- +goose StatementEnd
```

伴生 `.down.sql`：

```sql
-- migrate:version   20260626120900
-- migrate:direction down
-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS uix_thingmodel_unit_base;
-- +goose StatementEnd
```

### 4.3 Go-fn 迁移接口

```go
// backend/migrations/20260626120600_seed_admin_user.go
package migrations

import (
    "context"
    "database/sql"

    "github.com/pressly/goose/v3"

    "go-wind-admin/pkg/migrate/mctx"
)

func init() {
    mctx.Register(mctx.Migration{
        Version:        20260626120600,
        Description:    "seed admin user with bcrypt password",
        Author:         "ttshang@example.com",
        Reversibility:  mctx.BestEffort,
        DataLossReason: "down 仅删除 username='admin' 的行；若该行后续被改名则无法识别。",
        Up:             upSeedAdminUser,
        Down:           downSeedAdminUser,
    })
    // 同时注册到 goose（goose 才真正调度它）
    goose.AddMigrationContext(upSeedAdminUser, downSeedAdminUser)
}

func upSeedAdminUser(ctx context.Context, tx *sql.Tx) error {
    // bcrypt(constants.DefaultAdminPassword) + INSERT ... ON CONFLICT(username,tenant_id) DO NOTHING
    return nil
}

func downSeedAdminUser(ctx context.Context, tx *sql.Tx) error {
    // DELETE FROM users WHERE username='admin' AND tenant_id=0
    return nil
}
```

要点：
- `Reversibility` 三态枚举与 `.sql` 注解 1:1 对应。
- `mctx.Register` 把**元数据**塞进自家注册表，供 BFF/UI 读取；`goose.AddMigrationContext` 才真正让 goose 跑它。两个调用都在 `init()` 中完成。
- Go-fn 迁移**默认事务**（goose 默认）；需无事务时用 `goose.AddMigrationNoTxContext`，同时元数据 `Tx: false`。
- Go-fn 的 checksum 取"go 源文件 + 所有它静态引用的常量"的 sha256（实现期细化）；调用 `seed.SeedThingmodelUnits` 等服务层代码时，checksum **不**追踪服务层代码变化——服务层的责任是"保持 upsert 语义稳定"。

### 4.4 CI 强制 lint 规则（`pkg/migrate/lint.go`）

CI 在 PR 检查阶段独立运行 `go test ./pkg/migrate -run TestLint`，规则：

1. 文件名格式（regex）；
2. `.sql` 文件头注解完备性（必填字段缺失即失败）；
3. **危险操作扫描**：以下模式命中即要求 `reversibility != reversible`：
   - `DROP\s+TABLE`、`DROP\s+COLUMN`、`DROP\s+TYPE`、`DROP\s+SCHEMA`
   - `TRUNCATE`、`DELETE\s+FROM\s+\w+\s*;`（无 WHERE）
   - `ALTER\s+TYPE\s+\w+\s+DROP\s+VALUE`
   - `ALTER\s+COLUMN\s+\w+\s+TYPE` 含非 USING 子句的隐式转换
4. `.sql` 必须含 `-- +goose Up`；伴生 `.down.sql` 必须含 `-- +goose Down`；
5. 不允许引用未声明的 `depends_on` 版本号；
6. Go-fn 迁移：`mctx.Register` 与 `goose.AddMigration*` 必须**配对出现**，且元数据 `Version` 与文件名时间戳一致。

### 4.5 replay 测试（CI）

`go test ./pkg/migrate -run TestReplay`：
- 在 CI 内拉起临时 Postgres 容器；
- 顺序 `up` 到 HEAD，记录 schema HCL（`atlas schema inspect`）；
- 顺序 `down` 到 0（跳过 `irreversible`，对 `best_effort` 加 `--allow-data-loss` 跳过断言）；
- 再 `up` 到 HEAD；
- 对比首末两次 schema HCL，diff 必须为空。

---

## 5 版本状态模型

### 5.1 两张表

**`goose_db_version`**（goose 自带，结构不动）：

```
id           BIGSERIAL PRIMARY KEY
version_id   BIGINT  NOT NULL          -- yyyymmddhhmmss
is_applied   BOOLEAN NOT NULL
tstamp       TIMESTAMP DEFAULT now()
```

> goose 自身只有"成功记录"与"无记录"两态；失败/审计字段在旁表补全。

**`schema_migration_audit`**（自家表，承载丰富审计字段；UI 主要读这张）：

```sql
CREATE TABLE schema_migration_audit (
    id                  BIGSERIAL  PRIMARY KEY,
    version_id          BIGINT     NOT NULL,
    direction           TEXT       NOT NULL CHECK (direction IN ('up','down','repair')),
    kind                TEXT       NOT NULL CHECK (kind IN ('sql','go')),
    status              TEXT       NOT NULL CHECK (status IN ('running','applied','failed','rolled_back','skipped')),
    checksum            TEXT       NOT NULL,
    reversibility       TEXT       NOT NULL CHECK (reversibility IN ('reversible','irreversible','best_effort')),
    description         TEXT,
    started_at          TIMESTAMPTZ NOT NULL,
    finished_at         TIMESTAMPTZ,
    duration_ms         BIGINT,
    actor               TEXT       NOT NULL,
    actor_source        TEXT       NOT NULL CHECK (actor_source IN ('cli','ui','startup','ci')),
    error_message       TEXT,
    error_stack         TEXT,
    correlation_id      TEXT,
    schema_hash_before  TEXT,
    schema_hash_after   TEXT,
    UNIQUE (version_id, started_at)
);
CREATE INDEX idx_smq_version    ON schema_migration_audit (version_id);
CREATE INDEX idx_smq_started_at ON schema_migration_audit (started_at DESC);
```

### 5.2 状态机

```
                    ┌────────────────┐
                    │   pending      │  ← 文件存在但 goose_db_version 中无记录
                    └───────┬────────┘
                            │ start up
                            ▼
                    ┌────────────────┐  ── error ──▶  ┌─────────────┐
                    │   running      │                │  failed     │
                    └───────┬────────┘                └──────┬──────┘
                            │ success                        │ repair: mark_applied / retry_up
                            ▼                                ▼
                    ┌────────────────┐                ┌─────────────┐
                    │   applied      │                │  applied    │
                    └───────┬────────┘                └─────────────┘
                            │ down (only if reversibility != irreversible)
                            ▼
                    ┌────────────────┐
                    │ rolled_back    │
                    └────────────────┘
```

- `skipped` 仅由 `repair --mark-skipped` 使用：紧急维护窗口里跳过某个非关键种子。
- 存在 `failed` 行时**阻塞**后续 `up`：runner 检测到 HEAD 之前有未恢复的 `failed`，整次 `up` 直接报错；UI 显示红色横幅"v20260626120300 失败，请先 Repair"。

### 5.3 checksum 漂移策略

- 每次 `up` 前对文件读 sha256，与 `schema_migration_audit` 最近一次 applied 行的 checksum 比对：
  - **strict（默认）**：漂移则启动 / `up` fail-fast；
  - **warn**：仅打告警 + UI 顶部黄条；
  - **ignore**：完全忽略（仅灾备恢复期短时使用，配置项需带 `ttl` 时间戳，过期回落 strict）。
- 配置项位置：`data.yaml -> data.database.migrate.checksum_policy`。

---

## 6 并发与多副本安全

### 6.1 Postgres advisory lock

```go
// pkg/migrate/lock.go
const lockKey int64 = 0x676F77_6D696772
func (r *Runner) withLock(ctx context.Context, fn func() error) error {
    var got bool
    if err := r.db.QueryRowContext(ctx,
        "SELECT pg_try_advisory_lock($1)", lockKey).Scan(&got); err != nil {
        return err
    }
    if !got { return ErrLockBusy }
    defer r.db.ExecContext(ctx, "SELECT pg_advisory_unlock($1)", lockKey)
    return fn()
}
```

- 所有**写**入口（Up/Down/Repair）一律包在 `withLock` 中。
- 会话级锁；session 断开自动释放，不会残留。
- 多副本部署时第一副本独占执行，其它副本立即拿到 `ErrLockBusy`，UI 显示"另一节点正在执行迁移，请稍候"。

### 6.2 启动期版本兼容性校验（不持锁）

```go
// pkg/migrate/compat.go
const MinRequiredVersion int64 = /* codegen 注入：当前 release 要求的最低 db 版本 */

func VerifyCompat(ctx context.Context, db *sql.DB) error {
    var cur sql.NullInt64
    if err := db.QueryRowContext(ctx,
        "SELECT MAX(version_id) FROM goose_db_version WHERE is_applied").Scan(&cur); err != nil {
        return err
    }
    if !cur.Valid || cur.Int64 < MinRequiredVersion {
        return fmt.Errorf("database schema is behind: required >= v%d, current = v%d; run `gow migrate up`",
            MinRequiredVersion, cur.Int64)
    }
    return nil
}
```

- 启动钩子（`NewEntClient` 或新的 `NewMigrationVerifier`）调用 `VerifyCompat`，失败 `log.Fatal`。
- 该路径**只读、不持锁**，多副本启动无竞争。
- `MinRequiredVersion` 由构建脚本在编译期写入（取 `migrations/` 中 HEAD 之前**最近一个 schema 改动版本**，种子失败不阻塞启动）。

### 6.3 长事务保护

- `.sql` 默认单文件单事务；
- 文件头 `migrate:tx false` 会被 goose 翻译成 `-- +goose NO TRANSACTION`；
- 无事务脚本必须标 `reversibility: best_effort`（CI lint 强制）。

---

## 7 开发者工作流

### 7.1 新增一张表（DDL）

```bash
# 1. 编辑 ent schema
$EDITOR backend/app/admin/service/internal/data/ent/schema/widget.go

# 2. 运行 ent codegen（不变）
cd backend && make ent

# 3. 由 Atlas diff 当前 schema 与 DB 状态，生成版本化 .sql
gow migrate diff "add widget table"
#   → 产出 backend/migrations/20260626130000_add_widget_table.sql
#   → 同时产出 .down.sql（Atlas 反向推断；reversibility=reversible 默认）

# 4. 人工 review .sql / .down.sql，调整或补 IF NOT EXISTS

# 5. 本地试跑
gow migrate up
gow migrate status

# 6. PR
git add backend/migrations/ && git commit -m "feat: add widget table"
```

### 7.2 新增 Go-fn 种子

```bash
gow migrate new go "seed widget_categories"
#   → 产出 backend/migrations/20260626140000_seed_widget_categories.go 骨架
#   → 同时把 import / init / mctx.Register / goose.AddMigrationContext 都填好

$EDITOR backend/migrations/20260626140000_seed_widget_categories.go
gow migrate up && gow migrate status
```

### 7.3 无事务 DDL（CREATE INDEX CONCURRENTLY 等）

```bash
gow migrate new sql --no-tx "concurrent index on widgets created_at"
```

CLI 自动在文件头注入 `migrate:tx false` 与 `migrate:reversibility best_effort`。

### 7.4 修改已 applied 的迁移

**禁止**。任何对已 applied 文件的修改都会被 checksum 校验拦截。

正确做法：新建一个版本号来"修正"（典型场景：上线后才发现 v20260626120400 漏建索引 → 新增 v20260626150000 单独建该索引）。

### 7.5 本地 / CI 重置数据库

```bash
gow migrate reset           # DROP DATABASE + CREATE DATABASE，然后 up 到 HEAD
gow migrate redo            # 把最近一个版本 down + up（开发期 iteration 用）
```

### 7.6 Atlas 配置（`atlas.hcl`）

```hcl
env "dev" {
  src = "ent://backend/app/admin/service/internal/data/ent/schema"
  url = env("DATABASE_URL")
  dev = "docker://postgres/15/dev?search_path=public"
  migration {
    dir    = "file://backend/migrations"
    format = "{{ now.UTC.Format \"20060102150405\" }}_{{ .Name }}.sql"
  }
}
```

- 文件名格式由 `format` 严格固定为 `yyyymmddhhmmss_<desc>.sql`，与 Q-用户指定的命名一致。
- `dev` 数据库为 Atlas diff 用临时容器，CI 自动起。

### 7.7 `gow migrate` CLI 子命令完整清单

| 子命令 | 用途 |
|---|---|
| `new sql <desc>` | 创建空 `.sql` + `.down.sql` 骨架 |
| `new go <desc>` | 创建 Go-fn 迁移骨架 |
| `diff <desc>` | 调 Atlas 对当前 ent schema 与 DB 状态做 diff，生成 `.sql` |
| `up` | 应用所有 pending 到 HEAD |
| `upto <version_id>` | 应用到指定版本（包含） |
| `down` | 回滚最近一条 |
| `downto <version_id>` | 回滚到指定版本（不包含） |
| `redo` | down + up 最近一条 |
| `status` | 表格输出所有版本及状态 |
| `verify` | checksum 漂移检查 + 兼容性校验（非破坏，只读） |
| `repair --mark-applied <v>` | 把 failed/pending 标为 applied（不执行 SQL） |
| `repair --mark-skipped <v>` | 把 pending 标为 skipped（不执行 SQL） |
| `repair --retry-up <v>` | 重新尝试 up（前提：上次状态为 failed） |
| `repair --retry-down <v>` | 重新尝试 down |
| `reset` | DROP+CREATE DB，再 up 到 HEAD（仅允许在配置 `allow_destructive: true` 的环境） |
| `lint` | 离线跑 CI lint 规则（不连库） |

CLI 所有命令都通过 `pkg/migrate.Runner` 实现，**不** shell out 到 atlas / goose 二进制。

---

## 8 运行时触发模型

### 8.1 启动期

```go
// backend/app/admin/service/internal/data/ent_client.go (改造后)
func NewEntClient(ctx *bootstrap.Context) (..., error) {
    // OLD: client.Schema.Create(ctx, migrate.WithForeignKeys(true))
    // OLD: seed.EnsurePartialIndexes(...)
    // OLD: seed.SeedThingmodelUnits(...)
    // OLD: seed.SeedThingmodelFeatures(...)
    //
    // NEW: 只做兼容性校验；可选自动 apply（dev only）
    if cfg.Data.Database.Migrate.AutoApply {            // 默认 false
        if err := runner.Up(ctx); err != nil { return nil, nil, err }
    }
    if err := runner.VerifyCompat(ctx); err != nil {
        log.Fatalf("[MIGRATE] %v", err)
    }
    // ... ent client 初始化照旧
}
```

### 8.2 CLI 入口

`cmd/gowmigrate` 是一个独立 cobra binary，编译进 `gow` 工具链；本地与 SRE 都用它。

### 8.3 UI 入口

`MigrationService` BFF（见 §9）暴露的方法被 React 页面调用：
- 所有**写**操作（`Up / UpTo / Down / DownTo / Repair`）要求：
  - 调用者持有角色 `sys:platform_admin`；
  - 请求带 **二次确认 token**（前端弹"输入 `MIGRATE <version_id>` 确认"，后端校验 token 与 version 的 HMAC 绑定）；
  - 自动注入 `actor=ui:user-id=<id>`，进 audit 表。
- **读**操作（`List / Status / Detail / Audit`）要求 `sys:platform_admin` 或 `sys:db_viewer` 角色（后者为本方案新增，仅有只读权限）。

### 8.4 配置项（`backend/app/admin/service/configs/data.yaml`）

```yaml
data:
  database:
    driver: "postgres"
    source: "..."
    migrate:
      auto_apply: false            # 启动是否自动 apply；prod 强制 false
      checksum_policy: strict      # strict | warn | ignore
      checksum_ignore_until: ""    # ignore 模式下的强制 TTL，例 "2026-07-01T00:00:00Z"
      allow_destructive: false     # `reset` 等破坏性命令开关
      lock_timeout: "30s"          # advisory lock 抢占超时
      statement_timeout: "10m"     # 单条 SQL 超时
```

### 8.5 自动 apply 模式（dev/local）

- 仅 `auto_apply: true` 时启动期会跑 `runner.Up(ctx)`；
- Prod 部署配置文件被 `allow_destructive: false` + `auto_apply: false` 锁死，外加 K8s ConfigMap 加 `immutable: true`；
- 部署流水线在 prod 上线时通过独立 Job 跑 `gow migrate up`，job 成功后再放行 Deployment（K8s pre-install/pre-upgrade hook）。

---

## 9 BFF API 契约

### 9.1 `i_migration.proto`（位于 `backend/api/protos/admin/service/v1/`）

```protobuf
syntax = "proto3";
package admin.service.v1;

import "google/api/annotations.proto";
import "google/protobuf/timestamp.proto";
import "google/protobuf/empty.proto";
import "validate/validate.proto";

service MigrationService {
  // 查询：列出所有版本（含 pending）
  rpc ListMigrations(ListMigrationsRequest) returns (ListMigrationsResponse) {
    option (google.api.http) = { get: "/admin/v1/migrations" };
  }
  // 查询：单个版本详情（含 audit 历史、文件预览）
  rpc GetMigration(GetMigrationRequest) returns (Migration) {
    option (google.api.http) = { get: "/admin/v1/migrations/{version_id}" };
  }
  // 查询：当前 HEAD 与应用差
  rpc GetMigrationStatus(google.protobuf.Empty) returns (MigrationStatus) {
    option (google.api.http) = { get: "/admin/v1/migrations/status" };
  }
  // 查询：审计流（按时间倒序）
  rpc ListMigrationAudit(ListMigrationAuditRequest) returns (ListMigrationAuditResponse) {
    option (google.api.http) = { get: "/admin/v1/migrations/audit" };
  }

  // 写：申请二次确认 token（向后端"声明意图"）
  rpc IssueConfirmToken(IssueConfirmTokenRequest) returns (IssueConfirmTokenResponse) {
    option (google.api.http) = {
      post: "/admin/v1/migrations/confirm_token"
      body: "*"
    };
  }
  // 写：应用到最新
  rpc Up(UpRequest) returns (OperationReceipt) {
    option (google.api.http) = { post: "/admin/v1/migrations/up" body: "*" };
  }
  rpc UpTo(UpToRequest) returns (OperationReceipt) {
    option (google.api.http) = { post: "/admin/v1/migrations/up_to" body: "*" };
  }
  // 写：回滚
  rpc Down(DownRequest) returns (OperationReceipt) {
    option (google.api.http) = { post: "/admin/v1/migrations/down" body: "*" };
  }
  rpc DownTo(DownToRequest) returns (OperationReceipt) {
    option (google.api.http) = { post: "/admin/v1/migrations/down_to" body: "*" };
  }
  // 写：repair
  rpc Repair(RepairRequest) returns (OperationReceipt) {
    option (google.api.http) = { post: "/admin/v1/migrations/repair" body: "*" };
  }

  // 实时进度（SSE）
  rpc StreamProgress(StreamProgressRequest) returns (stream ProgressEvent) {
    option (google.api.http) = { get: "/admin/v1/migrations/progress" };
  }
}

message Migration {
  int64 version_id = 1;
  string description = 2;
  string kind = 3;            // sql | go
  string status = 4;          // pending | running | applied | failed | rolled_back | skipped
  string reversibility = 5;
  string irreversible_reason = 6;
  bool data_loss = 7;
  string author = 8;
  google.protobuf.Timestamp applied_at = 9;
  int64 duration_ms = 10;
  string checksum = 11;
  string error_message = 12;
  // 文件预览（最多前 200 行 + 后 200 行；UI 详情用）
  string up_preview = 13;
  string down_preview = 14;
}

message MigrationStatus {
  int64 head_version = 1;          // 文件系统中最新版本
  int64 current_version = 2;       // DB 中已应用最新版本
  int32 pending_count = 3;
  int32 failed_count = 4;
  bool  has_drift = 5;             // checksum 漂移
  string drift_message = 6;
  bool  lock_busy = 7;             // 当前有别的副本/会话在持锁
}

message IssueConfirmTokenRequest {
  string action = 1;               // "up" | "down" | "repair"
  int64 target_version = 2;        // 0 = "all pending"
}
message IssueConfirmTokenResponse {
  string token = 1;                // HMAC(action, target_version, actor, exp=now+2min)
  google.protobuf.Timestamp expires_at = 2;
}

message UpRequest      { string confirm_token = 1; }
message UpToRequest    { string confirm_token = 1; int64 target_version = 2; }
message DownRequest    { string confirm_token = 1; bool accept_data_loss = 2; }
message DownToRequest  { string confirm_token = 1; int64 target_version = 2; bool accept_data_loss = 3; }
message RepairRequest  { string confirm_token = 1; int64 version_id = 2; string mode = 3; } // mark_applied|mark_skipped|retry_up|retry_down

message OperationReceipt {
  string correlation_id = 1;       // 用 StreamProgress 拉日志
  string status = 2;               // "accepted" | "rejected"
  string reason = 3;
}

message ProgressEvent {
  string correlation_id = 1;
  int64 version_id = 2;
  string phase = 3;                // "start" | "executing" | "success" | "failure" | "done"
  string log_line = 4;             // 单行结构化日志
  google.protobuf.Timestamp ts = 5;
}
```

### 9.2 权限码

按 `backend/CLAUDE.md` 既定规则，BFF 路径自动派生：
- `system:migrations:view`（绑定 `sys:platform_admin` + 新增的 `sys:db_viewer`）
- `system:migrations:up`、`system:migrations:down`、`system:migrations:repair`（仅 `sys:platform_admin`）

新增的 `sys:db_viewer` 角色与权限组同样作为一个 Go-fn 种子迁移随本方案上线。

### 9.3 二次确认 token

- `IssueConfirmToken` 颁发的 token 是 `HMAC-SHA256(secret, "<action>|<target_version>|<actor_user_id>|<exp_unix>")` + base64。
- Secret 取 `auth.yaml -> migration_confirm_secret`（部署时与 JWT secret 等同级注入）。
- 写操作请求**必须**携带 token，后端校验：
  1. HMAC 通过；
  2. `exp` 未过期（默认 120 秒）；
  3. `(action, target_version)` 与请求一致；
  4. `actor_user_id` 与当前请求用户一致。
- 任一不通过即返回 `OperationReceipt{status:"rejected"}`，不实际触发 runner。

### 9.4 SSE 实时进度

- `StreamProgress` 按 `correlation_id` 订阅；
- BFF 内部用 `chan ProgressEvent` 多路复用（同进程 fan-out），runner 通过回调把日志 push 进 channel；
- 多副本场景：本副本只能听到本副本的 runner 进度（因为只有一个副本持有 advisory lock，前端拿到 `OperationReceipt.correlation_id` 后直接连那个副本所在的 BFF 实例即可，Ingress 用 sticky session）；
- 前端 UI 在 `OperationReceipt` 返回后 5 秒内**强制**建 SSE，否则退化为 1s 短轮询 `ListMigrationAudit`。

---

## 10 React UI 信息架构

### 10.1 菜单注入

在 `backend/pkg/constants/default_data.go` 的 `DefaultMenus` 中新增（作为一个 Go-fn 种子版本上线）：

```
"系统管理" → "数据库升级"
  path:     /system/migrations
  icon:     database
  authority: ["system:migrations:view"]
  sort:     999  (置于系统管理底部)
```

权限码 `system:migrations:view / :up / :down / :repair` 由 `SyncPermissions` 从 `menu.path` + 操作类型自动派生（与项目既有规则一致）。

### 10.2 页面结构

```
/system/migrations
├── 顶部状态条 StatusBar
│   ├── 当前 DB 版本：v20260626120800
│   ├── 文件系统 HEAD：v20260626130000
│   ├── 待应用：2 条 [Up to HEAD ▶]
│   ├── 失败：0 条
│   └── checksum drift 告警（漂移时显红）
│
├── 主表 MigrationsTable
│   列：version_id | description | kind | status | reversibility | author | duration | applied_at | actions
│   状态 chip 颜色：
│       pending=灰   running=蓝(动画)   applied=绿
│       failed=红    rolled_back=橙     skipped=灰斜体
│   reversibility 列：
│       reversible=✓绿  best_effort=⚠橙  irreversible=✗红（hover 显示 reason）
│   actions：
│       applied  → [详情] [回滚 ▼]（irreversible 按钮置灰）
│       failed   → [详情] [Repair ▼]
│       pending  → [详情]（仅在 HEAD 之前的 pending 才有意义；HEAD 用顶部 Up 按钮）
│
├── MigrationDetailDrawer（侧滑）
│   ├── 元数据卡片
│   ├── up SQL / Go 源预览（语法高亮，最多 200 行 + 折叠）
│   ├── down 预览
│   ├── 该版本的 audit 历史时间线（多次 up/down/repair 都列出）
│   └── 错误堆栈（如果 failed）
│
├── ConfirmUpToDialog
│   - 列出"即将应用的版本"（带 reversibility 图标）
│   - 让用户输入 "MIGRATE <target_version>" 确认
│   - 调用 IssueConfirmToken → UpTo
│
├── ConfirmRollbackDialog
│   - 三色 banner：
│       reversible      → 绿条 "可安全回滚"
│       best_effort     → 橙条 "回滚可能造成数据丢失"，强制勾选 "我已备份"
│       irreversible    → 红条 + 按钮变成 "不可回滚"，仅显示原因（按钮禁用）
│   - 让用户输入 "ROLLBACK <version>" 确认
│
└── AuditTimeline（独立标签页）
    - 全库 audit 流，可按 actor / direction / status 过滤
```

### 10.3 三态契约的视觉表达

| 维度 | reversible | best_effort | irreversible |
|---|---|---|---|
| 状态 chip | ✓ green | ⚠ amber | ✗ red |
| 回滚按钮 | 可点（绿） | 可点（红色二次确认） | **置灰**，hover tooltip 显示 reason |
| 列表行底色 | 默认 | 默认 | 略微红色底纹（视障友好用横线斜纹） |
| 详情页 banner | 无 | 橙色 "Rolling back this version may discard data: <reason>" | 红色 "This migration cannot be rolled back: <reason>" |

### 10.4 数据获取

- 列表/状态：`useListMigrations()` + `useMigrationStatus()`（React Query，stale 10s，windowFocus 重新拉）
- 详情：`useGetMigration(versionId)`（onclick 时拉）
- 写操作：`useUpMigrate / useDownMigrate / useRepairMigrate` mutation hooks；按 `CLAUDE.md` 既定规则，body 用 `{ data: ... }` 形式
- 进度：`useMigrationProgress(correlationId)` — 内部用 `EventSource` 连 SSE；连接失败回退 `setInterval` 拉 audit

### 10.5 i18n

新增 i18n 键空间 `system.migrations.*`；中英文同步落地：
- `system.migrations.title` / `.head` / `.current` / `.pending` / `.failed`
- `system.migrations.status.*`（六个状态）
- `system.migrations.reversibility.*`（三个）
- `system.migrations.action.*`（up/upTo/down/downTo/repair/detail）
- `system.migrations.confirm.up.template` = `"请输入 MIGRATE <版本号> 以确认"` / `"Type MIGRATE <version> to confirm"`

---

## 11 基线与现有 Go 种子的迁移计划

### 11.1 v1 baseline 生成步骤

1. 在一台**干净的、最近一次 `make ent && make run` 自动迁移完成**的 dev 环境上：

   ```bash
   pg_dump --schema-only --no-owner --no-privileges --no-comments \
     --schema=public gwa > /tmp/baseline_raw.sql
   ```

2. 整理 `/tmp/baseline_raw.sql` 为幂等 DDL：
   - `CREATE TABLE` → `CREATE TABLE IF NOT EXISTS`
   - `CREATE INDEX` → `CREATE INDEX IF NOT EXISTS`
   - `CREATE TYPE ... AS ENUM` → 包一层 `DO $$ BEGIN ... EXCEPTION WHEN duplicate_object THEN null; END $$;`
   - `ALTER TABLE ... ADD CONSTRAINT` → 包一层 `DO $$ BEGIN ... EXCEPTION WHEN duplicate_object THEN null; END $$;`
   - 移除 `pg_dump` 加的 SET 语句（`SET statement_timeout = 0;` 等）
3. 在头部加 migrate 注解：
   ```sql
   -- migrate:version            20260626120000
   -- migrate:author             ttshang@example.com
   -- migrate:kind               sql
   -- migrate:reversibility      best_effort
   -- migrate:data_loss          true
   -- migrate:tx                 true
   -- migrate:description        "v1 baseline: schema-only snapshot from gwa dev"
   ```
4. 编写伴生 `.down.sql`：对所有表 `DROP TABLE IF EXISTS ... CASCADE`，对所有 type `DROP TYPE IF EXISTS ...`。声明 `best_effort` + `data_loss=true`。
5. 跑一次 replay 测试：空库 → up → down → up，比对 `atlas schema inspect` HCL 应一致。
6. 提交 PR：`feat(migrate): introduce v1 baseline schema`。

### 11.2 现有 Go 种子的拆分（v2 起）

| 版本号 | 名称 | 来源 | 形态 | reversibility |
|---|---|---|---|---|
| v20260626120100 | seed_languages | `service.LanguageService.init()` → `constants.DefaultLanguages` | Go-fn | best_effort |
| v20260626120200 | seed_permission_groups | `service.PermissionGroupService.init()` → `constants.DefaultPermissionGroups` | Go-fn | best_effort |
| v20260626120300 | seed_permissions | `service.PermissionService.init()` → `constants.DefaultPermissions` | Go-fn | best_effort |
| v20260626120400 | seed_menus | `service.MenuService.init()` → `constants.DefaultMenus` | Go-fn | best_effort |
| v20260626120500 | seed_roles | `service.RoleService.init()` → `constants.DefaultRoles` | Go-fn | best_effort |
| v20260626120600 | seed_admin_user | `service.UserService.init()` → `constants.DefaultUsers` + bcrypt | Go-fn | best_effort |
| v20260626120700 | seed_thingmodel_units | `internal/data/seed/unit_seed.go` | Go-fn | best_effort |
| v20260626120800 | seed_thingmodel_features | `internal/data/seed/feature_seed.go` | Go-fn | best_effort |
| v20260626120900 | create_partial_index_unit_base | `internal/data/seed.EnsurePartialIndexes` | `.sql`（PG only） | reversible |
| v20260626121000 | seed_migration_module_menu | 本方案新增的"数据库升级"菜单与权限码 | Go-fn | best_effort |
| v20260626121100 | seed_db_viewer_role | 新增 `sys:db_viewer` 只读角色 | Go-fn | best_effort |

**关键约束**：
- 顺序严格遵守依赖关系：language → permission_group → permission → menu → role → user；
- 每个 Go-fn 迁移内部 upsert 逻辑**复用** `pkg/constants/default_data.go` 与 `internal/data/seed/*` 已经验证过的代码路径（避免重新发明轮子）；
- `reference_count` 字段在 thingmodel unit 种子里**保留**"不覆盖"的现有语义。

### 11.3 现有 service.init() 启动期注入的下线

下线步骤分两阶段（详见 §12 分期路线）：

**阶段 P-2**：在每个 `service.init()` 的注入入口前加 feature flag：

```go
func (s *MenuService) init() {
    if !s.cfg.Migrate.LegacyStartupSeed {
        return  // 默认 false：交由版本化迁移负责
    }
    // ... 旧逻辑保留作为回退
}
```

`LegacyStartupSeed: true` 仅在迁移过渡窗口用于灰度，老环境如有问题可临时打开。

**阶段 P-4**：所有环境观察 ≥ 2 个 release 无回退后，**物理删除** `init()` 注入代码与 `LegacyStartupSeed` flag；保留 `constants.Default*` 列表作为种子迁移的数据源。

### 11.4 thingmodel partial index 单独成版本

原 `seed.EnsurePartialIndexes` 是启动期 `IF NOT EXISTS` 创建，本方案拆为 v20260626120900 单独版本：

```sql
-- migrate:version            20260626120900
-- migrate:reversibility      reversible
-- migrate:description        "partial unique index: one base unit per (tenant_id, category_id)"

-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX IF NOT EXISTS uix_thingmodel_unit_base
  ON thingmodel_units (tenant_id, category_id)
  WHERE is_base = true;
-- +goose StatementEnd
```

`.down.sql`：`DROP INDEX IF EXISTS uix_thingmodel_unit_base;`。

### 11.5 老环境平滑过渡 runbook

线上已存在的环境升级到本方案的步骤：

1. 部署新 binary（带 `pkg/migrate` 与所有 v1..vN 文件），配置 `auto_apply: false`。
2. SRE 手工执行：
   ```bash
   gow migrate verify       # 应报"current_version=0, drift=none, lock=free"
   gow migrate up           # 应用 v1（baseline，IF NOT EXISTS 全部跳过实际创建）+ v2..vN（种子 upsert）
   gow migrate status       # 验证全部 applied
   ```
3. 重启 Deployment（Pod 重启后 `VerifyCompat` 通过，应用正常起服务）。
4. 观察 ≥ 1 周；期间 `LegacyStartupSeed` 配置项保持 `false`。
5. 进入 §11.3 P-4 阶段，物理删除 `init()` 注入。

---

## 12 可观测性与运维

### 12.1 结构化日志

`pkg/migrate` 所有日志走 Kratos `log` + 字段：`module=migrate version=v20260626120300 direction=up actor=ui:user-id=42 correlation=<uuid> phase=executing`。

关键日志事件：
- `migrate.lock.acquired` / `migrate.lock.busy` / `migrate.lock.released`
- `migrate.version.start` / `migrate.version.success` / `migrate.version.failure`
- `migrate.checksum.drift` / `migrate.checksum.ok`
- `migrate.compat.ok` / `migrate.compat.behind`

### 12.2 Prometheus 指标

```
gow_migrate_versions_total{status="applied|failed|rolled_back|pending"}        gauge
gow_migrate_apply_duration_ms{version_id, direction, kind}                     histogram
gow_migrate_lock_wait_ms                                                       histogram
gow_migrate_compat_violations_total                                            counter
gow_migrate_checksum_drift_total                                               counter
gow_migrate_audit_writes_total{outcome="ok|error"}                             counter
```

### 12.3 OpenTelemetry trace

每次 `Up/Down/Repair` 调用是一个 root span：
- 子 span 包：`acquire_lock`、`load_files`、`apply_v<id>`（每个版本一个）、`write_audit`、`release_lock`
- span attributes：`migrate.actor`、`migrate.correlation_id`、`migrate.target_version`

### 12.4 告警（建议 SRE 接入）

| 告警 | 触发条件 | 严重度 |
|---|---|---|
| `MigrateCompatViolation` | `gow_migrate_compat_violations_total` 1 分钟内 > 0 | P1（应用版本与 DB 不匹配，启动失败） |
| `MigrateChecksumDrift` | `gow_migrate_checksum_drift_total` 任意非零 | P1（脚本被改） |
| `MigrateFailedVersion` | `gow_migrate_versions_total{status="failed"} > 0` | P2 |
| `MigrateLockHeldTooLong` | `gow_migrate_lock_wait_ms` p99 > 5min | P3 |

---

## 13 失败处理与 Repair 场景

### 13.1 失败状态分类

| 类型 | 触发场景 | Repair 路径 |
|---|---|---|
| **A. SQL 语法错误 / 约束冲突** | 脚本本身 bug | 改文件 → CI fail → 不会上线（理想）；若已上线：`repair --retry-up` 不会修复（文件没变）→ 必须发新版本"修复版"，老 failed 行用 `repair --mark-skipped` 跳过 |
| **B. 临时性故障（外部依赖、网络）** | bcrypt 库失败、外部 service 调用失败 | `repair --retry-up <v>` 直接重试 |
| **C. 部分应用（事务一半失败）** | `tx: false` 脚本中途断电 | DB 状态半中半；`repair --retry-up` 需要脚本本身**幂等**（强制要求 `tx: false` 必须幂等）；不幂等的需手工 SQL 兜底后 `repair --mark-applied` |
| **D. checksum 漂移** | 已 applied 脚本被人改 | 优先回滚改动；不能回滚时 `checksum_policy: warn` 临时绕过 + 安排专项修复 |

### 13.2 Repair 三种模式

```
mark_applied   把 failed/pending 标为 applied（不实际跑 SQL）
                ↑ 用于"已手工 SQL 修复完毕、只是状态没同步"
mark_skipped   把 pending 标为 skipped（不实际跑 SQL）
                ↑ 用于"知道这个版本不该跑了，跳过它"
retry_up       重新尝试 up（前提：上次 status=failed）
retry_down     重新尝试 down（前提：上次 down 失败）
```

所有 Repair 操作：
- 必须经 二次确认 token（UI）；
- 必须 actor=`sys:platform_admin`；
- 写一行 `direction='repair'` 的 audit；
- 含 `reason` 字段（必填）。

### 13.3 灾备恢复后的 checksum 重对账

从备份恢复 DB 后，可能出现"`goose_db_version` 是恢复时刻的，但 `migrations/` 文件已经更新到更新版本"的局面。Runbook：

1. `gow migrate verify --checksum-policy=ignore-once`：runner 打印**所有**漂移的版本，但不阻塞；
2. 对每个漂移版本人工 review：
   - 如果是新版本（DB 中有 applied，文件已删）→ 跳过（应该不会出现，因为 migrations/ 是单调增长）；
   - 如果是已存在版本（文件被改）→ 重新计算 sha256，UPDATE `schema_migration_audit` 最新 applied 行的 checksum；
3. 切回 `checksum_policy: strict`，重启服务确认。

---

## 14 分期实施路线

| 阶段 | 范围 | 交付物 | 退出条件 |
|---|---|---|---|
| **P-1** Runner & Baseline | `pkg/migrate` 基础设施；v1 baseline；CLI；启动校验；CI lint & replay | 一台 dev 库能从空起 `gow migrate up` 到 v1；旧环境能"无副作用 up v1" | 单元测试覆盖率 ≥ 80%；replay test 通过 |
| **P-2** Go 种子拆分 | v2..vN Go-fn 种子；`LegacyStartupSeed` flag；与现有 service.init() 共存 | dev 环境 `gow migrate up` 后所有种子到位；老路径仍可回退 | dev/staging 各跑 1 周无 issue |
| **P-3** BFF & React UI | `i_migration.proto`；MigrationService 后端；React 页面；菜单与权限码（作为 v20260626121000 上线） | UI 能看版本、能触发 up/down/repair；SSE 进度通道工作 | 三个 P0 UI 验收用例通过 |
| **P-4** 启动注入下线 | 物理删除 `service.init()` 中的注入逻辑与 `LegacyStartupSeed` flag | 所有环境只剩版本化路径 | 全量 release 观察 ≥ 2 周无回退 |
| **P-5** SRE 接入 | 告警 / 仪表盘 / runbook / 部署文档 | Prometheus / Grafana / Wiki / on-call 手册 | SRE 签字接入 |

---

## 15 风险与缓解

| 风险 | 影响 | 缓解 |
|---|---|---|
| **Atlas 自动生成的 down 不正确** | 上线后回滚反向操作打错表/字段 | CI replay test 强制 `up→down→up` schema 一致；review 时人工检查 `.down.sql` |
| **goose checksum 与自家 audit 表双向漂移** | UI 显示与 goose 行为不一致 | runner 每次写入两表用同一事务；启动时校验两表一致性（不一致写 P2 告警） |
| **Go-fn 迁移依赖的服务层代码后续被改动** | 历史种子重跑结果不一致 | 约定：被 Go-fn 调用的"种子 upsert helper"标 `// stable since vXXXXX, do not change behavior`；CI 跑 `git log -p` lint，发现改动需 reviewer 二人确认 |
| **多副本启动同时执行 advisory lock 竞争** | 副本 A 跑、副本 B 反复重启失败 | 副本 B 在 `VerifyCompat` 失败时退化为"等待 30 秒 + 重试 3 次"再 fatal；K8s readinessProbe 仍可接管 |
| **已有线上库与 baseline SQL 的细微差异（脏数据、手工补丁）** | `gow migrate up` 命中已存在的不同约束 | 上线前在每个 staging 跑 `gow migrate verify --dry-run`；差异修补脚本作为 hotfix 版本号补入时间线 |
| **Atlas / goose 版本升级行为变化** | 历史脚本格式失效 | `go.mod` 锁版本；任何升级 PR 必须跑全量 replay test |
| **二次确认 token 被截获重放** | 攻击者通过 BFF 触发回滚 | token 含 actor_user_id 绑定；2min 过期；HTTPS only；BFF 端单 token 单次使用（消耗后从内存表删除） |
| **UI 触发的大型迁移（>10min）SSE 中断** | 进度看不到 | SSE 断线自动退化 1s 短轮询 audit；correlation_id 是稳定的，最终一致 |

---

## 16 演进方向（Future）

1. **向 C 自研 orchestrator 迁移**：当 goose 表达力不足（如需要"声明式依赖图、并行无关迁移、按 tenant 分库迁移"）时，把 `pkg/migrate` 内部从 goose 切换到自家 runner。BFF/CLI/UI 表面不变。
2. **Atlas Pro lint 引入**：把 Atlas 的 `atlas migrate lint` 接入 CI，覆盖更多语义级风险（如 `ALTER TABLE` 在大表上的 lock 行为）。
3. **跨服务"数据库变更广播"**：未来若拆出多个服务（`thingmodel-service` 独立），让每个服务有自己的 `migrations/` 与版本表，BFF 聚合展示。
4. **Vue Element / Vue Vben UI 复刻**：本期 Q9/C 仅做 React，后续按既有"三套等价复刻"惯例补齐。
5. **租户级 schema 隔离**：当业务需要 `schema-per-tenant` 时，runner 需扩展为"每个 schema 一份 `goose_db_version`"，并提供 BFF 批量推进能力。

---

## 17 范围外 / 待决（Open Questions）

本期**显式不做**，但后续可能需要单独立项：

- 物理备份与 PITR 接入；
- 灰度发布（部分 pod 用新 schema、其它用旧）；
- 流式数据迁移（大表迁移时的双写、读切流量）；
- 业务级"种子热重载"（不重启进程也能补种子）；
- 多语言 i18n 文案的版本化（当前 `constants.DefaultLanguages` 走 Go-fn 种子，更细粒度的 i18n 资源管理另立 spec）。

---

## 附录 A：与现有代码的接触点清单

| 文件 | 改动类型 | 说明 |
|---|---|---|
| `backend/app/admin/service/internal/data/ent_client.go` | 改造 | 删 `client.Schema.Create` + 三个 seed 调用；加 `runner.VerifyCompat` |
| `backend/app/admin/service/internal/data/seed/unit_seed.go` | 保留 | 作为 Go-fn 迁移的实现复用 |
| `backend/app/admin/service/internal/data/seed/feature_seed.go` | 保留 | 同上 |
| `backend/pkg/constants/default_data.go` | 保留 | 作为种子数据源，service.init() 的注入路径下线 |
| `backend/app/admin/service/internal/service/menu_service.go` | 改造 | `init()` 注入加 `LegacyStartupSeed` flag，P-4 删除 |
| `backend/app/admin/service/internal/service/permission_service.go` | 改造 | 同上 |
| `backend/app/admin/service/internal/service/permission_group_service.go` | 改造 | 同上 |
| `backend/app/admin/service/internal/service/role_service.go` | 改造 | 同上 |
| `backend/app/admin/service/internal/service/user_service.go` | 改造 | 同上 |
| `backend/app/admin/service/internal/service/language_service.go` | 改造 | 同上 |
| `backend/app/admin/service/configs/data.yaml` | 新增字段 | `database.migrate.*` 配置块 |
| `backend/Makefile` | 新增目标 | `make migrate-new / migrate-diff / migrate-up / ...` |
| `backend/api/protos/admin/service/v1/i_migration.proto` | 新增 | BFF 契约 |
| `frontend/admin/react/src/router/routes/system.tsx` | 新增路由 | `/system/migrations` |

## 附录 B：与 `CLAUDE.md` 已有规则的对齐

- ✅ **菜单与权限码**：本方案新增的"数据库升级"菜单作为 v20260626121000 种子上线，权限码由 `SyncPermissions` 从 `menu.path` 自动派生；同步绑定 `sys:platform_admin` + 新增 `sys:db_viewer`（符合 `CLAUDE.md Step 12`）。
- ✅ **proto 优先**：BFF 路径走 `backend/api/protos/admin/service/v1/i_migration.proto`，`make api && make ts` 走标准流水线。
- ✅ **三入口共享 Go 库**：CLI、startup hook、BFF 共享 `pkg/migrate`，遵守"never re-implement, never shell-out"。
- ✅ **不手编生成代码**：`backend/api/gen/`、`frontend/.../api/generated/` 仍由工具生成，本方案不动。
- ✅ **JSON 安全**：本方案不涉及 protobuf 消息走 `field.JSON` 存储，`CLAUDE.md Step 13` 的限制不适用。

---

## 修订记录

| 日期 | 版本 | 作者 | 变更 |
|---|---|---|---|
| 2026-06-26 | v1.0 | ttshang | 初稿，brainstorming Q1-Q9 + 簇 1-2 用户确认通过，簇 3-5 由作者按用户授权一次性落稿 |
