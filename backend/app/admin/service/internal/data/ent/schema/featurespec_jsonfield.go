package schema

import (
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"

	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"
)

// FeatureSpecField 包装 *thingmodelV1.FeatureSpec，让 Ent 的 field.JSON
// 通过 protojson 正确处理 oneof（以及 protobuf 的其它非标准 JSON 特性）。
//
// FeatureSpecField wraps *thingmodelV1.FeatureSpec so that Ent's field.JSON
// (which uses encoding/json) routes through protojson instead.
//
// 背景 / Background:
//
//	protobuf 的 oneof 在 Go 中表现为接口字段（type isFeatureSpec_Spec interface{...}），
//	且生成代码不带 json struct tag。encoding/json 无法把 JSON 反序列化回接口实例，
//	也会序列化出错乱的字段名（如 "Spec" 而非 protobuf 期望的 "spec"）。
//	直接把 *FeatureSpec 放进 field.JSON 会在 SELECT 时报：
//	  "cannot unmarshal object into Go struct field FeatureSpec.Spec of type isFeatureSpec_Spec"
//
// 解决 / Solution:
//
//	包装类型实现 json.Marshaler / json.Unmarshaler，内部委托给 protojson。
//	这是项目里所有 ent field.JSON 存 protobuf message 的共性规范，
//	见 backend/CLAUDE.md「Step 13: ent + proto message JSON 字段规范」。
type FeatureSpecField struct {
	*thingmodelV1.FeatureSpec
}

// 编译期断言实现了 json 接口 / Compile-time interface checks.
var (
	_ json.Marshaler   = (*FeatureSpecField)(nil)
	_ json.Unmarshaler = (*FeatureSpecField)(nil)
)

// MarshalJSON 将内部 FeatureSpec 通过 protojson 编码。
// MarshalJSON encodes inner FeatureSpec via protojson.
func (f *FeatureSpecField) MarshalJSON() ([]byte, error) {
	if f == nil || f.FeatureSpec == nil {
		return []byte("null"), nil
	}
	// EmitUnpopulated=false：保持简洁；UseProtoNames=false 使用 camelCase（与 proto json_name 一致）
	return protojson.MarshalOptions{
		UseProtoNames:   false,
		EmitUnpopulated: false,
	}.Marshal(f.FeatureSpec)
}

// UnmarshalJSON 从 JSON 通过 protojson 解码到内部 FeatureSpec。
// UnmarshalJSON decodes JSON into inner FeatureSpec via protojson.
func (f *FeatureSpecField) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		f.FeatureSpec = nil
		return nil
	}
	if f.FeatureSpec == nil {
		f.FeatureSpec = &thingmodelV1.FeatureSpec{}
	}
	return protojson.UnmarshalOptions{
		DiscardUnknown: true, // 兼容向前演化：忽略 DB 旧数据多余字段
	}.Unmarshal(data, f.FeatureSpec)
}

// WrapFeatureSpec 把 proto FeatureSpec 包装为 FeatureSpecField。
// WrapFeatureSpec wraps a proto FeatureSpec for storage.
func WrapFeatureSpec(s *thingmodelV1.FeatureSpec) *FeatureSpecField {
	if s == nil {
		return nil
	}
	return &FeatureSpecField{FeatureSpec: s}
}

// UnwrapFeatureSpec 取出内部 proto FeatureSpec（处理 nil）。
// UnwrapFeatureSpec extracts the inner FeatureSpec, handling nil wrapper.
func UnwrapFeatureSpec(f *FeatureSpecField) *thingmodelV1.FeatureSpec {
	if f == nil {
		return nil
	}
	return f.FeatureSpec
}
