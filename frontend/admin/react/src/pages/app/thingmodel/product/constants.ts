import type { ProSchemaValueEnumObj } from '@ant-design/pro-components';

type TFn = (key: string, options?: Record<string, any>) => string;

/**
 * 产品发布状态枚举映射 / Product status value enum.
 *
 * 工厂函数：调用方传入 `t`，确保跟随语言切换实时更新。
 * - DRAFT = 灰；PUBLISHED = 绿
 * - PRODUCT_STATUS_UNSPECIFIED 兜底显示"未设置"，避免后端 0 值漏到 UI 露出 enum 字面量。
 */
export function getStatusValueEnum(t: TFn): ProSchemaValueEnumObj {
  return {
    DRAFT: { text: t('statusMap.DRAFT'), status: 'Default' },
    PUBLISHED: { text: t('statusMap.PUBLISHED'), status: 'Success' },
    PRODUCT_STATUS_UNSPECIFIED: {
      text: t('statusMap.PRODUCT_STATUS_UNSPECIFIED'),
      status: 'Default',
    },
  };
}

/**
 * 启用/停用枚举映射 / Enabled value enum.
 */
export function getEnabledValueEnum(t: TFn): ProSchemaValueEnumObj {
  return {
    true: { text: t('common:enabled'), status: 'Success' },
    false: { text: t('common:disabled'), status: 'Default' },
  };
}
