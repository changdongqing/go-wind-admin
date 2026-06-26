package seed

import (
	"strings"
	"testing"

	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"
)

// TestUnitSeedData_Counts 验证 41+1 个分类（含拆出的 electric_charge / battery_energy）与单位总数。
// Verify category/unit counts per design doc 07 §1.
func TestUnitSeedData_Counts(t *testing.T) {
	wantCats := 42
	if got := len(UnitSeedData); got != wantCats {
		t.Fatalf("expected %d categories, got %d", wantCats, got)
	}

	totalUnits := 0
	for _, c := range UnitSeedData {
		totalUnits += len(c.Units)
	}
	// 设计文档 §1 合计约 225；浮动 ±5 内即可
	if totalUnits < 220 || totalUnits > 240 {
		t.Fatalf("total units %d out of expected range [220,240]", totalUnits)
	}
	t.Logf("seed totals: %d categories, %d units", wantCats, totalUnits)
}

// TestUnitSeedData_UniqueCodes 验证分类 code 与每分类内单位 code 的唯一性。
// Verify category-code uniqueness and per-category unit-code uniqueness.
func TestUnitSeedData_UniqueCodes(t *testing.T) {
	seenCat := map[string]struct{}{}
	for _, c := range UnitSeedData {
		if c.Code == "" {
			t.Fatalf("category code empty: %+v", c)
		}
		if _, dup := seenCat[c.Code]; dup {
			t.Fatalf("duplicate category code: %s", c.Code)
		}
		seenCat[c.Code] = struct{}{}

		seenUnit := map[string]struct{}{}
		for _, u := range c.Units {
			if u.Code == "" {
				t.Fatalf("[%s] empty unit code", c.Code)
			}
			if _, dup := seenUnit[u.Code]; dup {
				t.Fatalf("[%s] duplicate unit code: %s", c.Code, u.Code)
			}
			seenUnit[u.Code] = struct{}{}
		}
	}
}

// TestUnitSeedData_ExactlyOneBasePerCategory 每个分类有且仅有一个 is_base=true 的单位。
// Each category MUST have exactly one is_base=true unit (matches partial unique index).
func TestUnitSeedData_ExactlyOneBasePerCategory(t *testing.T) {
	for _, c := range UnitSeedData {
		count := 0
		for _, u := range c.Units {
			if u.IsBase {
				count++
			}
		}
		if count != 1 {
			t.Fatalf("category %s: expected exactly 1 base unit, got %d", c.Code, count)
		}
	}
}

// TestUnitSeedData_BaseFactorValid 基准单位必须 factor=1, offset=0（与 validateUnit 一致）。
// Base unit must have factor=1 and offset=0 (matches Service.validateUnit rule).
func TestUnitSeedData_BaseFactorValid(t *testing.T) {
	for _, c := range UnitSeedData {
		for _, u := range c.Units {
			if !u.IsBase {
				continue
			}
			if u.Factor != 1 {
				t.Fatalf("[%s/%s] base unit factor must be 1, got %v", c.Code, u.Code, u.Factor)
			}
			if u.Offset != 0 {
				t.Fatalf("[%s/%s] base unit offset must be 0, got %v", c.Code, u.Code, u.Offset)
			}
		}
	}
}

// TestUnitSeedData_LinearOffsetMustBeZero LINEAR 单位必须 offset==0。
// LINEAR units must have offset=0.
func TestUnitSeedData_LinearOffsetMustBeZero(t *testing.T) {
	for _, c := range UnitSeedData {
		for _, u := range c.Units {
			if u.ConversionType == thingmodelV1.ConversionType_LINEAR && u.Offset != 0 {
				t.Fatalf("[%s/%s] LINEAR unit offset must be 0, got %v", c.Code, u.Code, u.Offset)
			}
		}
	}
}

// TestUnitSeedData_NonZeroFactorForConvertibles LINEAR/AFFINE 单位 factor 必须非零。
// LINEAR/AFFINE units MUST have non-zero factor (else convert engine divides by zero).
func TestUnitSeedData_NonZeroFactorForConvertibles(t *testing.T) {
	for _, c := range UnitSeedData {
		for _, u := range c.Units {
			switch u.ConversionType {
			case thingmodelV1.ConversionType_LINEAR, thingmodelV1.ConversionType_AFFINE:
				if u.Factor == 0 {
					t.Fatalf("[%s/%s] convertible unit factor cannot be 0", c.Code, u.Code)
				}
			}
		}
	}
}

// TestUnitSeedData_RequiredFields 必填字段（code/symbol/name）非空。
// Required fields (code/symbol/name) must be non-empty.
func TestUnitSeedData_RequiredFields(t *testing.T) {
	for _, c := range UnitSeedData {
		if strings.TrimSpace(c.Name) == "" || strings.TrimSpace(c.BaseUnitSymbol) == "" {
			t.Fatalf("category %s: name/base_unit_symbol required", c.Code)
		}
		for _, u := range c.Units {
			if strings.TrimSpace(u.Symbol) == "" || strings.TrimSpace(u.Name) == "" {
				t.Fatalf("[%s/%s] symbol/name required", c.Code, u.Code)
			}
		}
	}
}

// TestUnitSeedData_ConversionTypeMapsToEnt 校验 protoConversionTypeToEnt 全部命中。
// Verify that protoConversionTypeToEnt maps every seed value correctly.
func TestUnitSeedData_ConversionTypeMapsToEnt(t *testing.T) {
	for _, c := range UnitSeedData {
		for _, u := range c.Units {
			if _, ok := protoConversionTypeToEnt(u.ConversionType); !ok {
				t.Fatalf("[%s/%s] conversion type %v cannot be mapped to ent enum", c.Code, u.Code, u.ConversionType)
			}
		}
	}
}

// TestUnitSeedData_KeyCategoriesPresent 关键拆分后的分类必须存在（治理点：electric_charge / battery_energy 拆自原"电池"）。
// Required categories per design §5 governance points.
func TestUnitSeedData_KeyCategoriesPresent(t *testing.T) {
	required := []string{
		"temperature", "pressure", "length", "area", "volume", "mass", "time", "speed",
		"voltage", "current", "power", "energy", "frequency", "humidity",
		"electric_charge", "battery_energy", // 拆自原"电池"
		"concentration", "dimensionless", "count",
	}
	codes := map[string]bool{}
	for _, c := range UnitSeedData {
		codes[c.Code] = true
	}
	for _, code := range required {
		if !codes[code] {
			t.Fatalf("required category %q missing from seed", code)
		}
	}
}
