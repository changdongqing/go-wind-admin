package service

import (
	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"
)

// product_validator.go：模型管理（产品 + 产品特征）的共享校验器。
//
// 设计依据 / Design ref:
//   - docs/thingmodel/sheji/模型管理/04-后端实现设计.md §3 + §5
//   - docs/thingmodel/sheji/修改记录/CR-001-结构化约束下沉到模型层.md
//
// CR-001 后变更：
//   - 删除 validateOverrideSpec / effectiveSpec / extractUnitID（override 概念已消失）；
//   - 保留 validateSourceRefMismatch（C12）与 validateSpecTypeMismatch（C13），
//     spec 字段类型现在直接是产品的最终 spec（不再叫 snapshot）。

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

// validateSpecTypeMismatch 校验 spec 与 feature_type 的一致性（C13）。
//   - PROPERTY 对应 spec.property 非空
//   - EVENT    对应 spec.event 非空
//   - SERVICE  对应 spec.service 非空
//   - RELATION 对应 spec.relation 非空
//
// 返回 "" 表示通过。
func validateSpecTypeMismatch(featureType thingmodelV1.FeatureType, spec *thingmodelV1.FeatureSpec) string {
	if spec == nil {
		// 调用方负责"spec 必填"语义；这里只做类型匹配，nil spec 视为通过
		return ""
	}
	switch featureType {
	case thingmodelV1.FeatureType_PROPERTY:
		if spec.GetProperty() == nil {
			return "feature_type=PROPERTY but spec.property is nil"
		}
	case thingmodelV1.FeatureType_EVENT:
		if spec.GetEvent() == nil {
			return "feature_type=EVENT but spec.event is nil"
		}
	case thingmodelV1.FeatureType_SERVICE:
		if spec.GetService() == nil {
			return "feature_type=SERVICE but spec.service is nil"
		}
	case thingmodelV1.FeatureType_RELATION:
		if spec.GetRelation() == nil {
			return "feature_type=RELATION but spec.relation is nil"
		}
	}
	return ""
}
