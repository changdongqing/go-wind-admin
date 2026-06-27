/**
 * 产品特征来源标签 / Product feature source tag.
 *
 * DEFAULT (默认模型) = green、GLOBAL (全局特征库) = blue、LOCAL (本地) = purple。
 * 用于产品特征列表 "来源" 列的视觉区分。
 *
 * 设计依据 / Design ref: docs/thingmodel/sheji/模型管理/05-前端实现设计.md §1.3
 */
import { Tag } from 'antd';
import { useTranslation } from 'react-i18next';

export type FeatureSourceValue = 'DEFAULT' | 'GLOBAL' | 'LOCAL';

interface Props {
  source?: FeatureSourceValue | string;
}

const palette: Record<FeatureSourceValue, string> = {
  DEFAULT: 'green',
  GLOBAL: 'blue',
  LOCAL: 'purple',
};

export const FeatureSourceTag = ({ source }: Props) => {
  const { t } = useTranslation('product-feature');
  // 后端零值（PRODUCT_FEATURE_SOURCE_UNSPECIFIED）或空值都视为"未知"，避免野露 proto 枚举字面量
  if (!source || source === 'PRODUCT_FEATURE_SOURCE_UNSPECIFIED') {
    return <Tag>{t('sourceMap.UNSPECIFIED', '-')}</Tag>;
  }
  const color = palette[source as FeatureSourceValue];
  if (!color) return <Tag>{source}</Tag>;
  return <Tag color={color}>{t(`sourceMap.${source}`, source)}</Tag>;
};

export default FeatureSourceTag;
