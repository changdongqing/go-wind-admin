// 模型管理种子程序：分类默认模型 + 产品 + 产品特征 的最小可演示链路。
// Model management seed: minimal end-to-end demo of category default model + product + product features.
//
// 设计依据 / Design ref:
//   - docs/thingmodel/sheji/模型管理/06-种子数据与实施计划.md
//   - docs/thingmodel/sheji/修改记录/CR-001-结构化约束下沉到模型层.md
//
// 链路 / Demo chain:
//   FACILITY 细类 "20010100 电动压缩式冷水机组"
//     └─ 默认模型：10 条 CDF（CR-001 后承载完整 spec）
//     └─ 示范产品 "GREE-LSBLG320 格力螺杆冷水机组 LSBLG320"
//         └─ 产品特征：8 条 source=DEFAULT（跳过 P-HVAC-0005 与 R-HVAC-0001 演示"逐个勾选"）
//                  + 1 条 source=LOCAL  L-NIGHT-MUTE "夜间静音模式"
//
// CR-001 变更：
//   - 原 CDF.override_spec / PF.feature_snapshot + override_spec 三字段融合为单一 spec；
//   - spec 由 feature_seed_data.go 中的 SeedFeature.Spec map 通过 BuildFeatureSpecFromMap 构造，
//     再合并 modelMgmtDefaultConstraints / modelMgmtProductConstraints 中的演示性收紧值。
//
// 幂等：所有 upsert 走 OnConflict.Ignore；重跑不报错、不重复插入。

package seed

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"go-wind-admin/app/admin/service/internal/data/ent"
	"go-wind-admin/app/admin/service/internal/data/ent/category"
	"go-wind-admin/app/admin/service/internal/data/ent/categorydefaultfeature"
	"go-wind-admin/app/admin/service/internal/data/ent/feature"
	"go-wind-admin/app/admin/service/internal/data/ent/product"
	"go-wind-admin/app/admin/service/internal/data/ent/productfeature"
	"go-wind-admin/app/admin/service/internal/data/ent/schema"
	appViewer "go-wind-admin/pkg/entgo/viewer"

	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"
)

const (
	sysTenantModelMgmt    uint32 = 0
	demoCategoryCode             = "20010100" // 电动压缩式冷水机组（FACILITY level=4）
	demoProductCode              = "GREE-LSBLG320"
	demoProductName              = "格力螺杆冷水机组 LSBLG320"
	demoLocalFeatureCode         = "L-NIGHT-MUTE"
	demoLocalFeatureIdent        = "nightMute"
	demoLocalFeatureName         = "夜间静音模式"
)

// modelMgmtFeatureCodes 默认模型应包含的全局特征 code 列表（10 条）
var modelMgmtFeatureCodes = []string{
	"P-RUN-0001",  // 开关状态 (通用 property)
	"P-RUN-0002",  // 运行模式 (通用 property)
	"P-HVAC-0001", // 供水温度（冷冻水出水）
	"P-HVAC-0002", // 回水温度（冷冻水回水）
	"P-HVAC-0005", // 水流量
	"P-HVAC-0007", // 设定温度
	"E-GEN-0001",  // 设备上线
	"E-HVAC-0001", // HVAC 故障事件
	"S-GEN-0001",  // 开机（通用服务）
	"R-HVAC-0001", // 冷源系统关系
}

// modelMgmtDefaultConstraints 分类层 spec 的演示性收紧（合并到 cdf.spec.property.constraints）。
// CR-001 后：取代原 override_spec，直接在 spec 构造时合并。
//
// P-HVAC-0001 (供水温度=冷冻水出水)：冷水机组下通常 4~15℃
// P-HVAC-0002 (回水温度=冷冻水回水)：通常 6~18℃
var modelMgmtDefaultConstraints = map[string]*thingmodelV1.ValueConstraints{
	"P-HVAC-0001": {Min: ptrFloat64(4), Max: ptrFloat64(15)},
	"P-HVAC-0002": {Min: ptrFloat64(6), Max: ptrFloat64(18)},
}

// modelMgmtSkipDefaultPull 产品创建时跳过拉取的 feature code（演示"逐个勾选"）
var modelMgmtSkipDefaultPull = map[string]bool{
	"P-HVAC-0005": true, // 水流量
	"R-HVAC-0001": true, // 冷源系统关系
}

// modelMgmtProductConstraints 产品层 spec 的二次收紧（覆盖 cdf 拷贝过来的 spec）
// 出水温度：分类层 4-15 → 产品层进一步收紧到 5-12（厂家工况）
var modelMgmtProductConstraints = map[string]*thingmodelV1.ValueConstraints{
	"P-HVAC-0001": {Min: ptrFloat64(5), Max: ptrFloat64(12)},
}

// SeedModelManagement 执行模型管理种子（必须在 categories/features/units 种子之后调用）。
//
// 失败容忍：依赖前置数据存在；若 category/feature 缺失则记 WARN 并跳过，不阻断启动。
func SeedModelManagement(ctx context.Context, client *ent.Client, logger *log.Helper) error {
	if client == nil {
		return fmt.Errorf("seed: ent client is nil")
	}
	// 系统视图：绕过 TenantPrivacy 过滤，允许读/写 tenant_id=0 全局数据
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

	// 2. 取 10 个特征骨架
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

	// 3. 构建 seed-data 索引（code → SeedFeature）：从 feature_seed_data*.go 拿 spec map
	seedByCode := buildSeedFeatureByCode()

	// 单位 code → id 索引 & feature identifier → id 索引（spec 内部解析用）
	unitIdx, err := BuildUnitCodeIndex(ctx, client)
	if err != nil {
		return fmt.Errorf("seed: build unit index: %w", err)
	}
	featIdx, err := BuildFeatureIdentifierIndex(ctx, client)
	if err != nil {
		return fmt.Errorf("seed: build feature identifier index: %w", err)
	}

	// 4. upsert 分类默认模型条目（10 条，CR-001 后承载完整 spec）
	createdCDFs := 0
	skippedCDFs := 0
	cdfSpecByCode := make(map[string]*thingmodelV1.FeatureSpec, len(modelMgmtFeatureCodes))
	for i, code := range modelMgmtFeatureCodes {
		f, ok := featByCode[code]
		if !ok {
			logger.Warnf("[model-mgmt-seed] feature %s missing in DB; skip default entry", code)
			skippedCDFs++
			continue
		}

		// 构造完整 spec：seed data → proto FeatureSpec → 合并分类层 constraints 收紧
		cdfSpec := buildCDFSpec(code, seedByCode, unitIdx, featIdx)
		applyConstraintsOverride(cdfSpec, modelMgmtDefaultConstraints[code])
		cdfSpecByCode[code] = cdfSpec

		builder := client.CategoryDefaultFeature.Create().
			SetTenantID(sysTenantModelMgmt).
			SetCategoryID(cat.ID).
			SetFeatureID(f.ID).
			SetSortOrder(uint32(i)).
			SetIsEnabled(true).
			SetCreatedAt(now)
		if cdfSpec != nil {
			builder.SetSpec(schema.WrapFeatureSpec(cdfSpec))
		}
		// 同步冗余特化列
		applyCDFSpecializedCols(builder, cdfSpec)

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

	// 5. upsert 产品（GREE-LSBLG320）
	productID, err := upsertDemoProduct(ctx, client, cat.ID, now)
	if err != nil {
		return fmt.Errorf("seed: upsert demo product: %w", err)
	}

	// 6. upsert 产品特征：8 条 DEFAULT + 1 条 LOCAL
	createdPFs := 0
	skippedPFs := 0
	for i, code := range modelMgmtFeatureCodes {
		if modelMgmtSkipDefaultPull[code] {
			continue
		}
		f, ok := featByCode[code]
		if !ok {
			continue
		}

		// CR-001：产品 spec 从 cdf.spec 深拷贝（buildCDFSpec 已经构造过）；
		// 再叠加 modelMgmtProductConstraints 做产品层收紧。
		pfSpec := cloneFeatureSpec(cdfSpecByCode[code])
		applyConstraintsOverride(pfSpec, modelMgmtProductConstraints[code])

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
		if pfSpec != nil {
			builder.SetSpec(schema.WrapFeatureSpec(pfSpec))
		}
		applyPFSpecializedCols(builder, pfSpec)

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

// ===========================================================================
// Helpers
// ===========================================================================

// buildSeedFeatureByCode 把 AllFeatureSeeds() 索引为 code → SeedFeature，便于 CDF 构造 spec。
func buildSeedFeatureByCode() map[string]SeedFeature {
	all := AllFeatureSeeds()
	m := make(map[string]SeedFeature, len(all))
	for _, sf := range all {
		m[sf.Code] = sf
	}
	return m
}

// buildCDFSpec 为单个 feature code 从 seed-data 拿 spec map 并构造完整 proto FeatureSpec。
// 同时解析 unit code → id（顶层 + 嵌套）以及 relation source/target identifier → id。
func buildCDFSpec(code string, seedByCode map[string]SeedFeature, unitIdx, featIdx map[string]uint32) *thingmodelV1.FeatureSpec {
	sf, ok := seedByCode[code]
	if !ok || sf.Spec == nil {
		return nil
	}
	// 解析 unit/relation 引用（在 map 阶段做，避免重复实现 proto 路径）
	if sf.FeatureType == ftProperty {
		ResolveUnitID(sf.Spec, unitIdx)
	} else if sf.FeatureType == ftRelation {
		ResolveRelationRefs(sf.Spec, featIdx)
	}
	return BuildFeatureSpecFromMap(sf.FeatureType.String(), sf.Spec)
}

// applyConstraintsOverride 在 spec.property.constraints 上叠加 override（仅 min/max/step，演示用）。
func applyConstraintsOverride(spec *thingmodelV1.FeatureSpec, over *thingmodelV1.ValueConstraints) {
	if spec == nil || over == nil {
		return
	}
	p := spec.GetProperty()
	if p == nil {
		return
	}
	if p.Constraints == nil {
		p.Constraints = &thingmodelV1.ValueConstraints{}
	}
	if over.Min != nil {
		p.Constraints.Min = over.Min
	}
	if over.Max != nil {
		p.Constraints.Max = over.Max
	}
	if over.Step != nil {
		p.Constraints.Step = over.Step
	}
}

// cloneFeatureSpec 深拷贝一份 spec（避免 CDF 和 PF 共享指针被改）。
func cloneFeatureSpec(s *thingmodelV1.FeatureSpec) *thingmodelV1.FeatureSpec {
	if s == nil {
		return nil
	}
	// proto.Clone 走 reflection 深拷贝（与 ent JSON 反序列化语义一致）。
	// 不导入 proto 包以减少依赖：手动重新构造一份 oneof。
	out := &thingmodelV1.FeatureSpec{}
	switch v := s.Spec.(type) {
	case *thingmodelV1.FeatureSpec_Property:
		if v.Property != nil {
			p := *v.Property
			if p.Constraints != nil {
				c := *p.Constraints
				p.Constraints = &c
			}
			if p.Unit != nil {
				u := *p.Unit
				p.Unit = &u
			}
			if p.BoolLabels != nil {
				bl := *p.BoolLabels
				p.BoolLabels = &bl
			}
			out.Spec = &thingmodelV1.FeatureSpec_Property{Property: &p}
		}
	case *thingmodelV1.FeatureSpec_Event:
		if v.Event != nil {
			e := *v.Event
			out.Spec = &thingmodelV1.FeatureSpec_Event{Event: &e}
		}
	case *thingmodelV1.FeatureSpec_Service:
		if v.Service != nil {
			svc := *v.Service
			out.Spec = &thingmodelV1.FeatureSpec_Service{Service: &svc}
		}
	case *thingmodelV1.FeatureSpec_Relation:
		if v.Relation != nil {
			r := *v.Relation
			out.Spec = &thingmodelV1.FeatureSpec_Relation{Relation: &r}
		}
	default:
		return nil
	}
	return out
}

// applyCDFSpecializedCols 同步 CDF 冗余特化列（从 spec 派生）。
func applyCDFSpecializedCols(b *ent.CategoryDefaultFeatureCreate, spec *thingmodelV1.FeatureSpec) {
	if spec == nil {
		return
	}
	if p := spec.GetProperty(); p != nil {
		if p.DataType != nil {
			b.SetDataType(categorydefaultfeature.DataType(p.DataType.String()))
		}
		if p.AccessMode != nil {
			b.SetAccessMode(categorydefaultfeature.AccessMode(p.AccessMode.String()))
		}
	}
	if e := spec.GetEvent(); e != nil && e.Level != nil {
		b.SetEventLevel(categorydefaultfeature.EventLevel(e.Level.String()))
	}
	if s := spec.GetService(); s != nil && s.CallMode != nil {
		b.SetCallMode(categorydefaultfeature.CallMode(s.CallMode.String()))
	}
	if r := spec.GetRelation(); r != nil && r.RelationType != nil {
		b.SetRelationType(*r.RelationType)
	}
}

// applyPFSpecializedCols 同步 ProductFeature 冗余特化列（从 spec 派生）。
func applyPFSpecializedCols(b *ent.ProductFeatureCreate, spec *thingmodelV1.FeatureSpec) {
	if spec == nil {
		return
	}
	if p := spec.GetProperty(); p != nil {
		if p.DataType != nil {
			b.SetDataType(productfeature.DataType(p.DataType.String()))
		}
		if p.AccessMode != nil {
			b.SetAccessMode(productfeature.AccessMode(p.AccessMode.String()))
		}
	}
	if e := spec.GetEvent(); e != nil && e.Level != nil {
		b.SetEventLevel(productfeature.EventLevel(e.Level.String()))
	}
	if s := spec.GetService(); s != nil && s.CallMode != nil {
		b.SetCallMode(productfeature.CallMode(s.CallMode.String()))
	}
	if r := spec.GetRelation(); r != nil && r.RelationType != nil {
		b.SetRelationType(*r.RelationType)
	}
}

// upsertDemoProduct upsert 示范产品并返回 id
func upsertDemoProduct(ctx context.Context, client *ent.Client, categoryID uint32, now time.Time) (uint32, error) {
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
		SetSpec(schema.WrapFeatureSpec(spec)).
		SetDataType(productfeature.DataTypeBool).
		SetAccessMode(productfeature.AccessModeRW).
		SetSortOrder(100).
		SetIsEnabled(true).
		SetCreatedAt(now).
		OnConflictColumns(productfeature.FieldProductID, productfeature.FieldCode).
		Ignore().
		Exec(ctx)
}

func ptrFloat64(v float64) *float64 { return &v }
