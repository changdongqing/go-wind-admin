-- ============================================================
-- 物模型特征种子数据 / Thing Model Feature Seed Data
-- 生成自: 14-特征种子数据-完整.xlsx
-- 目标表: public.thingmodel_features
-- ============================================================

-- 幂等插入: 使用 ON CONFLICT (tenant_id, code) DO NOTHING
-- 避免重复导入 / Idempotent insert with conflict handling

BEGIN;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 1, 0, 'EVENT', 'E-GEN-0001', 'deviceOnline', '设备上线', 'Device Online', '设备连接上线', '通用', NULL, NULL, 'INFO', NULL, NULL, '{"level":"INFO","outputParams":[{"dataType":"DATE","key":"timestamp"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 1, 0, 'SERVICE', 'S-GEN-0001', 'turnOn', '开机', 'Turn On', '设备开机', '通用', NULL, NULL, NULL, 'ASYNC', NULL, '{"callMode":"ASYNC","outputParams":[{"dataType":"BOOL","key":"success"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 2, 0, 'EVENT', 'E-GEN-0002', 'deviceOffline', '设备离线', 'Device Offline', '设备断开连接', '通用', NULL, NULL, 'INFO', NULL, NULL, '{"level":"INFO","outputParams":[{"dataType":"DATE","key":"timestamp"},{"dataType":"DOUBLE","key":"offlineDuration","unit":{"unitCode":"second","unitSymbol":"s"}}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 2, 0, 'SERVICE', 'S-GEN-0002', 'turnOff', '关机', 'Turn Off', '设备关机', '通用', NULL, NULL, NULL, 'ASYNC', NULL, '{"callMode":"ASYNC","outputParams":[{"dataType":"BOOL","key":"success"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 3, 0, 'EVENT', 'E-GEN-0003', 'runStatusChanged', '运行状态变更', 'Run Status Changed', '运行状态切换', '通用', NULL, NULL, 'INFO', NULL, NULL, '{"level":"INFO","outputParams":[{"dataType":"ENUM","key":"oldStatus"},{"dataType":"ENUM","key":"newStatus"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 3, 0, 'SERVICE', 'S-GEN-0003', 'restart', '重启', 'Restart', '延时重启设备', '通用', NULL, NULL, NULL, 'ASYNC', NULL, '{"callMode":"ASYNC","inputParams":[{"constraints":{"max":3600,"min":0},"dataType":"DOUBLE","key":"delay","unit":{"unitCode":"second","unitSymbol":"s"}}],"outputParams":[{"dataType":"BOOL","key":"success"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 4, 0, 'EVENT', 'E-GEN-0004', 'parameterOverLimit', '参数越限告警', 'Parameter Over Limit', '测量参数超出设定阈值', '通用', NULL, NULL, 'ALERT', NULL, NULL, '{"level":"ALERT","outputParams":[{"dataType":"TEXT","key":"parameterName"},{"dataType":"DOUBLE","key":"currentValue"},{"dataType":"DOUBLE","key":"limitValue"},{"dataType":"TEXT","key":"unit"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 4, 0, 'SERVICE', 'S-GEN-0004', 'setRunMode', '设置运行模式', 'Set Run Mode', '切换运行模式', '通用', NULL, NULL, NULL, 'ASYNC', NULL, '{"callMode":"ASYNC","inputParams":[{"dataType":"ENUM","key":"runMode"}],"outputParams":[{"dataType":"BOOL","key":"success"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 5, 0, 'EVENT', 'E-GEN-0005', 'communicationError', '通信异常告警', 'Communication Error', '通信链路异常', '通用', NULL, NULL, 'ALERT', NULL, NULL, '{"level":"ALERT","outputParams":[{"dataType":"ENUM","key":"errorType"},{"dataType":"TEXT","key":"detail"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 5, 0, 'SERVICE', 'S-GEN-0005', 'faultReset', '故障复位', 'Fault Reset', '清除故障状态', '通用', NULL, NULL, NULL, 'ASYNC', NULL, '{"callMode":"ASYNC","outputParams":[{"dataType":"BOOL","key":"success"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6, 0, 'EVENT', 'E-GEN-0006', 'deviceFault', '设备故障', 'Device Fault', '设备发生故障', '通用', NULL, NULL, 'ERROR', NULL, NULL, '{"level":"ERROR","outputParams":[{"dataType":"INT","key":"faultCode"},{"dataType":"TEXT","key":"faultDesc"},{"dataType":"DATE","key":"faultTime"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6, 0, 'SERVICE', 'S-GEN-0006', 'getProperty', '读取属性', 'Get Property', '同步读取指定属性当前值', '通用', NULL, NULL, NULL, 'SYNC', NULL, '{"callMode":"SYNC","inputParams":[{"dataType":"TEXT","key":"propertyIdentifier"}],"outputParams":[{"dataType":"STRUCT","key":"propertyValue"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7, 0, 'EVENT', 'E-GEN-0007', 'faultRecovery', '故障恢复', 'Fault Recovery', '故障已恢复', '通用', NULL, NULL, 'INFO', NULL, NULL, '{"level":"INFO","outputParams":[{"dataType":"INT","key":"faultCode"},{"dataType":"DOUBLE","key":"faultDuration","unit":{"unitCode":"minute","unitSymbol":"min"}}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7, 0, 'SERVICE', 'S-GEN-0007', 'setProperty', '设置属性', 'Set Property', '设置指定属性值', '通用', NULL, NULL, NULL, 'ASYNC', NULL, '{"callMode":"ASYNC","inputParams":[{"dataType":"TEXT","key":"propertyIdentifier"},{"dataType":"STRUCT","key":"propertyValue"}],"outputParams":[{"dataType":"BOOL","key":"success"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 8, 0, 'EVENT', 'E-GEN-0008', 'maintenanceReminder', '维保提醒', 'Maintenance Reminder', '设备需要维护保养', '通用', NULL, NULL, 'INFO', NULL, NULL, '{"level":"INFO","outputParams":[{"dataType":"ENUM","key":"maintenanceType"},{"dataType":"DOUBLE","key":"remainingTime","unit":{"unitCode":"hour","unitSymbol":"h"}}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 8, 0, 'SERVICE', 'S-GEN-0008', 'queryHistory', '查询历史数据', 'Query History', '查询指定时间段历史数据', '通用', NULL, NULL, NULL, 'SYNC', NULL, '{"callMode":"SYNC","inputParams":[{"dataType":"DATE","key":"startTime"},{"dataType":"DATE","key":"endTime"},{"dataType":"ARRAY","key":"propertyIdentifiers"}],"outputParams":[{"dataType":"ARRAY","key":"historyData"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 101, 0, 'PROPERTY', 'P-RUN-0001', 'powerSwitch', '开关状态', 'Power Switch', '设备开关', '通用', 'BOOL', 'RW', NULL, NULL, NULL, '{"accessMode":"RW","boolLabels":{"false":"关","true":"开"},"category":"runtime","dataType":"BOOL"}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 102, 0, 'PROPERTY', 'P-RUN-0002', 'runMode', '运行模式', 'Run Mode', '运行模式选择', '通用', 'ENUM', 'RW', NULL, NULL, NULL, '{"accessMode":"RW","category":"runtime","dataType":"ENUM","enumItems":[{"label":"停机","value":0},{"label":"手动","value":1},{"label":"自动","value":2},{"label":"远程","value":3}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 103, 0, 'PROPERTY', 'P-RUN-0003', 'runStatus', '运行状态', 'Run Status', '设备当前运行状态', '通用', 'ENUM', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"runtime","dataType":"ENUM","enumItems":[{"label":"停止","value":0},{"label":"运行","value":1},{"label":"待机","value":2},{"label":"故障","value":3}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 104, 0, 'PROPERTY', 'P-RUN-0004', 'faultCode', '故障代码', 'Fault Code', '0=无故障，其他=故障代码', '通用', 'INT', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"runtime","dataType":"INT"}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 105, 0, 'PROPERTY', 'P-RUN-0005', 'faultDesc', '故障描述', 'Fault Description', '故障文本描述', '通用', 'TEXT', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"runtime","dataType":"TEXT","textMaxLength":256}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 106, 0, 'PROPERTY', 'P-RUN-0006', 'localRemote', '本地/远程', 'Local/Remote', '控制权位置', '通用', 'ENUM', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"runtime","dataType":"ENUM","enumItems":[{"label":"本地","value":0},{"label":"远程","value":1}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 107, 0, 'PROPERTY', 'P-RUN-0007', 'autoManual', '手自动', 'Auto/Manual', '手/自动模式', '通用', 'ENUM', 'RW', NULL, NULL, NULL, '{"accessMode":"RW","category":"runtime","dataType":"ENUM","enumItems":[{"label":"手动","value":0},{"label":"自动","value":1}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 201, 0, 'PROPERTY', 'P-RAT-0001', 'model', '设备型号', 'Model', '出厂型号', '通用', 'TEXT', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","dataType":"TEXT","isRated":true,"textMaxLength":128}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 202, 0, 'PROPERTY', 'P-RAT-0002', 'manufacturer', '制造商', 'Manufacturer', '设备生产厂家', '通用', 'TEXT', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","dataType":"TEXT","isRated":true,"textMaxLength":128}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 203, 0, 'PROPERTY', 'P-RAT-0003', 'manufactureDate', '出厂日期', 'Manufacture Date', '出厂日期', '通用', 'DATE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","dataType":"DATE","isRated":true}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 204, 0, 'PROPERTY', 'P-RAT-0004', 'serialNumber', '序列号', 'Serial Number', '出厂序列号', '通用', 'TEXT', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","dataType":"TEXT","isRated":true,"textMaxLength":64}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 205, 0, 'PROPERTY', 'P-RAT-0005', 'ratedVoltage', '额定电压', 'Rated Voltage', '额定电压', '通用', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":10000,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"volt","unitSymbol":"V"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 206, 0, 'PROPERTY', 'P-RAT-0006', 'ratedCurrent', '额定电流', 'Rated Current', '额定电流', '通用', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":10000,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"ampere","unitSymbol":"A"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 207, 0, 'PROPERTY', 'P-RAT-0007', 'ratedPower', '额定功率', 'Rated Power', '额定功率', '通用', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":100000,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"kilowatt","unitSymbol":"kW"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 208, 0, 'PROPERTY', 'P-RAT-0008', 'ratedFrequency', '额定频率', 'Rated Frequency', '额定频率', '通用', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":100,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"hertz","unitSymbol":"Hz"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 209, 0, 'PROPERTY', 'P-RAT-0009', 'ipRating', '防护等级', 'IP Rating', '如 IP42/IP54', '通用', 'TEXT', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","dataType":"TEXT","isRated":true,"textMaxLength":16}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 210, 0, 'PROPERTY', 'P-RAT-0010', 'insulationClass', '绝缘等级', 'Insulation Class', '绝缘耐热等级', '通用', 'ENUM', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","dataType":"ENUM","enumItems":[{"label":"A级","value":0},{"label":"E级","value":1},{"label":"B级","value":2},{"label":"F级","value":3},{"label":"H级","value":4},{"label":"C级","value":5}],"isRated":true}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 301, 0, 'PROPERTY', 'P-STA-0001', 'totalRunTime', '累计运行时间', 'Total Run Time', '累计运行小时数', '通用', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"statistic","dataType":"DOUBLE","unit":{"unitCode":"hour","unitSymbol":"h"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 302, 0, 'PROPERTY', 'P-STA-0002', 'totalEnergy', '累计能耗', 'Total Energy', '累计耗电量', '通用', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"statistic","dataType":"DOUBLE","unit":{"unitCode":"kilowatt_hour","unitSymbol":"kWh"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 303, 0, 'PROPERTY', 'P-STA-0003', 'startCount', '启动次数', 'Start Count', '累计启动次数', '通用', 'INT', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"statistic","dataType":"INT"}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 304, 0, 'PROPERTY', 'P-STA-0004', 'faultCount', '故障次数', 'Fault Count', '累计故障次数', '通用', 'INT', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"statistic","dataType":"INT"}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 305, 0, 'PROPERTY', 'P-STA-0005', 'todayRunTime', '今日运行时间', 'Today Run Time', '当日累计运行小时', '通用', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"statistic","constraints":{"max":24,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"hour","unitSymbol":"h"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 306, 0, 'PROPERTY', 'P-STA-0006', 'todayEnergy', '今日能耗', 'Today Energy', '当日累计能耗', '通用', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"statistic","dataType":"DOUBLE","unit":{"unitCode":"kilowatt_hour","unitSymbol":"kWh"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 401, 0, 'PROPERTY', 'P-ENV-0001', 'ambientTemp', '环境温度', 'Ambient Temp', '设备环境温度', '通用', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"environment","constraints":{"max":80,"min":-40},"dataType":"DOUBLE","unit":{"unitCode":"celsius","unitSymbol":"℃"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 402, 0, 'PROPERTY', 'P-ENV-0002', 'ambientHumidity', '环境湿度', 'Ambient Humidity', '设备环境湿度', '通用', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"environment","constraints":{"max":100,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"percent_rh","unitSymbol":"%RH"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6001, 0, 'PROPERTY', 'P-HVAC-0001', 'supplyWaterTemp', '供水温度', 'Supply Water Temp', '供水温度', '冷机/锅炉/换热器', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":200,"min":-40},"dataType":"DOUBLE","unit":{"unitCode":"celsius","unitSymbol":"℃"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6002, 0, 'PROPERTY', 'P-HVAC-0002', 'returnWaterTemp', '回水温度', 'Return Water Temp', '回水温度', '冷机/锅炉/换热器', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":200,"min":-40},"dataType":"DOUBLE","unit":{"unitCode":"celsius","unitSymbol":"℃"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6003, 0, 'PROPERTY', 'P-HVAC-0003', 'supplyWaterPressure', '供水压力', 'Supply Water Pressure', '供水压力', '冷机/水泵', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":10,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"megapascal","unitSymbol":"MPa"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6004, 0, 'PROPERTY', 'P-HVAC-0004', 'returnWaterPressure', '回水压力', 'Return Water Pressure', '回水压力', '冷机/水泵', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":10,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"megapascal","unitSymbol":"MPa"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6005, 0, 'PROPERTY', 'P-HVAC-0005', 'waterFlowRate', '水流量', 'Water Flow Rate', '水流量', '冷机/水泵', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":10000,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"cubic_meter_per_hour","unitSymbol":"m³/h"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6006, 0, 'PROPERTY', 'P-HVAC-0006', 'operationMode', '供冷/供热模式', 'Operation Mode', '运行模式', '冷机/热泵/空调末端', 'ENUM', 'RW', NULL, NULL, NULL, '{"accessMode":"RW","category":"setting","dataType":"ENUM","enumItems":[{"label":"制冷","value":0},{"label":"制热","value":1},{"label":"通风","value":2},{"label":"除湿","value":3},{"label":"自动","value":4}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6007, 0, 'PROPERTY', 'P-HVAC-0007', 'setTemperature', '设定温度', 'Set Temperature', '设定温度', '空调末端/新风机组', 'DOUBLE', 'RW', NULL, NULL, NULL, '{"accessMode":"RW","category":"setting","constraints":{"max":60,"min":-20},"dataType":"DOUBLE","unit":{"unitCode":"celsius","unitSymbol":"℃"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6008, 0, 'PROPERTY', 'P-HVAC-0008', 'supplyAirTemp', '送风温度', 'Supply Air Temp', '送风温度', '空调末端/新风机组', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":80,"min":-40},"dataType":"DOUBLE","unit":{"unitCode":"celsius","unitSymbol":"℃"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6009, 0, 'PROPERTY', 'P-HVAC-0009', 'returnAirTemp', '回风温度', 'Return Air Temp', '回风温度', '空调末端', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":80,"min":-40},"dataType":"DOUBLE","unit":{"unitCode":"celsius","unitSymbol":"℃"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6010, 0, 'PROPERTY', 'P-HVAC-0010', 'supplyAirHumidity', '送风湿度', 'Supply Air Humidity', '送风湿度', '空调末端/新风机组', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":100,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"percent_rh","unitSymbol":"%RH"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6011, 0, 'PROPERTY', 'P-HVAC-0011', 'damperPosition', '风阀开度', 'Damper Position', '风阀开度', '通风机/空调末端', 'DOUBLE', 'RW', NULL, NULL, NULL, '{"accessMode":"RW","category":"setting","constraints":{"max":100,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"percent","unitSymbol":"%"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6012, 0, 'PROPERTY', 'P-HVAC-0012', 'valvePosition', '水阀开度', 'Valve Position', '水阀开度', '空调末端/换热器', 'DOUBLE', 'RW', NULL, NULL, NULL, '{"accessMode":"RW","category":"setting","constraints":{"max":100,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"percent","unitSymbol":"%"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6013, 0, 'PROPERTY', 'P-HVAC-0013', 'fanFrequency', '风机频率', 'Fan Frequency', '变频风机频率', '通风机/空调末端', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":100,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"hertz","unitSymbol":"Hz"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6014, 0, 'PROPERTY', 'P-HVAC-0014', 'fanSpeed', '风机转速', 'Fan Speed', '风机转速', '通风机', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":30000,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"rpm","unitSymbol":"rpm"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6015, 0, 'PROPERTY', 'P-HVAC-0015', 'filterPressureDiff', '过滤器压差', 'Filter Pressure Diff', '过滤器压差', '通风机/新风机组', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":2000,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"pascal","unitSymbol":"Pa"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6016, 0, 'PROPERTY', 'P-HVAC-0016', 'compressorStatus', '压缩机运行状态', 'Compressor Status', '压缩机状态', '制冷机组', 'ENUM', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"runtime","dataType":"ENUM","enumItems":[{"label":"停止","value":0},{"label":"运行","value":1},{"label":"卸载","value":2},{"label":"故障","value":3}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6017, 0, 'PROPERTY', 'P-HVAC-0017', 'compressorLoadRate', '压缩机负载率', 'Compressor Load Rate', '压缩机负载率', '制冷机组', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":100,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"percent","unitSymbol":"%"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6018, 0, 'PROPERTY', 'P-HVAC-0018', 'evaporatingTemp', '蒸发温度', 'Evaporating Temp', '蒸发温度', '制冷机组', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":30,"min":-50},"dataType":"DOUBLE","unit":{"unitCode":"celsius","unitSymbol":"℃"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6019, 0, 'PROPERTY', 'P-HVAC-0019', 'condensingTemp', '冷凝温度', 'Condensing Temp', '冷凝温度', '制冷机组', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":80,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"celsius","unitSymbol":"℃"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6020, 0, 'PROPERTY', 'P-HVAC-0020', 'condenserInletTemp', '冷却水进水温度', 'Condenser Inlet Temp', '冷却水进水温度', '制冷机组/冷却塔', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":60,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"celsius","unitSymbol":"℃"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6021, 0, 'PROPERTY', 'P-HVAC-0021', 'condenserOutletTemp', '冷却水出水温度', 'Condenser Outlet Temp', '冷却水出水温度', '制冷机组/冷却塔', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":50,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"celsius","unitSymbol":"℃"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6022, 0, 'PROPERTY', 'P-HVAC-0022', 'currentCoolingCapacity', '实时制冷量', 'Current Cooling Capacity', '实时制冷量', '制冷机组', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":50000,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"kilowatt","unitSymbol":"kW"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6023, 0, 'PROPERTY', 'P-HVAC-0023', 'currentCOP', '实时COP', 'Current COP', '实时能效', '制冷机组', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":15,"min":0},"dataType":"DOUBLE"}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6024, 0, 'PROPERTY', 'P-HVAC-0024', 'currentPower', '实时功率', 'Current Power', '实时功率', '通用', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":100000,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"kilowatt","unitSymbol":"kW"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6025, 0, 'PROPERTY', 'P-HVAC-0025', 'freshAirTemp', '新风温度', 'Fresh Air Temp', '新风温度', '新风机组', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":60,"min":-40},"dataType":"DOUBLE","unit":{"unitCode":"celsius","unitSymbol":"℃"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6026, 0, 'PROPERTY', 'P-HVAC-0026', 'freshAirHumidity', '新风湿度', 'Fresh Air Humidity', '新风湿度', '新风机组', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":100,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"percent_rh","unitSymbol":"%RH"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6027, 0, 'PROPERTY', 'P-HVAC-0027', 'co2Concentration', 'CO2浓度', 'CO2 Concentration', '二氧化碳浓度', '新风机组/空调末端', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":5000,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"ppm","unitSymbol":"ppm"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6028, 0, 'PROPERTY', 'P-HVAC-0028', 'setFanSpeed', '设定风速', 'Set Fan Speed', '设定风速档位', '空调末端', 'ENUM', 'RW', NULL, NULL, NULL, '{"accessMode":"RW","category":"setting","dataType":"ENUM","enumItems":[{"label":"低速","value":0},{"label":"中速","value":1},{"label":"高速","value":2},{"label":"自动","value":3}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6201, 0, 'PROPERTY', 'P-WP-0001', 'waterLevel', '水箱液位', 'Water Level', '水箱液位', '水箱/水池', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":20,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"meter","unitSymbol":"m"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6202, 0, 'PROPERTY', 'P-WP-0002', 'waterLevelPercent', '液位百分比', 'Water Level Percent', '液位百分比', '水箱/水池', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":100,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"percent","unitSymbol":"%"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6203, 0, 'PROPERTY', 'P-WP-0003', 'pipePressure', '管道压力', 'Pipe Pressure', '管道压力', '水泵/管网', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":10,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"megapascal","unitSymbol":"MPa"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6204, 0, 'PROPERTY', 'P-WP-0004', 'instantFlowRate', '瞬时流量', 'Instant Flow Rate', '瞬时流量', '水泵', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":10000,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"cubic_meter_per_hour","unitSymbol":"m³/h"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6205, 0, 'PROPERTY', 'P-WP-0005', 'totalFlowVolume', '累计流量', 'Total Flow Volume', '累计流量', '水泵', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"statistic","dataType":"DOUBLE","unit":{"unitCode":"cubic_meter","unitSymbol":"m³"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6206, 0, 'PROPERTY', 'P-WP-0006', 'pumpFrequency', '水泵频率', 'Pump Frequency', '变频水泵频率', '变频水泵', 'DOUBLE', 'RW', NULL, NULL, NULL, '{"accessMode":"RW","category":"setting","constraints":{"max":100,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"hertz","unitSymbol":"Hz"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6207, 0, 'PROPERTY', 'P-WP-0007', 'pumpSpeed', '水泵转速', 'Pump Speed', '水泵转速', '水泵', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":30000,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"rpm","unitSymbol":"rpm"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6208, 0, 'PROPERTY', 'P-WP-0008', 'outletWaterTemp', '出水温度', 'Outlet Water Temp', '热水出水温度', '热水设备', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":100,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"celsius","unitSymbol":"℃"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6209, 0, 'PROPERTY', 'P-WP-0009', 'setWaterTemp', '设定水温', 'Set Water Temp', '设定热水温度', '热水设备', 'DOUBLE', 'RW', NULL, NULL, NULL, '{"accessMode":"RW","category":"setting","constraints":{"max":80,"min":20},"dataType":"DOUBLE","unit":{"unitCode":"celsius","unitSymbol":"℃"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6210, 0, 'PROPERTY', 'P-WP-0010', 'chlorineConcentration', '余氯浓度', 'Chlorine Concentration', '余氯浓度', '水处理设备', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":10,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"mg_per_l","unitSymbol":"mg/L"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6211, 0, 'PROPERTY', 'P-WP-0011', 'turbidity', '浊度', 'Turbidity', '浊度', '水处理设备', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":100,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"ntu","unitSymbol":"NTU"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6212, 0, 'PROPERTY', 'P-WP-0012', 'phValue', 'pH值', 'pH Value', '酸碱度', '水处理设备', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":14,"min":0},"dataType":"DOUBLE"}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6301, 0, 'PROPERTY', 'P-EL-0001', 'voltage', '电压', 'Voltage', '工作电压', '变压器/配电柜', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":10000,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"volt","unitSymbol":"V"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6302, 0, 'PROPERTY', 'P-EL-0002', 'current', '电流', 'Current', '工作电流', '变压器/配电柜', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":100000,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"ampere","unitSymbol":"A"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6303, 0, 'PROPERTY', 'P-EL-0003', 'activePower', '有功功率', 'Active Power', '有功功率', '变压器/配电柜', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":100000,"min":-100000},"dataType":"DOUBLE","unit":{"unitCode":"kilowatt","unitSymbol":"kW"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6304, 0, 'PROPERTY', 'P-EL-0004', 'reactivePower', '无功功率', 'Reactive Power', '无功功率', '变压器/配电柜', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":100000,"min":-100000},"dataType":"DOUBLE","unit":{"unitCode":"kilovar","unitSymbol":"kvar"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6305, 0, 'PROPERTY', 'P-EL-0005', 'powerFactor', '功率因数', 'Power Factor', '功率因数', '变压器/配电柜', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":1,"min":-1},"dataType":"DOUBLE"}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6306, 0, 'PROPERTY', 'P-EL-0006', 'frequency', '频率', 'Frequency', '工作频率', '变压器/发电机组', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":100,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"hertz","unitSymbol":"Hz"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6307, 0, 'PROPERTY', 'P-EL-0007', 'activeEnergy', '有功电能', 'Active Energy', '有功电能', '电表', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"statistic","dataType":"DOUBLE","unit":{"unitCode":"kilowatt_hour","unitSymbol":"kWh"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6308, 0, 'PROPERTY', 'P-EL-0008', 'reactiveEnergy', '无功电能', 'Reactive Energy', '无功电能', '电表', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"statistic","dataType":"DOUBLE","unit":{"unitCode":"kvarh","unitSymbol":"kvarh"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6309, 0, 'PROPERTY', 'P-EL-0009', 'transformerTemp', '变压器温度', 'Transformer Temp', '变压器温度', '变压器', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":200,"min":-40},"dataType":"DOUBLE","unit":{"unitCode":"celsius","unitSymbol":"℃"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6310, 0, 'PROPERTY', 'P-EL-0010', 'loadRate', '负载率', 'Load Rate', '负载率', '变压器', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":150,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"percent","unitSymbol":"%"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6311, 0, 'PROPERTY', 'P-EL-0011', 'batteryVoltage', '电池电压', 'Battery Voltage', '电池电压', 'UPS/EPS', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":1000,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"volt","unitSymbol":"V"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6312, 0, 'PROPERTY', 'P-EL-0012', 'batteryRemaining', '电池剩余容量', 'Battery Remaining', '电池剩余容量', 'UPS/EPS', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":100,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"percent","unitSymbol":"%"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6313, 0, 'PROPERTY', 'P-EL-0013', 'remainingBackupTime', '剩余后备时间', 'Remaining Backup Time', '剩余后备时间', 'UPS/EPS', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":1440,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"minute","unitSymbol":"min"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6314, 0, 'PROPERTY', 'P-EL-0014', 'mainsStatus', '市电状态', 'Mains Status', '市电状态', 'UPS/EPS/发电机组', 'ENUM', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"runtime","dataType":"ENUM","enumItems":[{"label":"正常","value":0},{"label":"异常","value":1},{"label":"中断","value":2}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6315, 0, 'PROPERTY', 'P-EL-0015', 'switchStatus', '开关状态', 'Switch Status', '断路器开关状态', '配电柜/箱', 'BOOL', 'R', NULL, NULL, NULL, '{"accessMode":"R","boolLabels":{"false":"分闸","true":"合闸"},"category":"runtime","dataType":"BOOL"}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6316, 0, 'PROPERTY', 'P-EL-0016', 'illuminance', '照度', 'Illuminance', '照度', '照明设备', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":100000,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"lux","unitSymbol":"Lux"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6317, 0, 'PROPERTY', 'P-EL-0017', 'dimmingLevel', '调光级别', 'Dimming Level', '调光级别', '照明设备', 'DOUBLE', 'RW', NULL, NULL, NULL, '{"accessMode":"RW","category":"setting","constraints":{"max":100,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"percent","unitSymbol":"%"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6401, 0, 'PROPERTY', 'P-FP-0001', 'firePipePressure', '管网压力', 'Pipe Pressure', '消防管网压力', '消防水泵/消火栓', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":5,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"megapascal","unitSymbol":"MPa"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6402, 0, 'PROPERTY', 'P-FP-0002', 'fireWaterLevel', '消防水池液位', 'Fire Water Level', '消防水池液位', '消防水箱', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":10,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"meter","unitSymbol":"m"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6403, 0, 'PROPERTY', 'P-FP-0003', 'alarmStatus', '报警状态', 'Alarm Status', '火警状态', '火灾报警系统', 'ENUM', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"runtime","dataType":"ENUM","enumItems":[{"label":"正常","value":0},{"label":"火警","value":1},{"label":"预警","value":2},{"label":"故障","value":3}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6404, 0, 'PROPERTY', 'P-FP-0004', 'alarmZone', '报警区域', 'Alarm Zone', '报警区域', '火灾报警系统', 'TEXT', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"runtime","dataType":"TEXT","textMaxLength":128}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6405, 0, 'PROPERTY', 'P-FP-0005', 'sprinklerStatus', '喷头状态', 'Sprinkler Status', '喷头状态', '自动喷水系统', 'ENUM', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"runtime","dataType":"ENUM","enumItems":[{"label":"正常","value":0},{"label":"动作","value":1},{"label":"故障","value":2}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6406, 0, 'PROPERTY', 'P-FP-0006', 'smokeValveStatus', '排烟阀状态', 'Smoke Valve Status', '排烟阀状态', '防排烟系统', 'ENUM', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"runtime","dataType":"ENUM","enumItems":[{"label":"关闭","value":0},{"label":"开启","value":1},{"label":"故障","value":2}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6407, 0, 'PROPERTY', 'P-FP-0007', 'fireDoorStatus', '防火门状态', 'Fire Door Status', '防火门状态', '防火门', 'ENUM', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"runtime","dataType":"ENUM","enumItems":[{"label":"关闭","value":0},{"label":"开启","value":1},{"label":"故障","value":2}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6408, 0, 'PROPERTY', 'P-FP-0008', 'fireShutterPosition', '防火卷帘位置', 'Fire Shutter Position', '防火卷帘位置', '防火卷帘', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":100,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"percent","unitSymbol":"%"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6409, 0, 'PROPERTY', 'P-FP-0009', 'gasSuppressionStatus', '气体灭火状态', 'Gas Suppression Status', '气体灭火状态', '气体灭火系统', 'ENUM', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"runtime","dataType":"ENUM","enumItems":[{"label":"正常","value":0},{"label":"喷放","value":1},{"label":"故障","value":2}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6410, 0, 'PROPERTY', 'P-FP-0010', 'selfTestStatus', '系统自检状态', 'Self Test Status', '系统自检状态', '火灾报警系统', 'ENUM', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"runtime","dataType":"ENUM","enumItems":[{"label":"正常","value":0},{"label":"自检中","value":1},{"label":"自检异常","value":2}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6501, 0, 'PROPERTY', 'P-ELV-0001', 'currentFloor', '当前楼层', 'Current Floor', '电梯当前楼层', '电梯', 'INT', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":200,"min":-10},"dataType":"INT"}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6502, 0, 'PROPERTY', 'P-ELV-0002', 'travelDirection', '运行方向', 'Travel Direction', '电梯运行方向', '电梯', 'ENUM', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"runtime","dataType":"ENUM","enumItems":[{"label":"停止","value":0},{"label":"上行","value":1},{"label":"下行","value":2}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6503, 0, 'PROPERTY', 'P-ELV-0003', 'carPassengerCount', '轿厢内人数', 'Car Passenger Count', '轿厢内人数', '电梯', 'INT', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":50,"min":0},"dataType":"INT"}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6504, 0, 'PROPERTY', 'P-ELV-0004', 'currentLoad', '当前载重', 'Current Load', '当前载重', '电梯', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":20000,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"kilogram","unitSymbol":"kg"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6505, 0, 'PROPERTY', 'P-ELV-0005', 'loadPercentage', '载重百分比', 'Load Percentage', '载重百分比', '电梯', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":110,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"percent","unitSymbol":"%"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6506, 0, 'PROPERTY', 'P-ELV-0006', 'doorStatus', '门状态', 'Door Status', '门状态', '电梯', 'ENUM', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"runtime","dataType":"ENUM","enumItems":[{"label":"关门","value":0},{"label":"开门","value":1},{"label":"开门中","value":2},{"label":"关门中","value":3}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6507, 0, 'PROPERTY', 'P-ELV-0007', 'currentSpeed', '运行速度', 'Current Speed', '运行速度', '电梯', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":20,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"meter_per_second","unitSymbol":"m/s"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6508, 0, 'PROPERTY', 'P-ELV-0008', 'lockStatus', '锁梯状态', 'Lock Status', '锁梯状态', '电梯', 'BOOL', 'RW', NULL, NULL, NULL, '{"accessMode":"RW","boolLabels":{"false":"解锁","true":"锁梯"},"category":"setting","dataType":"BOOL"}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6509, 0, 'PROPERTY', 'P-ELV-0009', 'fireReturnStatus', '消防返回状态', 'Fire Return Status', '消防返回状态', '电梯', 'BOOL', 'R', NULL, NULL, NULL, '{"accessMode":"R","boolLabels":{"false":"正常","true":"消防返回"},"category":"runtime","dataType":"BOOL"}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6510, 0, 'PROPERTY', 'P-ELV-0010', 'escalatorDirection', '扶梯运行方向', 'Escalator Direction', '扶梯运行方向', '自动扶梯', 'ENUM', 'RW', NULL, NULL, NULL, '{"accessMode":"RW","category":"setting","dataType":"ENUM","enumItems":[{"label":"停止","value":0},{"label":"上行","value":1},{"label":"下行","value":2}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6601, 0, 'PROPERTY', 'P-SM-0001', 'onlineStatus', '在线状态', 'Online Status', '在线状态', '摄像机/门禁/传感器', 'BOOL', 'R', NULL, NULL, NULL, '{"accessMode":"R","boolLabels":{"false":"离线","true":"在线"},"category":"runtime","dataType":"BOOL"}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6602, 0, 'PROPERTY', 'P-SM-0002', 'signalStrength', '信号强度', 'Signal Strength', '信号强度', '无线设备', 'INT', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":0,"min":-120},"dataType":"INT","unit":{"unitCode":"dbm","unitSymbol":"dBm"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6603, 0, 'PROPERTY', 'P-SM-0003', 'cpuUsage', 'CPU使用率', 'CPU Usage', 'CPU使用率', '控制器/服务器', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":100,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"percent","unitSymbol":"%"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6604, 0, 'PROPERTY', 'P-SM-0004', 'memoryUsage', '内存使用率', 'Memory Usage', '内存使用率', '控制器/服务器', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":100,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"percent","unitSymbol":"%"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6605, 0, 'PROPERTY', 'P-SM-0005', 'storageUsage', '存储使用率', 'Storage Usage', '存储使用率', 'NVR/服务器', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","constraints":{"max":100,"min":0},"dataType":"DOUBLE","unit":{"unitCode":"percent","unitSymbol":"%"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6606, 0, 'PROPERTY', 'P-SM-0006', 'smDoorStatus', '门状态', 'Door Status', '门禁门状态', '门禁', 'ENUM', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"runtime","dataType":"ENUM","enumItems":[{"label":"关闭","value":0},{"label":"开启","value":1},{"label":"未关","value":2},{"label":"超时未关","value":3}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6607, 0, 'PROPERTY', 'P-SM-0007', 'armStatus', '布防状态', 'Arm Status', '入侵报警布防状态', '入侵报警', 'ENUM', 'RW', NULL, NULL, NULL, '{"accessMode":"RW","category":"setting","dataType":"ENUM","enumItems":[{"label":"撤防","value":0},{"label":"外出布防","value":1},{"label":"留守布防","value":2}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6608, 0, 'PROPERTY', 'P-SM-0008', 'measuredValue', '测量值', 'Measured Value', '传感器测量值', '传感器', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"measurement","dataType":"DOUBLE"}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6609, 0, 'PROPERTY', 'P-SM-0009', 'controlOutput', '控制输出值', 'Control Output', '控制器输出', '控制器', 'DOUBLE', 'RW', NULL, NULL, NULL, '{"accessMode":"RW","category":"setting","constraints":{"max":100,"min":0},"dataType":"DOUBLE"}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 6610, 0, 'PROPERTY', 'P-SM-0010', 'networkStatus', '网络连接状态', 'Network Status', '网络状态', '通信设备', 'ENUM', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"runtime","dataType":"ENUM","enumItems":[{"label":"断开","value":0},{"label":"已连接","value":1},{"label":"连接中","value":2}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7101, 0, 'EVENT', 'E-HVAC-0001', 'supplyTempHigh', '供水温度超高', 'Supply Temp High', '供水温度超过设定上限', '冷机/锅炉', NULL, NULL, 'ALERT', NULL, NULL, '{"level":"ALERT","outputParams":[{"dataType":"DOUBLE","key":"supplyTemp","unit":{"unitCode":"celsius","unitSymbol":"℃"}},{"dataType":"DOUBLE","key":"limitValue","unit":{"unitCode":"celsius","unitSymbol":"℃"}}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7102, 0, 'EVENT', 'E-HVAC-0002', 'supplyTempLow', '供水温度超低', 'Supply Temp Low', '供水温度低于设定下限', '冷机/锅炉', NULL, NULL, 'ALERT', NULL, NULL, '{"level":"ALERT","outputParams":[{"dataType":"DOUBLE","key":"supplyTemp","unit":{"unitCode":"celsius","unitSymbol":"℃"}},{"dataType":"DOUBLE","key":"limitValue","unit":{"unitCode":"celsius","unitSymbol":"℃"}}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7103, 0, 'EVENT', 'E-HVAC-0003', 'compressorHighPressure', '压缩机高压保护', 'Compressor High Pressure', '压缩机排气压力过高', '制冷机组', NULL, NULL, 'ERROR', NULL, NULL, '{"level":"ERROR","outputParams":[{"dataType":"DOUBLE","key":"pressure","unit":{"unitCode":"megapascal","unitSymbol":"MPa"}},{"dataType":"DOUBLE","key":"limitValue","unit":{"unitCode":"megapascal","unitSymbol":"MPa"}}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7104, 0, 'EVENT', 'E-HVAC-0004', 'compressorLowPressure', '压缩机低压保护', 'Compressor Low Pressure', '压缩机吸气压力过低', '制冷机组', NULL, NULL, 'ERROR', NULL, NULL, '{"level":"ERROR","outputParams":[{"dataType":"DOUBLE","key":"pressure","unit":{"unitCode":"megapascal","unitSymbol":"MPa"}},{"dataType":"DOUBLE","key":"limitValue","unit":{"unitCode":"megapascal","unitSymbol":"MPa"}}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7105, 0, 'EVENT', 'E-HVAC-0005', 'freezeProtection', '防冻保护', 'Freeze Protection', '水温接近冰点', '制冷机组/空调末端', NULL, NULL, 'ALERT', NULL, NULL, '{"level":"ALERT","outputParams":[{"dataType":"DOUBLE","key":"waterTemp","unit":{"unitCode":"celsius","unitSymbol":"℃"}}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7106, 0, 'EVENT', 'E-HVAC-0006', 'filterClogged', '过滤器堵塞', 'Filter Clogged', '过滤器压差过大', '通风机/新风机组', NULL, NULL, 'ALERT', NULL, NULL, '{"level":"ALERT","outputParams":[{"dataType":"DOUBLE","key":"pressureDiff","unit":{"unitCode":"pascal","unitSymbol":"Pa"}},{"dataType":"DOUBLE","key":"limitValue","unit":{"unitCode":"pascal","unitSymbol":"Pa"}}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7107, 0, 'EVENT', 'E-HVAC-0007', 'waterFlowSwitchOff', '水流开关断开', 'Water Flow Switch Off', '水流中断保护', '制冷机组/空调末端', NULL, NULL, 'ERROR', NULL, NULL, '{"level":"ERROR","outputParams":null}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7108, 0, 'EVENT', 'E-HVAC-0008', 'coolingTowerFanFault', '冷却塔风机故障', 'Cooling Tower Fan Fault', '冷却塔风机异常', '冷却塔', NULL, NULL, 'ERROR', NULL, NULL, '{"level":"ERROR","outputParams":[{"dataType":"INT","key":"fanIndex"},{"dataType":"TEXT","key":"faultDesc"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7201, 0, 'EVENT', 'E-WP-0001', 'highWaterLevel', '高液位告警', 'High Water Level', '液位超过高限', '水箱/水池', NULL, NULL, 'ALERT', NULL, NULL, '{"level":"ALERT","outputParams":[{"dataType":"DOUBLE","key":"waterLevel","unit":{"unitCode":"meter","unitSymbol":"m"}},{"dataType":"DOUBLE","key":"limitValue","unit":{"unitCode":"meter","unitSymbol":"m"}}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7202, 0, 'EVENT', 'E-WP-0002', 'lowWaterLevel', '低液位告警', 'Low Water Level', '液位低于低限', '水箱/水池', NULL, NULL, 'ALERT', NULL, NULL, '{"level":"ALERT","outputParams":[{"dataType":"DOUBLE","key":"waterLevel","unit":{"unitCode":"meter","unitSymbol":"m"}},{"dataType":"DOUBLE","key":"limitValue","unit":{"unitCode":"meter","unitSymbol":"m"}}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7203, 0, 'EVENT', 'E-WP-0003', 'overflowAlarm', '溢流告警', 'Overflow Alarm', '液位溢流', '水箱/水池', NULL, NULL, 'ERROR', NULL, NULL, '{"level":"ERROR","outputParams":[{"dataType":"DOUBLE","key":"waterLevel","unit":{"unitCode":"meter","unitSymbol":"m"}}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7204, 0, 'EVENT', 'E-WP-0004', 'pipeOverPressure', '管网超压', 'Pipe Over Pressure', '管道压力超高', '水泵/管网', NULL, NULL, 'ALERT', NULL, NULL, '{"level":"ALERT","outputParams":[{"dataType":"DOUBLE","key":"pressure","unit":{"unitCode":"megapascal","unitSymbol":"MPa"}},{"dataType":"DOUBLE","key":"limitValue","unit":{"unitCode":"megapascal","unitSymbol":"MPa"}}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7205, 0, 'EVENT', 'E-WP-0005', 'pipePressureLoss', '管网失压', 'Pipe Pressure Loss', '管道压力过低', '水泵/管网', NULL, NULL, 'ALERT', NULL, NULL, '{"level":"ALERT","outputParams":[{"dataType":"DOUBLE","key":"pressure","unit":{"unitCode":"megapascal","unitSymbol":"MPa"}},{"dataType":"DOUBLE","key":"limitValue","unit":{"unitCode":"megapascal","unitSymbol":"MPa"}}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7206, 0, 'EVENT', 'E-WP-0006', 'pumpOverload', '水泵过载', 'Pump Overload', '电机过载保护', '水泵', NULL, NULL, 'ERROR', NULL, NULL, '{"level":"ERROR","outputParams":[{"dataType":"DOUBLE","key":"current","unit":{"unitCode":"ampere","unitSymbol":"A"}},{"dataType":"DOUBLE","key":"ratedCurrent","unit":{"unitCode":"ampere","unitSymbol":"A"}}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7301, 0, 'EVENT', 'E-EL-0001', 'overCurrentProtection', '过流保护', 'Over Current Protection', '电流超过额定值', '配电柜', NULL, NULL, 'ERROR', NULL, NULL, '{"level":"ERROR","outputParams":[{"dataType":"DOUBLE","key":"current","unit":{"unitCode":"ampere","unitSymbol":"A"}},{"dataType":"DOUBLE","key":"ratedCurrent","unit":{"unitCode":"ampere","unitSymbol":"A"}}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7302, 0, 'EVENT', 'E-EL-0002', 'overVoltageProtection', '过压保护', 'Over Voltage Protection', '电压超过上限', '配电柜', NULL, NULL, 'ALERT', NULL, NULL, '{"level":"ALERT","outputParams":[{"dataType":"DOUBLE","key":"voltage","unit":{"unitCode":"volt","unitSymbol":"V"}},{"dataType":"DOUBLE","key":"limitValue","unit":{"unitCode":"volt","unitSymbol":"V"}}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7303, 0, 'EVENT', 'E-EL-0003', 'underVoltageProtection', '欠压保护', 'Under Voltage Protection', '电压低于下限', '配电柜', NULL, NULL, 'ALERT', NULL, NULL, '{"level":"ALERT","outputParams":[{"dataType":"DOUBLE","key":"voltage","unit":{"unitCode":"volt","unitSymbol":"V"}},{"dataType":"DOUBLE","key":"limitValue","unit":{"unitCode":"volt","unitSymbol":"V"}}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7304, 0, 'EVENT', 'E-EL-0004', 'transformerOverTemp', '变压器超温', 'Transformer Over Temp', '变压器温度过高', '变压器', NULL, NULL, 'ERROR', NULL, NULL, '{"level":"ERROR","outputParams":[{"dataType":"DOUBLE","key":"temperature","unit":{"unitCode":"celsius","unitSymbol":"℃"}},{"dataType":"DOUBLE","key":"limitValue","unit":{"unitCode":"celsius","unitSymbol":"℃"}}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7305, 0, 'EVENT', 'E-EL-0005', 'batteryLow', '电池电量低', 'Battery Low', '电池剩余容量不足', 'UPS/EPS', NULL, NULL, 'ALERT', NULL, NULL, '{"level":"ALERT","outputParams":[{"dataType":"DOUBLE","key":"remaining","unit":{"unitCode":"percent","unitSymbol":"%"}},{"dataType":"DOUBLE","key":"limitValue","unit":{"unitCode":"percent","unitSymbol":"%"}}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7306, 0, 'EVENT', 'E-EL-0006', 'mainsPowerLoss', '市电中断', 'Mains Power Loss', '市电供电中断', 'UPS/EPS/发电机组', NULL, NULL, 'ALERT', NULL, NULL, '{"level":"ALERT","outputParams":[{"dataType":"DOUBLE","key":"duration","unit":{"unitCode":"second","unitSymbol":"s"}}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7307, 0, 'EVENT', 'E-EL-0007', 'leakageAlarm', '漏电告警', 'Leakage Alarm', '漏电电流超标', '配电柜', NULL, NULL, 'ALERT', NULL, NULL, '{"level":"ALERT","outputParams":[{"dataType":"DOUBLE","key":"leakageCurrent","unit":{"unitCode":"milliampere","unitSymbol":"mA"}}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7401, 0, 'EVENT', 'E-FP-0001', 'fireAlarm', '火灾报警', 'Fire Alarm', '火灾探测器报警', '火灾报警系统', NULL, NULL, 'ERROR', NULL, NULL, '{"level":"ERROR","outputParams":[{"dataType":"TEXT","key":"zone"},{"dataType":"ENUM","key":"detectorType"},{"dataType":"TEXT","key":"detectorId"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7402, 0, 'EVENT', 'E-FP-0002', 'pipeUnderPressure', '管网欠压', 'Pipe Under Pressure', '消防管网压力不足', '消防水系统', NULL, NULL, 'ALERT', NULL, NULL, '{"level":"ALERT","outputParams":[{"dataType":"DOUBLE","key":"pressure","unit":{"unitCode":"megapascal","unitSymbol":"MPa"}},{"dataType":"DOUBLE","key":"limitValue","unit":{"unitCode":"megapascal","unitSymbol":"MPa"}}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7403, 0, 'EVENT', 'E-FP-0003', 'sprinklerActivated', '喷头动作', 'Sprinkler Activated', '喷头破裂喷水', '自动喷水系统', NULL, NULL, 'ALERT', NULL, NULL, '{"level":"ALERT","outputParams":[{"dataType":"TEXT","key":"zone"},{"dataType":"TEXT","key":"sprinklerId"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7404, 0, 'EVENT', 'E-FP-0004', 'gasDischarged', '气体喷放', 'Gas Discharged', '气体灭火剂喷放', '气体灭火系统', NULL, NULL, 'ERROR', NULL, NULL, '{"level":"ERROR","outputParams":[{"dataType":"TEXT","key":"zone"},{"dataType":"ENUM","key":"agentType"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7405, 0, 'EVENT', 'E-FP-0005', 'fireDoorNotClosed', '防火门未关闭', 'Fire Door Not Closed', '防火门超时未关', '防火门', NULL, NULL, 'ALERT', NULL, NULL, '{"level":"ALERT","outputParams":[{"dataType":"TEXT","key":"doorId"},{"dataType":"DOUBLE","key":"duration","unit":{"unitCode":"second","unitSymbol":"s"}}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7406, 0, 'EVENT', 'E-FP-0006', 'detectorOffline', '探测器离线', 'Detector Offline', '探测器通信中断', '火灾报警系统', NULL, NULL, 'ALERT', NULL, NULL, '{"level":"ALERT","outputParams":[{"dataType":"TEXT","key":"detectorId"},{"dataType":"TEXT","key":"zone"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7501, 0, 'EVENT', 'E-ELV-0001', 'trappedPassenger', '困人事件', 'Trapped Passenger', '电梯困人', '电梯', NULL, NULL, 'ERROR', NULL, NULL, '{"level":"ERROR","outputParams":[{"dataType":"INT","key":"floor"},{"dataType":"TEXT","key":"carId"},{"dataType":"DOUBLE","key":"duration","unit":{"unitCode":"minute","unitSymbol":"min"}}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7502, 0, 'EVENT', 'E-ELV-0002', 'overloadAlarm', '超载告警', 'Overload Alarm', '轿厢超载', '电梯', NULL, NULL, 'ALERT', NULL, NULL, '{"level":"ALERT","outputParams":[{"dataType":"DOUBLE","key":"currentLoad","unit":{"unitCode":"kilogram","unitSymbol":"kg"}},{"dataType":"INT","key":"ratedCapacity","unit":{"unitCode":"kilogram","unitSymbol":"kg"}}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7503, 0, 'EVENT', 'E-ELV-0003', 'doorLockFault', '门锁故障', 'Door Lock Fault', '层门/轿门锁故障', '电梯', NULL, NULL, 'ERROR', NULL, NULL, '{"level":"ERROR","outputParams":[{"dataType":"INT","key":"floor"},{"dataType":"TEXT","key":"faultDesc"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7504, 0, 'EVENT', 'E-ELV-0004', 'speedGovernorTriggered', '限速器动作', 'Speed Governor Triggered', '超速保护动作', '电梯', NULL, NULL, 'ERROR', NULL, NULL, '{"level":"ERROR","outputParams":[{"dataType":"DOUBLE","key":"currentSpeed","unit":{"unitCode":"meter_per_second","unitSymbol":"m/s"}},{"dataType":"DOUBLE","key":"ratedSpeed","unit":{"unitCode":"meter_per_second","unitSymbol":"m/s"}}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7505, 0, 'EVENT', 'E-ELV-0005', 'escalatorReverse', '扶梯逆行', 'Escalator Reverse', '扶梯非预期反向运行', '自动扶梯', NULL, NULL, 'ERROR', NULL, NULL, '{"level":"ERROR","outputParams":[{"dataType":"ENUM","key":"direction"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7601, 0, 'EVENT', 'E-SM-0001', 'illegalEntry', '非法闯入', 'Illegal Entry', '未授权开门', '门禁/入侵报警', NULL, NULL, 'ALERT', NULL, NULL, '{"level":"ALERT","outputParams":[{"dataType":"TEXT","key":"doorId"},{"dataType":"ENUM","key":"authMethod"},{"dataType":"DATE","key":"timestamp"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7602, 0, 'EVENT', 'E-SM-0002', 'doorOpenTimeout', '门超时未关', 'Door Open Timeout', '门长时间未关闭', '门禁', NULL, NULL, 'ALERT', NULL, NULL, '{"level":"ALERT","outputParams":[{"dataType":"TEXT","key":"doorId"},{"dataType":"DOUBLE","key":"duration","unit":{"unitCode":"second","unitSymbol":"s"}}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7603, 0, 'EVENT', 'E-SM-0003', 'videoSignalLost', '视频信号丢失', 'Video Signal Lost', '视频信号中断', '摄像机', NULL, NULL, 'ERROR', NULL, NULL, '{"level":"ERROR","outputParams":[{"dataType":"TEXT","key":"cameraId"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7604, 0, 'EVENT', 'E-SM-0004', 'storageInsufficient', '存储空间不足', 'Storage Insufficient', '存储即将满', 'NVR/服务器', NULL, NULL, 'ALERT', NULL, NULL, '{"level":"ALERT","outputParams":[{"dataType":"DOUBLE","key":"usage","unit":{"unitCode":"percent","unitSymbol":"%"}},{"dataType":"DOUBLE","key":"remaining","unit":{"unitCode":"gigabyte","unitSymbol":"GB"}}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7605, 0, 'EVENT', 'E-SM-0005', 'sensorValueAbnormal', '传感器数值异常', 'Sensor Value Abnormal', '测量值超出正常范围', '传感器', NULL, NULL, 'ALERT', NULL, NULL, '{"level":"ALERT","outputParams":[{"dataType":"TEXT","key":"parameterName"},{"dataType":"DOUBLE","key":"value"},{"dataType":"STRUCT","key":"normalRange"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 7606, 0, 'EVENT', 'E-SM-0006', 'controllerOffline', '控制器离线', 'Controller Offline', '控制器通信中断', 'DDC/PLC', NULL, NULL, 'ALERT', NULL, NULL, '{"level":"ALERT","outputParams":[{"dataType":"TEXT","key":"controllerId"},{"dataType":"DATE","key":"lastOnlineTime"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 8101, 0, 'SERVICE', 'S-HVAC-0001', 'setTemperatureSvc', '设定温度', 'Set Temperature', '设定目标温度', '空调末端', NULL, NULL, NULL, 'ASYNC', NULL, '{"callMode":"ASYNC","inputParams":[{"constraints":{"max":60,"min":-20},"dataType":"DOUBLE","key":"temperature","unit":{"unitCode":"celsius","unitSymbol":"℃"}}],"outputParams":[{"dataType":"BOOL","key":"success"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 8102, 0, 'SERVICE', 'S-HVAC-0002', 'setOperationMode', '切换运行模式', 'Set Operation Mode', '切换制冷/制热', '冷机/热泵/空调末端', NULL, NULL, NULL, 'ASYNC', NULL, '{"callMode":"ASYNC","inputParams":[{"dataType":"ENUM","key":"mode"}],"outputParams":[{"dataType":"BOOL","key":"success"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 8103, 0, 'SERVICE', 'S-HVAC-0003', 'setDamperPosition', '设定风阀开度', 'Set Damper Position', '调节风阀开度', '通风机/空调末端', NULL, NULL, NULL, 'ASYNC', NULL, '{"callMode":"ASYNC","inputParams":[{"constraints":{"max":100,"min":0},"dataType":"DOUBLE","key":"position","unit":{"unitCode":"percent","unitSymbol":"%"}}],"outputParams":[{"dataType":"BOOL","key":"success"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 8104, 0, 'SERVICE', 'S-HVAC-0004', 'setValvePosition', '设定水阀开度', 'Set Valve Position', '调节水阀开度', '空调末端/换热器', NULL, NULL, NULL, 'ASYNC', NULL, '{"callMode":"ASYNC","inputParams":[{"constraints":{"max":100,"min":0},"dataType":"DOUBLE","key":"position","unit":{"unitCode":"percent","unitSymbol":"%"}}],"outputParams":[{"dataType":"BOOL","key":"success"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 8105, 0, 'SERVICE', 'S-HVAC-0005', 'setFanFrequency', '设定风机频率', 'Set Fan Frequency', '调节变频器频率', '通风机/变频空调', NULL, NULL, NULL, 'ASYNC', NULL, '{"callMode":"ASYNC","inputParams":[{"constraints":{"max":100,"min":0},"dataType":"DOUBLE","key":"frequency","unit":{"unitCode":"hertz","unitSymbol":"Hz"}}],"outputParams":[{"dataType":"BOOL","key":"success"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 8106, 0, 'SERVICE', 'S-HVAC-0006', 'setFanSpeedSvc', '设定风速', 'Set Fan Speed', '设定风机档位', '空调末端', NULL, NULL, NULL, 'ASYNC', NULL, '{"callMode":"ASYNC","inputParams":[{"dataType":"ENUM","key":"speed"}],"outputParams":[{"dataType":"BOOL","key":"success"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 8201, 0, 'SERVICE', 'S-WP-0001', 'setPumpFrequency', '设定水泵频率', 'Set Pump Frequency', '调节水泵变频器频率', '变频水泵', NULL, NULL, NULL, 'ASYNC', NULL, '{"callMode":"ASYNC","inputParams":[{"constraints":{"max":100,"min":0},"dataType":"DOUBLE","key":"frequency","unit":{"unitCode":"hertz","unitSymbol":"Hz"}}],"outputParams":[{"dataType":"BOOL","key":"success"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 8202, 0, 'SERVICE', 'S-WP-0002', 'setWaterTempSvc', '设定水温', 'Set Water Temp', '设定目标热水温度', '热水设备', NULL, NULL, NULL, 'ASYNC', NULL, '{"callMode":"ASYNC","inputParams":[{"constraints":{"max":80,"min":20},"dataType":"DOUBLE","key":"temperature","unit":{"unitCode":"celsius","unitSymbol":"℃"}}],"outputParams":[{"dataType":"BOOL","key":"success"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 8203, 0, 'SERVICE', 'S-WP-0003', 'switchStandbyPump', '切换备用泵', 'Switch Standby Pump', '切换至指定备用泵', '水泵组', NULL, NULL, NULL, 'ASYNC', NULL, '{"callMode":"ASYNC","inputParams":[{"dataType":"INT","key":"pumpIndex"}],"outputParams":[{"dataType":"BOOL","key":"success"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 8301, 0, 'SERVICE', 'S-EL-0001', 'closeSwitch', '合闸', 'Close Switch', '合上指定回路开关', '配电柜/箱', NULL, NULL, NULL, 'ASYNC', NULL, '{"callMode":"ASYNC","inputParams":[{"dataType":"INT","key":"circuitIndex"}],"outputParams":[{"dataType":"BOOL","key":"success"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 8302, 0, 'SERVICE', 'S-EL-0002', 'openSwitch', '分闸', 'Open Switch', '断开指定回路开关', '配电柜/箱', NULL, NULL, NULL, 'ASYNC', NULL, '{"callMode":"ASYNC","inputParams":[{"dataType":"INT","key":"circuitIndex"}],"outputParams":[{"dataType":"BOOL","key":"success"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 8303, 0, 'SERVICE', 'S-EL-0003', 'setDimmingLevel', '设定调光级别', 'Set Dimming Level', '调节照明亮度', '照明设备', NULL, NULL, NULL, 'ASYNC', NULL, '{"callMode":"ASYNC","inputParams":[{"constraints":{"max":100,"min":0},"dataType":"DOUBLE","key":"level","unit":{"unitCode":"percent","unitSymbol":"%"}}],"outputParams":[{"dataType":"BOOL","key":"success"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 8304, 0, 'SERVICE', 'S-EL-0004', 'setLightingScene', '切换照明场景', 'Set Lighting Scene', '切换预设照明场景', '照明设备', NULL, NULL, NULL, 'ASYNC', NULL, '{"callMode":"ASYNC","inputParams":[{"dataType":"TEXT","key":"sceneId"}],"outputParams":[{"dataType":"BOOL","key":"success"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 8401, 0, 'SERVICE', 'S-FP-0001', 'startFirePump', '启动消防泵', 'Start Fire Pump', '远程启动消防水泵', '消防水泵', NULL, NULL, NULL, 'ASYNC', NULL, '{"callMode":"ASYNC","inputParams":[{"dataType":"INT","key":"pumpIndex"}],"outputParams":[{"dataType":"BOOL","key":"success"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 8402, 0, 'SERVICE', 'S-FP-0002', 'systemSelfTest', '系统自检', 'System Self Test', '触发系统自检', '火灾报警系统', NULL, NULL, NULL, 'ASYNC', NULL, '{"callMode":"ASYNC","outputParams":[{"dataType":"STRUCT","key":"result"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 8403, 0, 'SERVICE', 'S-FP-0003', 'silenceAlarm', '消音', 'Silence Alarm', '消除报警声响', '火灾报警系统', NULL, NULL, NULL, 'ASYNC', NULL, '{"callMode":"ASYNC","outputParams":[{"dataType":"BOOL","key":"success"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 8404, 0, 'SERVICE', 'S-FP-0004', 'resetSystem', '复位', 'Reset System', '系统复位', '火灾报警系统', NULL, NULL, NULL, 'ASYNC', NULL, '{"callMode":"ASYNC","outputParams":[{"dataType":"BOOL","key":"success"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 8501, 0, 'SERVICE', 'S-ELV-0001', 'callElevator', '呼梯', 'Call Elevator', '外呼电梯', '电梯', NULL, NULL, NULL, 'ASYNC', NULL, '{"callMode":"ASYNC","inputParams":[{"dataType":"INT","key":"floor"},{"dataType":"ENUM","key":"direction"}],"outputParams":[{"dataType":"BOOL","key":"success"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 8502, 0, 'SERVICE', 'S-ELV-0002', 'lockElevator', '锁梯', 'Lock Elevator', '锁定/解锁电梯', '电梯', NULL, NULL, NULL, 'ASYNC', NULL, '{"callMode":"ASYNC","inputParams":[{"dataType":"BOOL","key":"lock"}],"outputParams":[{"dataType":"BOOL","key":"success"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 8503, 0, 'SERVICE', 'S-ELV-0003', 'fireReturn', '消防返回', 'Fire Return', '触发消防迫降', '电梯', NULL, NULL, NULL, 'ASYNC', NULL, '{"callMode":"ASYNC","outputParams":[{"dataType":"BOOL","key":"success"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 8504, 0, 'SERVICE', 'S-ELV-0004', 'setEscalatorDirection', '切换扶梯方向', 'Set Escalator Direction', '切换扶梯运行方向', '自动扶梯', NULL, NULL, NULL, 'ASYNC', NULL, '{"callMode":"ASYNC","inputParams":[{"dataType":"ENUM","key":"direction"}],"outputParams":[{"dataType":"BOOL","key":"success"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 8601, 0, 'SERVICE', 'S-SM-0001', 'remoteOpenDoor', '远程开门', 'Remote Open Door', '远程开启指定门', '门禁', NULL, NULL, NULL, 'ASYNC', NULL, '{"callMode":"ASYNC","inputParams":[{"dataType":"TEXT","key":"doorId"}],"outputParams":[{"dataType":"BOOL","key":"success"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 8602, 0, 'SERVICE', 'S-SM-0002', 'armSystem', '布防', 'Arm System', '设定布防模式', '入侵报警', NULL, NULL, NULL, 'ASYNC', NULL, '{"callMode":"ASYNC","inputParams":[{"dataType":"ENUM","key":"armMode"}],"outputParams":[{"dataType":"BOOL","key":"success"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 8603, 0, 'SERVICE', 'S-SM-0003', 'ptzControl', '云台控制', 'PTZ Control', '控制云台方向和速度', '摄像机', NULL, NULL, NULL, 'ASYNC', NULL, '{"callMode":"ASYNC","inputParams":[{"dataType":"ENUM","key":"action"},{"constraints":{"max":10,"min":1},"dataType":"INT","key":"speed"}],"outputParams":[{"dataType":"BOOL","key":"success"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 8604, 0, 'SERVICE', 'S-SM-0004', 'setControlOutput', '设定控制输出', 'Set Control Output', '设定模拟输出值', '控制器', NULL, NULL, NULL, 'ASYNC', NULL, '{"callMode":"ASYNC","inputParams":[{"dataType":"INT","key":"outputIndex"},{"constraints":{"max":100,"min":0},"dataType":"DOUBLE","key":"value"}],"outputParams":[{"dataType":"BOOL","key":"success"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 8605, 0, 'SERVICE', 'S-SM-0005', 'publishMessage', '发布信息', 'Publish Message', '发布显示内容', '信息发布', NULL, NULL, NULL, 'ASYNC', NULL, '{"callMode":"ASYNC","inputParams":[{"dataType":"TEXT","key":"content"},{"dataType":"INT","key":"duration","unit":{"unitCode":"second","unitSymbol":"s"}}],"outputParams":[{"dataType":"BOOL","key":"success"}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 9101, 0, 'RELATION', 'R-HVAC-0001', 'copDerivedFrom', '实时COP派生关系', 'COP Derived From', '实时COP由实时制冷量与功率派生 (currentCOP = currentCoolingCapacity / currentPower)', '制冷机组', NULL, NULL, NULL, NULL, 'derivedFrom', '{"cardinality":"oneToOne","directional":true,"relationType":"derivedFrom","source":{"identifier":"currentCOP","kind":"feature"},"target":{"identifier":"currentCoolingCapacity","kind":"feature"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 9102, 0, 'RELATION', 'R-HVAC-0002', 'supplyReturnTempPair', '供回水温度关系', 'Supply-Return Temp Pair', '供水温度与回水温度的关联，用于温差计算', '冷机/锅炉/换热器', NULL, NULL, NULL, NULL, 'relatedTo', '{"cardinality":"oneToOne","directional":false,"relationType":"relatedTo","source":{"identifier":"supplyWaterTemp","kind":"feature"},"target":{"identifier":"returnWaterTemp","kind":"feature"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 9103, 0, 'RELATION', 'R-HVAC-0003', 'compressorStatusDependsOnPower', '压缩机状态依赖开关', 'Compressor Depends On Power', '压缩机运行状态依赖于设备开关', '制冷机组', NULL, NULL, NULL, NULL, 'dependsOn', '{"cardinality":"manyToOne","directional":true,"relationType":"dependsOn","source":{"identifier":"compressorStatus","kind":"feature"},"target":{"identifier":"powerSwitch","kind":"feature"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 9301, 0, 'RELATION', 'R-EL-0001', 'loadRateFromActivePower', '负载率派生关系', 'Load Rate Derived', '变压器负载率由有功功率/额定功率派生', '变压器', NULL, NULL, NULL, NULL, 'derivedFrom', '{"cardinality":"oneToOne","directional":true,"relationType":"derivedFrom","source":{"identifier":"loadRate","kind":"feature"},"target":{"identifier":"activePower","kind":"feature"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 9401, 0, 'RELATION', 'R-FP-0001', 'fireAlarmTriggersDoorRelease', '火警触发防火门释放', 'Fire Alarm Triggers Door Release', '火灾报警时触发防火门释放（控制链）', '消防系统', NULL, NULL, NULL, NULL, 'controls', '{"cardinality":"oneToMany","directional":true,"relationType":"controls","source":{"identifier":"alarmStatus","kind":"feature"},"target":{"identifier":"fireDoorStatus","kind":"feature"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 9501, 0, 'RELATION', 'R-ELV-0001', 'fireReturnControlsLockStatus', '消防返回控制锁梯', 'Fire Return Controls Lock', '消防返回触发时电梯进入锁梯状态', '电梯', NULL, NULL, NULL, NULL, 'controls', '{"cardinality":"oneToOne","directional":true,"relationType":"controls","source":{"identifier":"fireReturnStatus","kind":"feature"},"target":{"identifier":"lockStatus","kind":"feature"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51101, 0, 'PROPERTY', 'RAT-HVAC-CH-001', 'ratedCoolingCapacity', '额定制冷量', 'Rated Cooling Capacity', '铭牌额定制冷量', '制冷机组', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":50000,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"kilowatt","unitSymbol":"kW"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51102, 0, 'PROPERTY', 'RAT-HVAC-CH-002', 'ratedHeatingCapacity', '额定制热量', 'Rated Heating Capacity', '热泵机组适用', '制冷机组', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":50000,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"kilowatt","unitSymbol":"kW"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51103, 0, 'PROPERTY', 'RAT-HVAC-CH-003', 'ratedInputPower', '额定输入功率', 'Rated Input Power', '额定工况输入功率', '制冷机组', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":10000,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"kilowatt","unitSymbol":"kW"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51104, 0, 'PROPERTY', 'RAT-HVAC-CH-004', 'ratedCOP', '额定COP', 'Rated COP', '额定性能系数', '制冷机组', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":15,"min":0},"dataType":"DOUBLE","isRated":true}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51105, 0, 'PROPERTY', 'RAT-HVAC-CH-005', 'ratedEER', '额定EER', 'Rated EER', '额定能效比', '制冷机组', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":15,"min":0},"dataType":"DOUBLE","isRated":true}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51106, 0, 'PROPERTY', 'RAT-HVAC-CH-006', 'ratedIPLV', '额定IPLV', 'Rated IPLV', '综合部分负荷性能系数', '制冷机组', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":15,"min":0},"dataType":"DOUBLE","isRated":true}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51107, 0, 'PROPERTY', 'RAT-HVAC-CH-007', 'refrigerantType', '制冷剂类型', 'Refrigerant Type', '制冷剂种类', '制冷机组', 'ENUM', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","dataType":"ENUM","enumItems":[{"label":"R410A","value":0},{"label":"R134a","value":1},{"label":"R32","value":2},{"label":"R407C","value":3},{"label":"R290","value":4},{"label":"R717","value":5},{"label":"其他","value":6}],"isRated":true}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51108, 0, 'PROPERTY', 'RAT-HVAC-CH-008', 'refrigerantCharge', '制冷剂充注量', 'Refrigerant Charge', '制冷剂充注质量', '制冷机组', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":5000,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"kilogram","unitSymbol":"kg"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51109, 0, 'PROPERTY', 'RAT-HVAC-CH-009', 'ratedChilledWaterFlow', '冷冻水额定流量', 'Rated Chilled Water Flow', '额定冷冻水流量', '制冷机组', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":10000,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"cubic_meter_per_hour","unitSymbol":"m³/h"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51110, 0, 'PROPERTY', 'RAT-HVAC-CH-010', 'ratedCondenserWaterFlow', '冷却水额定流量', 'Rated Condenser Water Flow', '额定冷却水流量', '制冷机组', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":10000,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"cubic_meter_per_hour","unitSymbol":"m³/h"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51111, 0, 'PROPERTY', 'RAT-HVAC-CH-011', 'ratedChilledWaterTemp', '冷冻水进出水温度', 'Rated Chilled Water Temp', '额定冷冻水供回水温度', '制冷机组', 'STRUCT', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","dataType":"STRUCT","isRated":true,"structFields":[{"constraints":{"max":40,"min":0},"dataType":"DOUBLE","key":"inlet","unit":{"unitCode":"celsius","unitSymbol":"℃"}},{"constraints":{"max":40,"min":0},"dataType":"DOUBLE","key":"outlet","unit":{"unitCode":"celsius","unitSymbol":"℃"}}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51112, 0, 'PROPERTY', 'RAT-HVAC-CH-012', 'ratedCondenserWaterTemp', '冷却水进出水温度', 'Rated Condenser Water Temp', '额定冷却水进出水温度', '制冷机组', 'STRUCT', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","dataType":"STRUCT","isRated":true,"structFields":[{"constraints":{"max":50,"min":0},"dataType":"DOUBLE","key":"inlet","unit":{"unitCode":"celsius","unitSymbol":"℃"}},{"constraints":{"max":50,"min":0},"dataType":"DOUBLE","key":"outlet","unit":{"unitCode":"celsius","unitSymbol":"℃"}}]}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51113, 0, 'PROPERTY', 'RAT-HVAC-CH-013', 'compressorType', '压缩机类型', 'Compressor Type', '压缩机形式', '制冷机组', 'ENUM', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","dataType":"ENUM","enumItems":[{"label":"离心式","value":0},{"label":"螺杆式","value":1},{"label":"涡旋式","value":2},{"label":"活塞式","value":3},{"label":"转子式","value":4}],"isRated":true}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51114, 0, 'PROPERTY', 'RAT-HVAC-CH-014', 'compressorCount', '压缩机数量', 'Compressor Count', '压缩机台数', '制冷机组', 'INT', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":20,"min":1},"dataType":"INT","isRated":true}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51115, 0, 'PROPERTY', 'RAT-HVAC-CH-015', 'ratedWaterResistance', '额定水阻力', 'Rated Water Resistance', '额定水侧压降', '制冷机组', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":1000,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"kpa","unitSymbol":"kPa"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51201, 0, 'PROPERTY', 'RAT-HVAC-BL-001', 'ratedThermalPower', '额定热功率', 'Rated Thermal Power', '额定热输出功率', '锅炉', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":100,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"megawatt","unitSymbol":"MW"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51202, 0, 'PROPERTY', 'RAT-HVAC-BL-002', 'ratedWorkingPressure', '额定工作压力', 'Rated Working Pressure', '额定蒸汽/热水压力', '锅炉', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":10,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"megapascal","unitSymbol":"MPa"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51203, 0, 'PROPERTY', 'RAT-HVAC-BL-003', 'ratedSupplyTemp', '额定供水温度', 'Rated Supply Temp', '额定出口水温', '锅炉', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":250,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"celsius","unitSymbol":"℃"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51204, 0, 'PROPERTY', 'RAT-HVAC-BL-004', 'ratedReturnTemp', '额定回水温度', 'Rated Return Temp', '额定回水温度', '锅炉', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":200,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"celsius","unitSymbol":"℃"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51205, 0, 'PROPERTY', 'RAT-HVAC-BL-005', 'ratedEfficiency', '额定热效率', 'Rated Efficiency', '额定热效率', '锅炉', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":100,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"percent","unitSymbol":"%"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51206, 0, 'PROPERTY', 'RAT-HVAC-BL-006', 'fuelType', '燃料类型', 'Fuel Type', '燃料种类', '锅炉', 'ENUM', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","dataType":"ENUM","enumItems":[{"label":"天然气","value":0},{"label":"柴油","value":1},{"label":"电","value":2},{"label":"生物质","value":3},{"label":"液化石油气","value":4},{"label":"煤","value":5}],"isRated":true}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51207, 0, 'PROPERTY', 'RAT-HVAC-BL-007', 'ratedFuelConsumption', '额定燃料耗量', 'Rated Fuel Consumption', '额定工况燃料消耗', '锅炉', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","dataType":"DOUBLE","isRated":true}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51208, 0, 'PROPERTY', 'RAT-HVAC-BL-008', 'ratedWaterVolume', '额定水容量', 'Rated Water Volume', '锅炉容水量', '锅炉', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":100,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"cubic_meter","unitSymbol":"m³"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51301, 0, 'PROPERTY', 'RAT-HVAC-HX-001', 'ratedHeatTransfer', '额定换热量', 'Rated Heat Transfer', '额定换热能力', '换热器', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":50000,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"kilowatt","unitSymbol":"kW"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51302, 0, 'PROPERTY', 'RAT-HVAC-HX-002', 'heatExchangerType', '换热器类型', 'Heat Exchanger Type', '换热器形式', '换热器', 'ENUM', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","dataType":"ENUM","enumItems":[{"label":"板式","value":0},{"label":"壳管式","value":1},{"label":"容积式","value":2},{"label":"翅片式","value":3}],"isRated":true}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51303, 0, 'PROPERTY', 'RAT-HVAC-HX-003', 'ratedPrimaryFlow', '一次侧额定流量', 'Rated Primary Flow', '一次侧额定流量', '换热器', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":10000,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"cubic_meter_per_hour","unitSymbol":"m³/h"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51304, 0, 'PROPERTY', 'RAT-HVAC-HX-004', 'ratedSecondaryFlow', '二次侧额定流量', 'Rated Secondary Flow', '二次侧额定流量', '换热器', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":10000,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"cubic_meter_per_hour","unitSymbol":"m³/h"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51307, 0, 'PROPERTY', 'RAT-HVAC-HX-007', 'hxRatedWorkingPressure', '额定工作压力', 'Rated Working Pressure (HX)', '额定承压', '换热器', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":10,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"megapascal","unitSymbol":"MPa"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51401, 0, 'PROPERTY', 'RAT-HVAC-TU-001', 'tuRatedCoolingCapacity', '空调末端额定制冷量', 'TU Rated Cooling Capacity', '末端额定制冷量', '空调末端', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":500,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"kilowatt","unitSymbol":"kW"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51402, 0, 'PROPERTY', 'RAT-HVAC-TU-002', 'tuRatedHeatingCapacity', '空调末端额定制热量', 'TU Rated Heating Capacity', '末端额定制热量', '空调末端', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":500,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"kilowatt","unitSymbol":"kW"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51403, 0, 'PROPERTY', 'RAT-HVAC-TU-003', 'ratedAirVolume', '额定风量', 'Rated Air Volume', '额定风量', '空调末端', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":100000,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"cubic_meter_per_hour","unitSymbol":"m³/h"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51404, 0, 'PROPERTY', 'RAT-HVAC-TU-004', 'ratedExternalPressure', '额定余压', 'Rated External Pressure', '额定机外余压', '空调末端', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":2000,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"pascal","unitSymbol":"Pa"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51501, 0, 'PROPERTY', 'RAT-HVAC-FN-001', 'fnRatedAirVolume', '通风机额定风量', 'Fan Rated Air Volume', '额定工况风量', '通风机', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":500000,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"cubic_meter_per_hour","unitSymbol":"m³/h"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51502, 0, 'PROPERTY', 'RAT-HVAC-FN-002', 'ratedTotalPressure', '额定全压', 'Rated Total Pressure', '额定全压', '通风机', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":10000,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"pascal","unitSymbol":"Pa"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51503, 0, 'PROPERTY', 'RAT-HVAC-FN-003', 'ratedStaticPressure', '额定静压', 'Rated Static Pressure', '额定静压', '通风机', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":10000,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"pascal","unitSymbol":"Pa"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51504, 0, 'PROPERTY', 'RAT-HVAC-FN-004', 'fanType', '风机类型', 'Fan Type', '风机形式', '通风机', 'ENUM', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","dataType":"ENUM","enumItems":[{"label":"离心式","value":0},{"label":"轴流式","value":1},{"label":"混流式","value":2},{"label":"斜流式","value":3},{"label":"屋顶式","value":4}],"isRated":true}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51601, 0, 'PROPERTY', 'RAT-HVAC-CT-001', 'ctRatedWaterFlow', '冷却塔额定水量', 'CT Rated Water Flow', '额定冷却水量', '冷却塔', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":10000,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"cubic_meter_per_hour","unitSymbol":"m³/h"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51602, 0, 'PROPERTY', 'RAT-HVAC-CT-002', 'ctRatedInletTemp', '冷却塔额定进水温度', 'CT Rated Inlet Temp', '额定进水温度', '冷却塔', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":60,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"celsius","unitSymbol":"℃"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51603, 0, 'PROPERTY', 'RAT-HVAC-CT-003', 'ctRatedOutletTemp', '冷却塔额定出水温度', 'CT Rated Outlet Temp', '额定出水温度', '冷却塔', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":50,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"celsius","unitSymbol":"℃"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 51604, 0, 'PROPERTY', 'RAT-HVAC-CT-004', 'ratedWetBulbTemp', '额定湿球温度', 'Rated Wet Bulb Temp', '设计湿球温度', '冷却塔', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":40,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"celsius","unitSymbol":"℃"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 52101, 0, 'PROPERTY', 'RAT-WP-PM-001', 'ratedFlowRate', '额定流量', 'Rated Flow Rate', '额定工况流量', '水泵', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":10000,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"cubic_meter_per_hour","unitSymbol":"m³/h"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 52102, 0, 'PROPERTY', 'RAT-WP-PM-002', 'ratedHead', '额定扬程', 'Rated Head', '额定扬程', '水泵', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":500,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"meter","unitSymbol":"m"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 52103, 0, 'PROPERTY', 'RAT-WP-PM-003', 'pumpType', '水泵类型', 'Pump Type', '水泵形式', '水泵', 'ENUM', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","dataType":"ENUM","enumItems":[{"label":"离心式","value":0},{"label":"轴流式","value":1},{"label":"混流式","value":2},{"label":"潜水式","value":3},{"label":"管道式","value":4},{"label":"隔膜式","value":5}],"isRated":true}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 53101, 0, 'PROPERTY', 'RAT-EL-TR-001', 'ratedCapacity', '额定容量', 'Rated Capacity', '额定视在功率', '变压器', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":100000,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"kva","unitSymbol":"kVA"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 53102, 0, 'PROPERTY', 'RAT-EL-TR-002', 'ratedPrimaryVoltage', '一次侧额定电压', 'Rated Primary Voltage', '高压侧额定电压', '变压器', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":500,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"kilovolt","unitSymbol":"kV"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 53103, 0, 'PROPERTY', 'RAT-EL-TR-003', 'ratedSecondaryVoltage', '二次侧额定电压', 'Rated Secondary Voltage', '低压侧额定电压', '变压器', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":100,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"kilovolt","unitSymbol":"kV"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 55101, 0, 'PROPERTY', 'RAT-ELV-EL-001', 'elvRatedCapacity', '电梯额定载重量', 'Elevator Rated Capacity', '额定载重', '电梯', 'INT', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":20000,"min":0},"dataType":"INT","isRated":true,"unit":{"unitCode":"kilogram","unitSymbol":"kg"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 55102, 0, 'PROPERTY', 'RAT-ELV-EL-002', 'ratedSpeed', '电梯额定速度', 'Elevator Rated Speed', '额定运行速度', '电梯', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":20,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"meter_per_second","unitSymbol":"m/s"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 55103, 0, 'PROPERTY', 'RAT-ELV-EL-003', 'floorCount', '层站数', 'Floor Count', '服务楼层总数', '电梯', 'INT', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":200,"min":1},"dataType":"INT","isRated":true}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 55104, 0, 'PROPERTY', 'RAT-ELV-EL-004', 'liftHeight', '提升高度', 'Lift Height', '总垂直行程', '电梯', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":600,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"meter","unitSymbol":"m"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 55105, 0, 'PROPERTY', 'RAT-ELV-EL-005', 'elevatorType', '电梯类型', 'Elevator Type', '电梯用途分类', '电梯', 'ENUM', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","dataType":"ENUM","enumItems":[{"label":"乘客","value":0},{"label":"载货","value":1},{"label":"医用","value":2},{"label":"汽车","value":3},{"label":"消防","value":4},{"label":"观光","value":5},{"label":"防爆","value":6}],"isRated":true}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 56101, 0, 'PROPERTY', 'RAT-SM-CM-001', 'resolution', '分辨率', 'Resolution', '视频分辨率', '摄像机', 'ENUM', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","dataType":"ENUM","enumItems":[{"label":"720P","value":0},{"label":"1080P","value":1},{"label":"2K","value":2},{"label":"4K","value":3},{"label":"8K","value":4}],"isRated":true}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 56102, 0, 'PROPERTY', 'RAT-SM-CM-002', 'sensorType', '传感器类型', 'Sensor Type', '图像传感器类型', '摄像机', 'ENUM', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","dataType":"ENUM","enumItems":[{"label":"CMOS","value":0},{"label":"CCD","value":1}],"isRated":true}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 57101, 0, 'PROPERTY', 'RAT-AX-PV-001', 'ratedPeakPower', '额定峰值功率', 'Rated Peak Power', '额定峰值功率', '光伏', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":10000,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"kwp","unitSymbol":"kWp"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 57102, 0, 'PROPERTY', 'RAT-AX-PV-002', 'moduleEfficiency', '组件转换效率', 'Module Efficiency', '光伏组件转换效率', '光伏', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":50,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"percent","unitSymbol":"%"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 57201, 0, 'PROPERTY', 'RAT-AX-CP-001', 'chargingType', '充电类型', 'Charging Type', '充电模式', '充电桩', 'ENUM', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","dataType":"ENUM","enumItems":[{"label":"交流慢充","value":0},{"label":"直流快充","value":1}],"isRated":true}')
ON CONFLICT (tenant_id, code) DO NOTHING;

INSERT INTO public.thingmodel_features (created_at, updated_at, is_enabled, sort_order, tenant_id, feature_type, code, identifier, name, name_en, description, applicable_scope, data_type, access_mode, event_level, call_mode, relation_type, spec)
VALUES (now(), now(), true, 57202, 0, 'PROPERTY', 'RAT-AX-CP-002', 'cpRatedPower', '充电桩额定功率', 'CP Rated Power', '额定充电功率', '充电桩', 'DOUBLE', 'R', NULL, NULL, NULL, '{"accessMode":"R","category":"rated","constraints":{"max":500,"min":0},"dataType":"DOUBLE","isRated":true,"unit":{"unitCode":"kilowatt","unitSymbol":"kW"}}')
ON CONFLICT (tenant_id, code) DO NOTHING;

COMMIT;

-- 共 255 条种子数据 / Total 255 seed records