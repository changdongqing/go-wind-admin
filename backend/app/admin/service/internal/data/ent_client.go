package data

import (
	"github.com/go-kratos/kratos/v2/log"

	"entgo.io/ent/dialect/sql"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"

	entCrud "github.com/tx7do/go-crud/entgo"

	"github.com/tx7do/kratos-bootstrap/bootstrap"
	entBootstrap "github.com/tx7do/kratos-bootstrap/database/ent"

	"go-wind-admin/app/admin/service/internal/data/ent"
	"go-wind-admin/app/admin/service/internal/data/ent/migrate"
	_ "go-wind-admin/app/admin/service/internal/data/ent/runtime"
	"go-wind-admin/app/admin/service/internal/data/seed"
)

// NewEntClient 创建Ent ORM数据库客户端
func NewEntClient(ctx *bootstrap.Context) (*entCrud.EntClient[*ent.Client], func(), error) {
	l := ctx.NewLoggerHelper("ent/data/admin-service")

	cfg := ctx.GetConfig()
	if cfg == nil || cfg.Data == nil {
		l.Fatalf("[ENT] failed getting config")
		return nil, func() {}, nil
	}

	cli, err := entBootstrap.NewEntClient(cfg, func(drv *sql.Driver) *ent.Client {
		client := ent.NewClient(
			ent.Driver(drv),
			ent.Log(func(a ...any) {
				l.Debug(a...)
			}),
		)
		if client == nil {
			l.Fatalf("[ENT] failed creating ent client")
			return nil
		}

		// run the auto migration tool
		if cfg.Data.Database.GetMigrate() {
			if err := client.Schema.Create(ctx.Context(), migrate.WithForeignKeys(true)); err != nil {
				l.Fatalf("[ENT] failed creating schema resources: %v", err)
			}

			// 物模型-单位模块的附加迁移与种子（幂等）
			// Thing-model unit module extra migration & seed (idempotent).
			//   1) partial unique index：每分类仅一个基准单位（Ent schema 难以表达）
			//   2) 出厂预置 42 分类 + ~225 单位（tenant_id=0，按 code upsert）
			// 失败仅记录日志，不阻断服务启动，便于线上回滚。
			// On failure we log only and continue, so service startup is not blocked.
			if seedErr := seed.EnsurePartialIndexes(ctx.Context(), drv, l); seedErr != nil {
				l.Warnf("[ENT] ensure thingmodel partial indexes failed: %v", seedErr)
			}
			if seedErr := seed.SeedThingmodelUnits(ctx.Context(), client, l); seedErr != nil {
				l.Warnf("[ENT] seed thingmodel units failed: %v", seedErr)
			}

			// 物模型-特征模块种子（幂等，依赖单位种子先行）
			// Thing-model feature module seed (idempotent; depends on unit seed).
			//   - ~270 条出厂特征（属性/事件/服务/关系），tenant_id=0，按 code upsert
			//   - property 的 spec.unit.unitCode 在 upsert 前解析为 unitId
			//   - relation 的 source/target.identifier 在第二遍 upsert 时回填 id
			if seedErr := seed.SeedThingmodelFeatures(ctx.Context(), client, l); seedErr != nil {
				// 用 Errorf 让用户在启动日志中第一时间能看到具体失败原因。
				// Errorf so users immediately notice seed failures in startup logs.
				l.Errorf("[ENT] seed thingmodel features failed: %v", seedErr)
			}

			// 物模型-分类模块种子（幂等）
			// Thing-model category module seed (idempotent).
			//   - 三套国标清单：智能系统(30-36)/空间(10)/设备设施(20-26)，约 600~750 节点
			//   - tenant_id=0、按 (tenant_id, kind, code) upsert
			//   - 单表 + kind 枚举承载多业务域；4 层固定，code 变长 2/4/6/8
			if seedErr := seed.SeedThingmodelCategories(ctx.Context(), client, l); seedErr != nil {
				l.Errorf("[ENT] seed thingmodel categories failed: %v", seedErr)
			}

			// 物模型-模型管理种子（幂等）— 必须最后跑，依赖前三个种子结果
			// Thing-model model management seed (idempotent).
			//   - 1 个示范分类默认模型：电动压缩式冷水机组 (20010100) → 10 条特征 (含 2 条 override)
			//   - 1 个示范产品：GREE-LSBLG320 (status=PUBLISHED)
			//   - 9 条产品特征：8 条 DEFAULT (跳过制冷量/所属系统) + 1 条 LOCAL (夜间静音)
			// 详见 docs/thingmodel/sheji/模型管理/06-种子数据与实施计划.md
			if seedErr := seed.SeedModelManagement(ctx.Context(), client, l); seedErr != nil {
				l.Errorf("[ENT] seed model management failed: %v", seedErr)
			}
		}

		return client
	})
	if err != nil {
		log.Fatalf("[ENT] failed creating ent client: %v", err)
		return nil, func() {}, err
	}

	return cli, func() {
		if cleanErr := cli.Close(); cleanErr != nil {
			log.Errorf("[ENT] failed closing ent client: %v", cleanErr)
		}
	}, nil
}
