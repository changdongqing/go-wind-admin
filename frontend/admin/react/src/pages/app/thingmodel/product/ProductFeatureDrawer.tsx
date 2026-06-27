/**
 * 产品特征编辑 Drawer（edit / create-global / create-local）。
 *
 * - edit：基础字段灰显（仅白名单可改）；PUBLISHED 产品 readonly=partial
 * - create-global：弹 FeaturePicker 选多条全局特征，逐个创建 source=GLOBAL
 * - create-local：完全自定义 + 选 dataType（BOOL/INT/DOUBLE/TEXT/ENUM/DATE），落库 source=LOCAL
 *
 * 设计依据 / Design ref: docs/thingmodel/sheji/模型管理/05-前端实现设计.md §1.5
 */
import { useEffect, useState } from 'react';
import { Drawer, Form, Input, Select, Alert, Button, App } from 'antd';
import { useTranslation } from 'react-i18next';
import OverrideSpecForm from '../_shared/OverrideSpecForm';
import FeaturePicker from '../_shared/FeaturePicker';
import {
  useCreateProductFeature,
  useUpdateProductFeature,
} from '@/api/hooks/product-feature';
import type {
  thingmodelservicev1_FeatureOverrideSpec,
  thingmodelservicev1_ProductFeature,
} from '@/api/generated/admin/service/v1';

interface Props {
  open: boolean;
  productId: number;
  feature: thingmodelservicev1_ProductFeature | null;
  mode: 'edit' | 'create-local' | 'create-global';
  readonly?: 'partial' | false;
  onClose: (reload?: boolean) => void;
}

const LOCAL_DATA_TYPES = ['INT', 'DOUBLE', 'BOOL', 'TEXT', 'ENUM', 'DATE'] as const;

const ProductFeatureDrawer = ({ open, productId, feature, mode, readonly, onClose }: Props) => {
  const { t } = useTranslation(['product-feature', 'common']);
  const { message } = App.useApp();
  const [form] = Form.useForm();
  const [override, setOverride] = useState<thingmodelservicev1_FeatureOverrideSpec | null>(null);
  const [pickerOpen, setPickerOpen] = useState(mode === 'create-global');

  const { mutate: doCreate, isPending: creating } = useCreateProductFeature();
  const { mutate: doUpdate, isPending: updating } = useUpdateProductFeature();

  useEffect(() => {
    if (!open) return;
    if (mode === 'edit' && feature) {
      form.setFieldsValue({
        code: feature.code,
        identifier: feature.identifier,
        name: feature.name,
        description: feature.description,
      });
      setOverride(feature.overrideSpec ?? null);
    } else if (mode === 'create-local') {
      form.resetFields();
      setOverride(null);
    } else if (mode === 'create-global') {
      setPickerOpen(true);
    }
  }, [feature, mode, form, open]);

  const handleSubmit = async () => {
    const v = await form.validateFields();
    if (mode === 'edit' && feature) {
      const values: Record<string, any> = {
        name: v.name,
        description: v.description,
        overrideSpec: override ?? undefined,
      };
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
            featureType: 'PROPERTY',
            code: v.code,
            identifier: v.identifier,
            name: v.name,
            description: v.description,
            featureSnapshot: {
              property: {
                dataType: v.dataType,
                accessMode: v.accessMode ?? 'RW',
              },
            },
            dataType: v.dataType,
            accessMode: v.accessMode ?? 'RW',
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

  const isEditing = mode === 'edit';
  const isPropertyLocal = mode === 'create-local';

  return (
    <>
      <Drawer
        size={720}
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
        {readonly === 'partial' && (
          <Alert
            type="warning"
            message={t('publishedReadonlyWarning')}
            showIcon
            style={{ marginBottom: 12 }}
          />
        )}
        <Form
          form={form}
          layout="vertical"
          disabled={readonly === 'partial' && isEditing}
        >
          <Form.Item label={t('code')} name="code" rules={[{ required: !isEditing }]}>
            <Input disabled={isEditing} />
          </Form.Item>
          <Form.Item
            label={t('identifier')}
            name="identifier"
            rules={[{ required: !isEditing }]}
          >
            <Input disabled={isEditing} />
          </Form.Item>
          <Form.Item label={t('name')} name="name" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item label={t('description')} name="description">
            <Input.TextArea rows={2} />
          </Form.Item>

          {isPropertyLocal && (
            <>
              <Form.Item label={t('dataType')} name="dataType" rules={[{ required: true }]}>
                <Select
                  options={LOCAL_DATA_TYPES.map((v) => ({ label: v, value: v }))}
                />
              </Form.Item>
              <Form.Item label={t('accessMode')} name="accessMode" initialValue="RW">
                <Select
                  options={[
                    { label: 'R', value: 'R' },
                    { label: 'RW', value: 'RW' },
                  ]}
                />
              </Form.Item>
            </>
          )}
        </Form>

        {(isEditing || isPropertyLocal) && (
          <>
            <h4>{t('overrideSection')}</h4>
            <OverrideSpecForm
              featureType={feature?.featureType ?? 'PROPERTY'}
              snapshot={feature?.featureSnapshot}
              value={override}
              onChange={setOverride}
              readonly={readonly === 'partial' ? false : undefined}
            />
          </>
        )}
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
