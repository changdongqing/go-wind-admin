package service

import (
	"testing"

	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"
)

// feature_service_test.go: ImportFeatures 行解析逻辑的单元测试（不依赖 DB）。
//
// CR-001（2026-06-29）：thing_features 不再含 spec / 5 个特化列，本测试相应瘦身——
// 仅验证 buildFeatureFromImportRow 把骨架字段正确组装为 Feature DTO。
//
// 详见 docs/thingmodel/sheji/修改记录/CR-001-结构化约束下沉到模型层.md。

// importRow 构造 ImportFeatureRow 测试夹具（CR-001 后仅骨架字段）。
func importRow(featureType, code, identifier string) *thingmodelV1.ImportFeatureRow {
	return &thingmodelV1.ImportFeatureRow{
		FeatureType: featureType,
		Code:        code,
		Identifier:  identifier,
		Name:        code,
		SortOrder:   1,
	}
}

func TestBuildFeatureFromImportRow_Property(t *testing.T) {
	row := importRow("PROPERTY", "P-T-0001", "testTemp")
	row.ApplicableScope = "冷机"
	row.SemanticTag = "temperature"

	f, err := buildFeatureFromImportRow(nil, row, 7)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if f.GetFeatureType() != thingmodelV1.FeatureType_PROPERTY {
		t.Errorf("featureType: want PROPERTY, got %v", f.GetFeatureType())
	}
	if f.GetCode() != "P-T-0001" {
		t.Errorf("code: want P-T-0001, got %s", f.GetCode())
	}
	if f.GetIdentifier() != "testTemp" {
		t.Errorf("identifier: want testTemp, got %s", f.GetIdentifier())
	}
	if f.GetApplicableScope() != "冷机" {
		t.Errorf("applicableScope: want 冷机, got %s", f.GetApplicableScope())
	}
	if f.GetSemanticTag() != "temperature" {
		t.Errorf("semanticTag: want temperature, got %s", f.GetSemanticTag())
	}
	if f.GetCreatedBy() != 7 {
		t.Errorf("createdBy: want 7, got %d", f.GetCreatedBy())
	}
}

func TestBuildFeatureFromImportRow_Event(t *testing.T) {
	f, err := buildFeatureFromImportRow(nil, importRow("EVENT", "E-T-0001", "testEvent"), 7)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if f.GetFeatureType() != thingmodelV1.FeatureType_EVENT {
		t.Errorf("featureType: want EVENT, got %v", f.GetFeatureType())
	}
}

func TestBuildFeatureFromImportRow_Service(t *testing.T) {
	f, err := buildFeatureFromImportRow(nil, importRow("SERVICE", "S-T-0001", "testService"), 7)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if f.GetFeatureType() != thingmodelV1.FeatureType_SERVICE {
		t.Errorf("featureType: want SERVICE, got %v", f.GetFeatureType())
	}
}

func TestBuildFeatureFromImportRow_Relation(t *testing.T) {
	f, err := buildFeatureFromImportRow(nil, importRow("RELATION", "R-T-0001", "testRelation"), 7)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if f.GetFeatureType() != thingmodelV1.FeatureType_RELATION {
		t.Errorf("featureType: want RELATION, got %v", f.GetFeatureType())
	}
}

func TestBuildFeatureFromImportRow_InvalidFeatureType(t *testing.T) {
	_, err := buildFeatureFromImportRow(nil, importRow("XXX", "X-0001", "x"), 7)
	if err == nil {
		t.Fatal("expected error for invalid featureType, got nil")
	}
}
