package seed

import (
	"context"
	"fmt"
	"math"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	"github.com/go-kratos/kratos/v2/log"

	"go-wind-admin/app/admin/service/internal/data/ent"
	"go-wind-admin/app/admin/service/internal/data/ent/unit"
	"go-wind-admin/app/admin/service/internal/data/ent/unitcategory"
	appViewer "go-wind-admin/pkg/entgo/viewer"
)

// piOver 返回 π/n，便于在数据表中表达角度换算系数。
// piOver returns π/n; used to express angle conversion factors in the data table.
func piOver(n float64) float64 { return math.Pi / n }

// SeedThingmodelUnits 将 UnitSeedData 写入数据库（幂等 upsert）。
// SeedThingmodelUnits writes UnitSeedData into the database idempotently.
//
//   - 仅维护系统预置数据（tenant_id = 0）；租户自建（tenant_id > 0）永不被覆盖。
//   - 以 (tenant_id, code) 为键 upsert：
//   - 分类：UnitCategory 的 (tenant_id, code) unique 唯一索引
//   - 单位：Unit 的 (tenant_id, code) unique 唯一索引
//   - reference_count 字段在 upsert 时**不覆盖**，避免清空运行时累计的引用计数。
//   - logger 可为 nil（用全局 log）。/ logger may be nil; falls back to global log.
//
// 重要：内部把 ctx 包成 SystemViewerContext，否则 TenantPrivacy 策略会拒绝写入
// (返回 "security: missing ViewerContext in context")。
// IMPORTANT: wraps ctx with SystemViewerContext, otherwise TenantPrivacy rejects writes.
func SeedThingmodelUnits(ctx context.Context, client *ent.Client, logger *log.Helper) error {
	if client == nil {
		return fmt.Errorf("seed: ent client is nil")
	}
	if logger == nil {
		logger = log.NewHelper(log.With(log.GetLogger(), "module", "thingmodel-unit-seed"))
	}

	// 系统视图：绕过 TenantPrivacy 过滤，允许写入 tenant_id=0 的系统预置数据。
	// System viewer context: bypasses TenantPrivacy so we can write tenant_id=0 system rows.
	ctx = appViewer.NewSystemViewerContext(ctx)

	now := time.Now()
	totalCats, totalUnits := 0, 0

	for _, cat := range UnitSeedData {
		catID, err := upsertCategory(ctx, client, cat, now)
		if err != nil {
			logger.Errorf("seed category %s failed: %v", cat.Code, err)
			return fmt.Errorf("seed category %s: %w", cat.Code, err)
		}
		totalCats++

		for _, u := range cat.Units {
			if err := upsertUnit(ctx, client, catID, u, now); err != nil {
				logger.Errorf("seed unit %s/%s failed: %v", cat.Code, u.Code, err)
				return fmt.Errorf("seed unit %s/%s: %w", cat.Code, u.Code, err)
			}
			totalUnits++
		}
	}

	logger.Infof("[seed] thingmodel units: %d categories, %d units upserted", totalCats, totalUnits)
	return nil
}

// upsertCategory 按 (tenant_id=0, code) upsert UnitCategory，返回分类 ID。
// upsertCategory upserts UnitCategory by (tenant_id=0, code) and returns its id.
func upsertCategory(ctx context.Context, client *ent.Client, c SeedCategory, now time.Time) (uint32, error) {
	const sysTenant uint32 = 0

	if err := client.UnitCategory.Create().
		SetTenantID(sysTenant).
		SetCode(c.Code).
		SetName(c.Name).
		SetNameEn(c.NameEn).
		SetQuantity(c.Quantity).
		SetBaseUnitSymbol(c.BaseUnitSymbol).
		SetDescription(c.Description).
		SetIsEnabled(true).
		SetSortOrder(c.SortOrder).
		SetCreatedAt(now).
		OnConflictColumns(unitcategory.FieldTenantID, unitcategory.FieldCode).
		UpdateNewValues().
		// updated_at 由 OnConflict 默认更新；created_at 由 NewValues 仅在 INSERT 列上更新
		Exec(ctx); err != nil {
		return 0, err
	}

	cat, err := client.UnitCategory.Query().
		Where(unitcategory.TenantIDEQ(sysTenant), unitcategory.CodeEQ(c.Code)).
		Only(ctx)
	if err != nil {
		return 0, fmt.Errorf("query category after upsert: %w", err)
	}
	return cat.ID, nil
}

// upsertUnit 按 (tenant_id=0, code) upsert Unit。
// upsertUnit upserts Unit by (tenant_id=0, code).
//
// 关键点 / Key points:
//   - 不写 reference_count（保留运行时累计值，由 thing_property 模块维护）。
//   - is_base=true 与 partial unique index `uix_thingmodel_unit_base` 的冲突由
//     调用方在 migrate 后建索引保证；种子内每分类恰好一个 is_base=true。
func upsertUnit(ctx context.Context, client *ent.Client, categoryID uint32, u SeedUnit, now time.Time) error {
	const sysTenant uint32 = 0

	ct, ok := protoConversionTypeToEnt(u.ConversionType)
	if !ok {
		// 不应发生 / Should not happen for seed data
		return fmt.Errorf("invalid conversion type for unit %s", u.Code)
	}

	return client.Unit.Create().
		SetTenantID(sysTenant).
		SetCategoryID(categoryID).
		SetCode(u.Code).
		SetSymbol(u.Symbol).
		SetName(u.Name).
		SetNameEn(u.NameEn).
		SetIsBase(u.IsBase).
		SetConversionType(ct).
		SetFactor(u.Factor).
		SetOffset(u.Offset).
		SetFormulaExpr(u.FormulaExpr).
		SetPrecision(u.Precision).
		SetIsSiUnit(u.IsSiUnit).
		SetIsLegalUnit(u.IsLegalUnit).
		SetIsEnabled(true).
		SetSortOrder(u.SortOrder).
		SetReferenceCount(0).
		SetCreatedAt(now).
		OnConflictColumns(unit.FieldTenantID, unit.FieldCode).
		Update(func(up *ent.UnitUpsert) {
			// 重要：不覆盖 reference_count，避免清空运行时累计值。
			// Note: do NOT overwrite reference_count to preserve runtime counters.
			up.UpdateCategoryID().
				UpdateSymbol().
				UpdateName().
				UpdateNameEn().
				UpdateIsBase().
				UpdateConversionType().
				UpdateFactor().
				UpdateOffset().
				UpdateFormulaExpr().
				UpdatePrecision().
				UpdateIsSiUnit().
				UpdateIsLegalUnit().
				UpdateIsEnabled().
				UpdateSortOrder().
				SetUpdatedAt(now)
		}).
		Exec(ctx)
}

// protoConversionTypeToEnt 将种子用的 proto enum 转为 ent enum。
// protoConversionTypeToEnt maps proto ConversionType to ent ConversionType.
func protoConversionTypeToEnt(t any) (unit.ConversionType, bool) {
	s := fmt.Sprintf("%v", t)
	switch s {
	case "LINEAR":
		return unit.ConversionTypeLinear, true
	case "AFFINE":
		return unit.ConversionTypeAffine, true
	case "LOGARITHMIC":
		return unit.ConversionTypeLogarithmic, true
	case "CONDITIONAL":
		return unit.ConversionTypeConditional, true
	case "NONE":
		return unit.ConversionTypeNone, true
	}
	return "", false
}

// EnsurePartialIndexes 通过传入的 sql.Driver 创建 partial unique index。
// EnsurePartialIndexes creates partial unique indexes through the provided sql driver.
//
// 索引：每个 (tenant_id, category_id) 仅一行 is_base=true 的单位。
// Index: at most one is_base=true unit per (tenant_id, category_id).
//
// PostgreSQL / SQLite 支持过滤索引；MySQL 不支持，跳过（依赖 unit_repo 中的事务切换）。
// PostgreSQL/SQLite support filtered unique indexes; MySQL falls back to app-level tx.
func EnsurePartialIndexes(ctx context.Context, drv *entsql.Driver, logger *log.Helper) error {
	if drv == nil {
		return fmt.Errorf("seed: sql driver is nil")
	}
	if logger == nil {
		logger = log.NewHelper(log.With(log.GetLogger(), "module", "thingmodel-unit-seed"))
	}

	exec := func(stmt string) error {
		_, err := drv.DB().ExecContext(ctx, stmt)
		return err
	}

	switch drv.Dialect() {
	case dialect.Postgres:
		stmt := `CREATE UNIQUE INDEX IF NOT EXISTS uix_thingmodel_unit_base ` +
			`ON thingmodel_units (tenant_id, category_id) WHERE is_base = true`
		if err := exec(stmt); err != nil {
			return fmt.Errorf("create partial unique index (postgres): %w", err)
		}
		logger.Info("[seed] partial unique index uix_thingmodel_unit_base ensured (PostgreSQL)")
	case dialect.SQLite:
		stmt := `CREATE UNIQUE INDEX IF NOT EXISTS uix_thingmodel_unit_base ` +
			`ON thingmodel_units (tenant_id, category_id) WHERE is_base = 1`
		if err := exec(stmt); err != nil {
			return fmt.Errorf("create partial unique index (sqlite): %w", err)
		}
		logger.Info("[seed] partial unique index uix_thingmodel_unit_base ensured (SQLite)")
	case dialect.MySQL:
		logger.Warn("[seed] MySQL detected: partial unique index NOT supported; relying on app-level base-switch tx")
	default:
		logger.Warnf("[seed] unknown dialect %q, skip partial unique index", drv.Dialect())
	}
	return nil
}
