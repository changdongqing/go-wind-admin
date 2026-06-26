/**
 * 单位模块通用常量 / Unit module shared constants.
 *
 * 复用 dict 模块的"启用/禁用"配色与翻译键约定，并新增换算类型映射。
 * Mirrors dict module's enable/disable helpers, plus conversion-type Tag colors.
 */

type TFn = (key: string, options?: Record<string, any>) => string;

// ========== 启用/禁用（与 dict 模块一致） ==========

export function getEnableColor(isEnabled: boolean): string {
  return isEnabled ? 'success' : 'error';
}

export function getEnableLabel(t: TFn, isEnabled: boolean): string {
  return isEnabled ? t('statusMap.true') : t('statusMap.false');
}

export function enableBoolOptions(t: TFn) {
  return [
    { label: t('statusMap.true'), value: 'true' },
    { label: t('statusMap.false'), value: 'false' },
  ];
}

export function enableBoolRadioOptions(t: TFn) {
  return [
    { label: t('statusMap.true'), value: true },
    { label: t('statusMap.false'), value: false },
  ];
}

// ========== 换算类型 / Conversion type ==========

export const CONVERSION_TYPES = [
  'LINEAR',
  'AFFINE',
  'LOGARITHMIC',
  'CONDITIONAL',
  'NONE',
] as const;

export type ConversionType = (typeof CONVERSION_TYPES)[number];

export const conversionTypeColor: Record<ConversionType, string> = {
  LINEAR: 'blue',
  AFFINE: 'purple',
  LOGARITHMIC: 'orange',
  CONDITIONAL: 'gold',
  NONE: 'default',
};

export function conversionTypeOptions(t: TFn) {
  return CONVERSION_TYPES.map((v) => ({
    label: t(`conversionTypeMap.${v}`),
    value: v,
  }));
}

export function getConversionTypeLabel(t: TFn, v?: string): string {
  if (!v) return '';
  return t(`conversionTypeMap.${v}`, { defaultValue: v });
}

// ========== 基准单位 Tag ==========

export const baseUnitTagColor = 'green';

// ========== 仅 LINEAR / AFFINE 可换算 ==========

export function isConvertibleType(v?: string): boolean {
  return v === 'LINEAR' || v === 'AFFINE';
}
