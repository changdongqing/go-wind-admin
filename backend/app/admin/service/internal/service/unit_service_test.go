package service

import (
	"math"
	"strings"
	"testing"

	"github.com/tx7do/go-utils/trans"

	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"
)

// ============================================================================
// 单位换算引擎测试 / Unit conversion engine tests
//
// 覆盖 docs/thingmodel/sheji/03-单位换算引擎设计.md 第 8 节"测试矩阵"中的 T1-T12：
//   T1 线性正向       T2 仿射正向         T3 跨单位同分类       T4 反向对称
//   T5 基准↔自身     T6 不同分类拒绝     T7 对数拒绝           T8 条件拒绝
//   T9 基准系数校验   T10 线性偏移校验    T11 删除约束(repo层)  T12 溢出
// ============================================================================

// ---------- 单位构造辅助 / Unit fixture helpers ----------

func newUnit(catID uint32, code, symbol string, ct thingmodelV1.ConversionType, factor, offset float64, precision int32, isBase bool) *thingmodelV1.Unit {
	return &thingmodelV1.Unit{
		CategoryId:     trans.Ptr(catID),
		Code:           trans.Ptr(code),
		Symbol:         trans.Ptr(symbol),
		Name:           trans.Ptr(code),
		ConversionType: trans.Ptr(ct),
		Factor:         trans.Ptr(factor),
		Offset:         trans.Ptr(offset),
		Precision:      trans.Ptr(precision),
		IsBase:         trans.Ptr(isBase),
	}
}

// 温度分类（catID=1）：开尔文(基准) / 摄氏度 / 华氏度
func kelvin() *thingmodelV1.Unit {
	return newUnit(1, "kelvin", "K", thingmodelV1.ConversionType_LINEAR, 1, 0, 2, true)
}
func celsius() *thingmodelV1.Unit {
	// 50℃ → 323.15K：base = 50·1 + 273.15
	return newUnit(1, "celsius", "℃", thingmodelV1.ConversionType_AFFINE, 1, 273.15, 2, false)
}
func fahrenheit() *thingmodelV1.Unit {
	// ℉→K: K = (℉+459.67)·5/9 = ℉·(5/9) + 459.67·(5/9) = ℉·0.5556 + 255.372
	return newUnit(1, "fahrenheit", "℉", thingmodelV1.ConversionType_AFFINE, 5.0/9.0, 459.67*5.0/9.0, 4, false)
}

// 长度分类（catID=2）：米(基准) / 厘米
func meter() *thingmodelV1.Unit {
	return newUnit(2, "meter", "m", thingmodelV1.ConversionType_LINEAR, 1, 0, 4, true)
}
func centimeter() *thingmodelV1.Unit {
	return newUnit(2, "centimeter", "cm", thingmodelV1.ConversionType_LINEAR, 0.01, 0, 4, false)
}

// 质量分类（catID=3）：千克(基准)
func kilogram() *thingmodelV1.Unit {
	return newUnit(3, "kilogram", "kg", thingmodelV1.ConversionType_LINEAR, 1, 0, 4, true)
}

// 声级（dB 系，不可线性换算）
func decibel() *thingmodelV1.Unit {
	return newUnit(4, "decibel", "dB", thingmodelV1.ConversionType_LOGARITHMIC, 1, 0, 2, true)
}
func dBm() *thingmodelV1.Unit {
	return newUnit(4, "dBm", "dBm", thingmodelV1.ConversionType_LOGARITHMIC, 1, 0, 2, false)
}

// 浓度（条件换算）
func ppm() *thingmodelV1.Unit {
	return newUnit(5, "ppm", "ppm", thingmodelV1.ConversionType_CONDITIONAL, 1, 0, 2, true)
}
func mgPerM3() *thingmodelV1.Unit {
	return newUnit(5, "mg_per_m3", "mg/m³", thingmodelV1.ConversionType_CONDITIONAL, 1, 0, 2, false)
}

// approxEqual 浮点近似相等 / approximate equality with tolerance
func approxEqual(a, b, eps float64) bool { return math.Abs(a-b) <= eps }

// ---------- T1: 线性正向 / Linear forward ----------

func TestConvert_T1_LinearForward_CmToM(t *testing.T) {
	resp, err := convertUnits(centimeter(), meter(), 5, -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != thingmodelV1.ConvertUnitStatus_CONVERT_OK {
		t.Fatalf("status=%v message=%s", resp.Status, resp.Message)
	}
	if !approxEqual(resp.Result, 0.05, 1e-9) {
		t.Fatalf("5 cm → m expected 0.05, got %v", resp.Result)
	}
}

// ---------- T2: 仿射正向 / Affine forward ----------

func TestConvert_T2_AffineForward_CelsiusToKelvin(t *testing.T) {
	resp, err := convertUnits(celsius(), kelvin(), 50, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != thingmodelV1.ConvertUnitStatus_CONVERT_OK {
		t.Fatalf("status=%v message=%s", resp.Status, resp.Message)
	}
	if !approxEqual(resp.Result, 323.15, 1e-6) {
		t.Fatalf("50℃ → K expected 323.15, got %v", resp.Result)
	}
}

// ---------- T3: 跨单位同分类 / Cross-unit within category ----------

func TestConvert_T3_CelsiusToFahrenheit(t *testing.T) {
	resp, err := convertUnits(celsius(), fahrenheit(), 50, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != thingmodelV1.ConvertUnitStatus_CONVERT_OK {
		t.Fatalf("status=%v message=%s", resp.Status, resp.Message)
	}
	// 50℃ = 122℉ (允许 0.001 容差应对 5/9 双精度展开)
	if !approxEqual(resp.Result, 122, 1e-3) {
		t.Fatalf("50℃ → ℉ expected 122, got %v", resp.Result)
	}
}

// ---------- T4: 反向对称性 / Reverse symmetry ----------

func TestConvert_T4_ReverseSymmetry(t *testing.T) {
	// 50℃ → ℉ → ℃，期望误差 < 1e-6（高精度，避开 round 误差）
	r1, err := convertUnits(celsius(), fahrenheit(), 50, 9)
	if err != nil || r1.Status != thingmodelV1.ConvertUnitStatus_CONVERT_OK {
		t.Fatalf("first leg failed: %+v err=%v", r1, err)
	}
	r2, err := convertUnits(fahrenheit(), celsius(), r1.Result, 9)
	if err != nil || r2.Status != thingmodelV1.ConvertUnitStatus_CONVERT_OK {
		t.Fatalf("second leg failed: %+v err=%v", r2, err)
	}
	if !approxEqual(r2.Result, 50, 1e-6) {
		t.Fatalf("round-trip ℃→℉→℃ expected 50, got %v", r2.Result)
	}
}

// ---------- T5: 基准↔自身 / Base to base ----------

func TestConvert_T5_BaseToSelf(t *testing.T) {
	resp, err := convertUnits(kelvin(), kelvin(), 100, -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != thingmodelV1.ConvertUnitStatus_CONVERT_OK {
		t.Fatalf("status=%v", resp.Status)
	}
	if !approxEqual(resp.Result, 100, 1e-9) {
		t.Fatalf("100K → K expected 100, got %v", resp.Result)
	}
}

// ---------- T6: 不同分类拒绝 / Different category rejection ----------

func TestConvert_T6_DifferentCategoryRejected(t *testing.T) {
	resp, err := convertUnits(meter(), kilogram(), 5, -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != thingmodelV1.ConvertUnitStatus_CONVERT_DIFFERENT_CATEGORY {
		t.Fatalf("expected DIFFERENT_CATEGORY, got %v (msg=%s)", resp.Status, resp.Message)
	}
}

// ---------- T7: 对数单位拒绝 / Logarithmic rejection ----------

func TestConvert_T7_LogarithmicRejected(t *testing.T) {
	resp, err := convertUnits(decibel(), dBm(), 70, -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != thingmodelV1.ConvertUnitStatus_CONVERT_NOT_CONVERTIBLE {
		t.Fatalf("expected NOT_CONVERTIBLE, got %v", resp.Status)
	}
	if !strings.Contains(resp.Message, "对数") && !strings.Contains(strings.ToLower(resp.Message), "logarithmic") {
		t.Fatalf("expected logarithmic hint in message, got: %s", resp.Message)
	}
}

// ---------- T8: 条件换算拒绝 / Conditional rejection ----------

func TestConvert_T8_ConditionalRejected(t *testing.T) {
	resp, err := convertUnits(ppm(), mgPerM3(), 100, -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != thingmodelV1.ConvertUnitStatus_CONVERT_NOT_CONVERTIBLE {
		t.Fatalf("expected NOT_CONVERTIBLE, got %v", resp.Status)
	}
	if !strings.Contains(resp.Message, "条件") && !strings.Contains(strings.ToLower(resp.Message), "conditional") {
		t.Fatalf("expected conditional hint in message, got: %s", resp.Message)
	}
}

// ---------- T9: 写入校验-基准系数 / Base factor validation ----------

func TestValidateUnit_T9_BaseFactorInvalid(t *testing.T) {
	u := newUnit(1, "bad_base", "X", thingmodelV1.ConversionType_LINEAR, 2 /* factor */, 0, 2, true /* is_base */)
	err := validateUnit(u)
	if err == nil {
		t.Fatalf("expected UNIT_BASE_FACTOR_INVALID error, got nil")
	}
	if !strings.Contains(err.Error(), "base unit") {
		t.Fatalf("expected base unit error, got: %v", err)
	}
}

// ---------- T10: 写入校验-线性偏移 / Linear offset validation ----------

func TestValidateUnit_T10_LinearOffsetMustBeZero(t *testing.T) {
	u := newUnit(1, "bad_linear", "Y", thingmodelV1.ConversionType_LINEAR, 0.5, 1 /* offset != 0 */, 2, false)
	err := validateUnit(u)
	if err == nil {
		t.Fatalf("expected UNIT_LINEAR_OFFSET_MUST_BE_ZERO, got nil")
	}
	if !strings.Contains(err.Error(), "linear unit offset") {
		t.Fatalf("expected linear offset error, got: %v", err)
	}
}

// ---------- T11: 删除约束 / Delete constraint
// 该规则依赖 repo.ReferencedIDs 查询 reference_count > 0，需 DB（本期 reference_count 恒为 0）。
// 此处仅做契约层 smoke：确认错误码存在、消息可读，避免 P1 提交时漏掉相应 proto 错误码。
func TestUnit_T11_InUseCannotDelete_ErrorCodeExists(t *testing.T) {
	err := thingmodelV1.ErrorUnitInUseCannotDelete("guard test: ids=%v", []uint32{1})
	if err == nil {
		t.Fatalf("ErrorUnitInUseCannotDelete must be available in generated code")
	}
	if !strings.Contains(err.Error(), "guard test") {
		t.Fatalf("error format failed: %v", err)
	}
}

// ---------- T12: 溢出 / Overflow ----------

func TestConvert_T12_Overflow(t *testing.T) {
	// 构造 factor=1e308 的线性单位（人为异常），与基准换算时 1e308 · 1e308 → +Inf
	huge := newUnit(2, "huge", "H", thingmodelV1.ConversionType_LINEAR, 1e308, 0, 2, false)
	resp, err := convertUnits(huge, meter(), 1e308, -1)
	if err == nil {
		t.Fatalf("expected UNIT_OVERFLOW error, got resp=%+v", resp)
	}
	if !strings.Contains(err.Error(), "overflow") {
		t.Fatalf("expected overflow message, got: %v", err)
	}
}

// ---------- 附加：写入校验-缺字段 ----------

func TestValidateUnit_MissingFields(t *testing.T) {
	cases := []struct {
		name string
		u    *thingmodelV1.Unit
	}{
		{"no_code", &thingmodelV1.Unit{Symbol: trans.Ptr("X"), Name: trans.Ptr("x"), CategoryId: trans.Ptr(uint32(1))}},
		{"no_symbol", &thingmodelV1.Unit{Code: trans.Ptr("c"), Name: trans.Ptr("x"), CategoryId: trans.Ptr(uint32(1))}},
		{"no_name", &thingmodelV1.Unit{Code: trans.Ptr("c"), Symbol: trans.Ptr("X"), CategoryId: trans.Ptr(uint32(1))}},
		{"no_category", &thingmodelV1.Unit{Code: trans.Ptr("c"), Symbol: trans.Ptr("X"), Name: trans.Ptr("x")}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := validateUnit(c.u); err == nil {
				t.Fatalf("expected validation error, got nil")
			}
		})
	}
}
