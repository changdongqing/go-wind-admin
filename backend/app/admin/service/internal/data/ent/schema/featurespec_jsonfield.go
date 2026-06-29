package schema

import (
	"database/sql/driver"
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
//
// CR-001 注意：本类型**不**实现 sql.Scanner。Ent 0.14+ 检测到 Scanner 接口后
// 会让 scanValues 返回 `new(FeatureSpecField)`，但 assignValues 仍按 []byte 走
// json.Unmarshal —— 生成代码出现 `len(*value)` 这种把 wrapper 当字节切片用的不一致。
// 因此本 wrapper 仅暴露 driver.Valuer（写路径，Upsert 必需）+ MarshalJSON/UnmarshalJSON
// （读路径，由 ent 生成代码里的 json.Unmarshal 自动路由到 UnmarshalJSON 走 protojson）。
var (
	_ json.Marshaler   = (*FeatureSpecField)(nil)
	_ json.Unmarshaler = (*FeatureSpecField)(nil)
	// driver.Valuer：必需，否则 ent 在 Upsert 路径会把 wrapper 结构体直接传给 SQL 驱动，
	// pgx 报 "unsupported type"。
	_ driver.Valuer = (*FeatureSpecField)(nil)
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

// Value 实现 database/sql/driver.Valuer，让 SQL 驱动能直接序列化本类型。
// Value implements driver.Valuer so the SQL driver can serialize the wrapper
// directly. This is the canonical path used by Ent's Upsert.Set, which does
// not route through sqlgraph's json.Marshal.
func (f *FeatureSpecField) Value() (driver.Value, error) {
	if f == nil || f.FeatureSpec == nil {
		return nil, nil
	}
	return protojson.MarshalOptions{
		UseProtoNames:   false,
		EmitUnpopulated: false,
	}.Marshal(f.FeatureSpec)
}

// CR-001 注意：本类型不实现 sql.Scanner（接口名 Scan）——读路径由 ent 生成代码里的
// `json.Unmarshal(*value, &_m.Spec)` 自动委派到 UnmarshalJSON，走 protojson。
// 若需在 service 层从 raw bytes 还原，请直接调用 (&FeatureSpecField{}).UnmarshalJSON(b)。

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
