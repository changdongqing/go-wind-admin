package seed

import (
	"fmt"
	"sort"
	"testing"

	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"
)

// feature_seed_test.go: 特征种子数据的回归测试。
//
// 历史背景 / History:
//   首期上线时"种子没初始化进去"，根因是 4 条 identifier 跨特征类型重复：
//     - setTemperature: property P-HVAC-0007 vs service S-HVAC-0001
//     - setFanSpeed:    property P-HVAC-0028 vs service S-HVAC-0006
//     - setWaterTemp:   property P-WP-0009   vs service S-WP-0002
//     - pipePressure:   property P-WP-0003   vs property P-FP-0001
//   schema 在 (tenant_id, identifier) 上有唯一索引 uix_thingmodel_feature_tenant_identifier，
//   但种子 upsert 只在 (tenant_id, code) 上 ON CONFLICT，PostgreSQL 不会用 identifier 索引兜底，
//   于是这 4 条触发唯一约束硬错误被跳过。修复方式：重命名 4 个冲突 identifier。
//   本文件锁定该修复，防止回退。

// TestFeatureSeedData_NoDuplicateIdentifiers 确保 (tenant_id, identifier) 唯一性不被破坏。
// Ensure no duplicate identifiers across feature types (regression for the first-phase seeding bug).
func TestFeatureSeedData_NoDuplicateIdentifiers(t *testing.T) {
	all := AllFeatureSeeds()
	seen := make(map[string][]string, len(all)) // identifier -> codes
	for _, f := range all {
		seen[f.Identifier] = append(seen[f.Identifier], f.Code)
	}

	type dup struct {
		identifier string
		codes      []string
	}
	var dups []dup
	for id, codes := range seen {
		if len(codes) > 1 {
			dups = append(dups, dup{id, codes})
		}
	}
	sort.Slice(dups, func(i, j int) bool { return dups[i].identifier < dups[j].identifier })

	if len(dups) > 0 {
		for _, d := range dups {
			t.Errorf("duplicate identifier %q used by codes: %v", d.identifier, d.codes)
		}
		t.Fatalf("found %d duplicate identifiers; rename one of each pair to restore seed seeding", len(dups))
	}
	t.Logf("OK: %d seeds, no duplicate identifiers", len(all))
}

// TestFeatureSeedData_NoDuplicateCodes 确保清单 code 唯一（upsert 按 code，重复会互相覆盖）。
func TestFeatureSeedData_NoDuplicateCodes(t *testing.T) {
	all := AllFeatureSeeds()
	seen := make(map[string]int, len(all))
	for _, f := range all {
		seen[f.Code]++
	}
	for code, n := range seen {
		if n > 1 {
			t.Errorf("duplicate code %q appears %d times", code, n)
		}
	}
	t.Logf("OK: %d seeds, no duplicate codes", len(all))
}

// TestFeatureSeedData_RenamedIdentifiersDoNotCollide 锁定 4 个改名后的 code→identifier 映射不变。
// 唯一性由 TestFeatureSeedData_NoDuplicateIdentifiers 覆盖；本测试防止有人把 identifier 改回旧名导致回退。
// Regression: the 4 renamed identifiers must keep their new values.
func TestFeatureSeedData_RenamedIdentifiersDoNotCollide(t *testing.T) {
	all := AllFeatureSeeds()
	allByID := make(map[string]string, len(all)) // code -> identifier
	for _, f := range all {
		allByID[f.Code] = f.Identifier
	}
	// 4 个历史冲突点修复后的期望 identifier；若有人回退修复，本断言会先失败并给出明确提示。
	renamed := []struct {
		code, identifier string
	}{
		{"S-HVAC-0001", "setTemperatureSvc"}, // 旧名 setTemperature 与 P-HVAC-0007 冲突
		{"S-HVAC-0006", "setFanSpeedSvc"},    // 旧名 setFanSpeed 与 P-HVAC-0028 冲突
		{"S-WP-0002", "setWaterTempSvc"},     // 旧名 setWaterTemp 与 P-WP-0009 冲突
		{"P-FP-0001", "firePipePressure"},    // 旧名 pipePressure 与 P-WP-0003 冲突
	}
	for _, r := range renamed {
		got, ok := allByID[r.code]
		if !ok {
			t.Errorf("expected seed code %s to exist", r.code)
			continue
		}
		if got != r.identifier {
			t.Errorf("code %s: expected renamed identifier %q, got %q (did someone revert the dedup fix?)",
				r.code, r.identifier, got)
		}
	}
}

// TestFeatureSeedData_AllSpecsBuildValidProto 确保每条种子的 spec 都能构造出合法 proto，
// 且特化列映射（dataType/accessMode/level/callMode）无未知枚举值——这是 upsert 成功的前置条件。
func TestFeatureSeedData_AllSpecsBuildValidProto(t *testing.T) {
	all := AllFeatureSeeds()
	var problems []string
	for _, f := range all {
		proto := buildFeatureSpecProto(f)
		if proto == nil {
			problems = append(problems, fmt.Sprintf("%s (%s): buildFeatureSpecProto returned nil", f.Code, f.Identifier))
			continue
		}
		// 特化列枚举值必须可解析（syncSpecCols / upsertSpecCols 依赖）
		switch f.FeatureType {
		case thingmodelV1.FeatureType_PROPERTY:
			if dt, ok := f.Spec["dataType"].(string); ok {
				if protoDataType(dt) == thingmodelV1.DataType_DATA_TYPE_UNSPECIFIED {
					problems = append(problems, fmt.Sprintf("%s (%s): unknown property dataType %q", f.Code, f.Identifier, dt))
				}
			} else {
				problems = append(problems, fmt.Sprintf("%s (%s): property missing dataType", f.Code, f.Identifier))
			}
			if am, ok := f.Spec["accessMode"].(string); ok {
				if protoAccessMode(am) == thingmodelV1.AccessMode_ACCESS_MODE_UNSPECIFIED {
					problems = append(problems, fmt.Sprintf("%s (%s): unknown property accessMode %q", f.Code, f.Identifier, am))
				}
			}
		case thingmodelV1.FeatureType_EVENT:
			if lv, ok := f.Spec["level"].(string); ok {
				if protoEventLevel(lv) == thingmodelV1.EventLevel_EVENT_LEVEL_UNSPECIFIED {
					problems = append(problems, fmt.Sprintf("%s (%s): unknown event level %q", f.Code, f.Identifier, lv))
				}
			}
		case thingmodelV1.FeatureType_SERVICE:
			if cm, ok := f.Spec["callMode"].(string); ok {
				if protoCallMode(cm) == thingmodelV1.CallMode_CALL_MODE_UNSPECIFIED {
					problems = append(problems, fmt.Sprintf("%s (%s): unknown service callMode %q", f.Code, f.Identifier, cm))
				}
			}
		}
	}
	if len(problems) > 0 {
		for _, p := range problems {
			t.Error(p)
		}
		t.Fatalf("total problems: %d out of %d seeds", len(problems), len(all))
	}
	t.Logf("OK: all %d seeds build valid proto and pass validation", len(all))
}
