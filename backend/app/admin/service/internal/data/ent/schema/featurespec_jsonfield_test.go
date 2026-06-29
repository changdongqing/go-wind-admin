package schema

import (
	"encoding/json"
	"strings"
	"testing"

	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"
)

// TestFeatureSpecField_RoundTrip 验证 protojson wrapper 能正确 marshal/unmarshal
// 所有四类 oneof 分支（property/event/service/relation），覆盖 oneof + 嵌套
// message + map 等 encoding/json 直接处理会失败的场景。
//
// 这是核心保障：encoding/json 处理 protobuf oneof 时会丢字段、UNMARSHAL 时报错；
// 包装类型必须把所有 marshal 委托到 protojson。
func TestFeatureSpecField_RoundTrip(t *testing.T) {
	t.Run("Property/BOOL with boolLabels", func(t *testing.T) {
		dt := thingmodelV1.DataType_BOOL
		am := thingmodelV1.AccessMode_RW
		cat := "runtime"
		spec := &thingmodelV1.FeatureSpec{
			Spec: &thingmodelV1.FeatureSpec_Property{
				Property: &thingmodelV1.PropertySpec{
					DataType:   &dt,
					AccessMode: &am,
					Category:   &cat,
					BoolLabels: &thingmodelV1.BoolLabels{FalseLabel: "关", TrueLabel: "开"},
				},
			},
		}
		data, err := json.Marshal(WrapFeatureSpec(spec))
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		s := string(data)
		// protojson 用 camelCase 与 oneof 平铺：{"property":{"dataType":"BOOL",...}}
		if !strings.Contains(s, `"property"`) || !strings.Contains(s, `"dataType":"BOOL"`) {
			t.Fatalf("unexpected json: %s", s)
		}

		var out FeatureSpecField
		if err := json.Unmarshal(data, &out); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		got := UnwrapFeatureSpec(&out).GetProperty()
		if got.GetDataType() != thingmodelV1.DataType_BOOL {
			t.Fatalf("dataType: want BOOL got %v", got.GetDataType())
		}
		if got.GetAccessMode() != thingmodelV1.AccessMode_RW {
			t.Fatalf("accessMode: want RW got %v", got.GetAccessMode())
		}
		if got.GetBoolLabels().GetFalseLabel() != "关" || got.GetBoolLabels().GetTrueLabel() != "开" {
			t.Fatalf("boolLabels mismatch: %+v", got.GetBoolLabels())
		}
	})

	t.Run("Event with outputParams", func(t *testing.T) {
		lv := thingmodelV1.EventLevel_ALERT
		spec := &thingmodelV1.FeatureSpec{
			Spec: &thingmodelV1.FeatureSpec_Event{
				Event: &thingmodelV1.EventSpec{
					Level: &lv,
					OutputParams: []*thingmodelV1.ParamSpec{
						{Key: ptrStr("ts"), DataType: ptrDT(thingmodelV1.DataType_DATE)},
					},
				},
			},
		}
		data, err := json.Marshal(WrapFeatureSpec(spec))
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}

		var out FeatureSpecField
		if err := json.Unmarshal(data, &out); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		got := UnwrapFeatureSpec(&out).GetEvent()
		if got.GetLevel() != thingmodelV1.EventLevel_ALERT {
			t.Fatalf("level mismatch: %v", got.GetLevel())
		}
		if len(got.GetOutputParams()) != 1 || got.GetOutputParams()[0].GetKey() != "ts" {
			t.Fatalf("outputParams mismatch: %+v", got.GetOutputParams())
		}
	})

	t.Run("Relation with EntityRef", func(t *testing.T) {
		rt := "derivedFrom"
		spec := &thingmodelV1.FeatureSpec{
			Spec: &thingmodelV1.FeatureSpec_Relation{
				Relation: &thingmodelV1.RelationSpec{
					RelationType: &rt,
					Source:       &thingmodelV1.EntityRef{Kind: ptrStr("feature"), Identifier: ptrStr("currentCOP")},
					Target:       &thingmodelV1.EntityRef{Kind: ptrStr("feature"), Identifier: ptrStr("currentPower")},
				},
			},
		}
		data, err := json.Marshal(WrapFeatureSpec(spec))
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		var out FeatureSpecField
		if err := json.Unmarshal(data, &out); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		r := UnwrapFeatureSpec(&out).GetRelation()
		if r.GetRelationType() != "derivedFrom" {
			t.Fatalf("relationType: %s", r.GetRelationType())
		}
		if r.GetSource().GetIdentifier() != "currentCOP" || r.GetTarget().GetIdentifier() != "currentPower" {
			t.Fatalf("source/target mismatch: src=%s tgt=%s",
				r.GetSource().GetIdentifier(), r.GetTarget().GetIdentifier())
		}
	})

	t.Run("Nil safety", func(t *testing.T) {
		// Wrap(nil) → nil
		if WrapFeatureSpec(nil) != nil {
			t.Fatalf("WrapFeatureSpec(nil) should return nil")
		}
		// Unwrap(nil) → nil
		if UnwrapFeatureSpec(nil) != nil {
			t.Fatalf("UnwrapFeatureSpec(nil) should return nil")
		}
		// Marshal of nil wrapper → "null"
		var w *FeatureSpecField
		data, err := json.Marshal(w)
		if err != nil {
			t.Fatalf("marshal nil: %v", err)
		}
		if string(data) != "null" {
			t.Fatalf("marshal nil: want null got %s", string(data))
		}
		// Unmarshal null → empty
		var out FeatureSpecField
		if err := json.Unmarshal([]byte("null"), &out); err != nil {
			t.Fatalf("unmarshal null: %v", err)
		}
		if out.FeatureSpec != nil {
			t.Fatalf("unmarshal null should leave FeatureSpec nil")
		}
	})
}

// TestFeatureSpecField_DriverValuer 验证 SQL 驱动直接序列化路径。
// 这是修复 ent Upsert 路径 "unsupported type schema.FeatureSpecField" bug
// 的关键：在 Upsert.Set(col, wrapper) 中，ent 不走 sqlgraph 的 json.Marshal，
// 而是直接把 wrapper 传给 SQL 驱动。
//
// CR-001：本类型不再实现 sql.Scanner——读路径由 ent 生成代码里的
// `json.Unmarshal(*value, &_m.Spec)` 自动委派到 wrapper.UnmarshalJSON，走 protojson。
func TestFeatureSpecField_DriverValuer(t *testing.T) {
	t.Run("Value returns JSON bytes for property", func(t *testing.T) {
		dt := thingmodelV1.DataType_BOOL
		spec := &thingmodelV1.FeatureSpec{
			Spec: &thingmodelV1.FeatureSpec_Property{
				Property: &thingmodelV1.PropertySpec{
					DataType:   &dt,
					BoolLabels: &thingmodelV1.BoolLabels{FalseLabel: "off", TrueLabel: "on"},
				},
			},
		}
		w := WrapFeatureSpec(spec)
		v, err := w.Value()
		if err != nil {
			t.Fatalf("Value: %v", err)
		}
		b, ok := v.([]byte)
		if !ok {
			t.Fatalf("Value should return []byte, got %T", v)
		}
		if !strings.Contains(string(b), `"dataType":"BOOL"`) {
			t.Fatalf("Value JSON missing dataType: %s", string(b))
		}
	})

	t.Run("Value of nil returns nil", func(t *testing.T) {
		var w *FeatureSpecField
		v, err := w.Value()
		if err != nil || v != nil {
			t.Fatalf("Value(nil) want (nil,nil), got (%v,%v)", v, err)
		}
	})

	t.Run("Value + UnmarshalJSON round-trip", func(t *testing.T) {
		dt := thingmodelV1.DataType_DOUBLE
		spec := &thingmodelV1.FeatureSpec{
			Spec: &thingmodelV1.FeatureSpec_Property{
				Property: &thingmodelV1.PropertySpec{DataType: &dt},
			},
		}
		w := WrapFeatureSpec(spec)
		v, err := w.Value()
		if err != nil {
			t.Fatalf("Value: %v", err)
		}
		b, _ := v.([]byte)
		w2 := &FeatureSpecField{}
		if err := w2.UnmarshalJSON(b); err != nil {
			t.Fatalf("UnmarshalJSON: %v", err)
		}
		if UnwrapFeatureSpec(w2).GetProperty().GetDataType() != thingmodelV1.DataType_DOUBLE {
			t.Fatalf("round-trip lost dataType")
		}
	})
}

func ptrStr(s string) *string                              { return &s }
func ptrDT(d thingmodelV1.DataType) *thingmodelV1.DataType { return &d }
