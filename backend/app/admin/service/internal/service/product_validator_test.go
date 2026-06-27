package service

import (
	"testing"

	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// product_validator_test.go: 覆写白名单与 effective_spec 合并的单元测试。

func TestValidateOverrideSpec_NonPropertyRejectsConstraints(t *testing.T) {
	eventSnap := &thingmodelV1.FeatureSpec{
		Spec: &thingmodelV1.FeatureSpec_Event{Event: &thingmodelV1.EventSpec{}},
	}
	min := 0.0
	max := 10.0
	over := &thingmodelV1.FeatureOverrideSpec{
		Constraints: &thingmodelV1.ValueConstraints{Min: &min, Max: &max},
	}
	errs := validateOverrideSpec(eventSnap, over)
	if len(errs) == 0 {
		t.Fatal("expected error when non-property overrides constraints")
	}
}

func TestValidateOverrideSpec_NonPropertyRejectsUnit(t *testing.T) {
	svcSnap := &thingmodelV1.FeatureSpec{
		Spec: &thingmodelV1.FeatureSpec_Service{Service: &thingmodelV1.ServiceSpec{}},
	}
	uid := uint32(1)
	over := &thingmodelV1.FeatureOverrideSpec{
		Unit: &thingmodelV1.UnitRef{UnitId: &uid},
	}
	errs := validateOverrideSpec(svcSnap, over)
	if len(errs) == 0 {
		t.Fatal("expected error when non-property overrides unit")
	}
}

func TestValidateOverrideSpec_MinGreaterThanMax(t *testing.T) {
	propSnap := &thingmodelV1.FeatureSpec{
		Spec: &thingmodelV1.FeatureSpec_Property{Property: &thingmodelV1.PropertySpec{}},
	}
	min := 100.0
	max := 50.0
	over := &thingmodelV1.FeatureOverrideSpec{
		Constraints: &thingmodelV1.ValueConstraints{Min: &min, Max: &max},
	}
	errs := validateOverrideSpec(propSnap, over)
	if len(errs) == 0 {
		t.Fatal("expected error when min > max")
	}
}

func TestValidateOverrideSpec_HappyPath(t *testing.T) {
	dt := thingmodelV1.DataType_DOUBLE
	propSnap := &thingmodelV1.FeatureSpec{
		Spec: &thingmodelV1.FeatureSpec_Property{Property: &thingmodelV1.PropertySpec{DataType: &dt}},
	}
	min := 5.0
	max := 12.0
	over := &thingmodelV1.FeatureOverrideSpec{
		Constraints: &thingmodelV1.ValueConstraints{Min: &min, Max: &max},
	}
	errs := validateOverrideSpec(propSnap, over)
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
}

func TestValidateOverrideSpec_NilOverride(t *testing.T) {
	propSnap := &thingmodelV1.FeatureSpec{
		Spec: &thingmodelV1.FeatureSpec_Property{Property: &thingmodelV1.PropertySpec{}},
	}
	if errs := validateOverrideSpec(propSnap, nil); errs != nil {
		t.Fatalf("nil override should pass, got %v", errs)
	}
}

func TestEffectiveSpec_NoOverride(t *testing.T) {
	dt := thingmodelV1.DataType_DOUBLE
	snap := &thingmodelV1.FeatureSpec{
		Spec: &thingmodelV1.FeatureSpec_Property{Property: &thingmodelV1.PropertySpec{DataType: &dt}},
	}
	got := effectiveSpec(snap, nil)
	if got != snap {
		t.Fatal("nil override should return snapshot as-is")
	}
}

func TestEffectiveSpec_MergesConstraints(t *testing.T) {
	dt := thingmodelV1.DataType_DOUBLE
	snapMin := -20.0
	snapMax := 60.0
	snap := &thingmodelV1.FeatureSpec{
		Spec: &thingmodelV1.FeatureSpec_Property{Property: &thingmodelV1.PropertySpec{
			DataType:    &dt,
			Constraints: &thingmodelV1.ValueConstraints{Min: &snapMin, Max: &snapMax},
		}},
	}
	over := &thingmodelV1.FeatureOverrideSpec{
		Constraints: &thingmodelV1.ValueConstraints{Min: ptrFloat(5), Max: ptrFloat(12)},
	}
	got := effectiveSpec(snap, over)
	if got.GetProperty().GetConstraints().GetMax() != 12 {
		t.Fatalf("expected max=12 after merge, got %v", got.GetProperty().GetConstraints().GetMax())
	}
	// snapshot 不应被修改
	if snap.GetProperty().GetConstraints().GetMax() != 60 {
		t.Fatalf("snapshot mutated! expected max=60, got %v", snap.GetProperty().GetConstraints().GetMax())
	}
}

func TestEffectiveSpec_MergesUnit(t *testing.T) {
	uidNew := uint32(20)
	snap := &thingmodelV1.FeatureSpec{
		Spec: &thingmodelV1.FeatureSpec_Property{Property: &thingmodelV1.PropertySpec{
			Unit: &thingmodelV1.UnitRef{UnitId: ptrUint32(12)},
		}},
	}
	over := &thingmodelV1.FeatureOverrideSpec{
		Unit: &thingmodelV1.UnitRef{UnitId: &uidNew},
	}
	got := effectiveSpec(snap, over)
	if got.GetProperty().GetUnit().GetUnitId() != 20 {
		t.Fatalf("expected unit_id=20, got %v", got.GetProperty().GetUnit().GetUnitId())
	}
}

func TestValidateSourceRefMismatch(t *testing.T) {
	cases := []struct {
		name    string
		source  thingmodelV1.ProductFeatureSource
		refID   uint32
		wantErr bool
	}{
		{"LOCAL with refID=0 ok", thingmodelV1.ProductFeatureSource_LOCAL, 0, false},
		{"LOCAL with refID=5 fails", thingmodelV1.ProductFeatureSource_LOCAL, 5, true},
		{"DEFAULT with refID=5 ok", thingmodelV1.ProductFeatureSource_DEFAULT, 5, false},
		{"DEFAULT with refID=0 fails", thingmodelV1.ProductFeatureSource_DEFAULT, 0, true},
		{"GLOBAL with refID=3 ok", thingmodelV1.ProductFeatureSource_GLOBAL, 3, false},
		{"UNSPECIFIED fails", thingmodelV1.ProductFeatureSource_PRODUCT_FEATURE_SOURCE_UNSPECIFIED, 0, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := validateSourceRefMismatch(c.source, c.refID)
			if (got != "") != c.wantErr {
				t.Fatalf("want err=%v, got err=%q", c.wantErr, got)
			}
		})
	}
}

func TestValidateSpecTypeMismatch(t *testing.T) {
	propSnap := &thingmodelV1.FeatureSpec{
		Spec: &thingmodelV1.FeatureSpec_Property{Property: &thingmodelV1.PropertySpec{}},
	}
	if e := validateSpecTypeMismatch(thingmodelV1.FeatureType_PROPERTY, propSnap); e != "" {
		t.Fatalf("expected pass, got %q", e)
	}
	if e := validateSpecTypeMismatch(thingmodelV1.FeatureType_EVENT, propSnap); e == "" {
		t.Fatal("expected mismatch when EVENT but spec.property non-nil")
	}
	if e := validateSpecTypeMismatch(thingmodelV1.FeatureType_PROPERTY, nil); e != "" {
		t.Fatalf("nil snapshot should pass, got %q", e)
	}
}

func TestExtractUnitID_OverrideWins(t *testing.T) {
	snap := &thingmodelV1.FeatureSpec{
		Spec: &thingmodelV1.FeatureSpec_Property{Property: &thingmodelV1.PropertySpec{
			Unit: &thingmodelV1.UnitRef{UnitId: ptrUint32(10)},
		}},
	}
	over := &thingmodelV1.FeatureOverrideSpec{
		Unit: &thingmodelV1.UnitRef{UnitId: ptrUint32(20)},
	}
	if id := extractUnitID(snap, over); id != 20 {
		t.Fatalf("override should win, got %d", id)
	}
}

func TestExtractUnitID_FallsBackToSnapshot(t *testing.T) {
	snap := &thingmodelV1.FeatureSpec{
		Spec: &thingmodelV1.FeatureSpec_Property{Property: &thingmodelV1.PropertySpec{
			Unit: &thingmodelV1.UnitRef{UnitId: ptrUint32(10)},
		}},
	}
	if id := extractUnitID(snap, nil); id != 10 {
		t.Fatalf("expected 10 from snapshot, got %d", id)
	}
	if id := extractUnitID(nil, nil); id != 0 {
		t.Fatalf("expected 0 from nil/nil, got %d", id)
	}
}

// === helpers ===

func ptrFloat(v float64) *float64 { return &v }
func ptrUint32(v uint32) *uint32  { return &v }
var _ = wrapperspb.Bool // ensure import used in test if needed for future cases