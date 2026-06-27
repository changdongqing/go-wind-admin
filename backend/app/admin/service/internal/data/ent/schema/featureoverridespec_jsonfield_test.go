package schema

import (
	"encoding/json"
	"strings"
	"testing"

	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// TestFeatureOverrideSpecField_RoundTrip 验证 protojson wrapper 能正确编解码 6 个白名单字段。
// 重点覆盖：
//   - 嵌套 ValueConstraints / UnitRef 等 message 字段
//   - google.protobuf.BoolValue（wrapperspb，encoding/json 不能 round-trip）
//   - displayName / description / defaultValue 等基础标量
func TestFeatureOverrideSpecField_RoundTrip(t *testing.T) {
	t.Run("Constraints + Unit + scalars", func(t *testing.T) {
		min := 5.0
		max := 12.0
		step := 0.1
		unitID := uint32(12)
		unitCode := "celsius"
		unitSym := "℃"
		src := &thingmodelV1.FeatureOverrideSpec{
			Constraints: &thingmodelV1.ValueConstraints{Min: &min, Max: &max, Step: &step},
			Unit:        &thingmodelV1.UnitRef{UnitId: &unitID, UnitCode: &unitCode, UnitSymbol: &unitSym},
			DefaultValue: ptrStr("7.0"),
			DisplayName:  ptrStr("出水温度"),
			Description:  ptrStr("冷冻水侧"),
		}
		w := WrapFeatureOverrideSpec(src)

		data, err := json.Marshal(w)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if !strings.Contains(string(data), `"displayName":"出水温度"`) {
			t.Fatalf("displayName not encoded as protojson camelCase: %s", string(data))
		}

		var out FeatureOverrideSpecField
		if err := json.Unmarshal(data, &out); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		got := UnwrapFeatureOverrideSpec(&out)
		if got.GetDisplayName() != "出水温度" {
			t.Fatalf("displayName lost: %q", got.GetDisplayName())
		}
		if got.GetConstraints().GetMax() != 12.0 {
			t.Fatalf("constraints.max lost: %v", got.GetConstraints().GetMax())
		}
		if got.GetUnit().GetUnitId() != 12 {
			t.Fatalf("unit.unitId lost: %v", got.GetUnit().GetUnitId())
		}
	})

	t.Run("BoolValue wrapper round-trip", func(t *testing.T) {
		src := &thingmodelV1.FeatureOverrideSpec{
			Required: wrapperspb.Bool(true),
		}
		data, err := json.Marshal(WrapFeatureOverrideSpec(src))
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		// protojson 把 wrapperspb.BoolValue 编为 true（不是 {"value":true}）
		if !strings.Contains(string(data), `"required":true`) {
			t.Fatalf("BoolValue not encoded as bare true: %s", string(data))
		}

		var out FeatureOverrideSpecField
		if err := json.Unmarshal(data, &out); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		got := UnwrapFeatureOverrideSpec(&out)
		if got.GetRequired() == nil || got.GetRequired().GetValue() != true {
			t.Fatalf("required not round-tripped: %+v", got.GetRequired())
		}
	})

	t.Run("Nil safety", func(t *testing.T) {
		if WrapFeatureOverrideSpec(nil) != nil {
			t.Fatalf("WrapFeatureOverrideSpec(nil) should return nil")
		}
		if UnwrapFeatureOverrideSpec(nil) != nil {
			t.Fatalf("UnwrapFeatureOverrideSpec(nil) should return nil")
		}

		var w *FeatureOverrideSpecField
		data, err := json.Marshal(w)
		if err != nil {
			t.Fatalf("marshal nil: %v", err)
		}
		if string(data) != "null" {
			t.Fatalf("marshal nil: want null got %s", string(data))
		}

		var out FeatureOverrideSpecField
		if err := json.Unmarshal([]byte("null"), &out); err != nil {
			t.Fatalf("unmarshal null: %v", err)
		}
		if out.FeatureOverrideSpec != nil {
			t.Fatalf("unmarshal null should leave inner nil")
		}
	})
}

// TestFeatureOverrideSpecField_DriverValuerScanner 验证 SQL 驱动直接序列化路径。
// 此路径是 Ent 在 Upsert.Set 中实际使用的路径——绕过 sqlgraph 的 json.Marshal，
// 把 wrapper 结构体直接传给 SQL 驱动；没有 Valuer 时 pgx 报 "unsupported type"。
func TestFeatureOverrideSpecField_DriverValuerScanner(t *testing.T) {
	t.Run("Value returns JSON bytes via protojson", func(t *testing.T) {
		min := 5.0
		max := 12.0
		src := &thingmodelV1.FeatureOverrideSpec{
			Constraints: &thingmodelV1.ValueConstraints{Min: &min, Max: &max},
			DisplayName: ptrStr("test"),
		}
		w := WrapFeatureOverrideSpec(src)

		v, err := w.Value()
		if err != nil {
			t.Fatalf("Value: %v", err)
		}
		b, ok := v.([]byte)
		if !ok {
			t.Fatalf("Value should return []byte, got %T", v)
		}
		s := string(b)
		if !strings.Contains(s, `"displayName":"test"`) || !strings.Contains(s, `"max":12`) {
			t.Fatalf("Value JSON missing fields: %s", s)
		}
	})

	t.Run("Value of nil returns nil", func(t *testing.T) {
		var w *FeatureOverrideSpecField
		v, err := w.Value()
		if err != nil {
			t.Fatalf("Value(nil wrapper): %v", err)
		}
		if v != nil {
			t.Fatalf("Value(nil) should return nil, got %v", v)
		}

		w2 := &FeatureOverrideSpecField{}
		v2, err := w2.Value()
		if err != nil {
			t.Fatalf("Value(empty wrapper): %v", err)
		}
		if v2 != nil {
			t.Fatalf("Value(empty) should return nil, got %v", v2)
		}
	})

	t.Run("Scan from []byte round-trips", func(t *testing.T) {
		jsonData := []byte(`{"displayName":"出水温度","constraints":{"min":5,"max":12}}`)
		w := &FeatureOverrideSpecField{}
		if err := w.Scan(jsonData); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		got := UnwrapFeatureOverrideSpec(w)
		if got.GetDisplayName() != "出水温度" {
			t.Fatalf("Scan displayName lost: %q", got.GetDisplayName())
		}
		if got.GetConstraints().GetMax() != 12 {
			t.Fatalf("Scan constraints.max lost: %v", got.GetConstraints().GetMax())
		}
	})

	t.Run("Scan from string round-trips", func(t *testing.T) {
		w := &FeatureOverrideSpecField{}
		if err := w.Scan(`{"displayName":"abc"}`); err != nil {
			t.Fatalf("Scan(string): %v", err)
		}
		if UnwrapFeatureOverrideSpec(w).GetDisplayName() != "abc" {
			t.Fatalf("Scan(string) lost data")
		}
	})

	t.Run("Scan nil clears wrapper", func(t *testing.T) {
		w := &FeatureOverrideSpecField{FeatureOverrideSpec: &thingmodelV1.FeatureOverrideSpec{DisplayName: ptrStr("x")}}
		if err := w.Scan(nil); err != nil {
			t.Fatalf("Scan(nil): %v", err)
		}
		if w.FeatureOverrideSpec != nil {
			t.Fatalf("Scan(nil) should clear inner, got %+v", w.FeatureOverrideSpec)
		}
	})

	t.Run("Scan unsupported type errors", func(t *testing.T) {
		w := &FeatureOverrideSpecField{}
		if err := w.Scan(12345); err == nil {
			t.Fatalf("Scan(int) should error")
		}
	})

	t.Run("Round-trip via Value then Scan", func(t *testing.T) {
		src := &thingmodelV1.FeatureOverrideSpec{
			DisplayName: ptrStr("hello"),
			Description: ptrStr("world"),
		}
		w := WrapFeatureOverrideSpec(src)
		v, err := w.Value()
		if err != nil {
			t.Fatalf("Value: %v", err)
		}
		w2 := &FeatureOverrideSpecField{}
		if err := w2.Scan(v); err != nil {
			t.Fatalf("Scan: %v", err)
		}
		got := UnwrapFeatureOverrideSpec(w2)
		if got.GetDisplayName() != "hello" || got.GetDescription() != "world" {
			t.Fatalf("round-trip lost data: %+v", got)
		}
	})
}
