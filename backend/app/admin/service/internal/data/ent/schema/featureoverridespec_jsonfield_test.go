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
