package schema

import (
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"

	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"
)

// FeatureOverrideSpecField 包装 *thingmodelV1.FeatureOverrideSpec，
// 与 FeatureSpecField 同义：让 Ent 的 field.JSON 通过 protojson 编解码，
// 而非 encoding/json。
//
// FeatureOverrideSpecField wraps *thingmodelV1.FeatureOverrideSpec, mirroring
// FeatureSpecField — Ent's field.JSON uses encoding/json by default, which
// fails on protobuf-only features (wkt wrappers, etc.). protojson is the
// project's house-rule serializer for any proto message held in a JSON column.
//
// 背景 / Background:
//
//	即便 FeatureOverrideSpec 不含 oneof，它仍含 google.protobuf.BoolValue 等
//	wrapper 类型，encoding/json 不会按 protobuf 期望的形态（{"value":true}）
//	往返。为保持与 FeatureSpecField 一致的规范，统一走 protojson。
//	详见 backend/CLAUDE.md「Step 13: ent + proto message JSON 字段规范」。
type FeatureOverrideSpecField struct {
	*thingmodelV1.FeatureOverrideSpec
}

// 编译期断言实现了 json 接口 / Compile-time interface checks.
var (
	_ json.Marshaler   = (*FeatureOverrideSpecField)(nil)
	_ json.Unmarshaler = (*FeatureOverrideSpecField)(nil)
)

// MarshalJSON 将内部 FeatureOverrideSpec 通过 protojson 编码。
// MarshalJSON encodes inner FeatureOverrideSpec via protojson.
func (f *FeatureOverrideSpecField) MarshalJSON() ([]byte, error) {
	if f == nil || f.FeatureOverrideSpec == nil {
		return []byte("null"), nil
	}
	return protojson.MarshalOptions{
		UseProtoNames:   false,
		EmitUnpopulated: false,
	}.Marshal(f.FeatureOverrideSpec)
}

// UnmarshalJSON 从 JSON 通过 protojson 解码到内部 FeatureOverrideSpec。
// UnmarshalJSON decodes JSON into inner FeatureOverrideSpec via protojson.
func (f *FeatureOverrideSpecField) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		f.FeatureOverrideSpec = nil
		return nil
	}
	if f.FeatureOverrideSpec == nil {
		f.FeatureOverrideSpec = &thingmodelV1.FeatureOverrideSpec{}
	}
	return protojson.UnmarshalOptions{
		DiscardUnknown: true, // 兼容向前演化
	}.Unmarshal(data, f.FeatureOverrideSpec)
}

// WrapFeatureOverrideSpec 把 proto FeatureOverrideSpec 包装为 wrapper 类型。
// WrapFeatureOverrideSpec wraps a proto FeatureOverrideSpec for storage.
func WrapFeatureOverrideSpec(s *thingmodelV1.FeatureOverrideSpec) *FeatureOverrideSpecField {
	if s == nil {
		return nil
	}
	return &FeatureOverrideSpecField{FeatureOverrideSpec: s}
}

// UnwrapFeatureOverrideSpec 取出内部 proto FeatureOverrideSpec（处理 nil）。
// UnwrapFeatureOverrideSpec extracts the inner FeatureOverrideSpec, handling nil wrapper.
func UnwrapFeatureOverrideSpec(f *FeatureOverrideSpecField) *thingmodelV1.FeatureOverrideSpec {
	if f == nil {
		return nil
	}
	return f.FeatureOverrideSpec
}
