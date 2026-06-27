/**
 * 产品特征来源标签 / Product feature source tag.
 *
 * DEF (默认模型) = green、GLO (全局特征库) = blue、LOC (本地) = purple。
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

const palette: Record<FeatureSourceValue, { color: string; abbr: string }> = {
  DEFAULT: { color: 'green', abbr: 'DEF' },
  GLOBAL: { color: 'blue', abbr: 'GLO' },
  LOCAL: { color: 'purple', abbr: 'LOC' },
};

export const FeatureSourceTag = ({ source }: Props) => {
  const { t } = useTranslation('product-feature');
  if (!source) return null;
  const cfg = palette[source as FeatureSourceValue];
  if (!cfg) return <Tag>{source}</Tag>;
  return (
    <Tag color={cfg.color} title={t(`source.${source}`)}>
      {cfg.abbr}
    </Tag>
  );
};

export default FeatureSourceTag;
