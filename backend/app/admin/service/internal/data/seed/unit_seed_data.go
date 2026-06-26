// Package seed 提供物模型基础数据的种子（幂等 upsert）。
// Package seed provides idempotent seed data for the thing-model module.
//
// 设计依据 / Design ref: docs/thingmodel/sheji/07-种子数据与实施计划.md §2.1
// 数据来源 / Data source: docs/thingmodel/物模型属性单位清单.md
//
// 幂等策略 / Idempotency strategy:
//   - 按 (tenant_id=0, code) upsert；重复执行不报错也不复制。
//   - 仅维护 tenant_id=0 的系统预置数据；租户自建（tenant_id>0）永不被覆盖。
package seed

import thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"

// SeedCategory 单个物理量分类（含其下所有单位）的种子定义。
// SeedCategory describes one physical-quantity category with all its units.
type SeedCategory struct {
	Code           string
	Name           string
	NameEn         string
	Quantity       string
	BaseUnitSymbol string
	Description    string
	SortOrder      uint32
	Units          []SeedUnit
}

// SeedUnit 单个单位的种子定义。
// SeedUnit describes one unit within a category.
type SeedUnit struct {
	Code           string
	Symbol         string
	Name           string
	NameEn         string
	IsBase         bool
	ConversionType thingmodelV1.ConversionType
	Factor         float64 // 正向系数 k：base = x·k + offset / Forward factor
	Offset         float64 // 偏移 b（仅 AFFINE 非 0） / Affine offset
	FormulaExpr    string  // 公式说明（仅展示） / Display-only formula
	Precision      int32   // 建议显示精度 / Display precision
	IsSiUnit       bool    // 是否 SI 单位 / Is SI unit
	IsLegalUnit    bool    // 是否中国法定计量单位 / Is PRC legal unit
	SortOrder      uint32
}

// 缩写：换算类型 / Conversion-type aliases used in the data tables below.
const (
	ctLinear      = thingmodelV1.ConversionType_LINEAR
	ctAffine      = thingmodelV1.ConversionType_AFFINE
	ctLogarithmic = thingmodelV1.ConversionType_LOGARITHMIC
	ctConditional = thingmodelV1.ConversionType_CONDITIONAL
	ctNone        = thingmodelV1.ConversionType_NONE
)

// UnitSeedData 全部分类与单位的种子数据。
// UnitSeedData holds all categories and units to seed.
//
// 数量 / Counts: 42 个分类、约 225 个单位（与设计文档 07 章 §1 对齐）。
// 治理点：原清单"电池"分类已拆为 electric_charge (Ah 基准) 与 battery_energy (Wh 基准)；
//   见设计文档 07 章 §5。
//
// 注意：数据顺序与清单尽量一致；基准单位 IsBase=true 且 Factor=1, Offset=0。
var UnitSeedData = []SeedCategory{
	// 1. 温度 / Temperature
	{
		Code: "temperature", Name: "温度", NameEn: "Temperature",
		Quantity: "热力学温度", BaseUnitSymbol: "K", SortOrder: 10,
		Units: []SeedUnit{
			{Code: "kelvin", Symbol: "K", Name: "开尔文", NameEn: "Kelvin", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 2, IsSiUnit: true, IsLegalUnit: true},
			{Code: "celsius", Symbol: "℃", Name: "摄氏度", NameEn: "Celsius", ConversionType: ctAffine, Factor: 1, Offset: 273.15, FormulaExpr: "x+273.15", Precision: 2, IsSiUnit: false, IsLegalUnit: true, SortOrder: 1},
			{Code: "fahrenheit", Symbol: "℉", Name: "华氏度", NameEn: "Fahrenheit", ConversionType: ctAffine, Factor: 5.0 / 9.0, Offset: 459.67 * 5.0 / 9.0, FormulaExpr: "(x+459.67)*5/9", Precision: 4, SortOrder: 2},
			{Code: "rankine", Symbol: "°R", Name: "兰氏度", NameEn: "Rankine", ConversionType: ctLinear, Factor: 5.0 / 9.0, FormulaExpr: "x*5/9", Precision: 4, SortOrder: 3},
		},
	},
	// 2. 压力 / Pressure
	{
		Code: "pressure", Name: "压力", NameEn: "Pressure",
		Quantity: "压强", BaseUnitSymbol: "Pa", SortOrder: 20,
		Units: []SeedUnit{
			{Code: "pascal", Symbol: "Pa", Name: "帕斯卡", NameEn: "Pascal", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 2, IsSiUnit: true, IsLegalUnit: true},
			{Code: "kilopascal", Symbol: "kPa", Name: "千帕", NameEn: "Kilopascal", ConversionType: ctLinear, Factor: 1000, FormulaExpr: "x*1000", Precision: 2, IsSiUnit: true, IsLegalUnit: true, SortOrder: 1},
			{Code: "megapascal", Symbol: "MPa", Name: "兆帕", NameEn: "Megapascal", ConversionType: ctLinear, Factor: 1_000_000, FormulaExpr: "x*1000000", Precision: 4, IsSiUnit: true, IsLegalUnit: true, SortOrder: 2},
			{Code: "bar", Symbol: "bar", Name: "巴", NameEn: "Bar", ConversionType: ctLinear, Factor: 100000, FormulaExpr: "x*100000", Precision: 4, SortOrder: 3},
			{Code: "millibar", Symbol: "mbar", Name: "毫巴", NameEn: "Millibar", ConversionType: ctLinear, Factor: 100, FormulaExpr: "x*100", Precision: 2, SortOrder: 4},
			{Code: "mmhg", Symbol: "mmHg", Name: "毫米汞柱", NameEn: "Millimeter of mercury", ConversionType: ctLinear, Factor: 133.322, FormulaExpr: "x*133.322", Precision: 4, SortOrder: 5},
			{Code: "atm", Symbol: "atm", Name: "标准大气压", NameEn: "Atmosphere", ConversionType: ctLinear, Factor: 101325, FormulaExpr: "x*101325", Precision: 4, SortOrder: 6},
			{Code: "psi", Symbol: "psi", Name: "磅力每平方英寸", NameEn: "Pound per square inch", ConversionType: ctLinear, Factor: 6894.757, FormulaExpr: "x*6894.757", Precision: 4, SortOrder: 7},
			{Code: "inh2o", Symbol: "inH2O", Name: "英寸水柱", NameEn: "Inch of water", ConversionType: ctLinear, Factor: 249.089, FormulaExpr: "x*249.089", Precision: 4, SortOrder: 8},
			{Code: "torr", Symbol: "Torr", Name: "托", NameEn: "Torr", ConversionType: ctLinear, Factor: 133.322, FormulaExpr: "x*133.322", Precision: 4, SortOrder: 9},
		},
	},
	// 3. 长度 / Length
	{
		Code: "length", Name: "长度", NameEn: "Length",
		Quantity: "长度", BaseUnitSymbol: "m", SortOrder: 30,
		Units: []SeedUnit{
			{Code: "meter", Symbol: "m", Name: "米", NameEn: "Meter", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 4, IsSiUnit: true, IsLegalUnit: true},
			{Code: "kilometer", Symbol: "km", Name: "千米", NameEn: "Kilometer", ConversionType: ctLinear, Factor: 1000, FormulaExpr: "x*1000", Precision: 4, IsSiUnit: true, IsLegalUnit: true, SortOrder: 1},
			{Code: "decimeter", Symbol: "dm", Name: "分米", NameEn: "Decimeter", ConversionType: ctLinear, Factor: 0.1, FormulaExpr: "x*0.1", Precision: 4, IsSiUnit: true, IsLegalUnit: true, SortOrder: 2},
			{Code: "centimeter", Symbol: "cm", Name: "厘米", NameEn: "Centimeter", ConversionType: ctLinear, Factor: 0.01, FormulaExpr: "x*0.01", Precision: 4, IsSiUnit: true, IsLegalUnit: true, SortOrder: 3},
			{Code: "millimeter", Symbol: "mm", Name: "毫米", NameEn: "Millimeter", ConversionType: ctLinear, Factor: 0.001, FormulaExpr: "x*0.001", Precision: 4, IsSiUnit: true, IsLegalUnit: true, SortOrder: 4},
			{Code: "micrometer", Symbol: "μm", Name: "微米", NameEn: "Micrometer", ConversionType: ctLinear, Factor: 1e-6, FormulaExpr: "x*0.000001", Precision: 6, IsSiUnit: true, IsLegalUnit: true, SortOrder: 5},
			{Code: "nanometer", Symbol: "nm", Name: "纳米", NameEn: "Nanometer", ConversionType: ctLinear, Factor: 1e-9, FormulaExpr: "x*0.000000001", Precision: 9, IsSiUnit: true, IsLegalUnit: true, SortOrder: 6},
			{Code: "inch", Symbol: "in", Name: "英寸", NameEn: "Inch", ConversionType: ctLinear, Factor: 0.0254, FormulaExpr: "x*0.0254", Precision: 4, SortOrder: 7},
			{Code: "foot", Symbol: "ft", Name: "英尺", NameEn: "Foot", ConversionType: ctLinear, Factor: 0.3048, FormulaExpr: "x*0.3048", Precision: 4, SortOrder: 8},
			{Code: "yard", Symbol: "yd", Name: "码", NameEn: "Yard", ConversionType: ctLinear, Factor: 0.9144, FormulaExpr: "x*0.9144", Precision: 4, SortOrder: 9},
			{Code: "mile", Symbol: "mi", Name: "英里", NameEn: "Mile", ConversionType: ctLinear, Factor: 1609.344, FormulaExpr: "x*1609.344", Precision: 4, SortOrder: 10},
			{Code: "nautical_mile", Symbol: "nmi", Name: "海里", NameEn: "Nautical mile", ConversionType: ctLinear, Factor: 1852, FormulaExpr: "x*1852", Precision: 4, IsLegalUnit: true, SortOrder: 11},
			{Code: "mil", Symbol: "mil", Name: "密耳", NameEn: "Mil", ConversionType: ctLinear, Factor: 0.0000254, FormulaExpr: "x*0.0000254", Precision: 7, SortOrder: 12},
		},
	},
	// 4. 面积 / Area
	{
		Code: "area", Name: "面积", NameEn: "Area",
		Quantity: "面积", BaseUnitSymbol: "m²", SortOrder: 40,
		Units: []SeedUnit{
			{Code: "square_meter", Symbol: "m²", Name: "平方米", NameEn: "Square meter", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 4, IsSiUnit: true, IsLegalUnit: true},
			{Code: "square_kilometer", Symbol: "km²", Name: "平方千米", NameEn: "Square kilometer", ConversionType: ctLinear, Factor: 1_000_000, FormulaExpr: "x*1000000", Precision: 4, IsSiUnit: true, IsLegalUnit: true, SortOrder: 1},
			{Code: "square_decimeter", Symbol: "dm²", Name: "平方分米", NameEn: "Square decimeter", ConversionType: ctLinear, Factor: 0.01, FormulaExpr: "x*0.01", Precision: 4, IsSiUnit: true, IsLegalUnit: true, SortOrder: 2},
			{Code: "square_centimeter", Symbol: "cm²", Name: "平方厘米", NameEn: "Square centimeter", ConversionType: ctLinear, Factor: 0.0001, FormulaExpr: "x*0.0001", Precision: 6, IsSiUnit: true, IsLegalUnit: true, SortOrder: 3},
			{Code: "square_millimeter", Symbol: "mm²", Name: "平方毫米", NameEn: "Square millimeter", ConversionType: ctLinear, Factor: 1e-6, FormulaExpr: "x*0.000001", Precision: 6, IsSiUnit: true, IsLegalUnit: true, SortOrder: 4},
			{Code: "hectare", Symbol: "ha", Name: "公顷", NameEn: "Hectare", ConversionType: ctLinear, Factor: 10000, FormulaExpr: "x*10000", Precision: 4, IsLegalUnit: true, SortOrder: 5},
			{Code: "mu", Symbol: "mu", Name: "亩", NameEn: "Mu (Chinese unit)", ConversionType: ctLinear, Factor: 2000.0 / 3.0, FormulaExpr: "x*2000/3", Precision: 4, SortOrder: 6},
			{Code: "square_foot", Symbol: "ft²", Name: "平方英尺", NameEn: "Square foot", ConversionType: ctLinear, Factor: 0.092903, FormulaExpr: "x*0.092903", Precision: 6, SortOrder: 7},
			{Code: "square_inch", Symbol: "in²", Name: "平方英寸", NameEn: "Square inch", ConversionType: ctLinear, Factor: 0.00064516, FormulaExpr: "x*0.00064516", Precision: 8, SortOrder: 8},
			{Code: "acre", Symbol: "ac", Name: "英亩", NameEn: "Acre", ConversionType: ctLinear, Factor: 4046.856, FormulaExpr: "x*4046.856", Precision: 4, SortOrder: 9},
		},
	},
	// 5. 体积 / Volume
	{
		Code: "volume", Name: "体积", NameEn: "Volume",
		Quantity: "体积", BaseUnitSymbol: "m³", SortOrder: 50,
		Units: []SeedUnit{
			{Code: "cubic_meter", Symbol: "m³", Name: "立方米", NameEn: "Cubic meter", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 4, IsSiUnit: true, IsLegalUnit: true},
			{Code: "liter", Symbol: "L", Name: "升", NameEn: "Liter", ConversionType: ctLinear, Factor: 0.001, FormulaExpr: "x*0.001", Precision: 4, IsLegalUnit: true, SortOrder: 1},
			{Code: "milliliter", Symbol: "mL", Name: "毫升", NameEn: "Milliliter", ConversionType: ctLinear, Factor: 1e-6, FormulaExpr: "x*0.000001", Precision: 6, IsLegalUnit: true, SortOrder: 2},
			{Code: "cubic_centimeter", Symbol: "cm³", Name: "立方厘米", NameEn: "Cubic centimeter", ConversionType: ctLinear, Factor: 1e-6, FormulaExpr: "x*0.000001", Precision: 6, IsSiUnit: true, IsLegalUnit: true, SortOrder: 3},
			{Code: "cubic_decimeter", Symbol: "dm³", Name: "立方分米", NameEn: "Cubic decimeter", ConversionType: ctLinear, Factor: 0.001, FormulaExpr: "x*0.001", Precision: 4, IsSiUnit: true, IsLegalUnit: true, SortOrder: 4},
			{Code: "gallon_us", Symbol: "gal(US)", Name: "美加仑", NameEn: "US gallon", ConversionType: ctLinear, Factor: 0.00378541, FormulaExpr: "x*0.00378541", Precision: 6, SortOrder: 5},
			{Code: "gallon_uk", Symbol: "gal(UK)", Name: "英加仑", NameEn: "UK gallon", ConversionType: ctLinear, Factor: 0.00454609, FormulaExpr: "x*0.00454609", Precision: 6, SortOrder: 6},
			{Code: "cubic_foot", Symbol: "ft³", Name: "立方英尺", NameEn: "Cubic foot", ConversionType: ctLinear, Factor: 0.0283168, FormulaExpr: "x*0.0283168", Precision: 6, SortOrder: 7},
			{Code: "cubic_inch", Symbol: "in³", Name: "立方英寸", NameEn: "Cubic inch", ConversionType: ctLinear, Factor: 0.0000163871, FormulaExpr: "x*0.0000163871", Precision: 8, SortOrder: 8},
			{Code: "barrel_oil", Symbol: "bbl", Name: "美桶(石油)", NameEn: "Oil barrel", ConversionType: ctLinear, Factor: 0.158987, FormulaExpr: "x*0.158987", Precision: 6, SortOrder: 9},
		},
	},
	// 6. 质量 / Mass
	{
		Code: "mass", Name: "质量", NameEn: "Mass",
		Quantity: "质量", BaseUnitSymbol: "kg", SortOrder: 60,
		Units: []SeedUnit{
			{Code: "kilogram", Symbol: "kg", Name: "千克", NameEn: "Kilogram", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 4, IsSiUnit: true, IsLegalUnit: true},
			{Code: "gram", Symbol: "g", Name: "克", NameEn: "Gram", ConversionType: ctLinear, Factor: 0.001, FormulaExpr: "x*0.001", Precision: 4, IsSiUnit: true, IsLegalUnit: true, SortOrder: 1},
			{Code: "milligram", Symbol: "mg", Name: "毫克", NameEn: "Milligram", ConversionType: ctLinear, Factor: 1e-6, FormulaExpr: "x*0.000001", Precision: 6, IsSiUnit: true, IsLegalUnit: true, SortOrder: 2},
			{Code: "microgram", Symbol: "μg", Name: "微克", NameEn: "Microgram", ConversionType: ctLinear, Factor: 1e-9, FormulaExpr: "x*0.000000001", Precision: 9, IsSiUnit: true, IsLegalUnit: true, SortOrder: 3},
			{Code: "ton", Symbol: "t", Name: "吨", NameEn: "Tonne", ConversionType: ctLinear, Factor: 1000, FormulaExpr: "x*1000", Precision: 4, IsLegalUnit: true, SortOrder: 4},
			{Code: "pound", Symbol: "lb", Name: "磅", NameEn: "Pound", ConversionType: ctLinear, Factor: 0.45359237, FormulaExpr: "x*0.45359237", Precision: 6, SortOrder: 5},
			{Code: "ounce_av", Symbol: "oz", Name: "盎司(常衡)", NameEn: "Ounce (avoirdupois)", ConversionType: ctLinear, Factor: 0.0283495, FormulaExpr: "x*0.0283495", Precision: 6, SortOrder: 6},
			{Code: "ounce_troy", Symbol: "oz(tr)", Name: "盎司(金衡)", NameEn: "Ounce (troy)", ConversionType: ctLinear, Factor: 0.0311035, FormulaExpr: "x*0.0311035", Precision: 6, SortOrder: 7},
			{Code: "carat", Symbol: "ct", Name: "克拉", NameEn: "Carat", ConversionType: ctLinear, Factor: 0.0002, FormulaExpr: "x*0.0002", Precision: 6, SortOrder: 8},
		},
	},
	// 7. 时间 / Time
	{
		Code: "time", Name: "时间", NameEn: "Time",
		Quantity: "时间", BaseUnitSymbol: "s", SortOrder: 70,
		Units: []SeedUnit{
			{Code: "second", Symbol: "s", Name: "秒", NameEn: "Second", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 4, IsSiUnit: true, IsLegalUnit: true},
			{Code: "millisecond", Symbol: "ms", Name: "毫秒", NameEn: "Millisecond", ConversionType: ctLinear, Factor: 0.001, FormulaExpr: "x*0.001", Precision: 4, IsSiUnit: true, IsLegalUnit: true, SortOrder: 1},
			{Code: "microsecond", Symbol: "μs", Name: "微秒", NameEn: "Microsecond", ConversionType: ctLinear, Factor: 1e-6, FormulaExpr: "x*0.000001", Precision: 6, IsSiUnit: true, IsLegalUnit: true, SortOrder: 2},
			{Code: "nanosecond", Symbol: "ns", Name: "纳秒", NameEn: "Nanosecond", ConversionType: ctLinear, Factor: 1e-9, FormulaExpr: "x*0.000000001", Precision: 9, IsSiUnit: true, IsLegalUnit: true, SortOrder: 3},
			{Code: "minute", Symbol: "min", Name: "分钟", NameEn: "Minute", ConversionType: ctLinear, Factor: 60, FormulaExpr: "x*60", Precision: 4, IsLegalUnit: true, SortOrder: 4},
			{Code: "hour", Symbol: "h", Name: "小时", NameEn: "Hour", ConversionType: ctLinear, Factor: 3600, FormulaExpr: "x*3600", Precision: 4, IsLegalUnit: true, SortOrder: 5},
			{Code: "day", Symbol: "d", Name: "天", NameEn: "Day", ConversionType: ctLinear, Factor: 86400, FormulaExpr: "x*86400", Precision: 4, IsLegalUnit: true, SortOrder: 6},
			{Code: "week", Symbol: "wk", Name: "周", NameEn: "Week", ConversionType: ctLinear, Factor: 604800, FormulaExpr: "x*604800", Precision: 4, SortOrder: 7},
		},
	},
	// 8. 速度 / Speed
	{
		Code: "speed", Name: "速度", NameEn: "Speed",
		Quantity: "速度", BaseUnitSymbol: "m/s", SortOrder: 80,
		Units: []SeedUnit{
			{Code: "meter_per_second", Symbol: "m/s", Name: "米每秒", NameEn: "Meter per second", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 4, IsSiUnit: true, IsLegalUnit: true},
			{Code: "kmh", Symbol: "km/h", Name: "千米每小时", NameEn: "Kilometer per hour", ConversionType: ctLinear, Factor: 1.0 / 3.6, FormulaExpr: "x/3.6", Precision: 4, IsLegalUnit: true, SortOrder: 1},
			{Code: "mph", Symbol: "mph", Name: "英里每小时", NameEn: "Mile per hour", ConversionType: ctLinear, Factor: 0.44704, FormulaExpr: "x*0.44704", Precision: 4, SortOrder: 2},
			{Code: "knot", Symbol: "kn", Name: "节", NameEn: "Knot", ConversionType: ctLinear, Factor: 0.514444, FormulaExpr: "x*0.514444", Precision: 4, IsLegalUnit: true, SortOrder: 3},
			{Code: "fps", Symbol: "ft/s", Name: "英尺每秒", NameEn: "Foot per second", ConversionType: ctLinear, Factor: 0.3048, FormulaExpr: "x*0.3048", Precision: 4, SortOrder: 4},
			{Code: "fpm", Symbol: "ft/min", Name: "英尺每分钟", NameEn: "Foot per minute", ConversionType: ctLinear, Factor: 0.00508, FormulaExpr: "x*0.00508", Precision: 4, SortOrder: 5},
			{Code: "mach", Symbol: "Mach", Name: "马赫", NameEn: "Mach", ConversionType: ctLinear, Factor: 340.3, FormulaExpr: "x*340.3", Precision: 4, SortOrder: 6},
		},
	},
	// 9. 电压 / Voltage
	{
		Code: "voltage", Name: "电压", NameEn: "Voltage",
		Quantity: "电位差", BaseUnitSymbol: "V", SortOrder: 90,
		Units: []SeedUnit{
			{Code: "volt", Symbol: "V", Name: "伏特", NameEn: "Volt", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 4, IsSiUnit: true, IsLegalUnit: true},
			{Code: "kilovolt", Symbol: "kV", Name: "千伏", NameEn: "Kilovolt", ConversionType: ctLinear, Factor: 1000, FormulaExpr: "x*1000", Precision: 4, IsSiUnit: true, IsLegalUnit: true, SortOrder: 1},
			{Code: "millivolt", Symbol: "mV", Name: "毫伏", NameEn: "Millivolt", ConversionType: ctLinear, Factor: 0.001, FormulaExpr: "x*0.001", Precision: 4, IsSiUnit: true, IsLegalUnit: true, SortOrder: 2},
			{Code: "microvolt", Symbol: "μV", Name: "微伏", NameEn: "Microvolt", ConversionType: ctLinear, Factor: 1e-6, FormulaExpr: "x*0.000001", Precision: 6, IsSiUnit: true, IsLegalUnit: true, SortOrder: 3},
		},
	},
	// 10. 电流 / Current
	{
		Code: "current", Name: "电流", NameEn: "Current",
		Quantity: "电流", BaseUnitSymbol: "A", SortOrder: 100,
		Units: []SeedUnit{
			{Code: "ampere", Symbol: "A", Name: "安培", NameEn: "Ampere", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 4, IsSiUnit: true, IsLegalUnit: true},
			{Code: "kiloampere", Symbol: "kA", Name: "千安", NameEn: "Kiloampere", ConversionType: ctLinear, Factor: 1000, FormulaExpr: "x*1000", Precision: 4, IsSiUnit: true, IsLegalUnit: true, SortOrder: 1},
			{Code: "milliampere", Symbol: "mA", Name: "毫安", NameEn: "Milliampere", ConversionType: ctLinear, Factor: 0.001, FormulaExpr: "x*0.001", Precision: 4, IsSiUnit: true, IsLegalUnit: true, SortOrder: 2},
			{Code: "microampere", Symbol: "μA", Name: "微安", NameEn: "Microampere", ConversionType: ctLinear, Factor: 1e-6, FormulaExpr: "x*0.000001", Precision: 6, IsSiUnit: true, IsLegalUnit: true, SortOrder: 3},
		},
	},
	// 11. 功率 / Power
	{
		Code: "power", Name: "功率", NameEn: "Power",
		Quantity: "功率", BaseUnitSymbol: "W",
		Description: "含视在功率(VA)/无功功率(var)等量纲单位；它们与有功功率(W)量纲相同但物理含义不同。 / VA, var share same dimension with W but are not physically interchangeable.",
		SortOrder:   110,
		Units: []SeedUnit{
			{Code: "watt", Symbol: "W", Name: "瓦特", NameEn: "Watt", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 4, IsSiUnit: true, IsLegalUnit: true},
			{Code: "kilowatt", Symbol: "kW", Name: "千瓦", NameEn: "Kilowatt", ConversionType: ctLinear, Factor: 1000, FormulaExpr: "x*1000", Precision: 4, IsSiUnit: true, IsLegalUnit: true, SortOrder: 1},
			{Code: "megawatt", Symbol: "MW", Name: "兆瓦", NameEn: "Megawatt", ConversionType: ctLinear, Factor: 1_000_000, FormulaExpr: "x*1000000", Precision: 4, IsSiUnit: true, IsLegalUnit: true, SortOrder: 2},
			{Code: "milliwatt", Symbol: "mW", Name: "毫瓦", NameEn: "Milliwatt", ConversionType: ctLinear, Factor: 0.001, FormulaExpr: "x*0.001", Precision: 4, IsSiUnit: true, IsLegalUnit: true, SortOrder: 3},
			{Code: "hp_metric", Symbol: "PS", Name: "马力(米制)", NameEn: "Metric horsepower", ConversionType: ctLinear, Factor: 735.499, FormulaExpr: "x*735.499", Precision: 4, SortOrder: 4},
			{Code: "hp_imperial", Symbol: "hp", Name: "马力(英制)", NameEn: "Imperial horsepower", ConversionType: ctLinear, Factor: 745.7, FormulaExpr: "x*745.7", Precision: 4, SortOrder: 5},
			{Code: "btu_per_hour", Symbol: "Btu/h", Name: "英热单位每小时", NameEn: "BTU per hour", ConversionType: ctLinear, Factor: 0.293071, FormulaExpr: "x*0.293071", Precision: 6, SortOrder: 6},
			{Code: "va", Symbol: "VA", Name: "伏安(视在功率)", NameEn: "Volt-ampere", ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1 (等量纲)", Precision: 4, SortOrder: 7},
			{Code: "kva", Symbol: "kVA", Name: "千伏安", NameEn: "Kilovolt-ampere", ConversionType: ctLinear, Factor: 1000, FormulaExpr: "x*1000 (等量纲)", Precision: 4, SortOrder: 8},
			{Code: "var", Symbol: "var", Name: "乏(无功功率)", NameEn: "Var (reactive power)", ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1 (等量纲)", Precision: 4, SortOrder: 9},
			{Code: "kvar", Symbol: "kvar", Name: "千乏", NameEn: "Kilovar", ConversionType: ctLinear, Factor: 1000, FormulaExpr: "x*1000 (等量纲)", Precision: 4, SortOrder: 10},
			{Code: "kwp", Symbol: "kWp", Name: "千瓦峰(光伏)", NameEn: "Kilowatt-peak", ConversionType: ctLinear, Factor: 1000, FormulaExpr: "x*1000 (等量纲)", Precision: 4, SortOrder: 11},
		},
	},
	// 12. 能量 / Energy
	{
		Code: "energy", Name: "能量", NameEn: "Energy",
		Quantity: "能、功、热", BaseUnitSymbol: "J", SortOrder: 120,
		Units: []SeedUnit{
			{Code: "joule", Symbol: "J", Name: "焦耳", NameEn: "Joule", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 4, IsSiUnit: true, IsLegalUnit: true},
			{Code: "kilojoule", Symbol: "kJ", Name: "千焦", NameEn: "Kilojoule", ConversionType: ctLinear, Factor: 1000, FormulaExpr: "x*1000", Precision: 4, IsSiUnit: true, IsLegalUnit: true, SortOrder: 1},
			{Code: "megajoule", Symbol: "MJ", Name: "兆焦", NameEn: "Megajoule", ConversionType: ctLinear, Factor: 1_000_000, FormulaExpr: "x*1000000", Precision: 4, IsSiUnit: true, IsLegalUnit: true, SortOrder: 2},
			{Code: "kwh", Symbol: "kWh", Name: "千瓦时", NameEn: "Kilowatt-hour", ConversionType: ctLinear, Factor: 3_600_000, FormulaExpr: "x*3600000", Precision: 4, IsLegalUnit: true, SortOrder: 3},
			{Code: "mwh", Symbol: "MWh", Name: "兆瓦时", NameEn: "Megawatt-hour", ConversionType: ctLinear, Factor: 3_600_000_000, FormulaExpr: "x*3600000000", Precision: 4, SortOrder: 4},
			{Code: "cal", Symbol: "cal", Name: "卡(热化学)", NameEn: "Calorie (thermochem.)", ConversionType: ctLinear, Factor: 4.184, FormulaExpr: "x*4.184", Precision: 4, SortOrder: 5},
			{Code: "kcal", Symbol: "kcal", Name: "千卡", NameEn: "Kilocalorie", ConversionType: ctLinear, Factor: 4184, FormulaExpr: "x*4184", Precision: 4, SortOrder: 6},
			{Code: "btu", Symbol: "Btu", Name: "英热单位", NameEn: "British thermal unit", ConversionType: ctLinear, Factor: 1055.056, FormulaExpr: "x*1055.056", Precision: 4, SortOrder: 7},
			{Code: "ev", Symbol: "eV", Name: "电子伏特", NameEn: "Electronvolt", ConversionType: ctLinear, Factor: 1.602176e-19, FormulaExpr: "x*1.602176e-19", Precision: 9, SortOrder: 8},
			{Code: "kvarh", Symbol: "kvarh", Name: "千乏时(无功电能)", NameEn: "Kilovar-hour", ConversionType: ctLinear, Factor: 3_600_000, FormulaExpr: "x*3600000 (等量纲)", Precision: 4, SortOrder: 9},
		},
	},
	// 13. 频率 / Frequency
	{
		Code: "frequency", Name: "频率", NameEn: "Frequency",
		Quantity: "频率", BaseUnitSymbol: "Hz", SortOrder: 130,
		Units: []SeedUnit{
			{Code: "hertz", Symbol: "Hz", Name: "赫兹", NameEn: "Hertz", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 4, IsSiUnit: true, IsLegalUnit: true},
			{Code: "kilohertz", Symbol: "kHz", Name: "千赫兹", NameEn: "Kilohertz", ConversionType: ctLinear, Factor: 1000, FormulaExpr: "x*1000", Precision: 4, IsSiUnit: true, IsLegalUnit: true, SortOrder: 1},
			{Code: "megahertz", Symbol: "MHz", Name: "兆赫兹", NameEn: "Megahertz", ConversionType: ctLinear, Factor: 1_000_000, FormulaExpr: "x*1000000", Precision: 4, IsSiUnit: true, IsLegalUnit: true, SortOrder: 2},
			{Code: "gigahertz", Symbol: "GHz", Name: "吉赫兹", NameEn: "Gigahertz", ConversionType: ctLinear, Factor: 1_000_000_000, FormulaExpr: "x*1000000000", Precision: 4, IsSiUnit: true, IsLegalUnit: true, SortOrder: 3},
			{Code: "rpm", Symbol: "rpm", Name: "转每分钟", NameEn: "Revolutions per minute", ConversionType: ctLinear, Factor: 1.0 / 60.0, FormulaExpr: "x/60", Precision: 6, IsLegalUnit: true, SortOrder: 4},
			{Code: "rps", Symbol: "rps", Name: "转每秒", NameEn: "Revolutions per second", ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 4, SortOrder: 5},
		},
	},
	// 14. 湿度 / Humidity
	{
		Code: "humidity", Name: "湿度", NameEn: "Humidity",
		Quantity:    "相对湿度",
		Description: "本分类内多个基准单位互不可换算（%RH/g·m⁻³/g·kg⁻¹/露点℃），跨基准时由 conversion_type 限制自然拒绝。 / Multiple non-interchangeable bases coexist; cross-base conversion is rejected.",
		BaseUnitSymbol: "%RH", SortOrder: 140,
		Units: []SeedUnit{
			{Code: "rh", Symbol: "%RH", Name: "相对湿度", NameEn: "Relative humidity", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 2},
			{Code: "humidity_absolute", Symbol: "g/m³", Name: "绝对湿度", NameEn: "Absolute humidity", ConversionType: ctNone, Factor: 1, FormulaExpr: "x*1", Precision: 4, SortOrder: 1},
			{Code: "humidity_specific", Symbol: "g/kg", Name: "含湿量", NameEn: "Specific humidity", ConversionType: ctNone, Factor: 1, FormulaExpr: "x*1", Precision: 4, SortOrder: 2},
			{Code: "dewpoint_celsius", Symbol: "℃(露点)", Name: "露点温度", NameEn: "Dew point (Celsius)", ConversionType: ctNone, Factor: 1, FormulaExpr: "x+273.15", Precision: 2, SortOrder: 3},
		},
	},
	// 15. 照度 / Illuminance
	{
		Code: "illuminance", Name: "照度", NameEn: "Illuminance",
		Quantity: "发光相关", BaseUnitSymbol: "lx", SortOrder: 150,
		Units: []SeedUnit{
			{Code: "lux", Symbol: "lx", Name: "勒克斯", NameEn: "Lux", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 4, IsSiUnit: true, IsLegalUnit: true},
			{Code: "kilolux", Symbol: "klx", Name: "千勒克斯", NameEn: "Kilolux", ConversionType: ctLinear, Factor: 1000, FormulaExpr: "x*1000", Precision: 4, IsSiUnit: true, IsLegalUnit: true, SortOrder: 1},
			{Code: "candela", Symbol: "cd", Name: "坎德拉", NameEn: "Candela", ConversionType: ctNone, Factor: 1, FormulaExpr: "x*1", Precision: 4, IsSiUnit: true, IsLegalUnit: true, SortOrder: 2},
			{Code: "lumen", Symbol: "lm", Name: "流明", NameEn: "Lumen", ConversionType: ctNone, Factor: 1, FormulaExpr: "x*1", Precision: 4, IsSiUnit: true, IsLegalUnit: true, SortOrder: 3},
			{Code: "cd_per_m2", Symbol: "cd/m²", Name: "坎德拉每平方米", NameEn: "Candela per square meter", ConversionType: ctNone, Factor: 1, FormulaExpr: "x*1", Precision: 4, IsSiUnit: true, IsLegalUnit: true, SortOrder: 4},
			{Code: "nit", Symbol: "nt", Name: "尼特", NameEn: "Nit", ConversionType: ctNone, Factor: 1, FormulaExpr: "x*1", Precision: 4, SortOrder: 5},
		},
	},
	// 16. 角度 / Angle
	{
		Code: "angle", Name: "角度", NameEn: "Angle",
		Quantity: "平面角", BaseUnitSymbol: "rad", SortOrder: 160,
		Units: []SeedUnit{
			{Code: "radian", Symbol: "rad", Name: "弧度", NameEn: "Radian", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 6, IsSiUnit: true, IsLegalUnit: true},
			{Code: "degree", Symbol: "°", Name: "度", NameEn: "Degree", ConversionType: ctLinear, Factor: piOver(180), FormulaExpr: "x*π/180", Precision: 6, IsLegalUnit: true, SortOrder: 1},
			{Code: "arcminute", Symbol: "′", Name: "角分", NameEn: "Arcminute", ConversionType: ctLinear, Factor: piOver(10800), FormulaExpr: "x*π/10800", Precision: 8, IsLegalUnit: true, SortOrder: 2},
			{Code: "arcsecond", Symbol: "″", Name: "角秒", NameEn: "Arcsecond", ConversionType: ctLinear, Factor: piOver(648000), FormulaExpr: "x*π/648000", Precision: 9, IsLegalUnit: true, SortOrder: 3},
			{Code: "gradian", Symbol: "grad", Name: "百分度", NameEn: "Gradian", ConversionType: ctLinear, Factor: piOver(200), FormulaExpr: "x*π/200", Precision: 6, SortOrder: 4},
		},
	},
	// 17. 浓度 / Concentration（含 ppm 基准 + 多个非可换算子群）
	{
		Code: "concentration", Name: "浓度", NameEn: "Concentration",
		Quantity:    "浓度（混合）",
		Description: "本分类内含 ppm / mg·m⁻³ / mol·L⁻¹ 等多个不同基准；跨子群依赖外部参数（M, T），由 CONDITIONAL/NONE 类型自然拒绝。 / Multiple bases; cross-subgroup needs M/T params.",
		BaseUnitSymbol: "ppm", SortOrder: 170,
		Units: []SeedUnit{
			{Code: "ppm", Symbol: "ppm", Name: "百万分率", NameEn: "Parts per million", IsBase: true, ConversionType: ctConditional, Factor: 1, FormulaExpr: "x*1 (跨基准需 M/T)", Precision: 4},
			{Code: "ppb", Symbol: "ppb", Name: "十亿分率", NameEn: "Parts per billion", ConversionType: ctLinear, Factor: 0.001, FormulaExpr: "x*0.001", Precision: 6, SortOrder: 1},
			{Code: "mg_per_m3", Symbol: "mg/m³", Name: "毫克每立方米", NameEn: "Milligram per cubic meter", ConversionType: ctConditional, Factor: 1, FormulaExpr: "mg/m³ = ppm·M/22.4 (STP)", Precision: 4, SortOrder: 2},
			{Code: "ug_per_m3", Symbol: "μg/m³", Name: "微克每立方米", NameEn: "Microgram per cubic meter", ConversionType: ctConditional, Factor: 0.001, FormulaExpr: "x*0.001", Precision: 6, SortOrder: 3},
			{Code: "mg_per_l", Symbol: "mg/L", Name: "毫克每升", NameEn: "Milligram per liter", ConversionType: ctConditional, Factor: 1, FormulaExpr: "x*1", Precision: 4, SortOrder: 4},
			{Code: "ug_per_l", Symbol: "μg/L", Name: "微克每升", NameEn: "Microgram per liter", ConversionType: ctConditional, Factor: 0.001, FormulaExpr: "x*0.001", Precision: 6, SortOrder: 5},
			{Code: "concentration_percent", Symbol: "%", Name: "百分比(浓度)", NameEn: "Percent (concentration)", ConversionType: ctNone, Factor: 1, FormulaExpr: "x*1", Precision: 4, SortOrder: 6},
			{Code: "mol_per_l", Symbol: "mol/L", Name: "摩尔每升", NameEn: "Mole per liter", ConversionType: ctConditional, Factor: 1, FormulaExpr: "x*1", Precision: 4, SortOrder: 7},
			{Code: "mmol_per_l", Symbol: "mmol/L", Name: "毫摩尔每升", NameEn: "Millimole per liter", ConversionType: ctLinear, Factor: 0.001, FormulaExpr: "x*0.001", Precision: 6, SortOrder: 8},
			{Code: "umol_per_l", Symbol: "μmol/L", Name: "微摩尔每升", NameEn: "Micromole per liter", ConversionType: ctLinear, Factor: 1e-6, FormulaExpr: "x*0.000001", Precision: 8, SortOrder: 9},
			{Code: "ph", Symbol: "pH", Name: "pH值", NameEn: "pH", ConversionType: ctNone, Factor: 1, FormulaExpr: "无量纲，不可换算", Precision: 2, SortOrder: 10},
			{Code: "ntu", Symbol: "NTU", Name: "浊度", NameEn: "Nephelometric turbidity unit", ConversionType: ctNone, Factor: 1, FormulaExpr: "x*1", Precision: 4, SortOrder: 11},
		},
	},
	// 18. 流量(体积) / Volumetric flow
	{
		Code: "flow_volumetric", Name: "流量(体积)", NameEn: "Volumetric flow",
		Quantity: "体积流量", BaseUnitSymbol: "m³/s", SortOrder: 180,
		Units: []SeedUnit{
			{Code: "m3_per_s", Symbol: "m³/s", Name: "立方米每秒", NameEn: "Cubic meter per second", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 6, IsSiUnit: true, IsLegalUnit: true},
			{Code: "m3_per_h", Symbol: "m³/h", Name: "立方米每小时", NameEn: "Cubic meter per hour", ConversionType: ctLinear, Factor: 1.0 / 3600.0, FormulaExpr: "x/3600", Precision: 8, SortOrder: 1},
			{Code: "l_per_s", Symbol: "L/s", Name: "升每秒", NameEn: "Liter per second", ConversionType: ctLinear, Factor: 0.001, FormulaExpr: "x*0.001", Precision: 6, SortOrder: 2},
			{Code: "l_per_min", Symbol: "L/min", Name: "升每分钟", NameEn: "Liter per minute", ConversionType: ctLinear, Factor: 0.001 / 60.0, FormulaExpr: "x*0.001/60", Precision: 8, SortOrder: 3},
			{Code: "l_per_h", Symbol: "L/h", Name: "升每小时", NameEn: "Liter per hour", ConversionType: ctLinear, Factor: 0.001 / 3600.0, FormulaExpr: "x*0.001/3600", Precision: 9, SortOrder: 4},
			{Code: "gpm", Symbol: "GPM", Name: "美加仑每分钟", NameEn: "US gallon per minute", ConversionType: ctLinear, Factor: 6.30902e-5, FormulaExpr: "x*6.30902e-5", Precision: 9, SortOrder: 5},
		},
	},
	// 19. 流量(质量) / Mass flow
	{
		Code: "flow_mass", Name: "流量(质量)", NameEn: "Mass flow",
		Quantity: "质量流量", BaseUnitSymbol: "kg/s", SortOrder: 190,
		Units: []SeedUnit{
			{Code: "kg_per_s", Symbol: "kg/s", Name: "千克每秒", NameEn: "Kilogram per second", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 6, IsSiUnit: true, IsLegalUnit: true},
			{Code: "kg_per_h", Symbol: "kg/h", Name: "千克每小时", NameEn: "Kilogram per hour", ConversionType: ctLinear, Factor: 1.0 / 3600.0, FormulaExpr: "x/3600", Precision: 8, SortOrder: 1},
			{Code: "t_per_h", Symbol: "t/h", Name: "吨每小时", NameEn: "Tonne per hour", ConversionType: ctLinear, Factor: 1000.0 / 3600.0, FormulaExpr: "x*1000/3600", Precision: 6, SortOrder: 2},
		},
	},
	// 20. 力 / Force
	{
		Code: "force", Name: "力", NameEn: "Force",
		Quantity: "力", BaseUnitSymbol: "N", SortOrder: 200,
		Units: []SeedUnit{
			{Code: "newton", Symbol: "N", Name: "牛顿", NameEn: "Newton", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 4, IsSiUnit: true, IsLegalUnit: true},
			{Code: "kilonewton", Symbol: "kN", Name: "千牛", NameEn: "Kilonewton", ConversionType: ctLinear, Factor: 1000, FormulaExpr: "x*1000", Precision: 4, IsSiUnit: true, IsLegalUnit: true, SortOrder: 1},
			{Code: "meganewton", Symbol: "MN", Name: "兆牛", NameEn: "Meganewton", ConversionType: ctLinear, Factor: 1_000_000, FormulaExpr: "x*1000000", Precision: 4, IsSiUnit: true, IsLegalUnit: true, SortOrder: 2},
			{Code: "kgf", Symbol: "kgf", Name: "千克力", NameEn: "Kilogram-force", ConversionType: ctLinear, Factor: 9.80665, FormulaExpr: "x*9.80665", Precision: 6, SortOrder: 3},
			{Code: "lbf", Symbol: "lbf", Name: "磅力", NameEn: "Pound-force", ConversionType: ctLinear, Factor: 4.448222, FormulaExpr: "x*4.448222", Precision: 6, SortOrder: 4},
			{Code: "dyne", Symbol: "dyn", Name: "达因", NameEn: "Dyne", ConversionType: ctLinear, Factor: 1e-5, FormulaExpr: "x*0.00001", Precision: 8, SortOrder: 5},
			{Code: "gf", Symbol: "gf", Name: "克力", NameEn: "Gram-force", ConversionType: ctLinear, Factor: 0.00980665, FormulaExpr: "x*0.00980665", Precision: 8, SortOrder: 6},
		},
	},
	// 21. 力矩 / Torque
	{
		Code: "torque", Name: "力矩", NameEn: "Torque",
		Quantity: "力矩", BaseUnitSymbol: "N·m", SortOrder: 210,
		Units: []SeedUnit{
			{Code: "newton_meter", Symbol: "N·m", Name: "牛顿米", NameEn: "Newton-meter", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 4, IsSiUnit: true, IsLegalUnit: true},
			{Code: "kilonewton_meter", Symbol: "kN·m", Name: "千牛米", NameEn: "Kilonewton-meter", ConversionType: ctLinear, Factor: 1000, FormulaExpr: "x*1000", Precision: 4, IsSiUnit: true, IsLegalUnit: true, SortOrder: 1},
			{Code: "kgf_meter", Symbol: "kgf·m", Name: "千克力米", NameEn: "Kilogram-force meter", ConversionType: ctLinear, Factor: 9.80665, FormulaExpr: "x*9.80665", Precision: 6, SortOrder: 2},
			{Code: "lbf_ft", Symbol: "lbf·ft", Name: "磅力英尺", NameEn: "Pound-force foot", ConversionType: ctLinear, Factor: 1.355818, FormulaExpr: "x*1.355818", Precision: 6, SortOrder: 3},
			{Code: "lbf_in", Symbol: "lbf·in", Name: "磅力英寸", NameEn: "Pound-force inch", ConversionType: ctLinear, Factor: 0.112985, FormulaExpr: "x*0.112985", Precision: 6, SortOrder: 4},
		},
	},
	// 22. 电阻 / Resistance
	{
		Code: "resistance", Name: "电阻", NameEn: "Resistance",
		Quantity: "电阻", BaseUnitSymbol: "Ω", SortOrder: 220,
		Units: []SeedUnit{
			{Code: "ohm", Symbol: "Ω", Name: "欧姆", NameEn: "Ohm", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 4, IsSiUnit: true, IsLegalUnit: true},
			{Code: "kilohm", Symbol: "kΩ", Name: "千欧", NameEn: "Kilohm", ConversionType: ctLinear, Factor: 1000, FormulaExpr: "x*1000", Precision: 4, IsSiUnit: true, IsLegalUnit: true, SortOrder: 1},
			{Code: "megohm", Symbol: "MΩ", Name: "兆欧", NameEn: "Megohm", ConversionType: ctLinear, Factor: 1_000_000, FormulaExpr: "x*1000000", Precision: 4, IsSiUnit: true, IsLegalUnit: true, SortOrder: 2},
			{Code: "milliohm", Symbol: "mΩ", Name: "毫欧", NameEn: "Milliohm", ConversionType: ctLinear, Factor: 0.001, FormulaExpr: "x*0.001", Precision: 4, IsSiUnit: true, IsLegalUnit: true, SortOrder: 3},
			{Code: "microohm", Symbol: "μΩ", Name: "微欧", NameEn: "Microohm", ConversionType: ctLinear, Factor: 1e-6, FormulaExpr: "x*0.000001", Precision: 6, IsSiUnit: true, IsLegalUnit: true, SortOrder: 4},
		},
	},
	// 23. 电容 / Capacitance
	{
		Code: "capacitance", Name: "电容", NameEn: "Capacitance",
		Quantity: "电容", BaseUnitSymbol: "F", SortOrder: 230,
		Units: []SeedUnit{
			{Code: "farad", Symbol: "F", Name: "法拉", NameEn: "Farad", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 4, IsSiUnit: true, IsLegalUnit: true},
			{Code: "millifarad", Symbol: "mF", Name: "毫法", NameEn: "Millifarad", ConversionType: ctLinear, Factor: 0.001, FormulaExpr: "x*0.001", Precision: 6, IsSiUnit: true, IsLegalUnit: true, SortOrder: 1},
			{Code: "microfarad", Symbol: "μF", Name: "微法", NameEn: "Microfarad", ConversionType: ctLinear, Factor: 1e-6, FormulaExpr: "x*0.000001", Precision: 8, IsSiUnit: true, IsLegalUnit: true, SortOrder: 2},
			{Code: "nanofarad", Symbol: "nF", Name: "纳法", NameEn: "Nanofarad", ConversionType: ctLinear, Factor: 1e-9, FormulaExpr: "x*1e-9", Precision: 9, IsSiUnit: true, IsLegalUnit: true, SortOrder: 3},
			{Code: "picofarad", Symbol: "pF", Name: "皮法", NameEn: "Picofarad", ConversionType: ctLinear, Factor: 1e-12, FormulaExpr: "x*1e-12", Precision: 9, IsSiUnit: true, IsLegalUnit: true, SortOrder: 4},
		},
	},
	// 24. 电感 / Inductance
	{
		Code: "inductance", Name: "电感", NameEn: "Inductance",
		Quantity: "电感", BaseUnitSymbol: "H", SortOrder: 240,
		Units: []SeedUnit{
			{Code: "henry", Symbol: "H", Name: "亨利", NameEn: "Henry", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 4, IsSiUnit: true, IsLegalUnit: true},
			{Code: "millihenry", Symbol: "mH", Name: "毫亨", NameEn: "Millihenry", ConversionType: ctLinear, Factor: 0.001, FormulaExpr: "x*0.001", Precision: 6, IsSiUnit: true, IsLegalUnit: true, SortOrder: 1},
			{Code: "microhenry", Symbol: "μH", Name: "微亨", NameEn: "Microhenry", ConversionType: ctLinear, Factor: 1e-6, FormulaExpr: "x*0.000001", Precision: 8, IsSiUnit: true, IsLegalUnit: true, SortOrder: 2},
		},
	},
	// 25. 声级 / Sound level
	{
		Code: "sound_level", Name: "声级", NameEn: "Sound level",
		Quantity:    "声压级（对数）",
		Description: "对数尺度，不可线性换算。 / Logarithmic, not linearly convertible.",
		BaseUnitSymbol: "dB", SortOrder: 250,
		Units: []SeedUnit{
			{Code: "db", Symbol: "dB", Name: "分贝", NameEn: "Decibel", IsBase: true, ConversionType: ctLogarithmic, Factor: 1, FormulaExpr: "对数单位，不可线性换算", Precision: 2, IsLegalUnit: true},
			{Code: "dbm", Symbol: "dBm", Name: "分贝毫瓦", NameEn: "dBm", ConversionType: ctLogarithmic, Factor: 1, FormulaExpr: "10^(x/10) mW", Precision: 2, SortOrder: 1},
			{Code: "dba", Symbol: "dB(A)", Name: "A计权分贝", NameEn: "dB(A)", ConversionType: ctLogarithmic, Factor: 1, FormulaExpr: "对数单位", Precision: 2, SortOrder: 2},
			{Code: "dbc", Symbol: "dB(C)", Name: "C计权分贝", NameEn: "dB(C)", ConversionType: ctLogarithmic, Factor: 1, FormulaExpr: "对数单位", Precision: 2, SortOrder: 3},
		},
	},
	// 26. 数据速率 / Data rate
	{
		Code: "data_rate", Name: "数据速率", NameEn: "Data rate",
		Quantity: "数据速率", BaseUnitSymbol: "bps", SortOrder: 260,
		Units: []SeedUnit{
			{Code: "bps", Symbol: "bps", Name: "比特每秒", NameEn: "Bit per second", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 4},
			{Code: "kbps", Symbol: "Kbps", Name: "千比特每秒", NameEn: "Kilobit per second", ConversionType: ctLinear, Factor: 1000, FormulaExpr: "x*1000", Precision: 4, SortOrder: 1},
			{Code: "mbps", Symbol: "Mbps", Name: "兆比特每秒", NameEn: "Megabit per second", ConversionType: ctLinear, Factor: 1_000_000, FormulaExpr: "x*1000000", Precision: 4, SortOrder: 2},
			{Code: "gbps", Symbol: "Gbps", Name: "吉比特每秒", NameEn: "Gigabit per second", ConversionType: ctLinear, Factor: 1_000_000_000, FormulaExpr: "x*1000000000", Precision: 4, SortOrder: 3},
			{Code: "kbyte_per_s", Symbol: "KB/s", Name: "千字节每秒", NameEn: "Kilobyte per second", ConversionType: ctLinear, Factor: 8000, FormulaExpr: "x*8000", Precision: 4, SortOrder: 4},
			{Code: "mbyte_per_s", Symbol: "MB/s", Name: "兆字节每秒", NameEn: "Megabyte per second", ConversionType: ctLinear, Factor: 8_000_000, FormulaExpr: "x*8000000", Precision: 4, SortOrder: 5},
		},
	},
	// 27. 数据大小 / Data size
	{
		Code: "data_size", Name: "数据大小", NameEn: "Data size",
		Quantity: "数据量", BaseUnitSymbol: "B", SortOrder: 270,
		Units: []SeedUnit{
			{Code: "byte", Symbol: "B", Name: "字节", NameEn: "Byte", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 4},
			{Code: "kilobyte", Symbol: "KB", Name: "千字节", NameEn: "Kilobyte", ConversionType: ctLinear, Factor: 1024, FormulaExpr: "x*1024", Precision: 4, SortOrder: 1},
			{Code: "megabyte", Symbol: "MB", Name: "兆字节", NameEn: "Megabyte", ConversionType: ctLinear, Factor: 1048576, FormulaExpr: "x*1048576", Precision: 4, SortOrder: 2},
			{Code: "gigabyte", Symbol: "GB", Name: "吉字节", NameEn: "Gigabyte", ConversionType: ctLinear, Factor: 1073741824, FormulaExpr: "x*1073741824", Precision: 4, SortOrder: 3},
			{Code: "terabyte", Symbol: "TB", Name: "太字节", NameEn: "Terabyte", ConversionType: ctLinear, Factor: 1099511627776, FormulaExpr: "x*1099511627776", Precision: 4, SortOrder: 4},
			{Code: "bit", Symbol: "bit", Name: "比特", NameEn: "Bit", ConversionType: ctLinear, Factor: 0.125, FormulaExpr: "x*0.125", Precision: 4, SortOrder: 5},
		},
	},
	// 28. 电荷量 / Electric charge（拆自原"电池"分类）
	{
		Code: "electric_charge", Name: "电荷量", NameEn: "Electric charge",
		Quantity:    "电荷量",
		Description: "拆自原电池分类，与电池能量(Wh)分开以保证同分类内可换算。1 Ah = 3600 C。 / Split from 'battery'; Ah base ensures linear convertibility (1 Ah = 3600 C).",
		BaseUnitSymbol: "Ah", SortOrder: 280,
		Units: []SeedUnit{
			{Code: "ampere_hour", Symbol: "Ah", Name: "安时", NameEn: "Ampere-hour", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 4},
			{Code: "milliampere_hour", Symbol: "mAh", Name: "毫安时", NameEn: "Milliampere-hour", ConversionType: ctLinear, Factor: 0.001, FormulaExpr: "x*0.001", Precision: 6, SortOrder: 1},
			{Code: "coulomb", Symbol: "C", Name: "库仑", NameEn: "Coulomb", ConversionType: ctLinear, Factor: 1.0 / 3600.0, FormulaExpr: "x/3600 (1 Ah = 3600 C)", Precision: 8, IsSiUnit: true, IsLegalUnit: true, SortOrder: 2},
		},
	},
	// 29. 电池能量 / Battery energy（拆自原"电池"分类）
	{
		Code: "battery_energy", Name: "电池能量", NameEn: "Battery energy",
		Quantity:    "电能（电池）",
		Description: "拆自原电池分类，量纲为电能(Wh)，与电荷量(Ah)不可直接换算。 / Split from 'battery'; energy in Wh, not interchangeable with charge (Ah).",
		BaseUnitSymbol: "Wh", SortOrder: 290,
		Units: []SeedUnit{
			{Code: "watt_hour", Symbol: "Wh", Name: "瓦时", NameEn: "Watt-hour", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 4},
		},
	},
	// 30. 粘度(动力) / Dynamic viscosity
	{
		Code: "viscosity_dynamic", Name: "粘度(动力)", NameEn: "Dynamic viscosity",
		Quantity: "动力粘度", BaseUnitSymbol: "Pa·s", SortOrder: 300,
		Units: []SeedUnit{
			{Code: "pascal_second", Symbol: "Pa·s", Name: "帕斯卡秒", NameEn: "Pascal-second", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 6, IsSiUnit: true, IsLegalUnit: true},
			{Code: "poise", Symbol: "P", Name: "泊", NameEn: "Poise", ConversionType: ctLinear, Factor: 0.1, FormulaExpr: "x*0.1", Precision: 6, SortOrder: 1},
			{Code: "centipoise", Symbol: "cP", Name: "厘泊", NameEn: "Centipoise", ConversionType: ctLinear, Factor: 0.001, FormulaExpr: "x*0.001", Precision: 6, SortOrder: 2},
		},
	},
	// 31. 粘度(运动) / Kinematic viscosity
	{
		Code: "viscosity_kinematic", Name: "粘度(运动)", NameEn: "Kinematic viscosity",
		Quantity: "运动粘度", BaseUnitSymbol: "m²/s", SortOrder: 310,
		Units: []SeedUnit{
			{Code: "m2_per_s", Symbol: "m²/s", Name: "平方米每秒", NameEn: "Square meter per second", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 8, IsSiUnit: true, IsLegalUnit: true},
			{Code: "stokes", Symbol: "St", Name: "斯托克斯", NameEn: "Stokes", ConversionType: ctLinear, Factor: 0.0001, FormulaExpr: "x*0.0001", Precision: 8, SortOrder: 1},
			{Code: "centistokes", Symbol: "cSt", Name: "厘斯托克斯", NameEn: "Centistokes", ConversionType: ctLinear, Factor: 1e-6, FormulaExpr: "x*0.000001", Precision: 9, SortOrder: 2},
		},
	},
	// 32. 密度 / Density
	{
		Code: "density", Name: "密度", NameEn: "Density",
		Quantity: "密度", BaseUnitSymbol: "kg/m³", SortOrder: 320,
		Units: []SeedUnit{
			{Code: "kg_per_m3", Symbol: "kg/m³", Name: "千克每立方米", NameEn: "Kilogram per cubic meter", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 4, IsSiUnit: true, IsLegalUnit: true},
			{Code: "g_per_cm3", Symbol: "g/cm³", Name: "克每立方厘米", NameEn: "Gram per cubic centimeter", ConversionType: ctLinear, Factor: 1000, FormulaExpr: "x*1000", Precision: 4, SortOrder: 1},
			{Code: "g_per_ml", Symbol: "g/mL", Name: "克每毫升", NameEn: "Gram per milliliter", ConversionType: ctLinear, Factor: 1000, FormulaExpr: "x*1000", Precision: 4, SortOrder: 2},
			{Code: "lb_per_ft3", Symbol: "lb/ft³", Name: "磅每立方英尺", NameEn: "Pound per cubic foot", ConversionType: ctLinear, Factor: 16.0185, FormulaExpr: "x*16.0185", Precision: 4, SortOrder: 3},
		},
	},
	// 33. 加速度 / Acceleration
	{
		Code: "acceleration", Name: "加速度", NameEn: "Acceleration",
		Quantity: "加速度", BaseUnitSymbol: "m/s²", SortOrder: 330,
		Units: []SeedUnit{
			{Code: "m_per_s2", Symbol: "m/s²", Name: "米每二次方秒", NameEn: "Meter per second squared", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 4, IsSiUnit: true, IsLegalUnit: true},
			{Code: "standard_gravity", Symbol: "g", Name: "标准重力加速度", NameEn: "Standard gravity", ConversionType: ctLinear, Factor: 9.80665, FormulaExpr: "x*9.80665", Precision: 6, SortOrder: 1},
			{Code: "gal", Symbol: "Gal", Name: "伽", NameEn: "Gal (cm/s²)", ConversionType: ctLinear, Factor: 0.01, FormulaExpr: "x*0.01", Precision: 6, SortOrder: 2},
		},
	},
	// 34. 热通量 / Heat flux
	{
		Code: "heat_flux", Name: "热通量", NameEn: "Heat flux",
		Quantity: "热通量密度", BaseUnitSymbol: "W/m²", SortOrder: 340,
		Units: []SeedUnit{
			{Code: "w_per_m2", Symbol: "W/m²", Name: "瓦每平方米", NameEn: "Watt per square meter", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 4, IsSiUnit: true, IsLegalUnit: true},
		},
	},
	// 35. 热导率 / Thermal conductivity
	{
		Code: "thermal_conductivity", Name: "热导率", NameEn: "Thermal conductivity",
		Quantity: "热导率", BaseUnitSymbol: "W/(m·K)", SortOrder: 350,
		Units: []SeedUnit{
			{Code: "w_per_mk", Symbol: "W/(m·K)", Name: "瓦每米开尔文", NameEn: "Watt per meter-kelvin", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 4, IsSiUnit: true, IsLegalUnit: true},
		},
	},
	// 36. 热容 / Heat capacity
	{
		Code: "heat_capacity", Name: "热容", NameEn: "Heat capacity",
		Quantity: "热容", BaseUnitSymbol: "J/K", SortOrder: 360,
		Units: []SeedUnit{
			{Code: "j_per_k", Symbol: "J/K", Name: "焦耳每开尔文", NameEn: "Joule per kelvin", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 4, IsSiUnit: true, IsLegalUnit: true},
		},
	},
	// 37. 比热容 / Specific heat
	{
		Code: "specific_heat", Name: "比热容", NameEn: "Specific heat capacity",
		Quantity: "比热容", BaseUnitSymbol: "J/(kg·K)", SortOrder: 370,
		Units: []SeedUnit{
			{Code: "j_per_kgk", Symbol: "J/(kg·K)", Name: "焦耳每千克开尔文", NameEn: "Joule per kilogram-kelvin", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 4, IsSiUnit: true, IsLegalUnit: true},
		},
	},
	// 38. 磁场强度 / Magnetic field strength
	{
		Code: "magnetic_field", Name: "磁场强度", NameEn: "Magnetic field strength",
		Quantity: "磁场强度", BaseUnitSymbol: "A/m", SortOrder: 380,
		Units: []SeedUnit{
			{Code: "a_per_m", Symbol: "A/m", Name: "安培每米", NameEn: "Ampere per meter", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 4, IsSiUnit: true, IsLegalUnit: true},
		},
	},
	// 39. 磁通量 / Magnetic flux
	{
		Code: "magnetic_flux", Name: "磁通量", NameEn: "Magnetic flux",
		Quantity: "磁通量", BaseUnitSymbol: "Wb", SortOrder: 390,
		Units: []SeedUnit{
			{Code: "weber", Symbol: "Wb", Name: "韦伯", NameEn: "Weber", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 4, IsSiUnit: true, IsLegalUnit: true},
		},
	},
	// 40. 磁通密度 / Magnetic flux density
	{
		Code: "magnetic_flux_density", Name: "磁通密度", NameEn: "Magnetic flux density",
		Quantity: "磁通密度", BaseUnitSymbol: "T", SortOrder: 400,
		Units: []SeedUnit{
			{Code: "tesla", Symbol: "T", Name: "特斯拉", NameEn: "Tesla", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 6, IsSiUnit: true, IsLegalUnit: true},
			{Code: "gauss", Symbol: "Gs", Name: "高斯", NameEn: "Gauss", ConversionType: ctLinear, Factor: 0.0001, FormulaExpr: "x*0.0001", Precision: 8, SortOrder: 1},
		},
	},
	// 41. 无量纲 / Dimensionless
	{
		Code: "dimensionless", Name: "无量纲", NameEn: "Dimensionless",
		Quantity: "无量纲", BaseUnitSymbol: "%", SortOrder: 410,
		Units: []SeedUnit{
			{Code: "percent", Symbol: "%", Name: "百分比", NameEn: "Percent", IsBase: true, ConversionType: ctLinear, Factor: 1, FormulaExpr: "x*1", Precision: 4},
			{Code: "permille", Symbol: "‰", Name: "千分比", NameEn: "Per mille", ConversionType: ctLinear, Factor: 0.1, FormulaExpr: "x*0.1", Precision: 4, SortOrder: 1},
			{Code: "unitless", Symbol: "—", Name: "无单位", NameEn: "Unitless", ConversionType: ctNone, Factor: 1, FormulaExpr: "不可换算", Precision: 4, SortOrder: 2},
			{Code: "power_factor", Symbol: "PF", Name: "功率因数", NameEn: "Power factor", ConversionType: ctNone, Factor: 1, FormulaExpr: "不可换算", Precision: 4, SortOrder: 3},
			{Code: "efficiency", Symbol: "η", Name: "效率", NameEn: "Efficiency", ConversionType: ctNone, Factor: 1, FormulaExpr: "不可换算", Precision: 4, SortOrder: 4},
		},
	},
	// 42. 计数 / Count
	{
		Code: "count", Name: "计数", NameEn: "Count",
		Quantity: "计数", BaseUnitSymbol: "次", SortOrder: 420,
		Units: []SeedUnit{
			{Code: "times", Symbol: "次", Name: "次", NameEn: "Times", IsBase: true, ConversionType: ctNone, Factor: 1, FormulaExpr: "x*1", Precision: 0},
			{Code: "set", Symbol: "台", Name: "台", NameEn: "Set", ConversionType: ctNone, Factor: 1, FormulaExpr: "x*1", Precision: 0, SortOrder: 1},
			{Code: "piece", Symbol: "个", Name: "个", NameEn: "Piece", ConversionType: ctNone, Factor: 1, FormulaExpr: "x*1", Precision: 0, SortOrder: 2},
			{Code: "channel", Symbol: "路", Name: "路", NameEn: "Channel", ConversionType: ctNone, Factor: 1, FormulaExpr: "x*1", Precision: 0, SortOrder: 3},
			{Code: "floor", Symbol: "层", Name: "层", NameEn: "Floor", ConversionType: ctNone, Factor: 1, FormulaExpr: "x*1", Precision: 0, SortOrder: 4},
			{Code: "door", Symbol: "门", Name: "门", NameEn: "Door", ConversionType: ctNone, Factor: 1, FormulaExpr: "x*1", Precision: 0, SortOrder: 5},
			{Code: "person", Symbol: "人", Name: "人", NameEn: "Person", ConversionType: ctNone, Factor: 1, FormulaExpr: "x*1", Precision: 0, SortOrder: 6},
			{Code: "strip", Symbol: "条", Name: "条", NameEn: "Strip", ConversionType: ctNone, Factor: 1, FormulaExpr: "x*1", Precision: 0, SortOrder: 7},
			{Code: "point", Symbol: "点", Name: "点", NameEn: "Point", ConversionType: ctNone, Factor: 1, FormulaExpr: "x*1", Precision: 0, SortOrder: 8},
			{Code: "grade", Symbol: "级", Name: "级", NameEn: "Grade", ConversionType: ctNone, Factor: 1, FormulaExpr: "x*1", Precision: 0, SortOrder: 9},
			{Code: "block", Symbol: "块", Name: "块", NameEn: "Block", ConversionType: ctNone, Factor: 1, FormulaExpr: "x*1", Precision: 0, SortOrder: 10},
		},
	},
}
