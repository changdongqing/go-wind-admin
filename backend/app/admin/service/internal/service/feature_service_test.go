package service

import (
	"testing"

	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"
)

// feature_service_test.go: ImportFeatures 行解析逻辑的单元测试（不依赖 DB）。
//
// 测试对象：buildFeatureFromImportRow —— 把一行导入数据组装为可落库 Feature。
// 覆盖：
//   - 合法 property/event/service/relation 行 → 成功，特化列已同步；
//   - featureType 非法 / specJson 非法 / spec 校验失败 → 明确错误；
//   - 空 specJson（部分 service/event 允许）→ 不应 panic。

// importRow 构造 ImportFeatureRow 测试夹具。
func importRow(featureType, code, identifier, specJSON string) *thingmodelV1.ImportFeatureRow {
	return &thingmodelV1.ImportFeatureRow{
		FeatureType: featureType,
		Code:        code,
		Identifier:  identifier,
		Name:        code,
		SortOrder:   1,
		SpecJson:    specJSON,
	}
}

func TestBuildFeatureFromImportRow_Property(t *testing.T) {
	// 一条合法 property（与种子 P-HVAC-0001 同构）
	spec := `{"dataType":"DOUBLE","accessMode":"R","category":"measurement","unit":{"unitCode":"celsius","unitSymbol":"℃"},"constraints":{"min":-40,"max":200}}`
	f, err := buildFeatureFromImportRow(nil, importRow("PROPERTY", "P-T-0001", "testTemp", spec), 7)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if f.GetFeatureType() != thingmodelV1.FeatureType_PROPERTY {
		t.Errorf("featureType: want PROPERTY, got %v", f.GetFeatureType())
	}
	// 特化列同步：dataType / accessMode
	if f.GetDataType() != thingmodelV1.DataType_DOUBLE {
		t.Errorf("dataType col: want DOUBLE, got %v", f.GetDataType())
	}
	if f.GetAccessMode() != thingmodelV1.AccessMode_R {
		t.Errorf("accessMode col: want R, got %v", f.GetAccessMode())
	}
	// spec oneof 分支正确
	if _, ok := f.Spec.Spec.(*thingmodelV1.FeatureSpec_Property); !ok {
		t.Errorf("spec oneof: want Property branch, got %T", f.Spec.Spec)
	}
	// 公共字段
	if f.GetCreatedBy() != 7 {
		t.Errorf("createdBy: want 7, got %v", f.GetCreatedBy())
	}
	if !f.GetIsEnabled() {
		t.Errorf("isEnabled should default true")
	}
}

// 防止 trans.Ptr 未使用告警（构建期保证）—— 已移除 trans 导入。

func TestBuildFeatureFromImportRow_Event(t *testing.T) {
	spec := `{"level":"ALERT","outputParams":[{"key":"value","dataType":"DOUBLE"},{"key":"limit","dataType":"DOUBLE"}]}`
	f, err := buildFeatureFromImportRow(nil, importRow("EVENT", "E-T-0001", "testAlarm", spec), 1)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if f.GetEventLevel() != thingmodelV1.EventLevel_ALERT {
		t.Errorf("eventLevel col: want ALERT, got %v", f.GetEventLevel())
	}
}

func TestBuildFeatureFromImportRow_Service(t *testing.T) {
	spec := `{"callMode":"ASYNC","inputParams":[{"key":"mode","dataType":"ENUM"}],"outputParams":[{"key":"success","dataType":"BOOL"}]}`
	f, err := buildFeatureFromImportRow(nil, importRow("SERVICE", "S-T-0001", "testSvc", spec), 1)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if f.GetCallMode() != thingmodelV1.CallMode_ASYNC {
		t.Errorf("callMode col: want ASYNC, got %v", f.GetCallMode())
	}
}

func TestBuildFeatureFromImportRow_Relation(t *testing.T) {
	spec := `{"relationType":"derivedFrom","cardinality":"oneToOne","directional":true,"source":{"kind":"feature","identifier":"a"},"target":{"kind":"feature","identifier":"b"}}`
	f, err := buildFeatureFromImportRow(nil, importRow("RELATION", "R-T-0001", "testRel", spec), 1)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if f.GetRelationType() != "derivedFrom" {
		t.Errorf("relationType col: want derivedFrom, got %q", f.GetRelationType())
	}
}

func TestBuildFeatureFromImportRow_InvalidFeatureType(t *testing.T) {
	_, err := buildFeatureFromImportRow(nil, importRow("NOT_A_TYPE", "X", "x", `{}`), 1)
	if err == nil {
		t.Fatal("expected error for invalid featureType, got nil")
	}
}

func TestBuildFeatureFromImportRow_BadSpecJson(t *testing.T) {
	// specJson 不是合法 JSON
	_, err := buildFeatureFromImportRow(nil, importRow("PROPERTY", "P-T-0002", "x", `{not json`), 1)
	if err == nil {
		t.Fatal("expected error for malformed specJson, got nil")
	}
}

func TestBuildFeatureFromImportRow_SpecValidationFails(t *testing.T) {
	// property 缺 accessMode（V2 校验失败）
	spec := `{"dataType":"DOUBLE"}`
	_, err := buildFeatureFromImportRow(nil, importRow("PROPERTY", "P-T-0003", "x", spec), 1)
	if err == nil {
		t.Fatal("expected validation error (missing accessMode), got nil")
	}
}

func TestBuildFeatureFromImportRow_EmptySpecAllowedForMinimalEvent(t *testing.T) {
	// event 至少需要 level；空 specJson → spec nil → 校验报 "spec is required"，
	// 这是预期行为（事件不能无 spec）。
	_, err := buildFeatureFromImportRow(nil, importRow("EVENT", "E-T-0002", "x", ""), 1)
	if err == nil {
		t.Fatal("expected error for empty spec on EVENT, got nil")
	}
}
