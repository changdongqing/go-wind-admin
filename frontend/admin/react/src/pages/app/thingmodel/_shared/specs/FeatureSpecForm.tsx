import { useMemo } from 'react';
import { Alert } from 'antd';
import { useTranslation } from 'react-i18next';

import PropertySpecForm from './PropertySpecForm';
import EventSpecForm from './EventSpecForm';
import ServiceSpecForm from './ServiceSpecForm';
import RelationSpecForm from './RelationSpecForm';
import type { FeatureType } from '../../feature/constants';

/**
 * FeatureSpecForm — CR-001 后的统一 spec 编辑器分派组件。
 *
 * 按 featureType 切换到对应的 PropertySpec/EventSpec/ServiceSpec/RelationSpec 子表单。
 * 字段路径前缀固定为 `['spec', '<branch>']`（与 proto FeatureSpec oneof 序列化一致）：
 *   - PROPERTY → spec.property.*
 *   - EVENT    → spec.event.*
 *   - SERVICE  → spec.service.*
 *   - RELATION → spec.relation.*
 *
 * 实现注意：
 *   - 不使用 `<Form.Item noStyle + render-prop>` 包裹——那种写法下 antd
 *     的子 Form.Item 字段挂载时机晚于祖先 Form 的 setFieldsValue，
 *     会导致初始值/提交值丢失。直接条件渲染即可。
 *   - `disabled` 通过 `<fieldset disabled>` 实现，原生属性，不影响 Form 注册。
 *
 * 设计依据 / Design ref:
 *   - docs/thingmodel/sheji/10-特征参数与spec设计.md §3
 *   - docs/thingmodel/sheji/修改记录/CR-001-结构化约束下沉到模型层.md
 *
 * Hosts:
 *   - 分类管理 · 默认模型抽屉（CategoryDefaultFeaturesDrawer）
 *   - 产品详情 · 特征抽屉（ProductFeatureDrawer）
 */
export interface FeatureSpecFormProps {
  featureType?: FeatureType | string;
  /** 是否整体禁用（PUBLISHED 产品的特征 spec 编辑） */
  disabled?: boolean;
}

const FeatureSpecForm: React.FC<FeatureSpecFormProps> = ({ featureType, disabled }) => {
  const { t } = useTranslation();

  const branch = useMemo<FeatureType | undefined>(() => {
    if (!featureType) return undefined;
    const ft = String(featureType).toUpperCase();
    if (ft === 'PROPERTY' || ft === 'EVENT' || ft === 'SERVICE' || ft === 'RELATION') {
      return ft as FeatureType;
    }
    return undefined;
  }, [featureType]);

  if (!branch) {
    return (
      <Alert
        type="info"
        showIcon
        message={t('common:tips.spec.unknownFeatureType', {
          defaultValue: '未识别的特征类型，无法渲染 spec 表单',
        })}
      />
    );
  }

  const inner = (() => {
    switch (branch) {
      case 'PROPERTY':
        return <PropertySpecForm namePath={['spec', 'property']} />;
      case 'EVENT':
        return <EventSpecForm namePath={['spec', 'event']} />;
      case 'SERVICE':
        return <ServiceSpecForm namePath={['spec', 'service']} />;
      case 'RELATION':
        return <RelationSpecForm namePath={['spec', 'relation']} />;
      default:
        return null;
    }
  })();

  return (
    <fieldset disabled={disabled} style={{ border: 0, padding: 0, margin: 0 }}>
      {inner}
    </fieldset>
  );
};

export default FeatureSpecForm;
