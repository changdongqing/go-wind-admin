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
