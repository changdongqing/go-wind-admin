// Package seed 提供物模型特征种子数据（幂等 upsert）。
// Package seed: thing-model feature seed data (idempotent upsert).
//
// 设计依据 / Design ref: docs/thingmodel/sheji/14-特征种子数据与实施计划.md
// 数据来源 / Data source: docs/thingmodel/物模型特征清单.md
//
// 数据覆盖：
//   - 通用属性 21（运行状态7 + 额定10 + 统计6 + 环境2）
//   - 设备专属属性 ~85（暖通28 + 给排水12 + 电气17 + 消防10 + 电梯10 + 智能化10）
//   - 额定特征 ~95（按设备大类铺开）
//   - 通用事件 8 + 设备专属事件 31
//   - 通用服务 8 + 设备专属服务 25
//   - 关系 ~6（核心示例）
package seed

import (
	thingmodelV1 "go-wind-admin/api/gen/go/thingmodel/service/v1"
)

// SeedFeature 单条特征种子定义（与设计文档 §3.1 对齐）。
// SeedFeature describes one feature row for seeding.
type SeedFeature struct {
	FeatureType     thingmodelV1.FeatureType
	Code            string
	Identifier      string
	Name            string
	NameEn          string
	Description     string
	ApplicableScope string
	SortOrder       uint32
	// 用 spec map（JSON-friendly）声明，upsert 时按类型构造 proto FeatureSpec oneof
	Spec map[string]any
}

// 缩写：特征类型常量
const (
	ftProperty = thingmodelV1.FeatureType_PROPERTY
	ftEvent    = thingmodelV1.FeatureType_EVENT
	ftService  = thingmodelV1.FeatureType_SERVICE
	ftRelation = thingmodelV1.FeatureType_RELATION
)

// ===========================================================================
// 构造辅助 / Spec construction helpers
// ===========================================================================

// pp 构造 property spec map（按 dataType / accessMode / category 组合）
// Build a property spec map; extras let callers append unit/constraints/enumItems/etc.
func pp(dataType, accessMode, category string, extras ...map[string]any) map[string]any {
	m := map[string]any{
		"dataType":   dataType,
		"accessMode": accessMode,
	}
	if category != "" {
		m["category"] = category
	}
	for _, e := range extras {
		for k, v := range e {
			m[k] = v
		}
	}
	return m
}

// uref 构造 unit ref（unitCode + unitSymbol；unitId 在 upsert 时按 code 解析）
func uref(code, symbol string) map[string]any {
	return map[string]any{"unitCode": code, "unitSymbol": symbol}
}

// cs 构造 constraints
func cs(minV, maxV float64) map[string]any {
	return map[string]any{"constraints": map[string]any{"min": minV, "max": maxV}}
}

// enums 构造 enumItems
func enums(items ...[2]any) map[string]any {
	arr := make([]map[string]any, 0, len(items))
	for _, it := range items {
		arr = append(arr, map[string]any{"value": it[0], "label": it[1]})
	}
	return map[string]any{"enumItems": arr}
}

// boolL 构造 boolLabels
func boolL(f, t string) map[string]any {
	return map[string]any{"boolLabels": map[string]any{"false": f, "true": t}}
}

// txtLen 构造 textMaxLength
func txtLen(n int) map[string]any {
	return map[string]any{"textMaxLength": n}
}

// rated 标记为额定铭牌（category=rated + isRated=true）
func rated() map[string]any {
	return map[string]any{"category": "rated", "isRated": true}
}

// strFields 构造 struct 子字段
func strFields(fields ...map[string]any) map[string]any {
	return map[string]any{"structFields": fields}
}

// fStruct 单个 struct 子字段
func fStruct(key, dataType string, unitMap map[string]any, extra ...map[string]any) map[string]any {
	m := map[string]any{"key": key, "dataType": dataType}
	if unitMap != nil {
		m["unit"] = unitMap
	}
	for _, e := range extra {
		for k, v := range e {
			m[k] = v
		}
	}
	return m
}

// ev 构造 event spec map
func ev(level string, outputs ...map[string]any) map[string]any {
	return map[string]any{
		"level":        level,
		"outputParams": outputs,
	}
}

// svc 构造 service spec map
func svc(callMode string, inputs, outputs []map[string]any) map[string]any {
	m := map[string]any{"callMode": callMode}
	if inputs != nil {
		m["inputParams"] = inputs
	}
	if outputs != nil {
		m["outputParams"] = outputs
	}
	return m
}

// pm 构造参数（事件输出/服务输入输出/struct 子字段通用）
func pm(key, dataType string, extras ...map[string]any) map[string]any {
	m := map[string]any{"key": key, "dataType": dataType}
	for _, e := range extras {
		for k, v := range e {
			m[k] = v
		}
	}
	return m
}

// rel 构造 relation spec
func rel(relationType, cardinality string, directional bool, source, target map[string]any) map[string]any {
	return map[string]any{
		"relationType": relationType,
		"cardinality":  cardinality,
		"directional":  directional,
		"source":       source,
		"target":       target,
	}
}

// refFeature 构造关系实体引用（kind=feature，用 identifier 解析 id）
func refFeature(identifier string) map[string]any {
	return map[string]any{"kind": "feature", "identifier": identifier}
}

// ===========================================================================
// 通用属性 / Common properties
// ===========================================================================

var commonProperties = []SeedFeature{
	// 2.1 运行状态
	{ftProperty, "P-RUN-0001", "powerSwitch", "开关状态", "Power Switch", "设备开关", "通用", 101,
		pp("BOOL", "RW", "runtime", boolL("关", "开"))},
	{ftProperty, "P-RUN-0002", "runMode", "运行模式", "Run Mode", "运行模式选择", "通用", 102,
		pp("ENUM", "RW", "runtime", enums([2]any{0, "停机"}, [2]any{1, "手动"}, [2]any{2, "自动"}, [2]any{3, "远程"}))},
	{ftProperty, "P-RUN-0003", "runStatus", "运行状态", "Run Status", "设备当前运行状态", "通用", 103,
		pp("ENUM", "R", "runtime", enums([2]any{0, "停止"}, [2]any{1, "运行"}, [2]any{2, "待机"}, [2]any{3, "故障"}))},
	{ftProperty, "P-RUN-0004", "faultCode", "故障代码", "Fault Code", "0=无故障，其他=故障代码", "通用", 104,
		pp("INT", "R", "runtime")},
	{ftProperty, "P-RUN-0005", "faultDesc", "故障描述", "Fault Description", "故障文本描述", "通用", 105,
		pp("TEXT", "R", "runtime", txtLen(256))},
	{ftProperty, "P-RUN-0006", "localRemote", "本地/远程", "Local/Remote", "控制权位置", "通用", 106,
		pp("ENUM", "R", "runtime", enums([2]any{0, "本地"}, [2]any{1, "远程"}))},
	{ftProperty, "P-RUN-0007", "autoManual", "手自动", "Auto/Manual", "手/自动模式", "通用", 107,
		pp("ENUM", "RW", "runtime", enums([2]any{0, "手动"}, [2]any{1, "自动"}))},

	// 2.2 通用额定
	{ftProperty, "P-RAT-0001", "model", "设备型号", "Model", "出厂型号", "通用", 201,
		pp("TEXT", "R", "rated", txtLen(128), map[string]any{"isRated": true})},
	{ftProperty, "P-RAT-0002", "manufacturer", "制造商", "Manufacturer", "设备生产厂家", "通用", 202,
		pp("TEXT", "R", "rated", txtLen(128), map[string]any{"isRated": true})},
	{ftProperty, "P-RAT-0003", "manufactureDate", "出厂日期", "Manufacture Date", "出厂日期", "通用", 203,
		pp("DATE", "R", "rated", map[string]any{"isRated": true})},
	{ftProperty, "P-RAT-0004", "serialNumber", "序列号", "Serial Number", "出厂序列号", "通用", 204,
		pp("TEXT", "R", "rated", txtLen(64), map[string]any{"isRated": true})},
	{ftProperty, "P-RAT-0005", "ratedVoltage", "额定电压", "Rated Voltage", "额定电压", "通用", 205,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("volt", "V"), "isRated": true}, cs(0, 10000))},
	{ftProperty, "P-RAT-0006", "ratedCurrent", "额定电流", "Rated Current", "额定电流", "通用", 206,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("ampere", "A"), "isRated": true}, cs(0, 10000))},
	{ftProperty, "P-RAT-0007", "ratedPower", "额定功率", "Rated Power", "额定功率", "通用", 207,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("kilowatt", "kW"), "isRated": true}, cs(0, 100000))},
	{ftProperty, "P-RAT-0008", "ratedFrequency", "额定频率", "Rated Frequency", "额定频率", "通用", 208,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("hertz", "Hz"), "isRated": true}, cs(0, 100))},
	{ftProperty, "P-RAT-0009", "ipRating", "防护等级", "IP Rating", "如 IP42/IP54", "通用", 209,
		pp("TEXT", "R", "rated", txtLen(16), map[string]any{"isRated": true})},
	{ftProperty, "P-RAT-0010", "insulationClass", "绝缘等级", "Insulation Class", "绝缘耐热等级", "通用", 210,
		pp("ENUM", "R", "rated", map[string]any{"isRated": true},
			enums([2]any{0, "A级"}, [2]any{1, "E级"}, [2]any{2, "B级"}, [2]any{3, "F级"}, [2]any{4, "H级"}, [2]any{5, "C级"}))},

	// 2.3 通用统计
	{ftProperty, "P-STA-0001", "totalRunTime", "累计运行时间", "Total Run Time", "累计运行小时数", "通用", 301,
		pp("DOUBLE", "R", "statistic", map[string]any{"unit": uref("hour", "h")})},
	{ftProperty, "P-STA-0002", "totalEnergy", "累计能耗", "Total Energy", "累计耗电量", "通用", 302,
		pp("DOUBLE", "R", "statistic", map[string]any{"unit": uref("kilowatt_hour", "kWh")})},
	{ftProperty, "P-STA-0003", "startCount", "启动次数", "Start Count", "累计启动次数", "通用", 303,
		pp("INT", "R", "statistic")},
	{ftProperty, "P-STA-0004", "faultCount", "故障次数", "Fault Count", "累计故障次数", "通用", 304,
		pp("INT", "R", "statistic")},
	{ftProperty, "P-STA-0005", "todayRunTime", "今日运行时间", "Today Run Time", "当日累计运行小时", "通用", 305,
		pp("DOUBLE", "R", "statistic", map[string]any{"unit": uref("hour", "h")}, cs(0, 24))},
	{ftProperty, "P-STA-0006", "todayEnergy", "今日能耗", "Today Energy", "当日累计能耗", "通用", 306,
		pp("DOUBLE", "R", "statistic", map[string]any{"unit": uref("kilowatt_hour", "kWh")})},

	// 2.4 通用环境
	{ftProperty, "P-ENV-0001", "ambientTemp", "环境温度", "Ambient Temp", "设备环境温度", "通用", 401,
		pp("DOUBLE", "R", "environment", map[string]any{"unit": uref("celsius", "℃")}, cs(-40, 80))},
	{ftProperty, "P-ENV-0002", "ambientHumidity", "环境湿度", "Ambient Humidity", "设备环境湿度", "通用", 402,
		pp("DOUBLE", "R", "environment", map[string]any{"unit": uref("percent_rh", "%RH")}, cs(0, 100))},
}

// ===========================================================================
// 通用事件 / Common events
// ===========================================================================

var commonEvents = []SeedFeature{
	{ftEvent, "E-GEN-0001", "deviceOnline", "设备上线", "Device Online", "设备连接上线", "通用", 1,
		ev("INFO", pm("timestamp", "DATE"))},
	{ftEvent, "E-GEN-0002", "deviceOffline", "设备离线", "Device Offline", "设备断开连接", "通用", 2,
		ev("INFO", pm("timestamp", "DATE"), pm("offlineDuration", "DOUBLE", map[string]any{"unit": uref("second", "s")}))},
	{ftEvent, "E-GEN-0003", "runStatusChanged", "运行状态变更", "Run Status Changed", "运行状态切换", "通用", 3,
		ev("INFO", pm("oldStatus", "ENUM"), pm("newStatus", "ENUM"))},
	{ftEvent, "E-GEN-0004", "parameterOverLimit", "参数越限告警", "Parameter Over Limit", "测量参数超出设定阈值", "通用", 4,
		ev("ALERT",
			pm("parameterName", "TEXT"),
			pm("currentValue", "DOUBLE"),
			pm("limitValue", "DOUBLE"),
			pm("unit", "TEXT"))},
	{ftEvent, "E-GEN-0005", "communicationError", "通信异常告警", "Communication Error", "通信链路异常", "通用", 5,
		ev("ALERT", pm("errorType", "ENUM"), pm("detail", "TEXT"))},
	{ftEvent, "E-GEN-0006", "deviceFault", "设备故障", "Device Fault", "设备发生故障", "通用", 6,
		ev("ERROR", pm("faultCode", "INT"), pm("faultDesc", "TEXT"), pm("faultTime", "DATE"))},
	{ftEvent, "E-GEN-0007", "faultRecovery", "故障恢复", "Fault Recovery", "故障已恢复", "通用", 7,
		ev("INFO", pm("faultCode", "INT"), pm("faultDuration", "DOUBLE", map[string]any{"unit": uref("minute", "min")}))},
	{ftEvent, "E-GEN-0008", "maintenanceReminder", "维保提醒", "Maintenance Reminder", "设备需要维护保养", "通用", 8,
		ev("INFO", pm("maintenanceType", "ENUM"), pm("remainingTime", "DOUBLE", map[string]any{"unit": uref("hour", "h")}))},
}

// ===========================================================================
// 通用服务 / Common services
// ===========================================================================

var commonServices = []SeedFeature{
	{ftService, "S-GEN-0001", "turnOn", "开机", "Turn On", "设备开机", "通用", 1,
		svc("ASYNC", nil, []map[string]any{pm("success", "BOOL")})},
	{ftService, "S-GEN-0002", "turnOff", "关机", "Turn Off", "设备关机", "通用", 2,
		svc("ASYNC", nil, []map[string]any{pm("success", "BOOL")})},
	{ftService, "S-GEN-0003", "restart", "重启", "Restart", "延时重启设备", "通用", 3,
		svc("ASYNC",
			[]map[string]any{pm("delay", "DOUBLE", map[string]any{"unit": uref("second", "s")}, cs(0, 3600))},
			[]map[string]any{pm("success", "BOOL")})},
	{ftService, "S-GEN-0004", "setRunMode", "设置运行模式", "Set Run Mode", "切换运行模式", "通用", 4,
		svc("ASYNC", []map[string]any{pm("runMode", "ENUM")}, []map[string]any{pm("success", "BOOL")})},
	{ftService, "S-GEN-0005", "faultReset", "故障复位", "Fault Reset", "清除故障状态", "通用", 5,
		svc("ASYNC", nil, []map[string]any{pm("success", "BOOL")})},
	{ftService, "S-GEN-0006", "getProperty", "读取属性", "Get Property", "同步读取指定属性当前值", "通用", 6,
		svc("SYNC",
			[]map[string]any{pm("propertyIdentifier", "TEXT")},
			[]map[string]any{pm("propertyValue", "STRUCT")})},
	{ftService, "S-GEN-0007", "setProperty", "设置属性", "Set Property", "设置指定属性值", "通用", 7,
		svc("ASYNC",
			[]map[string]any{pm("propertyIdentifier", "TEXT"), pm("propertyValue", "STRUCT")},
			[]map[string]any{pm("success", "BOOL")})},
	{ftService, "S-GEN-0008", "queryHistory", "查询历史数据", "Query History", "查询指定时间段历史数据", "通用", 8,
		svc("SYNC",
			[]map[string]any{pm("startTime", "DATE"), pm("endTime", "DATE"), pm("propertyIdentifiers", "ARRAY")},
			[]map[string]any{pm("historyData", "ARRAY")})},
}
