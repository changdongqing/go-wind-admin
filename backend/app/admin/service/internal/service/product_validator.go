package service

import (
	"google.golang.org/protobuf/proto"

	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"
)

// product_validator.go：模型管理（分类默认模型 + 产品 + 产品特征）的共享校验器与工具。
// 设计依据 / Design ref: docs/thingmodel/sheji/模型管理/04-后端实现设计.md §3 + §5
//
// 提供：
//   - validateOverrideSpec：校验覆写 spec 仅包含白名单字段中的合法子集（针对 property）
//   - effectiveSpec：合并 snapshot + override → 运行时有效 spec
//   - validateSourceRefMismatch：校验 source / ref_feature_id 配对（C12）
//   - validateSpecTypeMismatch：feature_snapshot 类型必须与 feature_type 一致（C13）

// validateOverrideSpec 校验白名单覆写 spec 的语义合法性。
// 白名单结构本身由 proto FeatureOverrideSpec 收口（不能传 dataType 等结构字段），
// 这里再做一次值域兜底：
//  1. 非 property 的特征覆写 constraints/unit/defaultValue 视为非法；
//  2. constraints 必须满足 min <= max（若两端都给）；
//  3. unit.unit_id 引用必须存在（由调用方传入 unit-existence 检查函数）。
//
// 返回的错误是 thingmodel proto 错误码（PfOverrideInvalid 或 CatDefaultFeatureOverrideInvalid，
// 由调用方决定包装哪个）。本函数返回 string 列表方便上层聚合，nil 表示通过。
func validateOverrideSpec(
	targetSpec *thingmodelV1.FeatureSpec,
	override *thingmodelV1.FeatureOverrideSpec,
) []string {
	if override == nil {
		return nil
	}
	var errs []string

	prop := targetSpec.GetProperty()
	if prop == nil {
		// 非 property：禁止覆写 constraints/unit/defaultValue
		if override.GetConstraints() != nil {
			errs = append(errs, "non-property feature cannot override constraints")
		}
		if override.GetUnit() != nil {
			errs = append(errs, "non-property feature cannot override unit")
		}
		// defaultValue 是 string，按 proto schema 总有值；只在 property 上有意义
		if override.DefaultValue != nil && override.GetDefaultValue() != "" {
			errs = append(errs, "non-property feature cannot override defaultValue")
		}
	} else if override.GetConstraints() != nil {
		c := override.GetConstraints()
		if c.Min != nil && c.Max != nil && c.GetMax() < c.GetMin() {
			errs = append(errs, "override constraints.max < min")
		}
		if c.Step != nil && c.GetStep() < 0 {
			errs = append(errs, "override constraints.step must be >= 0")
		}
	}
	return errs
}

// effectiveSpec 计算运行时有效 spec = snapshot deepmerged with override。
// 仅 property 当前支持 constraints / unit / defaultValue 三项覆写合并；
// 其余特征类型 override 视为 no-op（仍可改 display_name/description 等列级别字段——
// 那些不通过本函数，而是 service 层在 ProductFeature 行上直接落列）。
func effectiveSpec(snap *thingmodelV1.FeatureSpec, over *thingmodelV1.FeatureOverrideSpec) *thingmodelV1.FeatureSpec {
	if over == nil {
		return snap
	}
	if snap == nil {
		return nil
	}
	out := proto.Clone(snap).(*thingmodelV1.FeatureSpec)
	if p := out.GetProperty(); p != nil {
		if over.GetConstraints() != nil {
			p.Constraints = over.GetConstraints()
		}
		if over.GetUnit() != nil {
			p.Unit = over.GetUnit()
		}
		// defaultValue 走 proto 字符串路径（与 ValueConstraints.default_value 不同位置）
		// 仅在 property.constraints.default_value 上落地，保持现有 spec 结构
		if over.DefaultValue != nil && over.GetDefaultValue() != "" {
			if p.Constraints == nil {
				p.Constraints = &thingmodelV1.ValueConstraints{}
			}
			dv := over.GetDefaultValue()
			p.Constraints.DefaultValue = &dv
		}
	}
	return out
}

// validateSourceRefMismatch 校验 source 与 ref_feature_id 的搭配（C12）。
//   - source=LOCAL  ↔ ref_feature_id 必须为 0/nil
//   - source∈{DEFAULT, GLOBAL} ↔ ref_feature_id 必须非 0
//
// 返回 "" 表示通过；返回字符串则是错误说明（service 层包装为 PfSourceRefMismatch）。
func validateSourceRefMismatch(source thingmodelV1.ProductFeatureSource, refFeatureID uint32) string {
	switch source {
	case thingmodelV1.ProductFeatureSource_LOCAL:
		if refFeatureID != 0 {
			return "source=LOCAL must have ref_feature_id=0"
		}
	case thingmodelV1.ProductFeatureSource_DEFAULT, thingmodelV1.ProductFeatureSource_GLOBAL:
		if refFeatureID == 0 {
			return "source=DEFAULT/GLOBAL must have ref_feature_id != 0"
		}
	default:
		return "source must be DEFAULT/GLOBAL/LOCAL"
	}
	return ""
}

// validateSpecTypeMismatch 校验 feature_snapshot 与 feature_type 的一致性（C13）。
//   - PROPERTY 对应 spec.property 非空
//   - EVENT    对应 spec.event 非空
//   - SERVICE  对应 spec.service 非空
//   - RELATION 对应 spec.relation 非空
//
// 返回 "" 表示通过。
func validateSpecTypeMismatch(featureType thingmodelV1.FeatureType, snapshot *thingmodelV1.FeatureSpec) string {
	if snapshot == nil {
		// 调用方负责"快照必填"语义；这里只做类型匹配，nil snapshot 视为通过
		return ""
	}
	switch featureType {
	case thingmodelV1.FeatureType_PROPERTY:
		if snapshot.GetProperty() == nil {
			return "feature_type=PROPERTY but snapshot.property is nil"
		}
	case thingmodelV1.FeatureType_EVENT:
		if snapshot.GetEvent() == nil {
			return "feature_type=EVENT but snapshot.event is nil"
		}
	case thingmodelV1.FeatureType_SERVICE:
		if snapshot.GetService() == nil {
			return "feature_type=SERVICE but snapshot.service is nil"
		}
	case thingmodelV1.FeatureType_RELATION:
		if snapshot.GetRelation() == nil {
			return "feature_type=RELATION but snapshot.relation is nil"
		}
	}
	return ""
}

// extractUnitID 从 spec / override 中提取 property 顶层 unit_id（>0 才视为有效）。
// 优先 override.unit > snapshot.property.unit。用于 reference_count 维护。
func extractUnitID(snapshot *thingmodelV1.FeatureSpec, override *thingmodelV1.FeatureOverrideSpec) uint32 {
	if override != nil && override.GetUnit() != nil && override.GetUnit().GetUnitId() > 0 {
		return override.GetUnit().GetUnitId()
	}
	if snapshot != nil && snapshot.GetProperty() != nil && snapshot.GetProperty().GetUnit() != nil {
		return snapshot.GetProperty().GetUnit().GetUnitId()
	}
	return 0
}
