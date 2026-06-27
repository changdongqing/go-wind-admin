/**
 * 特征模块通用常量 / Feature module shared constants.
 *
 * 复用 unit 模块的"启用/禁用"配色与翻译键约定；
 * 新增特征类型、数据类型、访问模式、事件级别、调用模式、关系类型、属性类别 等映射。
 *
 * Mirrors unit module's enable/disable helpers, plus feature-domain enums.
 */

type TFn = (key: string, options?: Record<string, any>) => string;

// ========== 启用/禁用（与 unit 模块一致，复用导出） ==========
export {
  getEnableColor,
  getEnableLabel,
  enableBoolOptions,
  enableBoolRadioOptions,
} from '../unit/constants';

// ========== 特征类型 / Feature type ==========

export const FEATURE_TYPES = ['PROPERTY', 'EVENT', 'SERVICE', 'RELATION'] as const;
export type FeatureType = (typeof FEATURE_TYPES)[number];

export const featureTypeColor: Record<FeatureType, string> = {
  PROPERTY: 'blue',
  EVENT: 'orange',
  SERVICE: 'green',
  RELATION: 'purple',
};

export const featureTypeIcon: Record<FeatureType, string> = {
  PROPERTY: 'lucide:circle-dot',
  EVENT: 'lucide:bell',
  SERVICE: 'lucide:cog',
  RELATION: 'lucide:share-2',
};

export function featureTypeOptions(t: TFn) {
  return FEATURE_TYPES.map((v) => ({
    label: t(`featureTypeMap.${v}`, { defaultValue: v }),
    value: v,
  }));
}

export function getFeatureTypeLabel(t: TFn, v?: string): string {
  if (!v) return '';
  return t(`featureTypeMap.${v}`, { defaultValue: v });
}

// ========== 数据类型 / Data type ==========

export const DATA_TYPES = [
  'INT',
  'FLOAT',
  'DOUBLE',
  'BOOL',
  'ENUM',
  'TEXT',
  'DATE',
  'STRUCT',
  'ARRAY',
] as const;
export type DataType = (typeof DATA_TYPES)[number];

export function dataTypeOptions(t: TFn) {
  return DATA_TYPES.map((v) => ({
    label: t(`dataTypeMap.${v}`, { defaultValue: v }),
    value: v,
  }));
}

export function getDataTypeLabel(t: TFn, v?: string): string {
  if (!v) return '';
  return t(`dataTypeMap.${v}`, { defaultValue: v });
}

// ========== 访问模式 / Access mode ==========

export const ACCESS_MODES = ['R', 'RW'] as const;
export type AccessMode = (typeof ACCESS_MODES)[number];

export const accessModeColor: Record<AccessMode, string> = {
  R: 'default',
  RW: 'processing',
};

export function accessModeOptions(t: TFn) {
  return ACCESS_MODES.map((v) => ({
    label: t(`accessModeMap.${v}`, { defaultValue: v }),
    value: v,
  }));
}

// ========== 事件级别 / Event level ==========

export const EVENT_LEVELS = ['INFO', 'ALERT', 'ERROR'] as const;
export type EventLevel = (typeof EVENT_LEVELS)[number];

export const eventLevelColor: Record<EventLevel, string> = {
  INFO: 'blue',
  ALERT: 'orange',
  ERROR: 'red',
};

export function eventLevelOptions(t: TFn) {
  return EVENT_LEVELS.map((v) => ({
    label: t(`eventLevelMap.${v}`, { defaultValue: v }),
    value: v,
  }));
}

// ========== 调用模式 / Call mode ==========

export const CALL_MODES = ['ASYNC', 'SYNC'] as const;
export type CallMode = (typeof CALL_MODES)[number];

export const callModeColor: Record<CallMode, string> = {
  ASYNC: 'cyan',
  SYNC: 'geekblue',
};

export function callModeOptions(t: TFn) {
  return CALL_MODES.map((v) => ({
    label: t(`callModeMap.${v}`, { defaultValue: v }),
    value: v,
  }));
}

// ========== 关系类型 / Relation type ==========

export const RELATION_TYPES = [
  'partOf',
  'feeds',
  'suppliedBy',
  'controls',
  'controlledBy',
  'derivedFrom',
  'dependsOn',
  'locatedIn',
  'monitors',
  'relatedTo',
] as const;
export type RelationType = (typeof RELATION_TYPES)[number];

export function relationTypeOptions(t: TFn) {
  return RELATION_TYPES.map((v) => ({
    label: t(`relationTypeMap.${v}`, { defaultValue: v }),
    value: v,
  }));
}

// ========== 基数 / Cardinality ==========

export const CARDINALITIES = ['oneToOne', 'oneToMany', 'manyToOne', 'manyToMany'] as const;
export type Cardinality = (typeof CARDINALITIES)[number];

export function cardinalityOptions(t: TFn) {
  return CARDINALITIES.map((v) => ({
    label: t(`cardinalityMap.${v}`, { defaultValue: v }),
    value: v,
  }));
}

// ========== 属性类别 / Property category ==========

export const PROPERTY_CATEGORIES = [
  'runtime',
  'measurement',
  'setting',
  'rated',
  'statistic',
  'environment',
] as const;
export type PropertyCategory = (typeof PROPERTY_CATEGORIES)[number];

export function propertyCategoryOptions(t: TFn) {
  return PROPERTY_CATEGORIES.map((v) => ({
    label: t(`propertyCategoryMap.${v}`, { defaultValue: v }),
    value: v,
  }));
}

// ========== EntityRef kind ==========

export const ENTITY_KINDS = ['feature', 'external'] as const;
export type EntityKind = (typeof ENTITY_KINDS)[number];

export function entityKindOptions(t: TFn) {
  return ENTITY_KINDS.map((v) => ({
    label: t(`entityKindMap.${v}`, { defaultValue: v }),
    value: v,
  }));
}
