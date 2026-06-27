/**
 * 物模型-分类管理常量
 * Thing-model category management constants
 */
import type { TFunction } from 'i18next';

// ProTable valueEnum 格式承载 kind 下拉选项
export const kindValueEnum = (t: TFunction) => ({
  SYSTEM:   { text: t('kind.SYSTEM') },
  SPACE:    { text: t('kind.SPACE') },
  FACILITY: { text: t('kind.FACILITY') },
});

export const levelValueEnum = (t: TFunction) => ({
  1: { text: t('level.1') },
  2: { text: t('level.2') },
  3: { text: t('level.3') },
  4: { text: t('level.4') },
});

export const levelColor: Record<number, string> = {
  1: 'magenta', // 大类
  2: 'volcano', // 中类
  3: 'blue', // 小类
  4: 'green', // 细类
};

export const kindColor: Record<string, string> = {
  SYSTEM:   'geekblue',
  SPACE:    'cyan',
  FACILITY: 'gold',
};

export const enableValueEnum = (t: TFunction) => ({
  true:  { text: t('tags.enabled'),  status: 'Success' as const },
  false: { text: t('tags.disabled'), status: 'Default' as const },
});
