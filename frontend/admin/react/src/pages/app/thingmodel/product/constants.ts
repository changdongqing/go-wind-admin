import type { ProSchemaValueEnumObj } from '@ant-design/pro-components';

/**
 * 产品发布状态枚举映射 / Product status value enum.
 *
 * DRAFT = 灰；PUBLISHED = 绿。
 */
export const statusValueEnum: ProSchemaValueEnumObj = {
  DRAFT: { text: '草稿', status: 'Default' },
  PUBLISHED: { text: '已发布', status: 'Success' },
};

/**
 * 启用/停用枚举映射 / Enabled value enum.
 */
export const enabledValueEnum: ProSchemaValueEnumObj = {
  true: { text: '启用', status: 'Success' },
  false: { text: '停用', status: 'Default' },
};
