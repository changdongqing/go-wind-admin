// 模型管理种子程序：分类默认模型 + 产品 + 产品特征 的最小可演示链路。
// Model management seed: minimal end-to-end demo of category default model + product + product features.
//
// 设计依据 / Design ref: docs/thingmodel/sheji/模型管理/06-种子数据与实施计划.md
//
// 链路 / Demo chain:
//   FACILITY 细类 "20010100 电动压缩式冷水机组"
//     └─ 默认模型：10 条特征（含 2 条 override）
//     └─ 示范产品 "GREE-LSBLG320 格力螺杆冷水机组 LSBLG320"
//         └─ 产品特征：8 条 source=DEFAULT（跳过制冷量 P-MEAS-0030 与 所属系统 R-PART-0001 演示"逐个勾选"）
//                  + 0 条 source=GLOBAL
//                  + 1 条 source=LOCAL  L-NIGHT-MUTE "夜间静音模式"
//
// 幂等：所有 upsert 走 OnConflict.Ignore；重跑不报错、不重复插入。

package seed

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/proto"

	"go-wind-admin/app/admin/service/internal/data/ent"
	"go-wind-admin/app/admin/service/internal/data/ent/categorydefaultfeature"
	"go-wind-admin/app/admin/service/internal/data/ent/category"
	"go-wind-admin/app/admin/service/internal/data/ent/feature"
	"go-wind-admin/app/admin/service/internal/data/ent/product"
	"go-wind-admin/app/admin/service/internal/data/ent/productfeature"
	"go-wind-admin/app/admin/service/internal/data/ent/schema"
	appViewer "go-wind-admin/pkg/entgo/viewer"

	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"
)

const (
	sysTenantModelMgmt    uint32 = 0
	demoCategoryCode             = "20010100"           // 电动压缩式冷水机组（FACILITY level=4）
	demoProductCode              = "GREE-LSBLG320"
	demoProductName              = "格力螺杆冷水机组 LSBLG320"
	demoLocalFeatureCode         = "L-NIGHT-MUTE"
	demoLocalFeatureIdent        = "nightMute"
	demoLocalFeatureName         = "夜间静音模式"
)

// modelMgmtFeatureCodes 默认模型应包含的全局特征 code 列表（10 条）
// 全部对应 feature_seed_data*.go 中实际存在的特征。
var modelMgmtFeatureCodes = []string{
	"P-RUN-0001",   // 开关状态 (通用 property)
	"P-RUN-0002",   // 运行模式 (通用 property)
	"P-HVAC-0001",  // 供水温度（冷冻水出水）
	"P-HVAC-0002",  // 回水温度（冷冻水回水）
	"P-HVAC-0005",  // 水流量
	"P-HVAC-0007",  // 设定温度
	"E-GEN-0001",   // 设备上线
	"E-HVAC-0001",  // HVAC 故障事件
	"S-GEN-0001",   // 开机（通用服务）
	"R-HVAC-0001",  // 冷源系统关系
}

// modelMgmtDefaultOverrides 分类层的轻量覆写（2 条收紧温度范围以演示）
//
// P-HVAC-0001 (供水温度=冷冻水出水)：冷水机组下通常 4~15℃
// P-HVAC-0002 (回水温度=冷冻水回水)：通常 6~18℃
var modelMgmtDefaultOverrides = map[string]*thingmodelV1.FeatureOverrideSpec{
	"P-HVAC-0001": {
		Constraints: &thingmodelV1.ValueConstraints{Min: ptrFloat64(4), Max: ptrFloat64(15)},
	},
	"P-HVAC-0002": {
		Constraints: &thingmodelV1.ValueConstraints{Min: ptrFloat64(6), Max: ptrFloat64(18)},
	},
}

// modelMgmtSkipDefaultPull 产品创建时跳过拉取的 feature code（演示"逐个勾选"）
var modelMgmtSkipDefaultPull = map[string]bool{
	"P-HVAC-0005":  true, // 水流量
	"R-HVAC-0001":  true, // 冷源系统关系
}

// modelMgmtProductOverrides 产品层在已拉取的特征上再次覆写（演示二次收紧）
// 出水温度：分类层 4-15 → 产品层进一步收紧到 5-12（厂家工况）
var modelMgmtProductOverrides = map[string]*thingmodelV1.FeatureOverrideSpec{
	"P-HVAC-0001": {
		Constraints: &thingmodelV1.ValueConstraints{Min: ptrFloat64(5), Max: ptrFloat64(12)},
	},
}

// SeedModelManagement 执行模型管理种子（必须在 categories/features/units 种子之后调用）。
//
// 失败容忍：依赖前置数据存在；若 category/feature 缺失则记 WARN 并跳过，不阻断启动。
func SeedModelManagement(ctx context.Context, client *ent.Client, logger *log.Helper) error {
	if client == nil {
		return fmt.Errorf("seed: ent client is nil")
	}
	// 系统视图：绕过 TenantPrivacy 过滤，允许读/写 tenant_id=0 全局数据
	// System viewer: bypasses TenantPrivacy, required for tenant_id=0 system seeds.
	ctx = appViewer.NewSystemViewerContext(ctx)
	now := time.Now()

	// 1. 取目标分类
	cat, err := client.Category.Query().
		Where(
			category.KindEQ(category.KindFacility),
			category.CodeEQ(demoCategoryCode),
			category.TenantIDEQ(sysTenantModelMgmt),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			logger.Warnf("[model-mgmt-seed] category %s not found; skip (categories seed must run first)", demoCategoryCode)
			return nil
		}
		return fmt.Errorf("seed: query category: %w", err)
	}

	// 2. 取 10 个特征
	feats, err := client.Feature.Query().
		Where(
			feature.CodeIn(modelMgmtFeatureCodes...),
			feature.TenantIDEQ(sysTenantModelMgmt),
		).
		All(ctx)
	if err != nil {
		return fmt.Errorf("seed: query features: %w", err)
	}
	if len(feats) == 0 {
		logger.Warnf("[model-mgmt-seed] no features found; skip (features seed must run first)")
		return nil
	}
	featByCode := make(map[string]*ent.Feature, len(feats))
	for _, f := range feats {
		if f.Code != nil {
			featByCode[*f.Code] = f
		}
	}

	// 3. upsert 分类默认模型条目（10 条）
	createdCDFs := 0
	skippedCDFs := 0
	for i, code := range modelMgmtFeatureCodes {
		f, ok := featByCode[code]
		if !ok {
			logger.Warnf("[model-mgmt-seed] feature %s missing in DB; skip default entry", code)
			skippedCDFs++
			continue
		}
		ov := modelMgmtDefaultOverrides[code]
		builder := client.CategoryDefaultFeature.Create().
			SetTenantID(sysTenantModelMgmt).
			SetCategoryID(cat.ID).
			SetFeatureID(f.ID).
			SetSortOrder(uint32(i)).
			SetIsEnabled(true).
			SetCreatedAt(now)
		if ov != nil {
			builder.SetOverrideSpec(schema.WrapFeatureOverrideSpec(ov))
		}
		err := builder.
			OnConflictColumns(
				categorydefaultfeature.FieldTenantID,
				categorydefaultfeature.FieldCategoryID,
				categorydefaultfeature.FieldFeatureID,
			).
			Ignore().
			Exec(ctx)
		if err != nil {
			logger.Errorf("[model-mgmt-seed] upsert cdf %s FAILED: %v", code, err)
			skippedCDFs++
			continue
		}
		createdCDFs++
	}
	logger.Infof("[model-mgmt-seed] category default features upserted=%d skipped=%d", createdCDFs, skippedCDFs)

	// 4. upsert 产品（GREE-LSBLG320）
	productID, err := upsertDemoProduct(ctx, client, cat.ID, now)
	if err != nil {
		return fmt.Errorf("seed: upsert demo product: %w", err)
	}

	// 5. upsert 产品特征：8 条 DEFAULT + 1 条 LOCAL
	createdPFs := 0
	skippedPFs := 0
	for i, code := range modelMgmtFeatureCodes {
		if modelMgmtSkipDefaultPull[code] {
			continue // 演示"逐个勾选"未拉取
		}
		f, ok := featByCode[code]
		if !ok {
			continue
		}
		// 合并 default 层 override 到 snapshot（值复制核心语义）
		snapshot := mergeOverrideForSeed(unwrapFeatureSpecRow(f), modelMgmtDefaultOverrides[code])

		builder := client.ProductFeature.Create().
			SetTenantID(sysTenantModelMgmt).
			SetProductID(productID).
			SetSource(productfeature.SourceDefault).
			SetRefFeatureID(f.ID).
			SetSortOrder(uint32(i)).
			SetIsEnabled(true).
			SetCreatedAt(now)
		if f.FeatureType != nil {
			builder.SetFeatureType(productfeature.FeatureType(*f.FeatureType))
		}
		if f.Code != nil {
			builder.SetCode(*f.Code)
		}
		if f.Identifier != nil {
			builder.SetIdentifier(*f.Identifier)
		}
		if f.Name != nil {
			builder.SetName(*f.Name)
		}
		if f.NameEn != nil {
			builder.SetNameEn(*f.NameEn)
		}
		if f.Description != nil {
			builder.SetDescription(*f.Description)
		}
		if f.DataType != nil {
			builder.SetDataType(productfeature.DataType(*f.DataType))
		}
		if f.AccessMode != nil {
			builder.SetAccessMode(productfeature.AccessMode(*f.AccessMode))
		}
		if f.EventLevel != nil {
			builder.SetEventLevel(productfeature.EventLevel(*f.EventLevel))
		}
		if f.CallMode != nil {
			builder.SetCallMode(productfeature.CallMode(*f.CallMode))
		}
		if f.RelationType != nil {
			builder.SetRelationType(*f.RelationType)
		}
		if snapshot != nil {
			builder.SetFeatureSnapshot(schema.WrapFeatureSpec(snapshot))
		}
		// 产品层二次覆写
		if ov, ok := modelMgmtProductOverrides[code]; ok && ov != nil {
			builder.SetOverrideSpec(schema.WrapFeatureOverrideSpec(ov))
		}
		err := builder.
			OnConflictColumns(productfeature.FieldProductID, productfeature.FieldCode).
			Ignore().
			Exec(ctx)
		if err != nil {
			logger.Errorf("[model-mgmt-seed] upsert pf %s FAILED: %v", code, err)
			skippedPFs++
			continue
		}
		createdPFs++
	}

	// LOCAL 特征：夜间静音
	if err := upsertLocalNightMute(ctx, client, productID, now); err != nil {
		logger.Errorf("[model-mgmt-seed] upsert local feature %s FAILED: %v", demoLocalFeatureCode, err)
		skippedPFs++
	} else {
		createdPFs++
	}
	logger.Infof("[model-mgmt-seed] product features upserted=%d skipped=%d", createdPFs, skippedPFs)
	logger.Infof("[model-mgmt-seed] complete: product=%d default_features=%d product_features=%d",
		productID, createdCDFs, createdPFs)
	return nil
}

// upsertDemoProduct upsert 示范产品并返回 id
func upsertDemoProduct(ctx context.Context, client *ent.Client, categoryID uint32, now time.Time) (uint32, error) {
	// 先查 (tenant_id, code) 看是否已存在
	existing, err := client.Product.Query().
		Where(
			product.TenantIDEQ(sysTenantModelMgmt),
			product.CodeEQ(demoProductCode),
		).
		First(ctx)
	if err == nil {
		return existing.ID, nil
	}
	if !ent.IsNotFound(err) {
		return 0, err
	}

	row, err := client.Product.Create().
		SetTenantID(sysTenantModelMgmt).
		SetCode(demoProductCode).
		SetName(demoProductName).
		SetCategoryID(categoryID).
		SetManufacturer("格力").
		SetModelNo("LSBLG320").
		SetStatus(product.StatusPublished).
		SetDescription("示范产品 — 演示从默认模型拉取与本地特征").
		SetIsEnabled(true).
		SetReferenceCount(0).
		SetCreatedAt(now).
		Save(ctx)
	if err != nil {
		return 0, err
	}
	return row.ID, nil
}

// upsertLocalNightMute LOCAL 特征种子：夜间静音模式（BOOL/RW）
func upsertLocalNightMute(ctx context.Context, client *ent.Client, productID uint32, now time.Time) error {
	dt := thingmodelV1.DataType_BOOL
	am := thingmodelV1.AccessMode_RW
	cat := "setting"
	spec := &thingmodelV1.FeatureSpec{
		Spec: &thingmodelV1.FeatureSpec_Property{
			Property: &thingmodelV1.PropertySpec{
				DataType:   &dt,
				AccessMode: &am,
				Category:   &cat,
				BoolLabels: &thingmodelV1.BoolLabels{FalseLabel: "关闭", TrueLabel: "开启"},
			},
		},
	}

	return client.ProductFeature.Create().
		SetTenantID(sysTenantModelMgmt).
		SetProductID(productID).
		SetSource(productfeature.SourceLocal).
		SetFeatureType(productfeature.FeatureTypeProperty).
		SetCode(demoLocalFeatureCode).
		SetIdentifier(demoLocalFeatureIdent).
		SetName(demoLocalFeatureName).
		SetDescription("厂家特有：夜间低噪运行模式").
		SetFeatureSnapshot(schema.WrapFeatureSpec(spec)).
		SetDataType(productfeature.DataTypeBool).
		SetAccessMode(productfeature.AccessModeRW).
		SetSortOrder(100).
		SetIsEnabled(true).
		SetCreatedAt(now).
		OnConflictColumns(productfeature.FieldProductID, productfeature.FieldCode).
		Ignore().
		Exec(ctx)
}

// mergeOverrideForSeed 合并 default 层 override 到 snapshot；仅 property 支持 constraints/unit/defaultValue
func mergeOverrideForSeed(snap *thingmodelV1.FeatureSpec, over *thingmodelV1.FeatureOverrideSpec) *thingmodelV1.FeatureSpec {
	if snap == nil {
		return nil
	}
	if over == nil {
		return snap
	}
	out := proto.Clone(snap).(*thingmodelV1.FeatureSpec)
	if p := out.GetProperty(); p != nil {
		if over.GetConstraints() != nil {
			p.Constraints = over.GetConstraints()
		}
		if over.GetUnit() != nil {
			p.Unit = over.GetUnit()
		}
	}
	return out
}

// unwrapFeatureSpecRow 从 ent.Feature 取出 *thingmodelV1.FeatureSpec
func unwrapFeatureSpecRow(f *ent.Feature) *thingmodelV1.FeatureSpec {
	if f == nil || f.Spec == nil {
		return nil
	}
	return schema.UnwrapFeatureSpec(f.Spec)
}

func ptrFloat64(v float64) *float64 { return &v }
