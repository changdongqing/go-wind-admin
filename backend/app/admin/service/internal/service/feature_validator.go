package service

import (
	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"
)

// feature_validator.go: 四类特征 spec 校验器（按 feature_type 分派）。
// 设计依据 / Design ref: docs/thingmodel/sheji/10-特征参数与spec设计.md §6
//
// 校验规则总表 V1~V17（V8/V15 需 DB 访问，在 service 层补充；纯函数仅校验结构）：
//   - V1..V9: PROPERTY
//   - V10:    EVENT
//   - V11..V12: SERVICE
//   - V13..V16: RELATION
//   - V17: 特化列与 spec 一致（由 syncSpecializedColumns 保证写入一致性）

// validateFeatureSpecForType 按 feature_type + FeatureSpec oneof 分派校验
// Dispatch validator by feature_type and FeatureSpec oneof branch.
func validateFeatureSpecForType(ft thingmodelV1.FeatureType, spec *thingmodelV1.FeatureSpec) []string {
	if spec == nil || spec.Spec == nil {
		return []string{"spec is required"}
	}
	switch ft {
	case thingmodelV1.FeatureType_PROPERTY:
		sp, ok := spec.Spec.(*thingmodelV1.FeatureSpec_Property)
		if !ok || sp.Property == nil {
			return []string{"spec.property required for PROPERTY feature"}
		}
		return validatePropertySpec(sp.Property)
	case thingmodelV1.FeatureType_EVENT:
		sp, ok := spec.Spec.(*thingmodelV1.FeatureSpec_Event)
		if !ok || sp.Event == nil {
			return []string{"spec.event required for EVENT feature"}
		}
		return validateEventSpec(sp.Event)
	case thingmodelV1.FeatureType_SERVICE:
		sp, ok := spec.Spec.(*thingmodelV1.FeatureSpec_Service)
		if !ok || sp.Service == nil {
			return []string{"spec.service required for SERVICE feature"}
		}
		return validateServiceSpec(sp.Service)
	case thingmodelV1.FeatureType_RELATION:
		sp, ok := spec.Spec.(*thingmodelV1.FeatureSpec_Relation)
		if !ok || sp.Relation == nil {
			return []string{"spec.relation required for RELATION feature"}
		}
		return validateRelationSpec(sp.Relation)
	default:
		// FEATURE_TYPE_UNSPECIFIED：按 oneof 自动推断
		return validateFeatureSpecAuto(spec)
	}
}

// validateFeatureSpecAuto 未给 featureType 时按 oneof 自动分派 / Auto dispatch by oneof
func validateFeatureSpecAuto(spec *thingmodelV1.FeatureSpec) []string {
	switch s := spec.Spec.(type) {
	case *thingmodelV1.FeatureSpec_Property:
		return validatePropertySpec(s.Property)
	case *thingmodelV1.FeatureSpec_Event:
		return validateEventSpec(s.Event)
	case *thingmodelV1.FeatureSpec_Service:
		return validateServiceSpec(s.Service)
	case *thingmodelV1.FeatureSpec_Relation:
		return validateRelationSpec(s.Relation)
	default:
		return []string{"unknown spec type"}
	}
}

// validatePropertySpec 属性 spec 校验（V1~V9）
// Property spec validator (rules V1~V9)
func validatePropertySpec(p *thingmodelV1.PropertySpec) []string {
	if p == nil {
		return []string{"property spec required"}
	}
	var errs []string
	dt := p.GetDataType()

	// V1: dataType 必填
	if dt == thingmodelV1.DataType_DATA_TYPE_UNSPECIFIED {
		errs = append(errs, "property.dataType required")
	}
	// V2: accessMode 必填
	if p.GetAccessMode() == thingmodelV1.AccessMode_ACCESS_MODE_UNSPECIFIED {
		errs = append(errs, "property.accessMode required")
	}
	// V3: ENUM 必须有 enumItems
	if dt == thingmodelV1.DataType_ENUM && len(p.GetEnumItems()) == 0 {
		errs = append(errs, "property.enumItems required for ENUM type")
	}
	// V4: STRUCT 必须有 structFields
	if dt == thingmodelV1.DataType_STRUCT && len(p.GetStructFields()) == 0 {
		errs = append(errs, "property.structFields required for STRUCT type")
	}
	// V5: ARRAY 必须有 arraySpec.element
	if dt == thingmodelV1.DataType_ARRAY {
		if p.GetArraySpec() == nil || p.GetArraySpec().GetElement() == nil {
			errs = append(errs, "property.arraySpec.element required for ARRAY type")
		}
	}
	// V7: constraints.min ≤ max（仅当两者都存在时）
	if c := p.GetConstraints(); c != nil {
		if c.GetMin() != 0 || c.GetMax() != 0 {
			if c.GetMin() > c.GetMax() {
				errs = append(errs, "property.constraints.min must be <= max")
			}
		}
	}
	// V9: category=rated 时 isRated=true
	if p.GetCategory() == "rated" && !p.GetIsRated() {
		errs = append(errs, "property.isRated must be true when category=rated")
	}

	// 递归 struct 子字段（轻量结构校验）
	for i, f := range p.GetStructFields() {
		for _, e := range validateParamSpec(f) {
			errs = append(errs, "property.structFields["+itoa(uint32(i))+"]: "+e)
		}
	}
	// 递归 array 元素
	if as := p.GetArraySpec(); as != nil && as.GetElement() != nil {
		for _, e := range validateParamSpec(as.GetElement()) {
			errs = append(errs, "property.arraySpec.element: "+e)
		}
	}
	return errs
}

// validateEventSpec 事件 spec 校验（V10）
func validateEventSpec(e *thingmodelV1.EventSpec) []string {
	if e == nil {
		return []string{"event spec required"}
	}
	var errs []string
	if e.GetLevel() == thingmodelV1.EventLevel_EVENT_LEVEL_UNSPECIFIED {
		errs = append(errs, "event.level required")
	}
	for i, p := range e.GetOutputParams() {
		for _, er := range validateParamSpec(p) {
			errs = append(errs, "event.outputParams["+itoa(uint32(i))+"]: "+er)
		}
	}
	return errs
}

// validateServiceSpec 服务 spec 校验（V11~V12）
func validateServiceSpec(s *thingmodelV1.ServiceSpec) []string {
	if s == nil {
		return []string{"service spec required"}
	}
	var errs []string
	if s.GetCallMode() == thingmodelV1.CallMode_CALL_MODE_UNSPECIFIED {
		errs = append(errs, "service.callMode required")
	}
	for i, p := range s.GetInputParams() {
		for _, e := range validateParamSpec(p) {
			errs = append(errs, "service.inputParams["+itoa(uint32(i))+"]: "+e)
		}
	}
	for i, p := range s.GetOutputParams() {
		for _, e := range validateParamSpec(p) {
			errs = append(errs, "service.outputParams["+itoa(uint32(i))+"]: "+e)
		}
	}
	return errs
}

// validateRelationSpec 关系 spec 校验（V13~V16）
// V15 的目标存在性校验需 DB 访问，在 service 层 validateRelationTargets 中补充。
func validateRelationSpec(r *thingmodelV1.RelationSpec) []string {
	if r == nil {
		return []string{"relation spec required"}
	}
	var errs []string
	if r.GetRelationType() == "" {
		errs = append(errs, "relation.relationType required")
	}
	if r.GetSource() == nil || r.GetTarget() == nil {
		errs = append(errs, "relation.source and target required")
	}
	// kind 校验
	checkKind := func(label string, ref *thingmodelV1.EntityRef) {
		if ref == nil {
			return
		}
		k := ref.GetKind()
		if k != "" && k != "feature" && k != "external" {
			errs = append(errs, label+".kind must be feature or external")
		}
		if k == "feature" && ref.GetId() == 0 && ref.GetCode() == "" && ref.GetIdentifier() == "" {
			errs = append(errs, label+": kind=feature requires id/code/identifier")
		}
	}
	checkKind("relation.source", r.GetSource())
	checkKind("relation.target", r.GetTarget())
	// 基数取值检查（非强制）
	if c := r.GetCardinality(); c != "" {
		switch c {
		case "oneToOne", "oneToMany", "manyToOne", "manyToMany":
		default:
			errs = append(errs, "relation.cardinality invalid: "+c)
		}
	}
	return errs
}

// validateParamSpec 通用参数 spec 校验（事件/服务参数、struct 子字段共用）
// Generic ParamSpec validator (used recursively by struct fields / event-service params).
func validateParamSpec(p *thingmodelV1.ParamSpec) []string {
	if p == nil {
		return []string{"param required"}
	}
	var errs []string
	if p.GetKey() == "" {
		errs = append(errs, "param.key required")
	}
	dt := p.GetDataType()
	if dt == thingmodelV1.DataType_DATA_TYPE_UNSPECIFIED {
		errs = append(errs, "param.dataType required")
	}
	if dt == thingmodelV1.DataType_ENUM && len(p.GetEnumItems()) == 0 {
		errs = append(errs, "param.enumItems required for ENUM type")
	}
	if dt == thingmodelV1.DataType_STRUCT && len(p.GetStructFields()) == 0 {
		errs = append(errs, "param.structFields required for STRUCT type")
	}
	if dt == thingmodelV1.DataType_ARRAY {
		if p.GetArraySpec() == nil || p.GetArraySpec().GetElement() == nil {
			errs = append(errs, "param.arraySpec.element required for ARRAY type")
		}
	}
	if c := p.GetConstraints(); c != nil {
		if (c.GetMin() != 0 || c.GetMax() != 0) && c.GetMin() > c.GetMax() {
			errs = append(errs, "param.constraints.min must be <= max")
		}
	}
	// 递归
	for i, f := range p.GetStructFields() {
		for _, e := range validateParamSpec(f) {
			errs = append(errs, "structFields["+itoa(uint32(i))+"]: "+e)
		}
	}
	if as := p.GetArraySpec(); as != nil && as.GetElement() != nil {
		for _, e := range validateParamSpec(as.GetElement()) {
			errs = append(errs, "arraySpec.element: "+e)
		}
	}
	return errs
}
