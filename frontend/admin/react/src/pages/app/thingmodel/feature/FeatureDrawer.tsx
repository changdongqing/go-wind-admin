import { useRef, useState, useEffect } from 'react';
import type { ProFormInstance } from '@ant-design/pro-components';
import {
  DrawerForm,
  ProFormText,
  ProFormDigit,
  ProFormRadio,
  ProFormSelect,
  ProFormTextArea,
} from '@ant-design/pro-components';
import { App, Divider, Form } from 'antd';
import { useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';

import { useCreateFeature, useUpdateFeature } from '@/api/hooks/feature';
import {
  enableBoolRadioOptions,
  featureTypeOptions,
  type FeatureType,
} from './constants';
import PropertySpecForm from './specs/PropertySpecForm';
import EventSpecForm from './specs/EventSpecForm';
import ServiceSpecForm from './specs/ServiceSpecForm';
import RelationSpecForm from './specs/RelationSpecForm';

interface FeatureDrawerProps {
  open: boolean;
  mode: 'create' | 'edit';
  /** 当前选中的特征类型；新建时作为默认 / 编辑时从 data 读取 */
  featureType: FeatureType;
  data?: any;
  onClose: () => void;
  onSuccess: () => void;
}

/**
 * 特征 新建/编辑 抽屉 / Feature create-edit drawer.
 *
 * 顶部公共字段 + 中部动态 spec 区（按 featureType 切换）+ 底部治理字段。
 * 提交前按 oneof 结构构造 `spec: { property: {...} }` / `{ event: {...} }` 等。
 */
const FeatureDrawer: React.FC<FeatureDrawerProps> = ({
  open,
  mode,
  featureType,
  data,
  onClose,
  onSuccess,
}) => {
  const { t } = useTranslation('feature');
  const formRef = useRef<ProFormInstance>(null);
  const queryClient = useQueryClient();
  const { message } = App.useApp();

  const [confirmLoading, setConfirmLoading] = useState(false);
  // 编辑场景下以 data.featureType 为准；新建以外部传入为准
  const [currentFeatureType, setCurrentFeatureType] = useState<FeatureType>(featureType);

  // 编辑模式回填
  useEffect(() => {
    if (!open) return;
    const ft = (data?.featureType as FeatureType) || featureType;
    setCurrentFeatureType(ft);

    setTimeout(() => {
      if (mode === 'edit' && data) {
        // 把 spec.oneof 拍平为 spec.{property|event|service|relation}
        const spec = data.spec || {};
        formRef.current?.setFieldsValue({
          featureType: ft,
          code: data.code,
          identifier: data.identifier,
          name: data.name,
          nameEn: data.nameEn,
          description: data.description,
          applicableScope: data.applicableScope,
          sortOrder: data.sortOrder ?? 0,
          isEnabled: data.isEnabled ?? true,
          spec: {
            property: spec.property,
            event: spec.event,
            service: spec.service,
            relation: spec.relation,
          },
        });
      } else {
        // 新建时初始化
        formRef.current?.resetFields();
        formRef.current?.setFieldsValue({
          featureType: ft,
          isEnabled: true,
          sortOrder: 0,
          spec:
            ft === 'PROPERTY'
              ? { property: { accessMode: 'R', dataType: 'DOUBLE' } }
              : ft === 'EVENT'
                ? { event: { level: 'INFO', outputParams: [] } }
                : ft === 'SERVICE'
                  ? { service: { callMode: 'ASYNC', inputParams: [], outputParams: [] } }
                  : { relation: { directional: true, source: { kind: 'feature' }, target: { kind: 'feature' } } },
        });
      }
    }, 0);
  }, [open, mode, data, featureType]);

  const createMutation = useCreateFeature({
    onSuccess: () => {
      message.success(t('createSuccess'));
      queryClient.invalidateQueries({ queryKey: ['listFeatures'] });
      queryClient.invalidateQueries({ queryKey: ['listFeaturesByType'] });
      onSuccess();
      onClose();
    },
    onError: (error: Error) => message.error(error.message || t('createFailed')),
  });

  const updateMutation = useUpdateFeature({
    onSuccess: () => {
      message.success(t('updateSuccess'));
      queryClient.invalidateQueries({ queryKey: ['listFeatures'] });
      queryClient.invalidateQueries({ queryKey: ['listFeaturesByType'] });
      onSuccess();
      onClose();
    },
    onError: (error: Error) => message.error(error.message || t('updateFailed')),
  });

  // 把表单 spec 字段重组为 proto oneof 形态：保留对应类型分支，其余清空
  const buildSpec = (formSpec: any, ft: FeatureType) => {
    if (!formSpec) return undefined;
    switch (ft) {
      case 'PROPERTY':
        return formSpec.property ? { property: formSpec.property } : undefined;
      case 'EVENT':
        return formSpec.event ? { event: formSpec.event } : undefined;
      case 'SERVICE':
        return formSpec.service ? { service: formSpec.service } : undefined;
      case 'RELATION':
        return formSpec.relation ? { relation: formSpec.relation } : undefined;
      default:
        return undefined;
    }
  };

  const handleSubmit = async (values: Record<string, any>) => {
    try {
      setConfirmLoading(true);
      const ft = (values.featureType as FeatureType) || currentFeatureType;
      const spec = buildSpec(values.spec, ft);

      if (mode === 'create') {
        await createMutation.mutateAsync({
          data: {
            featureType: ft as any,
            code: values.code,
            identifier: values.identifier,
            name: values.name,
            nameEn: values.nameEn,
            description: values.description,
            applicableScope: values.applicableScope,
            sortOrder: values.sortOrder,
            isEnabled: values.isEnabled,
            spec: spec as any,
          },
        });
      } else {
        await updateMutation.mutateAsync({
          id: data?.id,
          values: {
            featureType: ft,
            identifier: values.identifier,
            name: values.name,
            nameEn: values.nameEn,
            description: values.description,
            applicableScope: values.applicableScope,
            sortOrder: values.sortOrder,
            isEnabled: values.isEnabled,
            spec,
          },
        });
      }
      return true;
    } finally {
      setConfirmLoading(false);
    }
  };

  return (
    <DrawerForm
      title={mode === 'create' ? t('create') : t('edit')}
      width={760}
      open={open}
      formRef={formRef}
      onOpenChange={(o) => {
        if (!o) onClose();
      }}
      onFinish={handleSubmit}
      drawerProps={{ destroyOnHidden: true, maskClosable: false }}
      submitter={{
        searchConfig: { submitText: t('common:button.ok'), resetText: t('common:button.cancel') },
        submitButtonProps: { loading: confirmLoading },
      }}
      autoFocusFirstInput
    >
      <Divider orientation="left" plain>
        {t('commonSection')}
      </Divider>

      <ProFormSelect
        name="featureType"
        label={t('featureType')}
        options={featureTypeOptions(t)}
        rules={[{ required: true }]}
        disabled={mode === 'edit'}
        fieldProps={{
          onChange: (v) => setCurrentFeatureType(v as FeatureType),
        }}
      />

      <ProFormText
        name="code"
        label={t('code')}
        placeholder={t('codePlaceholder')}
        rules={[{ required: true, message: t('requiredCode') }]}
        readonly={mode === 'edit'}
      />
      <ProFormText
        name="identifier"
        label={t('identifier')}
        placeholder={t('identifierPlaceholder')}
        rules={[{ required: true, message: t('requiredIdentifier') }]}
      />
      <ProFormText
        name="name"
        label={t('name')}
        placeholder={t('namePlaceholder')}
        rules={[{ required: true, message: t('requiredName') }]}
      />
      <ProFormText
        name="nameEn"
        label={t('nameEn')}
        placeholder={t('nameEnPlaceholder')}
      />
      <ProFormTextArea
        name="description"
        label={t('description')}
        placeholder={t('descriptionPlaceholder')}
        fieldProps={{ rows: 2 }}
      />
      <ProFormText
        name="applicableScope"
        label={t('applicableScope')}
        placeholder={t('applicableScopePlaceholder')}
      />

      <Divider orientation="left" plain>
        {t('specSection')}
      </Divider>

      {currentFeatureType === 'PROPERTY' && <PropertySpecForm namePath={['spec', 'property']} />}
      {currentFeatureType === 'EVENT' && <EventSpecForm namePath={['spec', 'event']} />}
      {currentFeatureType === 'SERVICE' && <ServiceSpecForm namePath={['spec', 'service']} />}
      {currentFeatureType === 'RELATION' && <RelationSpecForm namePath={['spec', 'relation']} />}

      <Divider orientation="left" plain>
        {t('governanceSection')}
      </Divider>

      <ProFormDigit
        name="sortOrder"
        label={t('sortOrder')}
        placeholder={t('sortOrderPlaceholder')}
        min={0}
        fieldProps={{ precision: 0 }}
      />
      <ProFormRadio.Group
        name="isEnabled"
        label={t('status')}
        options={enableBoolRadioOptions(t)}
      />
    </DrawerForm>
  );
};

export default FeatureDrawer;
