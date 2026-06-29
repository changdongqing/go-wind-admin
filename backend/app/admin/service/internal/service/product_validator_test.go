package service

import (
	"testing"

	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"
)

// product_validator_test.go: 模型管理共享校验器的单元测试。
//
// CR-001（2026-06-29）：删除 effectiveSpec / validateOverrideSpec / extractUnitID 三组测试
// （对应的功能已彻底移除——spec 现在单一存放在 CDF/PF.spec 列，不再有 snapshot/override 合并语义）。
// 保留 validateSourceRefMismatch 与 validateSpecTypeMismatch。

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
	propSpec := &thingmodelV1.FeatureSpec{
		Spec: &thingmodelV1.FeatureSpec_Property{Property: &thingmodelV1.PropertySpec{}},
	}
	if e := validateSpecTypeMismatch(thingmodelV1.FeatureType_PROPERTY, propSpec); e != "" {
		t.Fatalf("expected pass, got %q", e)
	}
	if e := validateSpecTypeMismatch(thingmodelV1.FeatureType_EVENT, propSpec); e == "" {
		t.Fatal("expected mismatch when EVENT but spec.property non-nil")
	}
	if e := validateSpecTypeMismatch(thingmodelV1.FeatureType_PROPERTY, nil); e != "" {
		t.Fatalf("nil spec should pass, got %q", e)
	}
}
