# 数据库脚本版本化升级方案 — 评审 + 优化版设计（spec）

| 字段 | 值 |
|---|---|
| 文档版本 | v1.1（评审优化版）|
| 评审对象 | `2026-06-26-database-versioning-design.md`（v1.0）|
| 撰写日期 | 2026-06-26 |
| 评审人 | Claude（架构评审）|
| 范围 | GoWind Admin 后端 + React 前端 |
| 状态 | 评审完成，v1.1 待作者复核 |
| 阅读约定 | 本文件**不改动** v1.0 原文。Part 1 是对 v1.0 的评审意见（含代码佐证），Part 2 是吸收评审结论后的**完整、自洽**优化版设计，可独立替换 v1.0。 |

---

## Part 1 · 对 v1.0 的评审意见

### 1.1 总体评价

v1.0 是一份**完成度很高**的 spec：决策表（Q1–Q9）清晰、端到端数据流完整、文件命名/注解/lint/replay 的工程纪律到位、三态回滚契约与 UI 视觉表达想得透彻、附录 B 对 `CLAUDE.md` 既有规则做了显式对齐。**"三入口一通道"（CLI / startup hook / BFF 共享 `pkg/migrate`，no shell-out）是本方案最强的不变量**，应予保留。

但落到本仓库的真实代码上，有几处**技术契约与现状对不上**或**被一笔带过**，若按 v1.0 直接进入 `writing-plans` 会在 P-1/P-2 卡壳。下面按严重度分级列出，每条都附**代码佐证**。

### 1.2 问题清单（按严重度）

#### P0 — 技术契约错误，必须修正否则方案不成立

**P0-1　Go-fn 的版本号无法被 goose 锚定到时间戳 —— 动摇 Q4/B"单一时间线"地基**

v1.0 §4.3 示例同时做了两件事：
```go
mctx.Register(mctx.Migration{Version: 20260626120600, ...})   // 自家注册表
goose.AddMigrationContext(upSeedAdminUser, downSeedAdminUser) // goose
```
`mctx` 里的 `Version` 是自家元数据，**goose 看不见**。而 goose 的 Go 注册 API（`goose.AddMigrationContext`）默认按注册顺序分配 **1, 2, 3 …** 的顺序版本号，并不从文件名读时间戳（时间戳解析只对 `.sql` 文件生效）。结果：同一时间线上 `.sql` 用 14 位时间戳、Go-fn 用顺序号，两套版本号**不能正确交错**，`up/down` 的排序与"按时间戳单调"契约直接冲突。

> 影响：Q4/B"`.sql` 与 Go-fn 共用同一张版本表与同一时间线"无法按字面落地。
> 处置：见 Part 2 §4.3（每个 Go-fn 配一个**只声明版本号的薄 `.sql` stub**，由 goose 解析时间戳并 `+goose NO TRANSACTION` 转交给 Go func；stub 是契约载体，Go func 是实现载体）。需在 P-1 做 spike 验证。

**P0-2　advisory lock 的 key 在 §2.2 与 §6.1 自相矛盾**

- §2.2：`L = hashtext('gow_migrate')::bigint`
- §6.1：`const lockKey int64 = 0x676F77_6D696772`

两处算出的 key **不同** → 若不同入口/不同副本各引用一处，`pg_try_advisory_lock` 互不互斥，多副本安全（Q6 的多副本语义）失效。此外 `hashtext()` 返回 `int4`（±2³¹），`::bigint` 只是拓宽存储、值域仍是 int4，与 `pg_advisory_lock(int4, int4)` 双参重载是不同函数。

> 处置：Part 2 §6.1 统一为单一 `const`，并在注释里写明"等价于 `hashtextextended('gow_migrate', 0)`"，用 `int8` 单参重载。

**P0-3　关闭 `client.Schema.Create` 后，Q3 承诺保留的"Ent codegen + SQLite 测试通道"失去了建表来源**

现状（`ent_client.go:46-49`）：
```go
if cfg.Data.Database.GetMigrate() {
    if err := client.Schema.Create(ctx.Context(), migrate.WithForeignKeys(true)); err != nil { ... }
```
Q1 = "完全替代 `Schema.Create`"。但 Q3 又承诺"保留 Ent codegen 测试通道（SQLite）"。SQLite 测试目前**正是靠 `Schema.Create` 建表**的 —— 一旦全局删除该调用，SQLite 测试要么没表、要么得自己跑一遍 goose 迁移（而 §4.4 的 PG-only DDL 在 SQLite 上不兼容）。

> 处置：Part 2 §3.1 + §8.1 明确"测试通道保留一条**仅测试可见**的 `Schema.Create` 旁路（build tag `migrate_test`），生产路径走 goose"。

#### P1 — 设计缺口，进入实施会返工

**P1-1（最关键）　Go-fn 迁移依赖 `*Repo` 与 `SystemViewerContext`，但 `pkg/migrate` 是个无 DI 的库 —— "复用已验证代码路径"无法字面兑现**

真实种子代码的依赖：
- 物模型种子（`unit_seed.go` / `feature_seed.go`）：参数是 `*ent.Client` + `appViewer.NewSystemViewerContext`，可被 `pkg/migrate` 直接复用 ✓
- 语言/权限组/权限/菜单/角色/用户种子：实现在 `service.*Service.createDefault*`，依赖 **Wire 注入的 `*data.*Repo`**，且走的是 `menuRepo.Create(...)` 这类 repo 方法，不是裸 SQL

v1.0 §3 说"迁移文件只写调用 + 幂等 upsert 的薄壳，不复制粘贴种子数据"，但**没解释一个运行在 `NewEntClient`（DI 容器尚未构建）里的 runner 如何拿到 `*MenuRepo`**。两条路都有代价：
- (a) 迁移改用 repo-free 的 ent.Client upsert → **重复实现** service 层逻辑，违背"不重新发明轮子"
- (b) 把 `runner.Up` 推迟到 DI 容器构建之后 → 改变 §8.1 的启动顺序，且 CLI/BFF 也要重新设计对象装配

> 处置：Part 2 §4.4 给出明确取舍——**数据型种子只依赖 `*ent.Client`（在 `pkg/migrate` 内重写薄 upsert，允许与 service 层有受控重复）**；需要 `*Repo` 的复杂种子（目前没有，menu/role/user 的 `createDefault*` 逻辑都很薄）下沉为"`*ent.Client` + `constants.Default*`"的直接 upsert，service 层的 `init()` 路径整体下线（见 P1-2）。这样 `pkg/migrate` 只依赖 `*ent.Client`，不引入 repo 耦合，DI 干净。

**P1-2　真实 `service.init()` 是 `count==0` 守卫且吞错误，与 v1.0 假设的"flag 守卫"模型不符**

真实代码（`menu_service.go:41-46`，其余 5 个同构）：
```go
func (s *MenuService) init() {
    ctx := appViewer.NewSystemViewerContext(context.Background())
    if count, _ := s.menuRepo.Count(ctx, nil); count == 0 {   // ← 吞掉 Count 的 err
        _ = s.createDefaultMenus(ctx)                          // ← 吞掉 create 的 err
    }
}
```
v1.0 §11.3 示范给 `init()` 套 `LegacyStartupSeed` flag，但**没注意到现状是"空表才灌"**。过渡期双重写入的**顺序**才是关键：runner 在 `NewEntClient`（Wire 早期）跑，service `init()` 在 `NewMenuService`（Wire 后期）跑。只要 runner 先跑、灌入行，`init()` 看到 `count>0` 自动空转——OK，但 v1.0 **没有把这层顺序依赖写进契约**。另外 `count, _ :=` 吞错误 vs 新体系 fail-fast，是一次**可观测性的行为变化**。

> 处置：Part 2 §5.3 明确"runner 必须在所有 `New*Service` 之前完成（Wire provider 顺序契约）"，并把 `init()` 的 flag 守卫改为"runner 已成功 apply 时整体跳过"。

**P1-3　`ApiService.init()` 被 v1.0 §11.2 的种子拆分表遗漏**

grep 实测有 **7** 个 `*Service.init()`：
```
api_service.go / language_service.go / permission_group_service.go /
menu_service.go / permission_service.go / role_service.go / user_service.go
```
v1.0 §11.2 只列了 6 个（漏 `ApiService`）。若 `ApiService.init()` 也灌默认数据，必须进迁移计划或显式声明排除。

> 处置：Part 2 §11.2 补一行并在实施前确认 `ApiService.init()` 的行为。

**P1-4　配置键命名冲突：现有 `data.database.migrate` 是 bool，v1.0 要把它变成 struct**

现状 `cfg.Data.Database.GetMigrate()` 返回 `bool`（即 `data.database.migrate: true`）。v1.0 §8.4 引入 `data.database.migrate.auto_apply / checksum_policy / ...`（struct）。**同一 key、不同类型** —— 老的 `data.yaml` 升级到新 binary 时，`migrate: true` 无法反序列化进 struct，**服务起不来**。v1.0 通篇没有"配置 schema 迁移"这一节。

> 处置：Part 2 §8.4 改用新键 `data.database.migrations.*`（复数，避开单数 bool），并在附录 A 增"配置迁移"改动点；旧 `migrate: true` 在读取层做一次性兼容（读到 bool true → 等价 `auto_apply:true`，打 deprecation 日志）。

**P1-5　特征种子的"单条失败不中断、累计上报"语义与迁移契约"一个 failed 阻塞后续 up"冲突**

现状（`feature_seed.go:59-125`）：故意 `continue` 收集 failures，末尾汇总返回 error。这套语义是为"240+ 条特征里一条 bug 不至于全挂"设计的。而迁移模型（§5.2）规定"HEAD 之前有未恢复 `failed` 即整次 `up` 报错"。若把 `seed_thingmodel_features` 做成**一个版本**，1/270 失败 → 整版 `failed` → 阻塞下游所有版本，被迫 repair。粒度不匹配。

> 处置：Part 2 §11.2 把特征种子拆成"骨架版本（property/event/service 三类）+ relation 版本"两段，并在迁移内部**保留 continue-on-error + 汇总**的容错（迁移契约补充："数据型种子允许 partial-success，状态记 `applied_with_warnings`，不阻塞下游"）。

**P1-6　`repair --mark-applied / --mark-skipped` 只动 audit 表，不动 `goose_db_version`，会与 goose 脱钩**

goose 的 `current_version` 来自 `goose_db_version`，不是 audit 表。v1.0 §13.2 的 repair 三模式只描述了 audit 表的状态变化，没说 goose 表怎么同步。结果：`mark-applied` 后 audit 显示 applied，但 goose 仍认为该版本未应用 → 下次 `up` 又跑一遍。§5.2 又说"failed 阻塞后续 up"，而 `mark-skipped` 若只在 audit 标 skipped、goose 不知情，**根本解不了阻塞**。

> 处置：Part 2 §13.2 规定 repair 必须在**同一事务**内写两张表：`mark-applied` → goose 插 `is_applied=true` + audit 标 applied；`mark-skipped` → goose 插 `is_applied=true` + audit 标 skipped（对 goose 而言即"视为已应用"）。

#### P2 — 正确性/健壮性改进

**P2-1　`MinRequiredVersion` 的 codegen 注入机制 + "schema 改动版本"判定未形式化**（v1.0 §6.2）。建议新增 `migrate:category schema|data` 注解，codegen 取"最近的 `category=schema` 版本"作为 `MinRequiredVersion`，种子失败不阻塞启动才成立。

**P2-2　baseline 的 `CREATE TABLE IF NOT EXISTS` 对"脏库"是掩盖而非修复**（v1.0 §11.1）。新库 OK；老库若有手工补丁/列差异，`IF NOT EXISTS` 静默成功留下错误 schema。建议 baseline apply 前强制跑 `verify --dry-run` 比对 `atlas schema inspect`，diff 非空则拒绝 apply（仅对新环境豁免）。

**P2-3　SSE + sticky session 假设脆弱，且对一个"低频手动运维"操作 ROI 低**（v1.0 §9.4）。round-robin Ingress 或 pod 重调度会让 SSE 连到非持锁副本。建议**P-3 先只做 1s 短轮询 audit**，SSE 推迟到独立子阶段，降低 P-3 爆炸半径。

**P2-4　`reset`/`redo` 的 `allow_destructive` 配置位防护不足**。配置可被有 ConfigMap 权限的人改。建议再加一层"仅 CI/dev Pod 注入的环境变量密钥"（`GOW_MIGRATE_DESTRUCTIVE_KEY`），双重校验。

**P2-5　缺少"切换本身"的回滚 runbook**（v1.0 §11.5 只有前进路径）。补一条：若 prod `gow migrate up` 中途失败，回滚 binary 到上一版是安全的——只要没跑破坏性 down，新 binary 的 `VerifyCompat` 是"只读、DB 落后才 fatal"，降级 binary 不会因 schema 超前而拒起。

**P2-6　迁移系统的测试策略过薄**。v1.0 只有 §4.5 replay + P-1 "覆盖率 80%"。建议补 fault-injection 用例：迁移中途 kill、advisory lock 过期/会话断开、up 过程中 checksum 漂移、repair 后两表一致性。

#### P3 — 表述/清晰度

- **P3-1** §4.1 "禁止中文、空格、连字符以外的标点"是双重否定，易误读。改为正则：`^[a-z][a-z0-9_]{0,59}$`。
- **P3-2** §4.3 Go-fn checksum "所有静态引用的常量"一般不可判定（reflect/泛型 map key）。改为：checksum = 迁移文件内容 sha256 ∪ 显式 `// migrate:checksum_deps` 列出的 `constants` 文件 sha256。
- **P3-3** §7.1 / 附录 A 的 `make ent / make migrate-*` 与本机环境（Windows git-bash 无 make，见项目记忆）冲突。`gow` 是跨平台正典，Make 目标标注为"CI/Linux 可选"。
- **P3-4** `schema_migration_audit` 的 `UNIQUE(version_id, started_at)` 在同微秒重试理论可撞；改用 `correlation_id` 唯一约束更稳。
- **P3-5** §13.1 type-A 修复路径与 P1-6 同源（skipped 解不了 goose 阻塞），已在 P1-6 一并修。

### 1.3 战略层观察（不强制改，供决策）

- **A. 双表模型（goose_db_version + schema_migration_audit）= 两份真相源。** v1.0 的理由（不动 goose 内部、升级安全）成立，但同步成本真实。可考虑的替代：用 goose 的 `SetTableName` + 一次性 ALTER 给 goose 表加列，**单表**承载。本次评审**建议保留双表**（升级期更安全），但把"两表一致性校验"从 §15 的 P2 告警**升级为启动期硬校验**（不一致即 fatal，避免 UI 显示与 goose 行为长期背离）。
- **B. P-3 爆炸半径偏大。** 一个阶段同时交付 BFF + React UI + SSE + confirm-token + 新角色 + 菜单种子。建议拆 P-3a（只读：list/status/audit，无写、无 SSE）+ P-3b（写 + confirm-token + 进度）。只读 UI 能更早产生价值。
- **C. "三入口一通道"是最强不变量，保留。** 这条是 v1.0 最值得坚持的设计，Part 2 原样继承。

---

## Part 2 · 优化版方案设计 v1.1

> 本部分自洽，可直接替换 v1.0。与 v1.0 一致的内容（决策表骨架、文件布局总形、proto 契约主体、UI 信息架构）不再逐字复述，只写**变更与新增**；未提及处沿用 v1.0。

### 2.0 决策快照（相对 v1.0 的差异用 △ 标注）

| 编号 | 决策 | v1.1 结论 |
|---|---|---|
| Q1 | 与 Ent auto-migrate 的关系 | 完全替代生产路径的 `Schema.Create`；**△** 测试通道保留 `//go:build migrate_test` 旁路 |
| Q2 | Ent schema 角色 | A1 不变 |
| Q3 | 方言范围 | Postgres-only 生产；**△** SQLite 测试通道靠测试旁路建表，不走 goose |
| Q4 | Go-fn 融入版本化 | B 不变；**△** Go-fn 经 `.sql` stub 锚定时间戳版本号（修 P0-1） |
| Q5 | 运行时底座 | goose（库内嵌）+ Atlas（仅开发期）不变 |
| Q6 | 触发模型 | C 不变 |
| Q7 | v1 基线 | B 不变；**△** apply 前强制 `verify --dry-run` 比对（修 P2-2） |
| Q8 | 回滚契约 | B 不变；**△** 新增 `applied_with_warnings` 状态承载种子 partial-success（修 P1-5） |
| Q9 | UI 范围 | C 不变；**△** P-3 拆 a/b 两子阶段，SSE 推迟（修 P2-3、战略 B） |
| **Q10**（新） | 配置键命名 | **△** 新键 `data.database.migrations.*`（复数），避开与旧 bool `migrate` 冲突（修 P1-4） |
| **Q11**（新） | Go-fn 与 DI 的边界 | **△** 数据型种子只依赖 `*ent.Client`，不引入 `*Repo` 耦合（修 P1-1） |
| **Q12**（新） | 两表一致性 | **△** 启动期硬校验 goose_db_version ↔ audit 不一致即 fatal（修战略 A） |

### 2.1 仓库布局（差异）

```
backend/
├── migrations/
│   ├── 20260626120000_baseline.sql                # v1（PG-only 幂等 DDL）
│   ├── 20260626120000_baseline.down.sql
│   ├── 20260626120100_seed_languages.sql          # △ stub：仅 -- +goose Up/Down + NO TRANSACTION
│   ├── 20260626120100_seed_languages.go           # △ Go 实现（被 stub 转交调用）
│   ├── ...（每个 Go-fn 种子 = .sql stub + .go 对）
│   ├── 20260626120900_create_partial_index_unit_base.sql   # 纯 DDL，无 Go
│   └── 20260626121000_seed_migration_module_menu.{sql,go}
│
├── pkg/migrate/
│   ├── runner.go
│   ├── registry.go            # △ Go-fn 注册表：key=version_id，由 stub 解析注入
│   ├── stub_dispatch.go       # △ 新：解析 .sql stub 的版本号与 +goose 指令，路由到 Go func
│   ├── lock.go                # △ 单一 lockKey 常量（修 P0-2）
│   ├── audit.go               # △ 写双表（goose_db_version + audit）同事务
│   ├── compat.go              # △ MinRequiredVersion + 两表一致性硬校验
│   ├── lint.go
│   ├── replay_test.go
│   ├── fault_injection_test.go # △ 新：kill/lock-expiry/checksum-drift 用例（修 P2-6）
│   └── embed.go
│
└── app/admin/service/internal/
    ├── data/ent_client.go             # 改造：删 Schema.Create（生产）+ 旧 seed 调用
    ├── data/ent_client_test.go        # △ //go:build migrate_test 旁路（修 P0-3）
    └── service/migration_service.go   # BFF
```

### 2.2 迁移文件规范（关键差异）

#### 2.2.1 Go-fn 经 stub 锚定版本号（修 P0-1）

每个 Go-fn 迁移**两个文件同时间戳**：

`20260626120100_seed_languages.sql`（stub，由 goose 解析版本号）：
```sql
-- migrate:version       20260626120100
-- migrate:kind          go
-- migrate:category      data
-- migrate:reversibility best_effort
-- migrate:tx            false
-- migrate:description   "seed system languages"
-- +goose Up
-- +goose NO TRANSACTION
SELECT stub_dispatch('20260626120100', 'up');   -- △ 路由到 Go func
-- +goose Down
SELECT stub_dispatch('20260626120100', 'down');
```

`20260626120100_seed_languages.go`（实现）：
```go
package migrations

func init() {
    mctx.Register(mctx.Migration{Version: 20260626120100, Kind: "go", /* ... */ })
}

// 由 stub_dispatch 调用；签名与 goose Go 迁移一致，但版本号来自 stub 文件名。
func upSeedLanguages(ctx context.Context, tx *sql.Tx) error { /* ent.Client upsert */ }
```

要点：
- **goose 只解析 `.sql` stub 的文件名拿时间戳**，Go-fn 不再独立注册到 goose（避免顺序号污染）。`stub_dispatch` 是一个 PG 函数或 goose provider 钩子，按 version_id 在 `mctx` 查表并调用。
- lint 规则新增：**每个 Go-fn 必须有同名同时间戳的 `.sql` stub，否则 CI 失败**（修 P0-1 的强制约束）。

#### 2.2.2 文件头注解（相对 v1.0 增项）

```sql
-- migrate:version            20260626120900
-- migrate:kind               sql | go
-- migrate:category           schema | data          -- △ 新：codegen 据 schema 类推 MinRequiredVersion
-- migrate:reversibility      reversible | irreversible | best_effort
-- migrate:irreversible_reason ""                    (irreversible 必填)
-- migrate:data_loss          false | true
-- migrate:tx                 true | false           (tx:false 必须幂等且 reversibility!=reversible)
-- migrate:depends_on         []
-- migrate:checksum_deps      ["pkg/constants/default_data.go"]  -- △ 新：Go-fn 显式声明 checksum 依赖文件
-- migrate:description        "..."
```

#### 2.2.3 Go-fn 与 DI 的边界（修 P1-1，落地 Q11）

- **`pkg/migrate` 只持有 `*ent.Client`（+ `*sql.DB` 用于 advisory lock / raw exec），不依赖任何 `*data.*Repo`。**
- 语言/菜单/权限/角色/用户种子：在迁移文件内用 `*ent.Client` 直接 upsert `constants.Default*`，**允许与 `service.createDefault*` 有受控的薄重复**（两者都是几行 upsert）。service 层 `init()` 在 P-4 物理删除后，重复自然消失。
- 物模型种子：直接调用现有 `seed.SeedThingmodelUnits(ctx, client, logger)` / `SeedThingmodelFeatures(...)`——这俩本来就只吃 `*ent.Client`，零改动复用 ✓。
- 所有 Go-fn 内部 **必须** `ctx = appViewer.NewSystemViewerContext(ctx)`（复刻现状，绕过 TenantPrivacy）。

### 2.3 版本状态模型（差异）

#### 2.3.1 `schema_migration_audit` 增列与状态

```sql
-- △ status 枚举新增 applied_with_warnings（修 P1-5）
status TEXT NOT NULL CHECK (status IN (
    'running','applied','applied_with_warnings',
    'failed','rolled_back','skipped'))
```
- `applied_with_warnings`：数据型种子 partial-success（如特征种子 269/270 成功）专用；**不阻塞下游 up**，UI 显橙色但不是 failed。
- `UNIQUE(version_id, started_at)` → **改为 `UNIQUE(correlation_id)`**（修 P3-4）。

#### 2.3.2 两表一致性硬校验（修战略 A，落地 Q12）

启动期 `VerifyCompat` 顺带做：
1. `goose_db_version` 的 `MAX(version_id where is_applied)` ≡ audit 表最近一次 `direction in ('up','repair')` 且 `status in ('applied','applied_with_warnings','skipped')` 的 `MAX(version_id)`；
2. 每个已应用版本的 audit `checksum` 与 goose 记录一致（goose 自带 checksum 列）。

不一致 → **`log.Fatal`**（不降级为告警），避免 UI 与 goose 长期背离。修复路径走 §2.6 repair。

### 2.4 并发与启动（差异）

#### 2.4.1 advisory lock 单一 key（修 P0-2）

```go
// pkg/migrate/lock.go
// 单一来源，等价于 SELECT hashtextextended('gow_migrate', 0)::bigint; 用 int8 单参重载。
const lockKey int64 = 6760438529838312553 // 'gow_migrate' 的稳定哈希，部署前用脚本固化
```
- 注释里写明算法；任何入口引用此常量，禁止再写 `hashtext(...)` 字面量。
- 会话级锁，session 断开自动释放（沿用 v1.0）。

#### 2.4.2 启动顺序契约（修 P1-2）

Wire provider 顺序必须保证：**`NewEntClient`（内含 runner.Up(auto-apply) + VerifyCompat）在所有 `New*Service` 之前完成**。理由：runner 先灌种子 → service `init()` 的 `count==0` 守卫自然空转，杜绝双重写入。该顺序写进 §2.7 路线 P-2 的验收项。

#### 2.4.3 `MinRequiredVersion` codegen（修 P2-1）

- 注解 `migrate:category schema|data` 由 lint 解析。
- 构建脚本扫 `migrations/`，取"最近一个 `category=schema` 的版本号"写入 `pkg/migrate/compat.go` 的 `MinRequiredVersion`（`go:generate`）。
- 效果：只有 DDL 落后才会启动 fatal；纯种子落后不阻塞（与 v1.0 意图一致，但机制形式化）。

### 2.5 BFF / UI（差异：拆 P-3，推迟 SSE）

#### 2.5.1 proto 契约

沿用 v1.0 §9.1，**仅改**：
- `Migration.status` 枚举增 `applied_with_warnings`；
- `StreamProgress` 标注 `// Phase P-3b`，v1.1 P-3a 阶段**不发布**该 rpc（前端用 `ListMigrationAudit` 1s 短轮询）。

#### 2.5.2 路线拆分（修战略 B）

| 子阶段 | 范围 | 退出条件 |
|---|---|---|
| **P-3a** 只读 UI | `ListMigrations/GetMigration/GetMigrationStatus/ListMigrationAudit` + React 列表/详情/审计时间线；**无写、无 SSE** | 只读 UI 验收用例通过；`sys:db_viewer` 可见 |
| **P-3b** 写 + 进度 | `IssueConfirmToken/Up/UpTo/Down/DownTo/Repair` + confirm 对话框 + SSE（或维持短轮询） | 三态回滚 UI 用例 + confirm-token 安全用例通过 |

P-3a 可在 P-2 完成后立即开工，提前产生运维可见性。

### 2.6 Repair 双表原子写（修 P1-6 / P3-5）

所有 repair 模式在**单一事务**内同时操作两张表：

| 模式 | `goose_db_version` | `schema_migration_audit` 新行 |
|---|---|---|
| `mark_applied` | `INSERT (version_id, is_applied=true)` | `direction=repair, status=applied, reason=必填` |
| `mark_skipped` | `INSERT (version_id, is_applied=true)` | `direction=repair, status=skipped, reason=必填` |
| `retry_up` | （由 goose 正常 up 写） | 先插 `direction=repair, status=running`，成功/失败再更新 |
| `retry_down` | 同上 | 同上 |

关键：对 goose 而言 `skipped` ≡ `applied`（goose 无 skipped 概念），**只有写进 goose 表才能真正解除"failed 阻塞 up"**。修完随即跑一次 §2.3.2 一致性校验。

### 2.7 现有 Go 种子拆分表（修 P1-3 / P1-5）

| 版本号 | 名称 | 来源 | 形态 | category | reversibility |
|---|---|---|---|---|---|
| 20260626120100 | seed_languages | `constants.DefaultLanguages` | Go-fn | data | best_effort |
| 20260626120200 | seed_permission_groups | `constants.DefaultPermissionGroups` | Go-fn | data | best_effort |
| 20260626120300 | seed_permissions | `constants.DefaultPermissions` | Go-fn | data | best_effort |
| 20260626120400 | seed_menus | `constants.DefaultMenus` | Go-fn | data | best_effort |
| 20260626120500 | seed_roles | `constants.DefaultRoles` | Go-fn | data | best_effort |
| 20260626120600 | seed_admin_user | `constants.DefaultUsers` + bcrypt | Go-fn | data | best_effort |
| **20260626120650** | **seed_api_seed**（**△ 新，待确认**）| `ApiService.init()`（**v1.0 遗漏，实施前核实其是否灌默认数据**）| Go-fn | data | best_effort |
| 20260626120700 | seed_thingmodel_units | `seed.SeedThingmodelUnits` | Go-fn | data | best_effort |
| 20260626120800 | seed_thingmodel_features_core | `seed.SeedThingmodelFeatures` 非关系部分 | Go-fn | data | best_effort（**△ partial-success → applied_with_warnings**）|
| 20260626120850 | seed_thingmodel_features_relation | 特征 RELATION 部分 | Go-fn | data | best_effort |
| 20260626120900 | create_partial_index_unit_base | `seed.EnsurePartialIndexes` | .sql（PG） | schema | reversible |
| 20260626121000 | seed_migration_module_menu | 本方案菜单+权限码 | Go-fn | data | best_effort |
| 20260626121100 | seed_db_viewer_role | `sys:db_viewer` | Go-fn | data | best_effort |

**依赖顺序**（Wire 契约 + 迁移 depends_on）：language → permission_group → permission → menu → role → user → (api?) → thingmodel_units → thingmodel_features_core → features_relation → partial_index。

### 2.8 配置（修 P1-4，落地 Q10）

```yaml
data:
  database:
    driver: "postgres"
    source: "..."
    # △ 旧键 migrate（bool）废弃；读取层做一次性兼容：migrate:true → migrations.auto_apply:true
    migrations:                       # △ 复数，新 struct
      auto_apply: false               # prod 强制 false
      checksum_policy: strict         # strict | warn | ignore
      checksum_ignore_until: ""
      allow_destructive: false
      destructive_key: ""             # △ 新：reset/redo 需额外校验此 env-injected 密钥（修 P2-4）
      lock_timeout: "30s"
      statement_timeout: "10m"
      legacy_startup_seed: false      # △ P-2/P-4 过渡期 flag（修 P1-2），P-4 删除
```

### 2.9 老环境切换 runbook（补 P2-5 回滚路径）

```
1. 部署新 binary（auto_apply:false），先不重启业务 Pod。
2. SRE：gow migrate verify            # 期望 current=0, drift=none, lock=free
3. SRE：gow migrate up                # v1 baseline（IF NOT EXISTS 跳过已存在）+ 种子 upsert
4. gow migrate status                 # 全 applied（或 applied_with_warnings）
5. 滚动重启 Deployment → VerifyCompat 通过 → 起服务
6. 观察 ≥ 1 周；legacy_startup_seed:false
7. 回滚预案（新增）：
   - 若步骤 3 失败：保留失败 audit，回滚 binary 到上一版（旧版不读 goose_db_version，可正常起）；
   - 已应用的版本无需回退（种子是 upsert，baseline 是 IF NOT EXISTS，均幂等无破坏）；
   - 排查后重试步骤 3，或对失败版本走 repair。
8. 进入 P-4：物理删除 service.init() 注入。
```

### 2.10 测试策略（补 P2-6）

| 层 | 用例 |
|---|---|
| 单元 | lint 规则全量；registry/stub_dispatch 路由；checksum_deps 命中 |
| replay（CI） | 空库 up→down→up，schema HCL diff 为空（沿用 v1.0 §4.5）|
| **△ fault-injection** | 迁移中途 `SIGKILL` 进程 → 重启后状态自愈；advisory lock 会话断开 → 锁释放；up 中途篡改文件触发 checksum 漂移 → fail-fast；repair 后跑 §2.3.2 一致性校验通过 |
| **△ 双表一致性** | 人为只改 audit 不改 goose → 启动 fatal |
| 切换演练 | staging 上模拟"脏库 + 手工补丁"，`verify --dry-run` 必须 diff 非空并拒绝 apply |

### 2.11 接触点清单（相对 v1.0 附录 A 的增项）

| 文件 | 改动 | 关联 |
|---|---|---|
| `backend/app/admin/service/internal/data/ent_client.go` | 删生产 `Schema.Create` + 旧 seed；加 runner | P0-3, P1-2 |
| `backend/app/admin/service/internal/data/ent_client_test.go` | **△ 新**：`//go:build migrate_test` 旁路保留 `Schema.Create` | P0-3 |
| `backend/app/admin/service/internal/service/api_service.go` | **△ 核实 init() 行为**，纳入迁移或排除 | P1-3 |
| `backend/app/admin/service/configs/data.yaml` | **△** `migrate`(bool) → `migrations`(struct) + 兼容读取 | P1-4 |
| `backend/pkg/migrate/compat.go` | **△** `MinRequiredVersion` codegen + 两表硬校验 | P2-1, Q12 |
| `backend/pkg/migrate/lock.go` | **△** 单一 lockKey | P0-2 |

---

## 修订记录

| 日期 | 版本 | 作者 | 变更 |
|---|---|---|---|
| 2026-06-26 | v1.0 | ttshang | 初稿（原文，未改动）|
| 2026-06-26 | v1.1 | Claude（评审） | Part 1 评审（3×P0 / 6×P1 / 6×P2 / 5×P3，均附代码佐证）；Part 2 吸收结论产出优化版设计，新增 Q10–Q12 三项决策 |

## 附：v1.0 → v1.1 关键差异速查

- **P0-1** Go-fn 版本号：经 `.sql` stub 锚定时间戳（不再依赖 goose 顺序号）
- **P0-2** advisory lock：单一常量，消除 §2.2/§6.1 矛盾
- **P0-3** SQLite 测试通道：`//go:build migrate_test` 旁路保留 `Schema.Create`
- **P1-1** DI 边界：`pkg/migrate` 只持 `*ent.Client`，不引 `*Repo`（Q11）
- **P1-2** 启动顺序契约：runner 必须先于所有 `New*Service`（Wire provider 顺序）
- **P1-3** 补 `ApiService.init()` 到种子计划（待确认）
- **P1-4** 配置键 `migrate`(bool) → `migrations`(struct) + 兼容（Q10）
- **P1-5** 新增 `applied_with_warnings` 状态 + 特征种子拆分
- **P1-6** repair 双表原子写（goose + audit 同事务）
- **P2-1** `migrate:category schema|data` 注解 + codegen 注入 `MinRequiredVersion`
- **P2-2** baseline apply 前强制 `verify --dry-run`
- **P2-3** P-3 拆 a/b，SSE 推迟
- **P2-4** `destructive_key` 双重校验
- **P2-5** 切换回滚 runbook
- **P2-6** fault-injection + 双表一致性测试
- **Q12** 两表不一致启动 fatal（非告警）
