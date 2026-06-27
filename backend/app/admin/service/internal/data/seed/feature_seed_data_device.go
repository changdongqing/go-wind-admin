// 设备专属特征种子数据（属性 / 事件 / 服务 / 关系）
// Device-specific features (property/event/service/relation) seed data.
//
// 拆分依据 / Split rationale: 通用部分见 feature_seed_data.go；本文件按设备大类组织。

package seed

// ===========================================================================
// 6.1 暖通空调（HVAC）专属属性 / HVAC properties
// ===========================================================================

var hvacProperties = []SeedFeature{
	{ftProperty, "P-HVAC-0001", "supplyWaterTemp", "供水温度", "Supply Water Temp", "供水温度", "冷机/锅炉/换热器", 6001,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("celsius", "℃")}, cs(-40, 200))},
	{ftProperty, "P-HVAC-0002", "returnWaterTemp", "回水温度", "Return Water Temp", "回水温度", "冷机/锅炉/换热器", 6002,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("celsius", "℃")}, cs(-40, 200))},
	{ftProperty, "P-HVAC-0003", "supplyWaterPressure", "供水压力", "Supply Water Pressure", "供水压力", "冷机/水泵", 6003,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("megapascal", "MPa")}, cs(0, 10))},
	{ftProperty, "P-HVAC-0004", "returnWaterPressure", "回水压力", "Return Water Pressure", "回水压力", "冷机/水泵", 6004,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("megapascal", "MPa")}, cs(0, 10))},
	{ftProperty, "P-HVAC-0005", "waterFlowRate", "水流量", "Water Flow Rate", "水流量", "冷机/水泵", 6005,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("cubic_meter_per_hour", "m³/h")}, cs(0, 10000))},
	{ftProperty, "P-HVAC-0006", "operationMode", "供冷/供热模式", "Operation Mode", "运行模式", "冷机/热泵/空调末端", 6006,
		pp("ENUM", "RW", "setting", enums([2]any{0, "制冷"}, [2]any{1, "制热"}, [2]any{2, "通风"}, [2]any{3, "除湿"}, [2]any{4, "自动"}))},
	{ftProperty, "P-HVAC-0007", "setTemperature", "设定温度", "Set Temperature", "设定温度", "空调末端/新风机组", 6007,
		pp("DOUBLE", "RW", "setting", map[string]any{"unit": uref("celsius", "℃")}, cs(-20, 60))},
	{ftProperty, "P-HVAC-0008", "supplyAirTemp", "送风温度", "Supply Air Temp", "送风温度", "空调末端/新风机组", 6008,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("celsius", "℃")}, cs(-40, 80))},
	{ftProperty, "P-HVAC-0009", "returnAirTemp", "回风温度", "Return Air Temp", "回风温度", "空调末端", 6009,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("celsius", "℃")}, cs(-40, 80))},
	{ftProperty, "P-HVAC-0010", "supplyAirHumidity", "送风湿度", "Supply Air Humidity", "送风湿度", "空调末端/新风机组", 6010,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("percent_rh", "%RH")}, cs(0, 100))},
	{ftProperty, "P-HVAC-0011", "damperPosition", "风阀开度", "Damper Position", "风阀开度", "通风机/空调末端", 6011,
		pp("DOUBLE", "RW", "setting", map[string]any{"unit": uref("percent", "%")}, cs(0, 100))},
	{ftProperty, "P-HVAC-0012", "valvePosition", "水阀开度", "Valve Position", "水阀开度", "空调末端/换热器", 6012,
		pp("DOUBLE", "RW", "setting", map[string]any{"unit": uref("percent", "%")}, cs(0, 100))},
	{ftProperty, "P-HVAC-0013", "fanFrequency", "风机频率", "Fan Frequency", "变频风机频率", "通风机/空调末端", 6013,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("hertz", "Hz")}, cs(0, 100))},
	{ftProperty, "P-HVAC-0014", "fanSpeed", "风机转速", "Fan Speed", "风机转速", "通风机", 6014,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("rpm", "rpm")}, cs(0, 30000))},
	{ftProperty, "P-HVAC-0015", "filterPressureDiff", "过滤器压差", "Filter Pressure Diff", "过滤器压差", "通风机/新风机组", 6015,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("pascal", "Pa")}, cs(0, 2000))},
	{ftProperty, "P-HVAC-0016", "compressorStatus", "压缩机运行状态", "Compressor Status", "压缩机状态", "制冷机组", 6016,
		pp("ENUM", "R", "runtime", enums([2]any{0, "停止"}, [2]any{1, "运行"}, [2]any{2, "卸载"}, [2]any{3, "故障"}))},
	{ftProperty, "P-HVAC-0017", "compressorLoadRate", "压缩机负载率", "Compressor Load Rate", "压缩机负载率", "制冷机组", 6017,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("percent", "%")}, cs(0, 100))},
	{ftProperty, "P-HVAC-0018", "evaporatingTemp", "蒸发温度", "Evaporating Temp", "蒸发温度", "制冷机组", 6018,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("celsius", "℃")}, cs(-50, 30))},
	{ftProperty, "P-HVAC-0019", "condensingTemp", "冷凝温度", "Condensing Temp", "冷凝温度", "制冷机组", 6019,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("celsius", "℃")}, cs(0, 80))},
	{ftProperty, "P-HVAC-0020", "condenserInletTemp", "冷却水进水温度", "Condenser Inlet Temp", "冷却水进水温度", "制冷机组/冷却塔", 6020,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("celsius", "℃")}, cs(0, 60))},
	{ftProperty, "P-HVAC-0021", "condenserOutletTemp", "冷却水出水温度", "Condenser Outlet Temp", "冷却水出水温度", "制冷机组/冷却塔", 6021,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("celsius", "℃")}, cs(0, 50))},
	{ftProperty, "P-HVAC-0022", "currentCoolingCapacity", "实时制冷量", "Current Cooling Capacity", "实时制冷量", "制冷机组", 6022,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("kilowatt", "kW")}, cs(0, 50000))},
	{ftProperty, "P-HVAC-0023", "currentCOP", "实时COP", "Current COP", "实时能效", "制冷机组", 6023,
		pp("DOUBLE", "R", "measurement", cs(0, 15))},
	{ftProperty, "P-HVAC-0024", "currentPower", "实时功率", "Current Power", "实时功率", "通用", 6024,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("kilowatt", "kW")}, cs(0, 100000))},
	{ftProperty, "P-HVAC-0025", "freshAirTemp", "新风温度", "Fresh Air Temp", "新风温度", "新风机组", 6025,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("celsius", "℃")}, cs(-40, 60))},
	{ftProperty, "P-HVAC-0026", "freshAirHumidity", "新风湿度", "Fresh Air Humidity", "新风湿度", "新风机组", 6026,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("percent_rh", "%RH")}, cs(0, 100))},
	{ftProperty, "P-HVAC-0027", "co2Concentration", "CO2浓度", "CO2 Concentration", "二氧化碳浓度", "新风机组/空调末端", 6027,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("ppm", "ppm")}, cs(0, 5000))},
	{ftProperty, "P-HVAC-0028", "setFanSpeed", "设定风速", "Set Fan Speed", "设定风速档位", "空调末端", 6028,
		pp("ENUM", "RW", "setting", enums([2]any{0, "低速"}, [2]any{1, "中速"}, [2]any{2, "高速"}, [2]any{3, "自动"}))},
}

// ===========================================================================
// 6.2 给排水（WP）专属属性
// ===========================================================================

var wpProperties = []SeedFeature{
	{ftProperty, "P-WP-0001", "waterLevel", "水箱液位", "Water Level", "水箱液位", "水箱/水池", 6201,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("meter", "m")}, cs(0, 20))},
	{ftProperty, "P-WP-0002", "waterLevelPercent", "液位百分比", "Water Level Percent", "液位百分比", "水箱/水池", 6202,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("percent", "%")}, cs(0, 100))},
	{ftProperty, "P-WP-0003", "pipePressure", "管道压力", "Pipe Pressure", "管道压力", "水泵/管网", 6203,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("megapascal", "MPa")}, cs(0, 10))},
	{ftProperty, "P-WP-0004", "instantFlowRate", "瞬时流量", "Instant Flow Rate", "瞬时流量", "水泵", 6204,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("cubic_meter_per_hour", "m³/h")}, cs(0, 10000))},
	{ftProperty, "P-WP-0005", "totalFlowVolume", "累计流量", "Total Flow Volume", "累计流量", "水泵", 6205,
		pp("DOUBLE", "R", "statistic", map[string]any{"unit": uref("cubic_meter", "m³")})},
	{ftProperty, "P-WP-0006", "pumpFrequency", "水泵频率", "Pump Frequency", "变频水泵频率", "变频水泵", 6206,
		pp("DOUBLE", "RW", "setting", map[string]any{"unit": uref("hertz", "Hz")}, cs(0, 100))},
	{ftProperty, "P-WP-0007", "pumpSpeed", "水泵转速", "Pump Speed", "水泵转速", "水泵", 6207,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("rpm", "rpm")}, cs(0, 30000))},
	{ftProperty, "P-WP-0008", "outletWaterTemp", "出水温度", "Outlet Water Temp", "热水出水温度", "热水设备", 6208,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("celsius", "℃")}, cs(0, 100))},
	{ftProperty, "P-WP-0009", "setWaterTemp", "设定水温", "Set Water Temp", "设定热水温度", "热水设备", 6209,
		pp("DOUBLE", "RW", "setting", map[string]any{"unit": uref("celsius", "℃")}, cs(20, 80))},
	{ftProperty, "P-WP-0010", "chlorineConcentration", "余氯浓度", "Chlorine Concentration", "余氯浓度", "水处理设备", 6210,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("mg_per_l", "mg/L")}, cs(0, 10))},
	{ftProperty, "P-WP-0011", "turbidity", "浊度", "Turbidity", "浊度", "水处理设备", 6211,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("ntu", "NTU")}, cs(0, 100))},
	{ftProperty, "P-WP-0012", "phValue", "pH值", "pH Value", "酸碱度", "水处理设备", 6212,
		pp("DOUBLE", "R", "measurement", cs(0, 14))},
}

// ===========================================================================
// 6.3 电气（EL）专属属性
// ===========================================================================

var elProperties = []SeedFeature{
	{ftProperty, "P-EL-0001", "voltage", "电压", "Voltage", "工作电压", "变压器/配电柜", 6301,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("volt", "V")}, cs(0, 10000))},
	{ftProperty, "P-EL-0002", "current", "电流", "Current", "工作电流", "变压器/配电柜", 6302,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("ampere", "A")}, cs(0, 100000))},
	{ftProperty, "P-EL-0003", "activePower", "有功功率", "Active Power", "有功功率", "变压器/配电柜", 6303,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("kilowatt", "kW")}, cs(-100000, 100000))},
	{ftProperty, "P-EL-0004", "reactivePower", "无功功率", "Reactive Power", "无功功率", "变压器/配电柜", 6304,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("kilovar", "kvar")}, cs(-100000, 100000))},
	{ftProperty, "P-EL-0005", "powerFactor", "功率因数", "Power Factor", "功率因数", "变压器/配电柜", 6305,
		pp("DOUBLE", "R", "measurement", cs(-1, 1))},
	{ftProperty, "P-EL-0006", "frequency", "频率", "Frequency", "工作频率", "变压器/发电机组", 6306,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("hertz", "Hz")}, cs(0, 100))},
	{ftProperty, "P-EL-0007", "activeEnergy", "有功电能", "Active Energy", "有功电能", "电表", 6307,
		pp("DOUBLE", "R", "statistic", map[string]any{"unit": uref("kilowatt_hour", "kWh")})},
	{ftProperty, "P-EL-0008", "reactiveEnergy", "无功电能", "Reactive Energy", "无功电能", "电表", 6308,
		pp("DOUBLE", "R", "statistic", map[string]any{"unit": uref("kvarh", "kvarh")})},
	{ftProperty, "P-EL-0009", "transformerTemp", "变压器温度", "Transformer Temp", "变压器温度", "变压器", 6309,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("celsius", "℃")}, cs(-40, 200))},
	{ftProperty, "P-EL-0010", "loadRate", "负载率", "Load Rate", "负载率", "变压器", 6310,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("percent", "%")}, cs(0, 150))},
	{ftProperty, "P-EL-0011", "batteryVoltage", "电池电压", "Battery Voltage", "电池电压", "UPS/EPS", 6311,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("volt", "V")}, cs(0, 1000))},
	{ftProperty, "P-EL-0012", "batteryRemaining", "电池剩余容量", "Battery Remaining", "电池剩余容量", "UPS/EPS", 6312,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("percent", "%")}, cs(0, 100))},
	{ftProperty, "P-EL-0013", "remainingBackupTime", "剩余后备时间", "Remaining Backup Time", "剩余后备时间", "UPS/EPS", 6313,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("minute", "min")}, cs(0, 1440))},
	{ftProperty, "P-EL-0014", "mainsStatus", "市电状态", "Mains Status", "市电状态", "UPS/EPS/发电机组", 6314,
		pp("ENUM", "R", "runtime", enums([2]any{0, "正常"}, [2]any{1, "异常"}, [2]any{2, "中断"}))},
	{ftProperty, "P-EL-0015", "switchStatus", "开关状态", "Switch Status", "断路器开关状态", "配电柜/箱", 6315,
		pp("BOOL", "R", "runtime", boolL("分闸", "合闸"))},
	{ftProperty, "P-EL-0016", "illuminance", "照度", "Illuminance", "照度", "照明设备", 6316,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("lux", "Lux")}, cs(0, 100000))},
	{ftProperty, "P-EL-0017", "dimmingLevel", "调光级别", "Dimming Level", "调光级别", "照明设备", 6317,
		pp("DOUBLE", "RW", "setting", map[string]any{"unit": uref("percent", "%")}, cs(0, 100))},
}

// ===========================================================================
// 6.4 消防（FP）专属属性
// ===========================================================================

var fpProperties = []SeedFeature{
	// identifier 使用 firePipePressure 以避免与给排水 property P-WP-0003 的 pipePressure 冲突
	// （二者语义不同：消防管网 vs 给排水管网，identifier 区分更清晰）。
	{ftProperty, "P-FP-0001", "firePipePressure", "管网压力", "Pipe Pressure", "消防管网压力", "消防水泵/消火栓", 6401,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("megapascal", "MPa")}, cs(0, 5))},
	{ftProperty, "P-FP-0002", "fireWaterLevel", "消防水池液位", "Fire Water Level", "消防水池液位", "消防水箱", 6402,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("meter", "m")}, cs(0, 10))},
	{ftProperty, "P-FP-0003", "alarmStatus", "报警状态", "Alarm Status", "火警状态", "火灾报警系统", 6403,
		pp("ENUM", "R", "runtime", enums([2]any{0, "正常"}, [2]any{1, "火警"}, [2]any{2, "预警"}, [2]any{3, "故障"}))},
	{ftProperty, "P-FP-0004", "alarmZone", "报警区域", "Alarm Zone", "报警区域", "火灾报警系统", 6404,
		pp("TEXT", "R", "runtime", txtLen(128))},
	{ftProperty, "P-FP-0005", "sprinklerStatus", "喷头状态", "Sprinkler Status", "喷头状态", "自动喷水系统", 6405,
		pp("ENUM", "R", "runtime", enums([2]any{0, "正常"}, [2]any{1, "动作"}, [2]any{2, "故障"}))},
	{ftProperty, "P-FP-0006", "smokeValveStatus", "排烟阀状态", "Smoke Valve Status", "排烟阀状态", "防排烟系统", 6406,
		pp("ENUM", "R", "runtime", enums([2]any{0, "关闭"}, [2]any{1, "开启"}, [2]any{2, "故障"}))},
	{ftProperty, "P-FP-0007", "fireDoorStatus", "防火门状态", "Fire Door Status", "防火门状态", "防火门", 6407,
		pp("ENUM", "R", "runtime", enums([2]any{0, "关闭"}, [2]any{1, "开启"}, [2]any{2, "故障"}))},
	{ftProperty, "P-FP-0008", "fireShutterPosition", "防火卷帘位置", "Fire Shutter Position", "防火卷帘位置", "防火卷帘", 6408,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("percent", "%")}, cs(0, 100))},
	{ftProperty, "P-FP-0009", "gasSuppressionStatus", "气体灭火状态", "Gas Suppression Status", "气体灭火状态", "气体灭火系统", 6409,
		pp("ENUM", "R", "runtime", enums([2]any{0, "正常"}, [2]any{1, "喷放"}, [2]any{2, "故障"}))},
	{ftProperty, "P-FP-0010", "selfTestStatus", "系统自检状态", "Self Test Status", "系统自检状态", "火灾报警系统", 6410,
		pp("ENUM", "R", "runtime", enums([2]any{0, "正常"}, [2]any{1, "自检中"}, [2]any{2, "自检异常"}))},
}

// ===========================================================================
// 6.5 电梯（ELV）专属属性
// ===========================================================================

var elvProperties = []SeedFeature{
	{ftProperty, "P-ELV-0001", "currentFloor", "当前楼层", "Current Floor", "电梯当前楼层", "电梯", 6501,
		pp("INT", "R", "measurement", cs(-10, 200))},
	{ftProperty, "P-ELV-0002", "travelDirection", "运行方向", "Travel Direction", "电梯运行方向", "电梯", 6502,
		pp("ENUM", "R", "runtime", enums([2]any{0, "停止"}, [2]any{1, "上行"}, [2]any{2, "下行"}))},
	{ftProperty, "P-ELV-0003", "carPassengerCount", "轿厢内人数", "Car Passenger Count", "轿厢内人数", "电梯", 6503,
		pp("INT", "R", "measurement", cs(0, 50))},
	{ftProperty, "P-ELV-0004", "currentLoad", "当前载重", "Current Load", "当前载重", "电梯", 6504,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("kilogram", "kg")}, cs(0, 20000))},
	{ftProperty, "P-ELV-0005", "loadPercentage", "载重百分比", "Load Percentage", "载重百分比", "电梯", 6505,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("percent", "%")}, cs(0, 110))},
	{ftProperty, "P-ELV-0006", "doorStatus", "门状态", "Door Status", "门状态", "电梯", 6506,
		pp("ENUM", "R", "runtime", enums([2]any{0, "关门"}, [2]any{1, "开门"}, [2]any{2, "开门中"}, [2]any{3, "关门中"}))},
	{ftProperty, "P-ELV-0007", "currentSpeed", "运行速度", "Current Speed", "运行速度", "电梯", 6507,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("meter_per_second", "m/s")}, cs(0, 20))},
	{ftProperty, "P-ELV-0008", "lockStatus", "锁梯状态", "Lock Status", "锁梯状态", "电梯", 6508,
		pp("BOOL", "RW", "setting", boolL("解锁", "锁梯"))},
	{ftProperty, "P-ELV-0009", "fireReturnStatus", "消防返回状态", "Fire Return Status", "消防返回状态", "电梯", 6509,
		pp("BOOL", "R", "runtime", boolL("正常", "消防返回"))},
	{ftProperty, "P-ELV-0010", "escalatorDirection", "扶梯运行方向", "Escalator Direction", "扶梯运行方向", "自动扶梯", 6510,
		pp("ENUM", "RW", "setting", enums([2]any{0, "停止"}, [2]any{1, "上行"}, [2]any{2, "下行"}))},
}

// ===========================================================================
// 6.6 智能化（SM）专属属性
// ===========================================================================

var smProperties = []SeedFeature{
	{ftProperty, "P-SM-0001", "onlineStatus", "在线状态", "Online Status", "在线状态", "摄像机/门禁/传感器", 6601,
		pp("BOOL", "R", "runtime", boolL("离线", "在线"))},
	{ftProperty, "P-SM-0002", "signalStrength", "信号强度", "Signal Strength", "信号强度", "无线设备", 6602,
		pp("INT", "R", "measurement", map[string]any{"unit": uref("dbm", "dBm")}, cs(-120, 0))},
	{ftProperty, "P-SM-0003", "cpuUsage", "CPU使用率", "CPU Usage", "CPU使用率", "控制器/服务器", 6603,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("percent", "%")}, cs(0, 100))},
	{ftProperty, "P-SM-0004", "memoryUsage", "内存使用率", "Memory Usage", "内存使用率", "控制器/服务器", 6604,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("percent", "%")}, cs(0, 100))},
	{ftProperty, "P-SM-0005", "storageUsage", "存储使用率", "Storage Usage", "存储使用率", "NVR/服务器", 6605,
		pp("DOUBLE", "R", "measurement", map[string]any{"unit": uref("percent", "%")}, cs(0, 100))},
	{ftProperty, "P-SM-0006", "smDoorStatus", "门状态", "Door Status", "门禁门状态", "门禁", 6606,
		pp("ENUM", "R", "runtime", enums([2]any{0, "关闭"}, [2]any{1, "开启"}, [2]any{2, "未关"}, [2]any{3, "超时未关"}))},
	{ftProperty, "P-SM-0007", "armStatus", "布防状态", "Arm Status", "入侵报警布防状态", "入侵报警", 6607,
		pp("ENUM", "RW", "setting", enums([2]any{0, "撤防"}, [2]any{1, "外出布防"}, [2]any{2, "留守布防"}))},
	{ftProperty, "P-SM-0008", "measuredValue", "测量值", "Measured Value", "传感器测量值", "传感器", 6608,
		pp("DOUBLE", "R", "measurement")},
	{ftProperty, "P-SM-0009", "controlOutput", "控制输出值", "Control Output", "控制器输出", "控制器", 6609,
		pp("DOUBLE", "RW", "setting", cs(0, 100))},
	{ftProperty, "P-SM-0010", "networkStatus", "网络连接状态", "Network Status", "网络状态", "通信设备", 6610,
		pp("ENUM", "R", "runtime", enums([2]any{0, "断开"}, [2]any{1, "已连接"}, [2]any{2, "连接中"}))},
}

// ===========================================================================
// 设备专属事件 / Device-specific events
// ===========================================================================

var deviceEvents = []SeedFeature{
	// 7.1 HVAC
	{ftEvent, "E-HVAC-0001", "supplyTempHigh", "供水温度超高", "Supply Temp High", "供水温度超过设定上限", "冷机/锅炉", 7101,
		ev("ALERT", pm("supplyTemp", "DOUBLE", map[string]any{"unit": uref("celsius", "℃")}), pm("limitValue", "DOUBLE", map[string]any{"unit": uref("celsius", "℃")}))},
	{ftEvent, "E-HVAC-0002", "supplyTempLow", "供水温度超低", "Supply Temp Low", "供水温度低于设定下限", "冷机/锅炉", 7102,
		ev("ALERT", pm("supplyTemp", "DOUBLE", map[string]any{"unit": uref("celsius", "℃")}), pm("limitValue", "DOUBLE", map[string]any{"unit": uref("celsius", "℃")}))},
	{ftEvent, "E-HVAC-0003", "compressorHighPressure", "压缩机高压保护", "Compressor High Pressure", "压缩机排气压力过高", "制冷机组", 7103,
		ev("ERROR", pm("pressure", "DOUBLE", map[string]any{"unit": uref("megapascal", "MPa")}), pm("limitValue", "DOUBLE", map[string]any{"unit": uref("megapascal", "MPa")}))},
	{ftEvent, "E-HVAC-0004", "compressorLowPressure", "压缩机低压保护", "Compressor Low Pressure", "压缩机吸气压力过低", "制冷机组", 7104,
		ev("ERROR", pm("pressure", "DOUBLE", map[string]any{"unit": uref("megapascal", "MPa")}), pm("limitValue", "DOUBLE", map[string]any{"unit": uref("megapascal", "MPa")}))},
	{ftEvent, "E-HVAC-0005", "freezeProtection", "防冻保护", "Freeze Protection", "水温接近冰点", "制冷机组/空调末端", 7105,
		ev("ALERT", pm("waterTemp", "DOUBLE", map[string]any{"unit": uref("celsius", "℃")}))},
	{ftEvent, "E-HVAC-0006", "filterClogged", "过滤器堵塞", "Filter Clogged", "过滤器压差过大", "通风机/新风机组", 7106,
		ev("ALERT", pm("pressureDiff", "DOUBLE", map[string]any{"unit": uref("pascal", "Pa")}), pm("limitValue", "DOUBLE", map[string]any{"unit": uref("pascal", "Pa")}))},
	{ftEvent, "E-HVAC-0007", "waterFlowSwitchOff", "水流开关断开", "Water Flow Switch Off", "水流中断保护", "制冷机组/空调末端", 7107,
		ev("ERROR")},
	{ftEvent, "E-HVAC-0008", "coolingTowerFanFault", "冷却塔风机故障", "Cooling Tower Fan Fault", "冷却塔风机异常", "冷却塔", 7108,
		ev("ERROR", pm("fanIndex", "INT"), pm("faultDesc", "TEXT"))},

	// 7.2 WP
	{ftEvent, "E-WP-0001", "highWaterLevel", "高液位告警", "High Water Level", "液位超过高限", "水箱/水池", 7201,
		ev("ALERT", pm("waterLevel", "DOUBLE", map[string]any{"unit": uref("meter", "m")}), pm("limitValue", "DOUBLE", map[string]any{"unit": uref("meter", "m")}))},
	{ftEvent, "E-WP-0002", "lowWaterLevel", "低液位告警", "Low Water Level", "液位低于低限", "水箱/水池", 7202,
		ev("ALERT", pm("waterLevel", "DOUBLE", map[string]any{"unit": uref("meter", "m")}), pm("limitValue", "DOUBLE", map[string]any{"unit": uref("meter", "m")}))},
	{ftEvent, "E-WP-0003", "overflowAlarm", "溢流告警", "Overflow Alarm", "液位溢流", "水箱/水池", 7203,
		ev("ERROR", pm("waterLevel", "DOUBLE", map[string]any{"unit": uref("meter", "m")}))},
	{ftEvent, "E-WP-0004", "pipeOverPressure", "管网超压", "Pipe Over Pressure", "管道压力超高", "水泵/管网", 7204,
		ev("ALERT", pm("pressure", "DOUBLE", map[string]any{"unit": uref("megapascal", "MPa")}), pm("limitValue", "DOUBLE", map[string]any{"unit": uref("megapascal", "MPa")}))},
	{ftEvent, "E-WP-0005", "pipePressureLoss", "管网失压", "Pipe Pressure Loss", "管道压力过低", "水泵/管网", 7205,
		ev("ALERT", pm("pressure", "DOUBLE", map[string]any{"unit": uref("megapascal", "MPa")}), pm("limitValue", "DOUBLE", map[string]any{"unit": uref("megapascal", "MPa")}))},
	{ftEvent, "E-WP-0006", "pumpOverload", "水泵过载", "Pump Overload", "电机过载保护", "水泵", 7206,
		ev("ERROR", pm("current", "DOUBLE", map[string]any{"unit": uref("ampere", "A")}), pm("ratedCurrent", "DOUBLE", map[string]any{"unit": uref("ampere", "A")}))},

	// 7.3 EL
	{ftEvent, "E-EL-0001", "overCurrentProtection", "过流保护", "Over Current Protection", "电流超过额定值", "配电柜", 7301,
		ev("ERROR", pm("current", "DOUBLE", map[string]any{"unit": uref("ampere", "A")}), pm("ratedCurrent", "DOUBLE", map[string]any{"unit": uref("ampere", "A")}))},
	{ftEvent, "E-EL-0002", "overVoltageProtection", "过压保护", "Over Voltage Protection", "电压超过上限", "配电柜", 7302,
		ev("ALERT", pm("voltage", "DOUBLE", map[string]any{"unit": uref("volt", "V")}), pm("limitValue", "DOUBLE", map[string]any{"unit": uref("volt", "V")}))},
	{ftEvent, "E-EL-0003", "underVoltageProtection", "欠压保护", "Under Voltage Protection", "电压低于下限", "配电柜", 7303,
		ev("ALERT", pm("voltage", "DOUBLE", map[string]any{"unit": uref("volt", "V")}), pm("limitValue", "DOUBLE", map[string]any{"unit": uref("volt", "V")}))},
	{ftEvent, "E-EL-0004", "transformerOverTemp", "变压器超温", "Transformer Over Temp", "变压器温度过高", "变压器", 7304,
		ev("ERROR", pm("temperature", "DOUBLE", map[string]any{"unit": uref("celsius", "℃")}), pm("limitValue", "DOUBLE", map[string]any{"unit": uref("celsius", "℃")}))},
	{ftEvent, "E-EL-0005", "batteryLow", "电池电量低", "Battery Low", "电池剩余容量不足", "UPS/EPS", 7305,
		ev("ALERT", pm("remaining", "DOUBLE", map[string]any{"unit": uref("percent", "%")}), pm("limitValue", "DOUBLE", map[string]any{"unit": uref("percent", "%")}))},
	{ftEvent, "E-EL-0006", "mainsPowerLoss", "市电中断", "Mains Power Loss", "市电供电中断", "UPS/EPS/发电机组", 7306,
		ev("ALERT", pm("duration", "DOUBLE", map[string]any{"unit": uref("second", "s")}))},
	{ftEvent, "E-EL-0007", "leakageAlarm", "漏电告警", "Leakage Alarm", "漏电电流超标", "配电柜", 7307,
		ev("ALERT", pm("leakageCurrent", "DOUBLE", map[string]any{"unit": uref("milliampere", "mA")}))},

	// 7.4 FP
	{ftEvent, "E-FP-0001", "fireAlarm", "火灾报警", "Fire Alarm", "火灾探测器报警", "火灾报警系统", 7401,
		ev("ERROR", pm("zone", "TEXT"), pm("detectorType", "ENUM"), pm("detectorId", "TEXT"))},
	{ftEvent, "E-FP-0002", "pipeUnderPressure", "管网欠压", "Pipe Under Pressure", "消防管网压力不足", "消防水系统", 7402,
		ev("ALERT", pm("pressure", "DOUBLE", map[string]any{"unit": uref("megapascal", "MPa")}), pm("limitValue", "DOUBLE", map[string]any{"unit": uref("megapascal", "MPa")}))},
	{ftEvent, "E-FP-0003", "sprinklerActivated", "喷头动作", "Sprinkler Activated", "喷头破裂喷水", "自动喷水系统", 7403,
		ev("ALERT", pm("zone", "TEXT"), pm("sprinklerId", "TEXT"))},
	{ftEvent, "E-FP-0004", "gasDischarged", "气体喷放", "Gas Discharged", "气体灭火剂喷放", "气体灭火系统", 7404,
		ev("ERROR", pm("zone", "TEXT"), pm("agentType", "ENUM"))},
	{ftEvent, "E-FP-0005", "fireDoorNotClosed", "防火门未关闭", "Fire Door Not Closed", "防火门超时未关", "防火门", 7405,
		ev("ALERT", pm("doorId", "TEXT"), pm("duration", "DOUBLE", map[string]any{"unit": uref("second", "s")}))},
	{ftEvent, "E-FP-0006", "detectorOffline", "探测器离线", "Detector Offline", "探测器通信中断", "火灾报警系统", 7406,
		ev("ALERT", pm("detectorId", "TEXT"), pm("zone", "TEXT"))},

	// 7.5 ELV
	{ftEvent, "E-ELV-0001", "trappedPassenger", "困人事件", "Trapped Passenger", "电梯困人", "电梯", 7501,
		ev("ERROR", pm("floor", "INT"), pm("carId", "TEXT"), pm("duration", "DOUBLE", map[string]any{"unit": uref("minute", "min")}))},
	{ftEvent, "E-ELV-0002", "overloadAlarm", "超载告警", "Overload Alarm", "轿厢超载", "电梯", 7502,
		ev("ALERT", pm("currentLoad", "DOUBLE", map[string]any{"unit": uref("kilogram", "kg")}), pm("ratedCapacity", "INT", map[string]any{"unit": uref("kilogram", "kg")}))},
	{ftEvent, "E-ELV-0003", "doorLockFault", "门锁故障", "Door Lock Fault", "层门/轿门锁故障", "电梯", 7503,
		ev("ERROR", pm("floor", "INT"), pm("faultDesc", "TEXT"))},
	{ftEvent, "E-ELV-0004", "speedGovernorTriggered", "限速器动作", "Speed Governor Triggered", "超速保护动作", "电梯", 7504,
		ev("ERROR", pm("currentSpeed", "DOUBLE", map[string]any{"unit": uref("meter_per_second", "m/s")}), pm("ratedSpeed", "DOUBLE", map[string]any{"unit": uref("meter_per_second", "m/s")}))},
	{ftEvent, "E-ELV-0005", "escalatorReverse", "扶梯逆行", "Escalator Reverse", "扶梯非预期反向运行", "自动扶梯", 7505,
		ev("ERROR", pm("direction", "ENUM"))},

	// 7.6 SM
	{ftEvent, "E-SM-0001", "illegalEntry", "非法闯入", "Illegal Entry", "未授权开门", "门禁/入侵报警", 7601,
		ev("ALERT", pm("doorId", "TEXT"), pm("authMethod", "ENUM"), pm("timestamp", "DATE"))},
	{ftEvent, "E-SM-0002", "doorOpenTimeout", "门超时未关", "Door Open Timeout", "门长时间未关闭", "门禁", 7602,
		ev("ALERT", pm("doorId", "TEXT"), pm("duration", "DOUBLE", map[string]any{"unit": uref("second", "s")}))},
	{ftEvent, "E-SM-0003", "videoSignalLost", "视频信号丢失", "Video Signal Lost", "视频信号中断", "摄像机", 7603,
		ev("ERROR", pm("cameraId", "TEXT"))},
	{ftEvent, "E-SM-0004", "storageInsufficient", "存储空间不足", "Storage Insufficient", "存储即将满", "NVR/服务器", 7604,
		ev("ALERT", pm("usage", "DOUBLE", map[string]any{"unit": uref("percent", "%")}), pm("remaining", "DOUBLE", map[string]any{"unit": uref("gigabyte", "GB")}))},
	{ftEvent, "E-SM-0005", "sensorValueAbnormal", "传感器数值异常", "Sensor Value Abnormal", "测量值超出正常范围", "传感器", 7605,
		ev("ALERT", pm("parameterName", "TEXT"), pm("value", "DOUBLE"), pm("normalRange", "STRUCT"))},
	{ftEvent, "E-SM-0006", "controllerOffline", "控制器离线", "Controller Offline", "控制器通信中断", "DDC/PLC", 7606,
		ev("ALERT", pm("controllerId", "TEXT"), pm("lastOnlineTime", "DATE"))},
}

// ===========================================================================
// 设备专属服务 / Device-specific services
// ===========================================================================

var deviceServices = []SeedFeature{
	// 8.1 HVAC
	// identifier 使用 setTemperatureSvc 以避免与同名 property P-HVAC-0007 冲突
	// (tenant_id, identifier) 有唯一索引；upsert 只按 code 处理冲突，跨类型重名会被 DB 拒绝。
	{ftService, "S-HVAC-0001", "setTemperatureSvc", "设定温度", "Set Temperature", "设定目标温度", "空调末端", 8101,
		svc("ASYNC",
			[]map[string]any{pm("temperature", "DOUBLE", map[string]any{"unit": uref("celsius", "℃")}, cs(-20, 60))},
			[]map[string]any{pm("success", "BOOL")})},
	{ftService, "S-HVAC-0002", "setOperationMode", "切换运行模式", "Set Operation Mode", "切换制冷/制热", "冷机/热泵/空调末端", 8102,
		svc("ASYNC", []map[string]any{pm("mode", "ENUM")}, []map[string]any{pm("success", "BOOL")})},
	{ftService, "S-HVAC-0003", "setDamperPosition", "设定风阀开度", "Set Damper Position", "调节风阀开度", "通风机/空调末端", 8103,
		svc("ASYNC",
			[]map[string]any{pm("position", "DOUBLE", map[string]any{"unit": uref("percent", "%")}, cs(0, 100))},
			[]map[string]any{pm("success", "BOOL")})},
	{ftService, "S-HVAC-0004", "setValvePosition", "设定水阀开度", "Set Valve Position", "调节水阀开度", "空调末端/换热器", 8104,
		svc("ASYNC",
			[]map[string]any{pm("position", "DOUBLE", map[string]any{"unit": uref("percent", "%")}, cs(0, 100))},
			[]map[string]any{pm("success", "BOOL")})},
	{ftService, "S-HVAC-0005", "setFanFrequency", "设定风机频率", "Set Fan Frequency", "调节变频器频率", "通风机/变频空调", 8105,
		svc("ASYNC",
			[]map[string]any{pm("frequency", "DOUBLE", map[string]any{"unit": uref("hertz", "Hz")}, cs(0, 100))},
			[]map[string]any{pm("success", "BOOL")})},
	// identifier 使用 setFanSpeedSvc 以避免与同名 property P-HVAC-0028 冲突（见上 S-HVAC-0001 说明）。
	{ftService, "S-HVAC-0006", "setFanSpeedSvc", "设定风速", "Set Fan Speed", "设定风机档位", "空调末端", 8106,
		svc("ASYNC", []map[string]any{pm("speed", "ENUM")}, []map[string]any{pm("success", "BOOL")})},

	// 8.2 WP
	{ftService, "S-WP-0001", "setPumpFrequency", "设定水泵频率", "Set Pump Frequency", "调节水泵变频器频率", "变频水泵", 8201,
		svc("ASYNC",
			[]map[string]any{pm("frequency", "DOUBLE", map[string]any{"unit": uref("hertz", "Hz")}, cs(0, 100))},
			[]map[string]any{pm("success", "BOOL")})},
	// identifier 使用 setWaterTempSvc 以避免与同名 property P-WP-0009 冲突（见上 S-HVAC-0001 说明）。
	{ftService, "S-WP-0002", "setWaterTempSvc", "设定水温", "Set Water Temp", "设定目标热水温度", "热水设备", 8202,
		svc("ASYNC",
			[]map[string]any{pm("temperature", "DOUBLE", map[string]any{"unit": uref("celsius", "℃")}, cs(20, 80))},
			[]map[string]any{pm("success", "BOOL")})},
	{ftService, "S-WP-0003", "switchStandbyPump", "切换备用泵", "Switch Standby Pump", "切换至指定备用泵", "水泵组", 8203,
		svc("ASYNC", []map[string]any{pm("pumpIndex", "INT")}, []map[string]any{pm("success", "BOOL")})},

	// 8.3 EL
	{ftService, "S-EL-0001", "closeSwitch", "合闸", "Close Switch", "合上指定回路开关", "配电柜/箱", 8301,
		svc("ASYNC", []map[string]any{pm("circuitIndex", "INT")}, []map[string]any{pm("success", "BOOL")})},
	{ftService, "S-EL-0002", "openSwitch", "分闸", "Open Switch", "断开指定回路开关", "配电柜/箱", 8302,
		svc("ASYNC", []map[string]any{pm("circuitIndex", "INT")}, []map[string]any{pm("success", "BOOL")})},
	{ftService, "S-EL-0003", "setDimmingLevel", "设定调光级别", "Set Dimming Level", "调节照明亮度", "照明设备", 8303,
		svc("ASYNC",
			[]map[string]any{pm("level", "DOUBLE", map[string]any{"unit": uref("percent", "%")}, cs(0, 100))},
			[]map[string]any{pm("success", "BOOL")})},
	{ftService, "S-EL-0004", "setLightingScene", "切换照明场景", "Set Lighting Scene", "切换预设照明场景", "照明设备", 8304,
		svc("ASYNC", []map[string]any{pm("sceneId", "TEXT")}, []map[string]any{pm("success", "BOOL")})},

	// 8.4 FP
	{ftService, "S-FP-0001", "startFirePump", "启动消防泵", "Start Fire Pump", "远程启动消防水泵", "消防水泵", 8401,
		svc("ASYNC", []map[string]any{pm("pumpIndex", "INT")}, []map[string]any{pm("success", "BOOL")})},
	{ftService, "S-FP-0002", "systemSelfTest", "系统自检", "System Self Test", "触发系统自检", "火灾报警系统", 8402,
		svc("ASYNC", nil, []map[string]any{pm("result", "STRUCT")})},
	{ftService, "S-FP-0003", "silenceAlarm", "消音", "Silence Alarm", "消除报警声响", "火灾报警系统", 8403,
		svc("ASYNC", nil, []map[string]any{pm("success", "BOOL")})},
	{ftService, "S-FP-0004", "resetSystem", "复位", "Reset System", "系统复位", "火灾报警系统", 8404,
		svc("ASYNC", nil, []map[string]any{pm("success", "BOOL")})},

	// 8.5 ELV
	{ftService, "S-ELV-0001", "callElevator", "呼梯", "Call Elevator", "外呼电梯", "电梯", 8501,
		svc("ASYNC", []map[string]any{pm("floor", "INT"), pm("direction", "ENUM")}, []map[string]any{pm("success", "BOOL")})},
	{ftService, "S-ELV-0002", "lockElevator", "锁梯", "Lock Elevator", "锁定/解锁电梯", "电梯", 8502,
		svc("ASYNC", []map[string]any{pm("lock", "BOOL")}, []map[string]any{pm("success", "BOOL")})},
	{ftService, "S-ELV-0003", "fireReturn", "消防返回", "Fire Return", "触发消防迫降", "电梯", 8503,
		svc("ASYNC", nil, []map[string]any{pm("success", "BOOL")})},
	{ftService, "S-ELV-0004", "setEscalatorDirection", "切换扶梯方向", "Set Escalator Direction", "切换扶梯运行方向", "自动扶梯", 8504,
		svc("ASYNC", []map[string]any{pm("direction", "ENUM")}, []map[string]any{pm("success", "BOOL")})},

	// 8.6 SM
	{ftService, "S-SM-0001", "remoteOpenDoor", "远程开门", "Remote Open Door", "远程开启指定门", "门禁", 8601,
		svc("ASYNC", []map[string]any{pm("doorId", "TEXT")}, []map[string]any{pm("success", "BOOL")})},
	{ftService, "S-SM-0002", "armSystem", "布防", "Arm System", "设定布防模式", "入侵报警", 8602,
		svc("ASYNC", []map[string]any{pm("armMode", "ENUM")}, []map[string]any{pm("success", "BOOL")})},
	{ftService, "S-SM-0003", "ptzControl", "云台控制", "PTZ Control", "控制云台方向和速度", "摄像机", 8603,
		svc("ASYNC",
			[]map[string]any{pm("action", "ENUM"), pm("speed", "INT", cs(1, 10))},
			[]map[string]any{pm("success", "BOOL")})},
	{ftService, "S-SM-0004", "setControlOutput", "设定控制输出", "Set Control Output", "设定模拟输出值", "控制器", 8604,
		svc("ASYNC",
			[]map[string]any{pm("outputIndex", "INT"), pm("value", "DOUBLE", cs(0, 100))},
			[]map[string]any{pm("success", "BOOL")})},
	{ftService, "S-SM-0005", "publishMessage", "发布信息", "Publish Message", "发布显示内容", "信息发布", 8605,
		svc("ASYNC", []map[string]any{pm("content", "TEXT"), pm("duration", "INT", map[string]any{"unit": uref("second", "s")})}, []map[string]any{pm("success", "BOOL")})},
}

// ===========================================================================
// 关系（核心示例）/ Relations (core examples)
// ===========================================================================

var coreRelations = []SeedFeature{
	{ftRelation, "R-HVAC-0001", "copDerivedFrom", "实时COP派生关系", "COP Derived From",
		"实时COP由实时制冷量与功率派生 (currentCOP = currentCoolingCapacity / currentPower)", "制冷机组", 9101,
		rel("derivedFrom", "oneToOne", true,
			refFeature("currentCOP"),
			refFeature("currentCoolingCapacity"))},

	{ftRelation, "R-HVAC-0002", "supplyReturnTempPair", "供回水温度关系", "Supply-Return Temp Pair",
		"供水温度与回水温度的关联，用于温差计算", "冷机/锅炉/换热器", 9102,
		rel("relatedTo", "oneToOne", false,
			refFeature("supplyWaterTemp"),
			refFeature("returnWaterTemp"))},

	{ftRelation, "R-HVAC-0003", "compressorStatusDependsOnPower", "压缩机状态依赖开关", "Compressor Depends On Power",
		"压缩机运行状态依赖于设备开关", "制冷机组", 9103,
		rel("dependsOn", "manyToOne", true,
			refFeature("compressorStatus"),
			refFeature("powerSwitch"))},

	{ftRelation, "R-EL-0001", "loadRateFromActivePower", "负载率派生关系", "Load Rate Derived",
		"变压器负载率由有功功率/额定功率派生", "变压器", 9301,
		rel("derivedFrom", "oneToOne", true,
			refFeature("loadRate"),
			refFeature("activePower"))},

	{ftRelation, "R-FP-0001", "fireAlarmTriggersDoorRelease", "火警触发防火门释放", "Fire Alarm Triggers Door Release",
		"火灾报警时触发防火门释放（控制链）", "消防系统", 9401,
		rel("controls", "oneToMany", true,
			refFeature("alarmStatus"),
			refFeature("fireDoorStatus"))},

	{ftRelation, "R-ELV-0001", "fireReturnControlsLockStatus", "消防返回控制锁梯", "Fire Return Controls Lock",
		"消防返回触发时电梯进入锁梯状态", "电梯", 9501,
		rel("controls", "oneToOne", true,
			refFeature("fireReturnStatus"),
			refFeature("lockStatus"))},
}

// ===========================================================================
// 额定参数（按设备大类铺开）/ Rated parameters by device category
// ===========================================================================

var ratedFeatures = []SeedFeature{
	// 5.1.1 制冷机组（CH）— 部分关键参数
	{ftProperty, "RAT-HVAC-CH-001", "ratedCoolingCapacity", "额定制冷量", "Rated Cooling Capacity", "铭牌额定制冷量", "制冷机组", 51101,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("kilowatt", "kW"), "isRated": true}, cs(0, 50000))},
	{ftProperty, "RAT-HVAC-CH-002", "ratedHeatingCapacity", "额定制热量", "Rated Heating Capacity", "热泵机组适用", "制冷机组", 51102,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("kilowatt", "kW"), "isRated": true}, cs(0, 50000))},
	{ftProperty, "RAT-HVAC-CH-003", "ratedInputPower", "额定输入功率", "Rated Input Power", "额定工况输入功率", "制冷机组", 51103,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("kilowatt", "kW"), "isRated": true}, cs(0, 10000))},
	{ftProperty, "RAT-HVAC-CH-004", "ratedCOP", "额定COP", "Rated COP", "额定性能系数", "制冷机组", 51104,
		pp("DOUBLE", "R", "rated", map[string]any{"isRated": true}, cs(0, 15))},
	{ftProperty, "RAT-HVAC-CH-005", "ratedEER", "额定EER", "Rated EER", "额定能效比", "制冷机组", 51105,
		pp("DOUBLE", "R", "rated", map[string]any{"isRated": true}, cs(0, 15))},
	{ftProperty, "RAT-HVAC-CH-006", "ratedIPLV", "额定IPLV", "Rated IPLV", "综合部分负荷性能系数", "制冷机组", 51106,
		pp("DOUBLE", "R", "rated", map[string]any{"isRated": true}, cs(0, 15))},
	{ftProperty, "RAT-HVAC-CH-007", "refrigerantType", "制冷剂类型", "Refrigerant Type", "制冷剂种类", "制冷机组", 51107,
		pp("ENUM", "R", "rated", map[string]any{"isRated": true},
			enums([2]any{0, "R410A"}, [2]any{1, "R134a"}, [2]any{2, "R32"}, [2]any{3, "R407C"}, [2]any{4, "R290"}, [2]any{5, "R717"}, [2]any{6, "其他"}))},
	{ftProperty, "RAT-HVAC-CH-008", "refrigerantCharge", "制冷剂充注量", "Refrigerant Charge", "制冷剂充注质量", "制冷机组", 51108,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("kilogram", "kg"), "isRated": true}, cs(0, 5000))},
	{ftProperty, "RAT-HVAC-CH-009", "ratedChilledWaterFlow", "冷冻水额定流量", "Rated Chilled Water Flow", "额定冷冻水流量", "制冷机组", 51109,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("cubic_meter_per_hour", "m³/h"), "isRated": true}, cs(0, 10000))},
	{ftProperty, "RAT-HVAC-CH-010", "ratedCondenserWaterFlow", "冷却水额定流量", "Rated Condenser Water Flow", "额定冷却水流量", "制冷机组", 51110,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("cubic_meter_per_hour", "m³/h"), "isRated": true}, cs(0, 10000))},
	{ftProperty, "RAT-HVAC-CH-011", "ratedChilledWaterTemp", "冷冻水进出水温度", "Rated Chilled Water Temp", "额定冷冻水供回水温度", "制冷机组", 51111,
		map[string]any{
			"dataType":   "STRUCT",
			"accessMode": "R",
			"category":   "rated",
			"isRated":    true,
			"structFields": []map[string]any{
				fStruct("inlet", "DOUBLE", uref("celsius", "℃"), cs(0, 40)),
				fStruct("outlet", "DOUBLE", uref("celsius", "℃"), cs(0, 40)),
			},
		}},
	{ftProperty, "RAT-HVAC-CH-012", "ratedCondenserWaterTemp", "冷却水进出水温度", "Rated Condenser Water Temp", "额定冷却水进出水温度", "制冷机组", 51112,
		map[string]any{
			"dataType":   "STRUCT",
			"accessMode": "R",
			"category":   "rated",
			"isRated":    true,
			"structFields": []map[string]any{
				fStruct("inlet", "DOUBLE", uref("celsius", "℃"), cs(0, 50)),
				fStruct("outlet", "DOUBLE", uref("celsius", "℃"), cs(0, 50)),
			},
		}},
	{ftProperty, "RAT-HVAC-CH-013", "compressorType", "压缩机类型", "Compressor Type", "压缩机形式", "制冷机组", 51113,
		pp("ENUM", "R", "rated", map[string]any{"isRated": true},
			enums([2]any{0, "离心式"}, [2]any{1, "螺杆式"}, [2]any{2, "涡旋式"}, [2]any{3, "活塞式"}, [2]any{4, "转子式"}))},
	{ftProperty, "RAT-HVAC-CH-014", "compressorCount", "压缩机数量", "Compressor Count", "压缩机台数", "制冷机组", 51114,
		pp("INT", "R", "rated", map[string]any{"isRated": true}, cs(1, 20))},
	{ftProperty, "RAT-HVAC-CH-015", "ratedWaterResistance", "额定水阻力", "Rated Water Resistance", "额定水侧压降", "制冷机组", 51115,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("kpa", "kPa"), "isRated": true}, cs(0, 1000))},

	// 5.1.2 锅炉（BL）
	{ftProperty, "RAT-HVAC-BL-001", "ratedThermalPower", "额定热功率", "Rated Thermal Power", "额定热输出功率", "锅炉", 51201,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("megawatt", "MW"), "isRated": true}, cs(0, 100))},
	{ftProperty, "RAT-HVAC-BL-002", "ratedWorkingPressure", "额定工作压力", "Rated Working Pressure", "额定蒸汽/热水压力", "锅炉", 51202,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("megapascal", "MPa"), "isRated": true}, cs(0, 10))},
	{ftProperty, "RAT-HVAC-BL-003", "ratedSupplyTemp", "额定供水温度", "Rated Supply Temp", "额定出口水温", "锅炉", 51203,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("celsius", "℃"), "isRated": true}, cs(0, 250))},
	{ftProperty, "RAT-HVAC-BL-004", "ratedReturnTemp", "额定回水温度", "Rated Return Temp", "额定回水温度", "锅炉", 51204,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("celsius", "℃"), "isRated": true}, cs(0, 200))},
	{ftProperty, "RAT-HVAC-BL-005", "ratedEfficiency", "额定热效率", "Rated Efficiency", "额定热效率", "锅炉", 51205,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("percent", "%"), "isRated": true}, cs(0, 100))},
	{ftProperty, "RAT-HVAC-BL-006", "fuelType", "燃料类型", "Fuel Type", "燃料种类", "锅炉", 51206,
		pp("ENUM", "R", "rated", map[string]any{"isRated": true},
			enums([2]any{0, "天然气"}, [2]any{1, "柴油"}, [2]any{2, "电"}, [2]any{3, "生物质"}, [2]any{4, "液化石油气"}, [2]any{5, "煤"}))},
	{ftProperty, "RAT-HVAC-BL-007", "ratedFuelConsumption", "额定燃料耗量", "Rated Fuel Consumption", "额定工况燃料消耗", "锅炉", 51207,
		pp("DOUBLE", "R", "rated", map[string]any{"isRated": true})},
	{ftProperty, "RAT-HVAC-BL-008", "ratedWaterVolume", "额定水容量", "Rated Water Volume", "锅炉容水量", "锅炉", 51208,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("cubic_meter", "m³"), "isRated": true}, cs(0, 100))},

	// 5.1.3 换热器（HX）
	{ftProperty, "RAT-HVAC-HX-001", "ratedHeatTransfer", "额定换热量", "Rated Heat Transfer", "额定换热能力", "换热器", 51301,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("kilowatt", "kW"), "isRated": true}, cs(0, 50000))},
	{ftProperty, "RAT-HVAC-HX-002", "heatExchangerType", "换热器类型", "Heat Exchanger Type", "换热器形式", "换热器", 51302,
		pp("ENUM", "R", "rated", map[string]any{"isRated": true},
			enums([2]any{0, "板式"}, [2]any{1, "壳管式"}, [2]any{2, "容积式"}, [2]any{3, "翅片式"}))},
	{ftProperty, "RAT-HVAC-HX-003", "ratedPrimaryFlow", "一次侧额定流量", "Rated Primary Flow", "一次侧额定流量", "换热器", 51303,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("cubic_meter_per_hour", "m³/h"), "isRated": true}, cs(0, 10000))},
	{ftProperty, "RAT-HVAC-HX-004", "ratedSecondaryFlow", "二次侧额定流量", "Rated Secondary Flow", "二次侧额定流量", "换热器", 51304,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("cubic_meter_per_hour", "m³/h"), "isRated": true}, cs(0, 10000))},
	{ftProperty, "RAT-HVAC-HX-007", "hxRatedWorkingPressure", "额定工作压力", "Rated Working Pressure (HX)", "额定承压", "换热器", 51307,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("megapascal", "MPa"), "isRated": true}, cs(0, 10))},

	// 5.1.4 空调末端（TU）
	{ftProperty, "RAT-HVAC-TU-001", "tuRatedCoolingCapacity", "空调末端额定制冷量", "TU Rated Cooling Capacity", "末端额定制冷量", "空调末端", 51401,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("kilowatt", "kW"), "isRated": true}, cs(0, 500))},
	{ftProperty, "RAT-HVAC-TU-002", "tuRatedHeatingCapacity", "空调末端额定制热量", "TU Rated Heating Capacity", "末端额定制热量", "空调末端", 51402,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("kilowatt", "kW"), "isRated": true}, cs(0, 500))},
	{ftProperty, "RAT-HVAC-TU-003", "ratedAirVolume", "额定风量", "Rated Air Volume", "额定风量", "空调末端", 51403,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("cubic_meter_per_hour", "m³/h"), "isRated": true}, cs(0, 100000))},
	{ftProperty, "RAT-HVAC-TU-004", "ratedExternalPressure", "额定余压", "Rated External Pressure", "额定机外余压", "空调末端", 51404,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("pascal", "Pa"), "isRated": true}, cs(0, 2000))},

	// 5.1.5 通风机（FN）
	{ftProperty, "RAT-HVAC-FN-001", "fnRatedAirVolume", "通风机额定风量", "Fan Rated Air Volume", "额定工况风量", "通风机", 51501,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("cubic_meter_per_hour", "m³/h"), "isRated": true}, cs(0, 500000))},
	{ftProperty, "RAT-HVAC-FN-002", "ratedTotalPressure", "额定全压", "Rated Total Pressure", "额定全压", "通风机", 51502,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("pascal", "Pa"), "isRated": true}, cs(0, 10000))},
	{ftProperty, "RAT-HVAC-FN-003", "ratedStaticPressure", "额定静压", "Rated Static Pressure", "额定静压", "通风机", 51503,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("pascal", "Pa"), "isRated": true}, cs(0, 10000))},
	{ftProperty, "RAT-HVAC-FN-004", "fanType", "风机类型", "Fan Type", "风机形式", "通风机", 51504,
		pp("ENUM", "R", "rated", map[string]any{"isRated": true},
			enums([2]any{0, "离心式"}, [2]any{1, "轴流式"}, [2]any{2, "混流式"}, [2]any{3, "斜流式"}, [2]any{4, "屋顶式"}))},

	// 5.1.6 冷却塔（CT）
	{ftProperty, "RAT-HVAC-CT-001", "ctRatedWaterFlow", "冷却塔额定水量", "CT Rated Water Flow", "额定冷却水量", "冷却塔", 51601,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("cubic_meter_per_hour", "m³/h"), "isRated": true}, cs(0, 10000))},
	{ftProperty, "RAT-HVAC-CT-002", "ctRatedInletTemp", "冷却塔额定进水温度", "CT Rated Inlet Temp", "额定进水温度", "冷却塔", 51602,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("celsius", "℃"), "isRated": true}, cs(0, 60))},
	{ftProperty, "RAT-HVAC-CT-003", "ctRatedOutletTemp", "冷却塔额定出水温度", "CT Rated Outlet Temp", "额定出水温度", "冷却塔", 51603,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("celsius", "℃"), "isRated": true}, cs(0, 50))},
	{ftProperty, "RAT-HVAC-CT-004", "ratedWetBulbTemp", "额定湿球温度", "Rated Wet Bulb Temp", "设计湿球温度", "冷却塔", 51604,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("celsius", "℃"), "isRated": true}, cs(0, 40))},

	// 5.2.1 水泵（PM）
	{ftProperty, "RAT-WP-PM-001", "ratedFlowRate", "额定流量", "Rated Flow Rate", "额定工况流量", "水泵", 52101,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("cubic_meter_per_hour", "m³/h"), "isRated": true}, cs(0, 10000))},
	{ftProperty, "RAT-WP-PM-002", "ratedHead", "额定扬程", "Rated Head", "额定扬程", "水泵", 52102,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("meter", "m"), "isRated": true}, cs(0, 500))},
	{ftProperty, "RAT-WP-PM-003", "pumpType", "水泵类型", "Pump Type", "水泵形式", "水泵", 52103,
		pp("ENUM", "R", "rated", map[string]any{"isRated": true},
			enums([2]any{0, "离心式"}, [2]any{1, "轴流式"}, [2]any{2, "混流式"}, [2]any{3, "潜水式"}, [2]any{4, "管道式"}, [2]any{5, "隔膜式"}))},

	// 5.3.1 变压器（TR）
	{ftProperty, "RAT-EL-TR-001", "ratedCapacity", "额定容量", "Rated Capacity", "额定视在功率", "变压器", 53101,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("kva", "kVA"), "isRated": true}, cs(0, 100000))},
	{ftProperty, "RAT-EL-TR-002", "ratedPrimaryVoltage", "一次侧额定电压", "Rated Primary Voltage", "高压侧额定电压", "变压器", 53102,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("kilovolt", "kV"), "isRated": true}, cs(0, 500))},
	{ftProperty, "RAT-EL-TR-003", "ratedSecondaryVoltage", "二次侧额定电压", "Rated Secondary Voltage", "低压侧额定电压", "变压器", 53103,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("kilovolt", "kV"), "isRated": true}, cs(0, 100))},

	// 5.5.1 电梯（EL）
	{ftProperty, "RAT-ELV-EL-001", "elvRatedCapacity", "电梯额定载重量", "Elevator Rated Capacity", "额定载重", "电梯", 55101,
		pp("INT", "R", "rated", map[string]any{"unit": uref("kilogram", "kg"), "isRated": true}, cs(0, 20000))},
	{ftProperty, "RAT-ELV-EL-002", "ratedSpeed", "电梯额定速度", "Elevator Rated Speed", "额定运行速度", "电梯", 55102,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("meter_per_second", "m/s"), "isRated": true}, cs(0, 20))},
	{ftProperty, "RAT-ELV-EL-003", "floorCount", "层站数", "Floor Count", "服务楼层总数", "电梯", 55103,
		pp("INT", "R", "rated", map[string]any{"isRated": true}, cs(1, 200))},
	{ftProperty, "RAT-ELV-EL-004", "liftHeight", "提升高度", "Lift Height", "总垂直行程", "电梯", 55104,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("meter", "m"), "isRated": true}, cs(0, 600))},
	{ftProperty, "RAT-ELV-EL-005", "elevatorType", "电梯类型", "Elevator Type", "电梯用途分类", "电梯", 55105,
		pp("ENUM", "R", "rated", map[string]any{"isRated": true},
			enums([2]any{0, "乘客"}, [2]any{1, "载货"}, [2]any{2, "医用"}, [2]any{3, "汽车"}, [2]any{4, "消防"}, [2]any{5, "观光"}, [2]any{6, "防爆"}))},

	// 5.6.1 摄像机（CM）
	{ftProperty, "RAT-SM-CM-001", "resolution", "分辨率", "Resolution", "视频分辨率", "摄像机", 56101,
		pp("ENUM", "R", "rated", map[string]any{"isRated": true},
			enums([2]any{0, "720P"}, [2]any{1, "1080P"}, [2]any{2, "2K"}, [2]any{3, "4K"}, [2]any{4, "8K"}))},
	{ftProperty, "RAT-SM-CM-002", "sensorType", "传感器类型", "Sensor Type", "图像传感器类型", "摄像机", 56102,
		pp("ENUM", "R", "rated", map[string]any{"isRated": true},
			enums([2]any{0, "CMOS"}, [2]any{1, "CCD"}))},

	// 5.7.1 光伏（PV）
	{ftProperty, "RAT-AX-PV-001", "ratedPeakPower", "额定峰值功率", "Rated Peak Power", "额定峰值功率", "光伏", 57101,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("kwp", "kWp"), "isRated": true}, cs(0, 10000))},
	{ftProperty, "RAT-AX-PV-002", "moduleEfficiency", "组件转换效率", "Module Efficiency", "光伏组件转换效率", "光伏", 57102,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("percent", "%"), "isRated": true}, cs(0, 50))},

	// 5.7.2 充电桩（CP）
	{ftProperty, "RAT-AX-CP-001", "chargingType", "充电类型", "Charging Type", "充电模式", "充电桩", 57201,
		pp("ENUM", "R", "rated", map[string]any{"isRated": true},
			enums([2]any{0, "交流慢充"}, [2]any{1, "直流快充"}))},
	{ftProperty, "RAT-AX-CP-002", "cpRatedPower", "充电桩额定功率", "CP Rated Power", "额定充电功率", "充电桩", 57202,
		pp("DOUBLE", "R", "rated", map[string]any{"unit": uref("kilowatt", "kW"), "isRated": true}, cs(0, 500))},
}

// ===========================================================================
// 聚合所有种子数据 / Aggregate all seeds
// ===========================================================================

// AllFeatureSeeds 返回所有特征种子数据（用于 SeedThingmodelFeatures 一次性入库）。
// AllFeatureSeeds returns the full set of seed features for one-shot import.
func AllFeatureSeeds() []SeedFeature {
	all := make([]SeedFeature, 0, 256)
	all = append(all, commonProperties...)
	all = append(all, hvacProperties...)
	all = append(all, wpProperties...)
	all = append(all, elProperties...)
	all = append(all, fpProperties...)
	all = append(all, elvProperties...)
	all = append(all, smProperties...)
	all = append(all, ratedFeatures...)
	all = append(all, commonEvents...)
	all = append(all, deviceEvents...)
	all = append(all, commonServices...)
	all = append(all, deviceServices...)
	all = append(all, coreRelations...)
	return all
}
