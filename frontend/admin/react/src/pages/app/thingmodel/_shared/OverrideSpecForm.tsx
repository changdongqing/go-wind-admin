/**
 * 特征覆写表单（白名单字段：constraints/unit/defaultValue/displayName/description/required）。
 * Feature override whitelist editor.
 *
 * 收口规则：proto FeatureOverrideSpec 本身只允许这 6 个字段；本表单是 UI 层最后一道收口。
 * 非 property 特征上 constraints/unit/defaultValue 灰显（后端会校验拒收）。
 *
 * 设计依据 / Design ref: docs/thingmodel/sheji/模型管理/05-前端实现设计.md §1.5 / §6.2
 */
import { Form, InputNumber, Input, Switch, Alert } from 'antd';
import { useTranslation } from 'react-i18next';
import type {
  thingmodelservicev1_FeatureOverrideSpec,
  thingmodelservicev1_FeatureSpec,
  thingmodelservicev1_FeatureType,
} from '@/api/generated/admin/service/v1';

interface Props {
  /** 用于决定能否覆写 constraints/unit/defaultValue 的目标 feature 类型 */
  featureType?: thingmodelservicev1_FeatureType;
  /** 当前 snapshot（只读展示原值用） */
  snapshot?: thingmodelservicev1_FeatureSpec;
  /** 受控 override 值 */
  value?: thingmodelservicev1_FeatureOverrideSpec | null;
  onChange?: (v: thingmodelservicev1_FeatureOverrideSpec | null) => void;
  /** 整个表单只读（PUBLISHED 产品的非白名单字段灰显） */
  readonly?: boolean;
}

export const OverrideSpecForm = ({
  featureType,
  snapshot,
  value,
  onChange,
  readonly,
}: Props) => {
  const { t } = useTranslation(['product-feature', 'common']);

  const isProperty = featureType === 'PROPERTY';

  const merge = (patch: Partial<thingmodelservicev1_FeatureOverrideSpec>) => {
    onChange?.({ ...(value ?? {}), ...patch });
  };

  const cur = value ?? {};

  return (
    <Form layout="vertical" disabled={readonly}>
      <Form.Item label={t('override.displayName')}>
        <Input
          value={cur.displayName}
          onChange={(e) => merge({ displayName: e.target.value })}
          placeholder={t('override.displayNamePlaceholder', '作用域内展示别名')}
        />
      </Form.Item>

      <Form.Item label={t('override.description')}>
        <Input.TextArea
          rows={2}
          value={cur.description}
          onChange={(e) => merge({ description: e.target.value })}
        />
      </Form.Item>

      {!isProperty && (
        <Alert
          type="info"
          showIcon
          message={t('override.nonPropertyHint', '非属性类型不支持覆写 constraints / unit / defaultValue')}
          style={{ marginBottom: 12 }}
        />
      )}

      <Form.Item label={t('override.constraintsRange', '取值范围 min/max/step')}>
        <Input.Group compact>
          <InputNumber
            placeholder="min"
            disabled={!isProperty}
            value={cur.constraints?.min}
            onChange={(v) =>
              merge({ constraints: { ...(cur.constraints ?? {}), min: v ?? undefined } })
            }
            style={{ width: '33%' }}
          />
          <InputNumber
            placeholder="max"
            disabled={!isProperty}
            value={cur.constraints?.max}
            onChange={(v) =>
              merge({ constraints: { ...(cur.constraints ?? {}), max: v ?? undefined } })
            }
            style={{ width: '33%' }}
          />
          <InputNumber
            placeholder="step"
            disabled={!isProperty}
            value={cur.constraints?.step}
            onChange={(v) =>
              merge({ constraints: { ...(cur.constraints ?? {}), step: v ?? undefined } })
            }
            style={{ width: '34%' }}
          />
        </Input.Group>
      </Form.Item>

      <Form.Item label={t('override.defaultValue')}>
        <Input
          disabled={!isProperty}
          value={cur.defaultValue}
          onChange={(e) => merge({ defaultValue: e.target.value })}
        />
      </Form.Item>

      <Form.Item label={t('override.required', 'required（service 输入参数用）')}>
        <Switch
          checked={cur.required ?? undefined}
          onChange={(checked) => merge({ required: checked })}
        />
      </Form.Item>

      {snapshot?.property && isProperty && (
        <Alert
          type="default"
          message={
            <span>
              {t('override.snapshotPreview', '原始范围')}：min=
              {snapshot.property.constraints?.min ?? '-'} max=
              {snapshot.property.constraints?.max ?? '-'}
            </span>
          }
        />
      )}
    </Form>
  );
};

export default OverrideSpecForm;
