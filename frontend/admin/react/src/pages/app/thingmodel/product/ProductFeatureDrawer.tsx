/**
 * 产品特征编辑 Drawer（edit / create-global / create-local）。
 *
 * - edit：DRAFT 全字段可改（含完整 spec）；PUBLISHED 整 spec 表单 disabled，仅 name/description 可改
 * - create-global：弹 FeaturePicker 选多条全局特征，逐个创建 source=GLOBAL（spec 在创建时由后端从骨架/CDF 推断或为空，后续在 edit 中补齐）
 * - create-local：完全自定义；解除 PROPERTY 唯一限制，允许 EVENT/SERVICE/RELATION
 *
 * 设计依据 / Design ref:
 *   - docs/thingmodel/sheji/模型管理/05-前端实现设计.md §1.5
 *   - docs/thingmodel/sheji/修改记录/CR-001-结构化约束下沉到模型层.md
 *
 * CR-001（2026-06-29）：
 *   - feature_snapshot/override_spec 合并为单一 spec；OverrideSpecForm 已删除；
 *   - 改为挂载共享 FeatureSpecForm 按 featureType 切换完整 spec 编辑器；
 *   - create-local 模式：featureType 由用户选择，不再硬编码 PROPERTY；
 *   - LOCAL_DATA_TYPES 限制解除（dataType 在 PropertySpecForm 内通过 spec.property.dataType 选择）。
 */
import { useEffect, useState } from 'react';
import { Drawer, Form, Input, Select, Alert, Button, App, Divider } from 'antd';
import { useTranslation } from 'react-i18next';
import FeatureSpecForm from '../_shared/specs/FeatureSpecForm';
import FeaturePicker from '../_shared/FeaturePicker';
import { featureTypeOptions, type FeatureType } from '../feature/constants';
import {
  useCreateProductFeature,
  useUpdateProductFeature,
} from '@/api/hooks/product-feature';
import type {
  thingmodelservicev1_FeatureSpec,
  thingmodelservicev1_ProductFeature,
} from '@/api/generated/admin/service/v1';

interface Props {
  open: boolean;
  productId: number;
  feature: thingmodelservicev1_ProductFeature | null;
  mode: 'edit' | 'create-local' | 'create-global';
  /** PUBLISHED 产品的 spec 编辑被冻结；name/description 仍可改 */
  readonly?: 'partial' | false;
  onClose: (reload?: boolean) => void;
}

const ProductFeatureDrawer = ({ open, productId, feature, mode, readonly, onClose }: Props) => {
  const { t } = useTranslation(['product-feature', 'feature', 'common']);
  const { message } = App.useApp();
  const [form] = Form.useForm<{
    featureType?: FeatureType;
    code?: string;
    identifier?: string;
    name?: string;
    description?: string;
    spec?: thingmodelservicev1_FeatureSpec;
  }>();
  const [pickerOpen, setPickerOpen] = useState(mode === 'create-global');
  const [currentFeatureType, setCurrentFeatureType] = useState<FeatureType | undefined>(undefined);

  const { mutate: doCreate, isPending: creating } = useCreateProductFeature();
  const { mutate: doUpdate, isPending: updating } = useUpdateProductFeature();

  useEffect(() => {
    if (!open) return;
    if (mode === 'edit' && feature) {
      form.setFieldsValue({
        featureType: feature.featureType as FeatureType,
        code: feature.code,
        identifier: feature.identifier,
        name: feature.name,
        description: feature.description,
        spec: feature.spec,
      });
      setCurrentFeatureType(feature.featureType as FeatureType);
    } else if (mode === 'create-local') {
      form.resetFields();
      // 默认 PROPERTY（用户可改）
      form.setFieldsValue({ featureType: 'PROPERTY' });
      setCurrentFeatureType('PROPERTY');
    } else if (mode === 'create-global') {
      setPickerOpen(true);
    }
  }, [feature, mode, form, open]);

  const isEditing = mode === 'edit';
  const isLocal = mode === 'create-local';
  const specFrozen = readonly === 'partial';

  const handleSubmit = async () => {
    const v = await form.validateFields();
    if (mode === 'edit' && feature) {
      const values: Record<string, any> = {
        name: v.name,
        description: v.description,
      };
      // PUBLISHED 时 spec 冻结，不发送
      if (!specFrozen && v.spec) {
        values.spec = v.spec;
      }
      doUpdate(
        { id: feature.id!, values },
        {
          onSuccess: () => {
            message.success(t('common:updateSuccess'));
            onClose(true);
          },
          onError: (err) => message.error(err.message),
        },
      );
      return;
    }
    if (mode === 'create-local') {
      doCreate(
        {
          data: {
            productId,
            source: 'LOCAL',
            featureType: v.featureType,
            code: v.code,
            identifier: v.identifier,
            name: v.name,
            description: v.description,
            spec: v.spec,
          },
        },
        {
          onSuccess: () => {
            message.success(t('common:createSuccess'));
            onClose(true);
          },
          onError: (err) => message.error(err.message),
        },
      );
    }
  };

  return (
    <>
      <Drawer
        size={780}
        open={open && (mode !== 'create-global' || !!feature)}
        onClose={() => onClose(false)}
        title={
          isEditing
            ? `${t('editFeature')} — ${feature?.name}`
            : mode === 'create-local'
              ? t('createLocalFeature')
              : t('createGlobalFeature')
        }
        destroyOnHidden
        extra={
          <Button type="primary" onClick={handleSubmit} loading={creating || updating}>
            {t('common:save')}
          </Button>
        }
      >
        {specFrozen && (
          <Alert
            type="warning"
            message={t('publishedReadonlyWarning', {
              defaultValue: '产品已发布：基础信息（name/description）可改，spec 已冻结。',
            })}
            showIcon
            style={{ marginBottom: 12 }}
          />
        )}
        <Form
          form={form}
          layout="vertical"
          onValuesChange={(changed) => {
            if (changed.featureType) {
              setCurrentFeatureType(changed.featureType as FeatureType);
              // 切换 featureType 时清空 spec 子分支（避免 oneof 错位）
              form.setFieldValue('spec', undefined);
            }
          }}
        >
          {isLocal && (
            <Form.Item
              label={t('feature:featureType')}
              name="featureType"
              rules={[{ required: true }]}
            >
              <Select options={featureTypeOptions(t)} />
            </Form.Item>
          )}
          <Form.Item label={t('code')} name="code" rules={[{ required: !isEditing }]}>
            <Input disabled={isEditing} />
          </Form.Item>
          <Form.Item label={t('identifier')} name="identifier" rules={[{ required: !isEditing }]}>
            <Input disabled={isEditing} />
          </Form.Item>
          <Form.Item label={t('name')} name="name" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item label={t('description')} name="description">
            <Input.TextArea rows={2} />
          </Form.Item>

          {(isEditing || isLocal) && (
            <>
              <Divider orientation="left" plain>
                {t('specSection', { defaultValue: '结构化约束 Spec' })}
              </Divider>
              <FeatureSpecForm
                featureType={currentFeatureType ?? (feature?.featureType as FeatureType)}
                disabled={specFrozen}
              />
            </>
          )}
        </Form>
      </Drawer>

      {mode === 'create-global' && (
        <FeaturePicker
          open={pickerOpen}
          onCancel={() => {
            setPickerOpen(false);
            onClose(false);
          }}
          onConfirm={(feats) => {
            setPickerOpen(false);
            if (feats.length === 0) {
              onClose(false);
              return;
            }
            let done = 0;
            let failed = 0;
            feats.forEach((f) => {
              doCreate(
                {
                  data: {
                    productId,
                    source: 'GLOBAL',
                    refFeatureId: f.id,
                  },
                },
                {
                  onSuccess: () => {
                    done++;
                    if (done + failed === feats.length) {
                      if (failed > 0) {
                        message.warning(t('batchAddPartial', { done, failed }));
                      } else {
                        message.success(t('common:createSuccess'));
                      }
                      onClose(true);
                    }
                  },
                  onError: (err) => {
                    failed++;
                    message.error(`${f.code}: ${err.message}`);
                    if (done + failed === feats.length) {
                      onClose(true);
                    }
                  },
                },
              );
            });
          }}
        />
      )}
    </>
  );
};

export default ProductFeatureDrawer;
